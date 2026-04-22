package config

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv             string
	Port               string
	InventoryBaseURL   string
	TransactionBaseURL string

	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration

	AuthEnabled      bool
	JWTSecret        string
	JWTIssuer        string
	JWTExpireMinutes int

	RateLimitEnabled      bool
	RateLimitGeneralLimit int
	RateLimitAuthLimit    int
	RateLimitWindow       time.Duration
}

func Load() Config {
	return Config{
		AppEnv:             getEnv("APP_ENV", "local"),
		Port:               getEnv("PORT", "8080"),
		InventoryBaseURL:   getEnv("INVENTORY_BASE_URL", "http://localhost:8082"),
		TransactionBaseURL: getEnv("TRANSACTION_BASE_URL", "http://localhost:8081"),

		ReadTimeout:     getEnvDurationMS("READ_TIMEOUT_MS", 5000),
		WriteTimeout:    getEnvDurationMS("WRITE_TIMEOUT_MS", 5000),
		IdleTimeout:     getEnvDurationMS("IDLE_TIMEOUT_MS", 30000),
		ShutdownTimeout: getEnvDurationMS("SHUTDOWN_TIMEOUT_MS", 5000),

		AuthEnabled:      getEnvBool("AUTH_ENABLED", true),
		JWTSecret:        getEnv("JWT_SECRET", "dev-secret-change-me"),
		JWTIssuer:        getEnv("JWT_ISSUER", "finpharm-gateway"),
		JWTExpireMinutes: getEnvInt("JWT_EXPIRE_MINUTES", 60),

		RateLimitEnabled:      getEnvBool("RATE_LIMIT_ENABLED", true),
		RateLimitGeneralLimit: getEnvInt("RATE_LIMIT_GENERAL_LIMIT", 60),
		RateLimitAuthLimit:    getEnvInt("RATE_LIMIT_AUTH_LIMIT", 20),
		RateLimitWindow:       getEnvDurationSeconds("RATE_LIMIT_WINDOW_SECONDS", 60),
	}
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.Port) == "" {
		return errConfig("PORT is required")
	}
	if !isPositiveIntegerString(c.Port) {
		return errConfig("PORT must be a positive integer")
	}
	if strings.TrimSpace(c.InventoryBaseURL) == "" {
		return errConfig("INVENTORY_BASE_URL is required")
	}
	if err := validateBaseURL(c.InventoryBaseURL, "INVENTORY_BASE_URL"); err != nil {
		return err
	}
	if strings.TrimSpace(c.TransactionBaseURL) == "" {
		return errConfig("TRANSACTION_BASE_URL is required")
	}
	if err := validateBaseURL(c.TransactionBaseURL, "TRANSACTION_BASE_URL"); err != nil {
		return err
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

	if c.AuthEnabled {
		if strings.TrimSpace(c.JWTSecret) == "" {
			return errConfig("JWT_SECRET is required when AUTH_ENABLED=true")
		}
		if strings.TrimSpace(c.JWTIssuer) == "" {
			return errConfig("JWT_ISSUER is required when AUTH_ENABLED=true")
		}
		if c.JWTExpireMinutes <= 0 {
			return errConfig("JWT_EXPIRE_MINUTES must be > 0")
		}
		if c.AppEnv != "local" && c.JWTSecret == "dev-secret-change-me" {
			return errConfig("JWT_SECRET must not use default dev value outside local environment")
		}
	}

	if c.RateLimitEnabled {
		if c.RateLimitGeneralLimit <= 0 {
			return errConfig("RATE_LIMIT_GENERAL_LIMIT must be > 0")
		}
		if c.RateLimitAuthLimit <= 0 {
			return errConfig("RATE_LIMIT_AUTH_LIMIT must be > 0")
		}
		if c.RateLimitWindow <= 0 {
			return errConfig("RATE_LIMIT_WINDOW_SECONDS must be > 0")
		}
	}

	return nil
}

// IsDebugEnabled dipertahankan karena masih dipakai oleh router/handler gateway.
func (c Config) IsDebugEnabled() bool {
	return strings.EqualFold(strings.TrimSpace(c.AppEnv), "local")
}

func errConfig(msg string) error {
	return fmt.Errorf("config validation error: %s", msg)
}

func validateBaseURL(raw string, field string) error {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return errConfig(field + " must be a valid absolute URL")
	}
	return nil
}

func isPositiveIntegerString(v string) bool {
	n, err := strconv.Atoi(strings.TrimSpace(v))
	return err == nil && n > 0
}

func getEnv(key string, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}

func getEnvInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}

	i, err := strconv.Atoi(v)
	if err != nil {
		slog.Warn("invalid_int_env", "key", key, "value", v, "fallback", fallback)
		return fallback
	}

	return i
}

func getEnvBool(key string, fallback bool) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if v == "" {
		return fallback
	}

	switch v {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		slog.Warn("invalid_bool_env", "key", key, "value", v, "fallback", fallback)
		return fallback
	}
}

func getEnvDurationMS(key string, fallbackMS int) time.Duration {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return time.Duration(fallbackMS) * time.Millisecond
	}

	i, err := strconv.Atoi(v)
	if err != nil {
		slog.Warn("invalid_duration_ms_env", "key", key, "value", v, "fallback_ms", fallbackMS)
		return time.Duration(fallbackMS) * time.Millisecond
	}

	return time.Duration(i) * time.Millisecond
}

func getEnvDurationSeconds(key string, fallbackSeconds int) time.Duration {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return time.Duration(fallbackSeconds) * time.Second
	}

	i, err := strconv.Atoi(v)
	if err != nil {
		slog.Warn("invalid_duration_seconds_env", "key", key, "value", v, "fallback_seconds", fallbackSeconds)
		return time.Duration(fallbackSeconds) * time.Second
	}

	return time.Duration(i) * time.Second
}