//go:build windows

package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/sys/windows/registry"
)

const autostartRegKey = `Software\Microsoft\Windows\CurrentVersion\Run`
const autostartValueName = "Huz CCTV Server"

// Windows process flags (not available in syscall).
const (
	createNewProcessGroup = 0x00000200 // CREATE_NEW_PROCESS_GROUP
	detachedProcess       = 0x00000008 // DETACHED_PROCESS
)

// setDetachAttr runs the child process detached from the console.
func setDetachAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: createNewProcessGroup | detachedProcess,
	}
}

func processAlive(pid int) bool {
	out, err := exec.Command("tasklist", "/FI", "PID eq "+strconv.Itoa(pid), "/NH").Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), strconv.Itoa(pid))
}

func terminate(pid int) error {
	return exec.Command("taskkill", "/PID", strconv.Itoa(pid)).Run()
}

func forceKill(pid int) error {
	return exec.Command("taskkill", "/F", "/PID", strconv.Itoa(pid)).Run()
}

func processUptime(pid int) (string, bool) {
	return "", false
}

func AutoStartEnable() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	k, _, err := registry.CreateKey(registry.CURRENT_USER, autostartRegKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	// cmd /c start "" /b to run in the background without showing a console window at login.
	value := fmt.Sprintf(`cmd.exe /c start "" /b "%s" serve`, exe)
	return k.SetStringValue(autostartValueName, value)
}

func AutoStartDisable() error {
	k, err := registry.OpenKey(registry.CURRENT_USER, autostartRegKey, registry.SET_VALUE)
	if err != nil {
		return nil // key does not exist -> treat as disabled
	}
	defer k.Close()
	return k.DeleteValue(autostartValueName)
}

func AutoStartStatus() (bool, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, autostartRegKey, registry.QUERY_VALUE)
	if err != nil {
		return false, nil
	}
	defer k.Close()
	_, _, err = k.GetStringValue(autostartValueName)
	return err == nil, nil
}

func openBrowser(url string) error {
	return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Run()
}
