package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port             string
	AdminUsername    string
	AdminPassword    string
	CookieSecure     bool
	SessionPersistent bool
	DBPath           string
}

func Load() *Config {
	cfg := &Config{
		Port:             "3300",
		AdminUsername:    "admin",
		AdminPassword:    "onemilusd",
		CookieSecure:     false,
		SessionPersistent: true,
		DBPath:           "data/app.db",
	}

	_ = loadDotEnv(cfg)
	cfg.Port = firstNonEmpty(os.Getenv("PORT"), cfg.Port)
	cfg.AdminUsername = firstNonEmpty(os.Getenv("ADMIN_USERNAME"), cfg.AdminUsername)
	cfg.AdminPassword = firstNonEmpty(os.Getenv("ADMIN_PASSWORD"), cfg.AdminPassword)
	cfg.CookieSecure = parseBool(os.Getenv("COOKIE_SECURE"), cfg.CookieSecure)
	cfg.SessionPersistent = parseBool(os.Getenv("SESSION_PERSISTENT"), cfg.SessionPersistent)
	cfg.DBPath = firstNonEmpty(os.Getenv("DB_PATH"), cfg.DBPath)

	if cfg.DBPath == "" {
		cfg.DBPath = "data/app.db"
	}
	return cfg
}

func loadDotEnv(cfg *Config) error {
	f, err := os.Open(".env")
	if err != nil {
		return err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		switch key {
		case "PORT":
			cfg.Port = val
		case "ADMIN_USERNAME":
			cfg.AdminUsername = val
		case "ADMIN_PASSWORD":
			cfg.AdminPassword = val
		case "COOKIE_SECURE":
			cfg.CookieSecure = parseBool(val, cfg.CookieSecure)
		case "SESSION_PERSISTENT":
			cfg.SessionPersistent = parseBool(val, cfg.SessionPersistent)
		case "DB_PATH":
			cfg.DBPath = val
		}
	}
	return s.Err()
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func parseBool(val string, fallback bool) bool {
	if strings.TrimSpace(val) == "" {
		return fallback
	}
	b, err := strconv.ParseBool(val)
	if err != nil {
		return fallback
	}
	return b
}
