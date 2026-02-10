package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port               string
	TransactionBaseURL string

	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

func Load() Config {
	port := getEnv("PORT", "8080")
	txURL := getEnv("TRANSACTION_BASE_URL", "http://localhost:8081")

	readMs := getEnvInt("READ_TIMEOUT_MS", 5000)
	writeMs := getEnvInt("WRITE_TIMEOUT_MS", 5000)
	idleMs := getEnvInt("IDLE_TIMEOUT_MS", 30000)
	shutdownMs := getEnvInt("SHUTDOWN_TIMEOUT_MS", 7000)

	return Config{
		Port:               port,
		TransactionBaseURL: txURL,
		ReadTimeout:        time.Duration(readMs) * time.Millisecond,
		WriteTimeout:       time.Duration(writeMs) * time.Millisecond,
		IdleTimeout:        time.Duration(idleMs) * time.Millisecond,
		ShutdownTimeout:    time.Duration(shutdownMs) * time.Millisecond,
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
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}
