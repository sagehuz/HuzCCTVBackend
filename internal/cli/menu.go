package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"huzbackend-go/internal/config"
)

// RunMenu shows the interactive menu. Used by the `huzbackend menu` command and
// when the binary is run without arguments from a terminal.
func RunMenu() int {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		printMenu()
		fmt.Print("Choose [1-9, q]: ")
		if !scanner.Scan() {
			break
		}
		switch strings.TrimSpace(scanner.Text()) {
		case "1":
			runStart(nil)
		case "2":
			runStop(nil)
		case "3":
			runRestart(nil)
		case "4":
			runStatus(nil)
		case "5":
			runLogs([]string{"-n", "50"})
		case "6":
			runOpen(nil)
		case "7":
			autostartMenu(scanner)
		case "8":
			configMenu(scanner)
		case "9", "q", "Q", "exit", "quit":
			fmt.Println("Goodbye!")
			return 0
		default:
			fmt.Println("Invalid choice.")
		}
	}
	return 0
}

func printMenu() {
	running, pid := serverRunning()
	cfg := config.Load()
	status := "🔴 STOPPED"
	if running {
		status = fmt.Sprintf("🟢 RUNNING (PID %d)", pid)
	}
	fmt.Println()
	fmt.Println("════════════════════════════════════════════════")
	fmt.Printf("   Huz CCTV Server %s — management\n", Version)
	fmt.Printf("   Status: %-22s | Port: %s\n", status, cfg.Port)
	fmt.Println("════════════════════════════════════════════════")
	fmt.Println("  1) Start server")
	fmt.Println("  2) Stop server")
	fmt.Println("  3) Restart server")
	fmt.Println("  4) Detailed status")
	fmt.Println("  5) View logs (last 50 lines)")
	fmt.Println("  6) Open web dashboard")
	fmt.Println("  7) Autostart at login")
	fmt.Println("  8) Configuration (.env)")
	fmt.Println("  9) Quit")
}

func autostartMenu(scanner *bufio.Scanner) {
	for {
		state := "?"
		if on, err := AutoStartStatus(); err == nil {
			if on {
				state = "ENABLED"
			} else {
				state = "DISABLED"
			}
		}
		fmt.Println("  ── Autostart at login (current: " + state + ") ──")
		fmt.Println("  1) Enable")
		fmt.Println("  2) Disable")
		fmt.Println("  3) Back")
		fmt.Print("Choose [1-3]: ")
		if !scanner.Scan() {
			return
		}
		switch strings.TrimSpace(scanner.Text()) {
		case "1":
			runAutostart([]string{"on"})
		case "2":
			runAutostart([]string{"off"})
		default:
			return
		}
	}
}

func configMenu(scanner *bufio.Scanner) {
	if running, _ := serverRunning(); running {
		fmt.Println("  (Note: the server is running — configuration changes need a 'Restart' to take effect.)")
	}
	for {
		fmt.Println("  ── Configuration (.env) ──")
		fmt.Println("  1) Show current configuration")
		fmt.Println("  2) Change PORT")
		fmt.Println("  3) Change ADMIN_USERNAME")
		fmt.Println("  4) Change ADMIN_PASSWORD")
		fmt.Println("  5) Change DB_PATH")
		fmt.Println("  6) Change COOKIE_SECURE (true/false)")
		fmt.Println("  7) Change SESSION_PERSISTENT (true/false)")
		fmt.Println("  8) Restore default configuration")
		fmt.Println("  9) Back")
		fmt.Print("Choose [1-9]: ")
		if !scanner.Scan() {
			return
		}
		switch strings.TrimSpace(scanner.Text()) {
		case "1":
			configList()
		case "2":
			promptSetConfig(scanner, "PORT", "Enter new PORT (e.g. 3300): ", validPort)
		case "3":
			promptSetConfig(scanner, "ADMIN_USERNAME", "Enter new ADMIN_USERNAME: ", nil)
		case "4":
			promptSetConfig(scanner, "ADMIN_PASSWORD", "Enter new ADMIN_PASSWORD: ", nil)
		case "5":
			promptSetConfig(scanner, "DB_PATH", "Enter new DB_PATH (e.g. data/app.db): ", nil)
		case "6":
			promptSetConfig(scanner, "COOKIE_SECURE", "COOKIE_SECURE (true/false): ", validBool)
		case "7":
			promptSetConfig(scanner, "SESSION_PERSISTENT", "SESSION_PERSISTENT (true/false): ", validBool)
		case "8":
			configReset()
		default:
			return
		}
	}
}

func promptSetConfig(scanner *bufio.Scanner, key, prompt string, validate func(string) error) {
	fmt.Print(prompt)
	if !scanner.Scan() {
		return
	}
	value := strings.TrimSpace(scanner.Text())
	if value == "" {
		fmt.Println("Skipped — empty value.")
		return
	}
	if validate != nil {
		if err := validate(value); err != nil {
			fmt.Println("Skipped — " + err.Error())
			return
		}
	}
	configSet(key, value)
}

func validPort(s string) error {
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 || n > 65535 {
		return errors.New("PORT must be a number between 1 and 65535")
	}
	return nil
}

func validBool(s string) error {
	switch strings.ToLower(s) {
	case "true", "false":
		return nil
	}
	return errors.New("only true or false is accepted")
}
