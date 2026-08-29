package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"huzbackend-go/internal/config"
)

// defaultEnvContent is used when there is no .env.example next to the binary.
const defaultEnvContent = `PORT=3300
ADMIN_USERNAME=admin
ADMIN_PASSWORD=onemilusd
COOKIE_SECURE=false
SESSION_PERSISTENT=true
DB_PATH=data/app.db
`

var allowedConfigKeys = map[string]bool{
	"PORT": true, "ADMIN_USERNAME": true, "ADMIN_PASSWORD": true,
	"COOKIE_SECURE": true, "SESSION_PERSISTENT": true, "DB_PATH": true,
}

func envFilePath() string        { return filepath.Join(mustCwd(), ".env") }
func envExampleFilePath() string { return filepath.Join(mustCwd(), ".env.example") }

func runConfig(args []string) int {
	if len(args) == 0 {
		printConfigUsage()
		return 2
	}
	switch args[0] {
	case "list":
		return configList()
	case "get":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Missing key. Usage: huzbackend config get <KEY>")
			return 2
		}
		return configGet(args[1])
	case "set":
		if len(args) < 3 {
			fmt.Fprintln(os.Stderr, "Missing arguments. Usage: huzbackend config set <KEY> <VALUE>")
			return 2
		}
		return configSet(args[1], args[2])
	case "reset":
		return configReset()
	default:
		printConfigUsage()
		return 2
	}
}

func printConfigUsage() {
	fmt.Println(`Usage:
  huzbackend config list                 Show current configuration
  huzbackend config get <KEY>            Show a single value (ADMIN_PASSWORD is masked)
  huzbackend config set <KEY> <VALUE>    Set a value, written to .env
  huzbackend config reset                Restore default configuration

Valid keys: PORT, ADMIN_USERNAME, ADMIN_PASSWORD, COOKIE_SECURE, SESSION_PERSISTENT, DB_PATH`)
}

func configList() int {
	cfg := config.Load()
	if _, err := os.Stat(envFilePath()); err != nil {
		fmt.Println("(.env file does not exist yet — using default values.)")
	}
	fmt.Printf("PORT                = %s\n", cfg.Port)
	fmt.Printf("ADMIN_USERNAME      = %s\n", cfg.AdminUsername)
	fmt.Printf("ADMIN_PASSWORD      = %s\n", maskSecret(cfg.AdminPassword))
	fmt.Printf("COOKIE_SECURE       = %t\n", cfg.CookieSecure)
	fmt.Printf("SESSION_PERSISTENT  = %t\n", cfg.SessionPersistent)
	fmt.Printf("DB_PATH             = %s\n", cfg.DBPath)
	fmt.Printf("(Configuration file: %s)\n", envFilePath())
	return 0
}

func configGet(key string) int {
	key = strings.ToUpper(strings.TrimSpace(key))
	if !allowedConfigKeys[key] {
		fmt.Fprintf(os.Stderr, "Invalid key: %s\n", key)
		printConfigUsage()
		return 2
	}
	cfg := config.Load()
	val := configValue(cfg, key)
	if key == "ADMIN_PASSWORD" {
		val = maskSecret(val)
	}
	fmt.Printf("%s=%s\n", key, val)
	return 0
}

func configValue(cfg *config.Config, key string) string {
	switch key {
	case "PORT":
		return cfg.Port
	case "ADMIN_USERNAME":
		return cfg.AdminUsername
	case "ADMIN_PASSWORD":
		return cfg.AdminPassword
	case "COOKIE_SECURE":
		return fmt.Sprintf("%t", cfg.CookieSecure)
	case "SESSION_PERSISTENT":
		return fmt.Sprintf("%t", cfg.SessionPersistent)
	case "DB_PATH":
		return cfg.DBPath
	}
	return ""
}

func maskSecret(string) string {
	return "•••••• (hidden — change with 'huzbackend config set ADMIN_PASSWORD ...')"
}

func configSet(key, value string) int {
	key = strings.ToUpper(strings.TrimSpace(key))
	value = strings.TrimSpace(value)
	if !allowedConfigKeys[key] {
		fmt.Fprintf(os.Stderr, "Invalid key: %s\n", key)
		printConfigUsage()
		return 2
	}
	if value == "" {
		fmt.Fprintln(os.Stderr, "Value must not be empty.")
		return 2
	}
	path := envFilePath()
	if err := ensureEnvFile(path); err != nil {
		fmt.Fprintf(os.Stderr, "Could not create .env: %v\n", err)
		return 1
	}
	if err := setEnvKey(path, key, value); err != nil {
		fmt.Fprintf(os.Stderr, "Could not write .env: %v\n", err)
		return 1
	}
	if key == "ADMIN_PASSWORD" {
		fmt.Printf("Set ADMIN_PASSWORD (hidden, file: %s)\n", path)
	} else {
		fmt.Printf("Set %s=%s (file: %s)\n", key, value, path)
	}
	return 0
}

func configReset() int {
	path := envFilePath()
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Could not remove .env: %v\n", err)
		return 1
	}
	if err := ensureEnvFile(path); err != nil {
		fmt.Fprintf(os.Stderr, "Could not recreate the default .env: %v\n", err)
		return 1
	}
	fmt.Printf("Restored the default configuration (%s).\n", path)
	return 0
}

// ensureEnvFile creates .env if it does not exist (prefers copying .env.example).
func ensureEnvFile(path string) error {
	if fileExists(path) {
		return nil
	}
	if example := envExampleFilePath(); fileExists(example) {
		if data, err := os.ReadFile(example); err == nil {
			return os.WriteFile(path, data, 0o644)
		}
	}
	return os.WriteFile(path, []byte(defaultEnvContent), 0o644)
}

// setEnvKey updates the KEY line in .env, keeping comments and ordering intact.
func setEnvKey(path, key, value string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var out []string
	found := false
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "#") && strings.Contains(trimmed, "=") {
			k := strings.TrimSpace(strings.SplitN(trimmed, "=", 2)[0])
			if k == key {
				out = append(out, key+"="+value)
				found = true
				continue
			}
		}
		out = append(out, line)
	}
	if !found {
		out = append(out, key+"="+value)
	}
	return os.WriteFile(path, []byte(strings.Join(out, "\n")+"\n"), 0o644)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
