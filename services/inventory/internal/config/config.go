package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv string
	Port   string

	StorageDriver string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

func Load() Config {
	appEnv := getEnv("APP_ENV", "local")
	port := getEnv("PORT", "8082")

	storageDriver := getEnv("STORAGE_DRIVER", "memory")

	dbHost := getEnv("DB_HOST", "127.0.0.1")
	dbPort := getEnv("DB_PORT", "55432")
	dbUser := getEnv("DB_USER", "finpharm")
	dbPassword := getEnv("DB_PASSWORD", "finpharm")
	dbName := getEnv("DB_NAME", "inventory_db")
	dbSSLMode := getEnv("DB_SSLMODE", "disable")

	readMs := getEnvInt("READ_TIMEOUT_MS", 5000)
	writeMs := getEnvInt("WRITE_TIMEOUT_MS", 5000)
	idleMs := getEnvInt("IDLE_TIMEOUT_MS", 30000)
	shutdownMs := getEnvInt("SHUTDOWN_TIMEOUT_MS", 7000)

	return Config{
		AppEnv:          appEnv,
		Port:            port,
		StorageDriver:   storageDriver,
		DBHost:          dbHost,
		DBPort:          dbPort,
		DBUser:          dbUser,
		DBPassword:      dbPassword,
		DBName:          dbName,
		DBSSLMode:       dbSSLMode,
		ReadTimeout:     time.Duration(readMs) * time.Millisecond,
		WriteTimeout:    time.Duration(writeMs) * time.Millisecond,
		IdleTimeout:     time.Duration(idleMs) * time.Millisecond,
		ShutdownTimeout: time.Duration(shutdownMs) * time.Millisecond,
	}
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.Port) == "" {
		return errConfig("PORT is required")
	}
	if !isPositiveIntegerString(c.Port) {
		return errConfig("PORT must be a positive integer")
	}

	driver := strings.ToLower(strings.TrimSpace(c.StorageDriver))
	switch driver {
	case "memory", "postgres":
	default:
		return errConfig("STORAGE_DRIVER must be either memory or postgres")
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

	if driver == "postgres" {
		if strings.TrimSpace(c.DBHost) == "" {
			return errConfig("DB_HOST is required when STORAGE_DRIVER=postgres")
		}
		if strings.TrimSpace(c.DBPort) == "" || !isPositiveIntegerString(c.DBPort) {
			return errConfig("DB_PORT must be a positive integer when STORAGE_DRIVER=postgres")
		}
		if strings.TrimSpace(c.DBUser) == "" {
			return errConfig("DB_USER is required when STORAGE_DRIVER=postgres")
		}
		if strings.TrimSpace(c.DBName) == "" {
			return errConfig("DB_NAME is required when STORAGE_DRIVER=postgres")
		}
		if strings.TrimSpace(c.DBSSLMode) == "" {
			return errConfig("DB_SSLMODE is required when STORAGE_DRIVER=postgres")
		}
	}

	return nil
}

func (c Config) IsDebugEnabled() bool {
	return c.AppEnv == "local" || c.AppEnv == "dev"
}

func (c Config) DBConnString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		url.QueryEscape(c.DBUser),
		url.QueryEscape(c.DBPassword),
		c.DBHost,
		c.DBPort,
		c.DBName,
		url.QueryEscape(c.DBSSLMode),
	)
}

func errConfig(msg string) error {
	return fmt.Errorf("config validation error: %s", msg)
}

func isPositiveIntegerString(v string) bool {
	n, err := strconv.Atoi(strings.TrimSpace(v))
	return err == nil && n > 0
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