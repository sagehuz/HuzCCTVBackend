package cli

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"huzbackend-go/internal/config"
)

func runStatus(args []string) int {
	cfg := config.Load()
	running, pid := serverRunning()
	if running {
		fmt.Println("Status: 🟢 RUNNING")
		fmt.Printf("PID: %d\n", pid)
		if up, ok := processUptime(pid); ok {
			fmt.Printf("Uptime: %s\n", up)
		}
		fmt.Printf("Port: %s\n", cfg.Port)
		fmt.Printf("Dashboard: http://127.0.0.1:%s\n", cfg.Port)
		fmt.Printf("Log: %s\n", logFilePath())
		return 0
	}
	fmt.Println("Status: 🔴 STOPPED")
	fmt.Printf("Port: %s (will listen once started)\n", cfg.Port)
	if pid > 0 {
		fmt.Printf("Stale PID file present: %d\n", pid)
	}
	if portOpen(cfg.Port) {
		fmt.Printf("Note: another process is already listening on port %s.\n", cfg.Port)
	}
	return 0
}

func readPID() (int, error) {
	data, err := os.ReadFile(pidFilePath())
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, err
	}
	return pid, nil
}

func serverRunning() (bool, int) {
	pid, err := readPID()
	if err != nil || pid <= 0 {
		return false, 0
	}
	if !processAlive(pid) {
		return false, pid
	}
	return true, pid
}

func portOpen(port string) bool {
	conn, err := net.DialTimeout("tcp", "127.0.0.1:"+port, 300*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}
