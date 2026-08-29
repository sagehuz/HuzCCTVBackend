//go:build linux

package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const systemdUnitName = "huzcctv.service"

func systemdUnitPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "systemd", "user", systemdUnitName), nil
}

func AutoStartEnable() error {
	unitPath, err := systemdUnitPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(unitPath), 0o755); err != nil {
		return err
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	dir := filepath.Dir(exe)
	content := fmt.Sprintf(`[Unit]
Description=Huz CCTV Server
After=network.target

[Service]
ExecStart=%s serve
WorkingDirectory=%s
Restart=on-failure

[Install]
WantedBy=default.target
`, exe, dir)
	if err := os.WriteFile(unitPath, []byte(content), 0o644); err != nil {
		return err
	}
	_ = exec.Command("systemctl", "--user", "daemon-reload").Run()
	if err := exec.Command("systemctl", "--user", "enable", systemdUnitName).Run(); err != nil {
		return fmt.Errorf("systemctl --user enable failed: %v", err)
	}
	return nil
}

func AutoStartDisable() error {
	unitPath, err := systemdUnitPath()
	if err != nil {
		return err
	}
	_ = exec.Command("systemctl", "--user", "disable", systemdUnitName).Run()
	_ = exec.Command("systemctl", "--user", "stop", systemdUnitName).Run()
	if err := os.Remove(unitPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func AutoStartStatus() (bool, error) {
	unitPath, err := systemdUnitPath()
	if err != nil {
		return false, err
	}
	if _, err := os.Stat(unitPath); err != nil {
		return false, nil
	}
	out, err := exec.Command("systemctl", "--user", "is-enabled", systemdUnitName).Output()
	if err != nil {
		return false, nil
	}
	return strings.TrimSpace(string(out)) == "enabled", nil
}

func openBrowser(url string) error {
	return exec.Command("xdg-open", url).Run()
}
