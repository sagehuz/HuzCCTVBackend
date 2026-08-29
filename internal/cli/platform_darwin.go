//go:build darwin

package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

const launchAgentLabel = "com.huzcctv.server"

func launchAgentPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "LaunchAgents", launchAgentLabel+".plist"), nil
}

func launchAgentContent() ([]byte, error) {
	exe, err := os.Executable()
	if err != nil {
		return nil, err
	}
	dir := filepath.Dir(exe)
	logPath := filepath.Join(dir, ".huzbackend.log")
	plist := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>%s</string>
	<key>ProgramArguments</key>
	<array>
		<string>%s</string>
		<string>serve</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<false/>
	<key>WorkingDirectory</key>
	<string>%s</string>
	<key>StandardOutPath</key>
	<string>%s</string>
	<key>StandardErrorPath</key>
	<string>%s</string>
</dict>
</plist>
`, launchAgentLabel, exe, dir, logPath, logPath)
	return []byte(plist), nil
}

func AutoStartEnable() error {
	if on, err := AutoStartStatus(); err == nil && on {
		return nil
	}
	plistPath, err := launchAgentPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(plistPath), 0o755); err != nil {
		return err
	}
	content, err := launchAgentContent()
	if err != nil {
		return err
	}
	if err := os.WriteFile(plistPath, content, 0o644); err != nil {
		return err
	}
	uid := strconv.Itoa(os.Getuid())
	if err := exec.Command("launchctl", "bootstrap", "gui/"+uid, plistPath).Run(); err != nil {
		// Older macOS: fall back to `load -w`.
		if lerr := exec.Command("launchctl", "load", "-w", plistPath).Run(); lerr != nil {
			return fmt.Errorf("launchctl bootstrap/load failed: %v / %v", err, lerr)
		}
	}
	return nil
}

func AutoStartDisable() error {
	plistPath, err := launchAgentPath()
	if err != nil {
		return err
	}
	uid := strconv.Itoa(os.Getuid())
	_ = exec.Command("launchctl", "bootout", "gui/"+uid+"/"+launchAgentLabel).Run()
	_ = exec.Command("launchctl", "unload", "-w", plistPath).Run()
	if err := os.Remove(plistPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func AutoStartStatus() (bool, error) {
	plistPath, err := launchAgentPath()
	if err != nil {
		return false, err
	}
	if _, err := os.Stat(plistPath); err != nil {
		return false, nil
	}
	uid := strconv.Itoa(os.Getuid())
	cmd := exec.Command("launchctl", "print", "gui/"+uid+"/"+launchAgentLabel)
	return cmd.Run() == nil, nil
}

func openBrowser(url string) error {
	return exec.Command("open", url).Run()
}
