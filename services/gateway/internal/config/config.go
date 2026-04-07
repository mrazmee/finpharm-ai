package config

import (
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

	InventoryBaseURL   string
	TransactionBaseURL string

	AuthEnabled      bool
	JWTSecret        string
	JWTIssuer        string
	JWTExpireMinutes int
}

func Load() Config {
	readMs := getEnvInt("READ_TIMEOUT_MS", 5000)
	writeMs := getEnvInt("WRITE_TIMEOUT_MS", 5000)
	idleMs := getEnvInt("IDLE_TIMEOUT_MS", 30000)
	shutdownMs := getEnvInt("SHUTDOWN_TIMEOUT_MS", 7000)

	return Config{
		AppEnv: getEnv("APP_ENV", "local"),
		Port:   getEnv("PORT", "8080"),

		ReadTimeout:     time.Duration(readMs) * time.Millisecond,
		WriteTimeout:    time.Duration(writeMs) * time.Millisecond,
		IdleTimeout:     time.Duration(idleMs) * time.Millisecond,
		ShutdownTimeout: time.Duration(shutdownMs) * time.Millisecond,

		InventoryBaseURL:   getEnv("INVENTORY_BASE_URL", "http://localhost:8082"),
		TransactionBaseURL: getEnv("TRANSACTION_BASE_URL", "http://localhost:8081"),

		AuthEnabled:      getEnvBool("AUTH_ENABLED", false),
		JWTSecret:        getEnv("JWT_SECRET", "finpharm-local-secret"),
		JWTIssuer:        getEnv("JWT_ISSUER", "finpharm-gateway"),
		JWTExpireMinutes: getEnvInt("JWT_EXPIRE_MINUTES", 60),
	}
}

func (c Config) IsDebugEnabled() bool {
	env := strings.ToLower(strings.TrimSpace(c.AppEnv))
	switch env {
	case "local", "dev", "development":
		return true
	default:
		return false
	}
}

func getEnv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func getEnvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
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