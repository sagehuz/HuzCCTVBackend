// Package cli provides the Huz CCTV Server management command set, embedded in
// the huzbackend binary itself so users do not need to touch scripts or complex
// terminal commands. Running `huzbackend` with no arguments still starts the
// server as before.
package cli

import (
	"fmt"
	"os"
	"runtime"
)

// Version is displayed by the `huzbackend version` command.
const Version = "1.0.0"

// Run executes a CLI command and returns the exit code for main().
func Run(args []string) int {
	if len(args) == 0 {
		printHelp()
		return 2
	}
	cmd, rest := args[0], args[1:]
	switch cmd {
	case "help", "--help", "-h":
		printHelp()
		return 0
	case "version":
		fmt.Printf("Huz CCTV Server %s (%s/%s)\n", Version, runtime.GOOS, runtime.GOARCH)
		return 0
	case "start":
		return runStart(rest)
	case "stop":
		return runStop(rest)
	case "restart":
		return runRestart(rest)
	case "status":
		return runStatus(rest)
	case "logs":
		return runLogs(rest)
	case "open":
		return runOpen(rest)
	case "autostart":
		return runAutostart(rest)
	case "config":
		return runConfig(rest)
	case "menu":
		return RunMenu()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %q\n", cmd)
		printHelp()
		return 2
	}
}

func printHelp() {
	fmt.Println(`Huz CCTV Server — management tool

Usage:
  huzbackend                                Open the interactive menu (when run from a terminal)
  huzbackend start                          Start the server in the background
  huzbackend stop                           Stop the server
  huzbackend restart                        Restart the server
  huzbackend status                         Show server status
  huzbackend logs [-n <lines>] [-f]         Show logs (default 50 lines; -f to follow)
  huzbackend open                           Open the web dashboard
  huzbackend autostart on|off|status        Start automatically at login
  huzbackend config list|get|set|reset      Manage .env configuration
  huzbackend menu                           Interactive menu mode
  huzbackend version                        Show version
  huzbackend help                           Show this help

Examples:
  huzbackend            # open the menu
  huzbackend status
  huzbackend autostart on
  huzbackend config set PORT 3301`)
}

// runAutostart enables, disables or shows the status of starting at login.
// Platform implementations: AutoStartEnable/AutoStartDisable/AutoStartStatus.
func runAutostart(args []string) int {
	if len(args) == 0 {
		printAutostartUsage()
		return 2
	}
	switch args[0] {
	case "on":
		if err := AutoStartEnable(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to enable autostart: %v\n", err)
			return 1
		}
		fmt.Println("Autostart enabled (the server will start automatically when you log in).")
	case "off":
		if err := AutoStartDisable(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to disable autostart: %v\n", err)
			return 1
		}
		fmt.Println("Autostart disabled.")
	case "status":
		on, err := AutoStartStatus()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not check autostart status: %v\n", err)
			return 1
		}
		if on {
			fmt.Println("Autostart at login: ENABLED")
		} else {
			fmt.Println("Autostart at login: DISABLED")
		}
	default:
		printAutostartUsage()
		return 2
	}
	return 0
}

func printAutostartUsage() {
	fmt.Println(`Usage:
  huzbackend autostart on       Enable autostart at login
  huzbackend autostart off      Disable autostart at login
  huzbackend autostart status   Show autostart status`)
}
