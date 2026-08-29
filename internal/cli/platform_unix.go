//go:build !windows

package cli

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// setDetachAttr detaches the child process from the terminal to run in the background.
func setDetachAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}

func processAlive(pid int) bool {
	if syscall.Kill(pid, 0) != nil {
		return false
	}
	// Remove zombies: an exited process not yet reaped by its parent
	// (kill(pid, 0) still succeeds for zombies). State 'Z' = zombie.
	out, err := exec.Command("ps", "-o", "stat=", "-p", strconv.Itoa(pid)).Output()
	if err != nil {
		return true // ps failed — treat as alive to be safe
	}
	stat := strings.TrimSpace(string(out))
	return stat != "" && !strings.HasPrefix(stat, "Z")
}

func terminate(pid int) error {
	return syscall.Kill(pid, syscall.SIGTERM)
}

func forceKill(pid int) error {
	return syscall.Kill(pid, syscall.SIGKILL)
}

func processUptime(pid int) (string, bool) {
	// Linux: etimes = seconds. macOS: etimes unsupported, use etime.
	out, err := exec.Command("ps", "-o", "etimes=", "-p", strconv.Itoa(pid)).Output()
	if err == nil {
		if secs, err := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64); err == nil && secs >= 0 {
			return formatDuration(time.Duration(secs) * time.Second), true
		}
	}
	out, err = exec.Command("ps", "-o", "etime=", "-p", strconv.Itoa(pid)).Output()
	if err != nil {
		return "", false
	}
	if d, ok := parseETime(string(out)); ok {
		return formatDuration(d), true
	}
	return "", false
}

// parseETime converts the [[DD-]hh:]mm:ss format of `ps -o etime` to a duration.
func parseETime(s string) (time.Duration, bool) {
	s = strings.TrimSpace(s)
	days := int64(0)
	if idx := strings.Index(s, "-"); idx >= 0 {
		d, err := strconv.ParseInt(s[:idx], 10, 64)
		if err != nil {
			return 0, false
		}
		days = d
		s = s[idx+1:]
	}
	parts := strings.Split(s, ":")
	if len(parts) < 1 || len(parts) > 3 {
		return 0, false
	}
	var secs int64
	mult := int64(1)
	for i := len(parts) - 1; i >= 0; i-- {
		v, err := strconv.ParseInt(strings.TrimSpace(parts[i]), 10, 64)
		if err != nil {
			return 0, false
		}
		secs += v * mult
		mult *= 60
	}
	return time.Duration(days*86400+secs) * time.Second, true
}

func formatDuration(d time.Duration) string {
	return fmt.Sprintf("%02d:%02d:%02d", int(d.Hours()), int(d.Minutes())%60, int(d.Seconds())%60)
}
