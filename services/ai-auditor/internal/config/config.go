package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv string
	Port   string

	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration

	AuditProvider string
	AuditFailOpen bool

	GeminiAPIKey  string
	GeminiModel   string
	GeminiTimeout time.Duration
}

func Load() Config {
	appEnv := getEnv("APP_ENV", "local")
	port := getEnv("PORT", "8083")

	readMs := getEnvInt("READ_TIMEOUT_MS", 5000)
	writeMs := getEnvInt("WRITE_TIMEOUT_MS", 5000)
	idleMs := getEnvInt("IDLE_TIMEOUT_MS", 30000)
	shutdownMs := getEnvInt("SHUTDOWN_TIMEOUT_MS", 7000)

	return Config{
		AppEnv:          appEnv,
		Port:            port,
		ReadTimeout:     time.Duration(readMs) * time.Millisecond,
		WriteTimeout:    time.Duration(writeMs) * time.Millisecond,
		IdleTimeout:     time.Duration(idleMs) * time.Millisecond,
		ShutdownTimeout: time.Duration(shutdownMs) * time.Millisecond,

		AuditProvider: getEnv("AUDIT_PROVIDER", "gemini"),
		AuditFailOpen: getEnvBool("AUDIT_FAIL_OPEN", false),

		GeminiAPIKey:  firstNonEmpty(os.Getenv("GEMINI_API_KEY"), os.Getenv("GOOGLE_API_KEY")),
		GeminiModel:   getEnv("GEMINI_MODEL", "gemini-2.5-flash"),
		GeminiTimeout: time.Duration(getEnvInt("GEMINI_TIMEOUT_MS", 3000)) * time.Millisecond,
	}
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.Port) == "" {
		return errConfig("PORT is required")
	}
	if !isPositiveIntegerString(c.Port) {
		return errConfig("PORT must be a positive integer")
	}

	if c.ReadTimeout <= 0 {
		return errConfig("READ_TIMEOUT_MS must be > 0")
	}
	if c.WriteTimeout <= 0 {
		return errConfig("WRITE_TIMEOUT_MS must be > 0")
	}
	if c.IdleTimeout <= 0 {
		return errConfig("IDLE_TIMEOUT_MS must be > 0")
	}
	if c.ShutdownTimeout <= 0 {
		return errConfig("SHUTDOWN_TIMEOUT_MS must be > 0")
	}
	if c.GeminiTimeout <= 0 {
		return errConfig("GEMINI_TIMEOUT_MS must be > 0")
	}

	provider := normalizeProvider(c.AuditProvider)
	switch provider {
	case "gemini", "rule-based", "fallback":
	default:
		return errConfig("AUDIT_PROVIDER must be one of gemini, rule-based, or fallback")
	}

	if provider == "gemini" {
		if strings.TrimSpace(c.GeminiModel) == "" {
			return errConfig("GEMINI_MODEL is required when AUDIT_PROVIDER=gemini")
		}
		if !isLocalEnv(c.AppEnv) && strings.TrimSpace(c.GeminiAPIKey) == "" {
			return errConfig("GEMINI_API_KEY or GOOGLE_API_KEY is required outside local environment when AUDIT_PROVIDER=gemini")
		}
	}

	return nil
}

func errConfig(msg string) error {
	return fmt.Errorf("config validation error: %s", msg)
}

func isPositiveIntegerString(v string) bool {
	n, err := strconv.Atoi(strings.TrimSpace(v))
	return err == nil && n > 0
}

func normalizeProvider(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "mock", "rule_based", "rule-based":
		return "rule-based"
	default:
		return strings.ToLower(strings.TrimSpace(v))
	}
}

func isLocalEnv(env string) bool {
	switch strings.ToLower(strings.TrimSpace(env)) {
	case "", "local", "dev", "development":
		return true
	default:
		return false
	}
}

func getEnv(key, def string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	return v
}

func getEnvInt(key string, def int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}

func getEnvBool(key string, def bool) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if v == "" {
		return def
	}
	switch v {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return def
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}