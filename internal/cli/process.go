package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"huzbackend-go/internal/config"
)

func mustCwd() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}

// pidFilePath and logFilePath live next to the binary (main chdirs to it).
func pidFilePath() string { return filepath.Join(mustCwd(), ".huzbackend.pid") }
func logFilePath() string { return filepath.Join(mustCwd(), ".huzbackend.log") }

// --- start ---

func runStart(args []string) int {
	if running, pid := serverRunning(); running {
		fmt.Printf("Server is already running (PID %d). Use 'huzbackend status' for details.\n", pid)
		return 0
	}
	if err := startServer(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start the server: %v\n", err)
		return 1
	}
	pid, _ := readPID()
	cfg := config.Load()

	// Wait up to ~15s for the server to finish starting. Confirm with the
	// server's own log: it must contain "Server is running on port ..." on two
	// consecutive checks with no bind error ("HTTP server error"). The long
	// timeout accounts for slow SQLite/WAL warm-up, which can take a few seconds.
	const (
		startMarker = "Server is running on port"
		errMarker   = "HTTP server error"
	)
	ok := false
	sawStart := false
	for i := 0; i < 150; i++ {
		if !processAlive(pid) {
			break
		}
		data, _ := os.ReadFile(logFilePath())
		logText := string(data)
		if strings.Contains(logText, errMarker) {
			break
		}
		if strings.Contains(logText, startMarker) {
			if sawStart {
				ok = true
				break
			}
			sawStart = true
		} else {
			sawStart = false
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !ok {
		alive := processAlive(pid)
		if alive {
			// The process may be in its shutdown phase (graceful shutdown after
			// a bind error). Wait briefly and re-check before concluding.
			time.Sleep(500 * time.Millisecond)
			alive = processAlive(pid)
		}
		fmt.Fprintf(os.Stderr, "Server (PID %d) did not confirm startup (alive=%v). Latest log:\n", pid, alive)
		if lines, err := tailLines(logFilePath(), 10); err == nil {
			for _, l := range lines {
				fmt.Fprintln(os.Stderr, "  "+l)
			}
		}
		// Only remove the PID file if the process is really dead; if it is still
		// alive the server may still be starting — the user should run `status`
		// to keep an eye on it.
		if !alive {
			_ = os.Remove(pidFilePath())
		}
		return 1
	}
	fmt.Printf("Server is running at http://127.0.0.1:%s (PID %d)\n", cfg.Port, pid)
	return 0
}

// startServer runs the server in the background by re-spawning the current
// binary with the "serve" argument, detached from the terminal, logging to
// .huzbackend.log.
func startServer() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	dir := filepath.Dir(exe)
	cmd := exec.Command(exe, "serve")
	cmd.Dir = dir
	setDetachAttr(cmd)

	logF, err := os.OpenFile(logFilePath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer logF.Close()
	cmd.Stdout = logF
	cmd.Stderr = logF
	cmd.Stdin = nil

	if err := cmd.Start(); err != nil {
		return err
	}
	pid := cmd.Process.Pid
	if err := os.WriteFile(pidFilePath(), []byte(strconv.Itoa(pid)), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not write PID file: %v\n", err)
	}
	return cmd.Process.Release()
}

// --- stop ---

func runStop(args []string) int {
	pid, err := readPID()
	if err != nil || pid <= 0 {
		fmt.Println("No PID file found — the server does not seem to be running (or has already stopped).")
		return 0
	}
	if !processAlive(pid) {
		fmt.Printf("PID %d is no longer active — cleaning up the old PID file.\n", pid)
		_ = os.Remove(pidFilePath())
		return 0
	}
	fmt.Printf("Stopping server (PID %d)...\n", pid)
	if err := terminate(pid); err != nil {
		fmt.Fprintf(os.Stderr, "Could not send the stop signal: %v — trying force kill...\n", err)
		if ferr := forceKill(pid); ferr != nil {
			fmt.Fprintf(os.Stderr, "Force kill failed: %v\n", ferr)
			return 1
		}
	}
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if !processAlive(pid) {
			fmt.Println("Server stopped.")
			_ = os.Remove(pidFilePath())
			return 0
		}
		time.Sleep(500 * time.Millisecond)
	}
	if processAlive(pid) {
		fmt.Println("Server did not exit within 10s — force killing.")
		if err := forceKill(pid); err != nil {
			fmt.Fprintf(os.Stderr, "Force kill failed: %v\n", err)
			return 1
		}
		fmt.Println("Server stopped.")
	}
	_ = os.Remove(pidFilePath())
	return 0
}

// --- restart ---

func runRestart(args []string) int {
	if code := runStop(args); code != 0 {
		return code
	}
	if running, _ := serverRunning(); running {
		fmt.Fprintln(os.Stderr, "Could not stop the server — aborting restart.")
		return 1
	}
	return runStart(args)
}
