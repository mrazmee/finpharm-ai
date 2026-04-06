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

	InventoryBaseURL string
	AIAuditorBaseURL string
	AIAuditorTimeout time.Duration

	RabbitMQURL                        string
	RabbitMQExchange                   string
	RabbitMQTransactionApprovedQueue   string
	RabbitMQTransactionApprovedRouting string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
}

func Load() Config {
	readMs := getEnvInt("READ_TIMEOUT_MS", 5000)
	writeMs := getEnvInt("WRITE_TIMEOUT_MS", 5000)
	idleMs := getEnvInt("IDLE_TIMEOUT_MS", 30000)
	shutdownMs := getEnvInt("SHUTDOWN_TIMEOUT_MS", 7000)
	aiAuditorTimeoutMs := getEnvInt("AI_AUDITOR_TIMEOUT_MS", 5000)

	return Config{
		AppEnv: getEnv("APP_ENV", "local"),
		Port:   getEnv("PORT", "8081"),

		ReadTimeout:     time.Duration(readMs) * time.Millisecond,
		WriteTimeout:    time.Duration(writeMs) * time.Millisecond,
		IdleTimeout:     time.Duration(idleMs) * time.Millisecond,
		ShutdownTimeout: time.Duration(shutdownMs) * time.Millisecond,

		InventoryBaseURL: getEnv("INVENTORY_BASE_URL", "http://localhost:8082"),
		AIAuditorBaseURL: getEnv("AI_AUDITOR_BASE_URL", "http://localhost:8083"),
		AIAuditorTimeout: time.Duration(aiAuditorTimeoutMs) * time.Millisecond,

		RabbitMQURL:                        getEnv("RABBITMQ_URL", "amqp://finpharm:finpharm@localhost:5672/"),
		RabbitMQExchange:                   getEnv("RABBITMQ_EXCHANGE", "finpharm.events"),
		RabbitMQTransactionApprovedQueue:   getEnv("RABBITMQ_TRANSACTION_APPROVED_QUEUE", "transaction.approved.queue"),
		RabbitMQTransactionApprovedRouting: getEnv("RABBITMQ_TRANSACTION_APPROVED_ROUTING_KEY", "transaction.approved"),

		DBHost:     getEnv("DB_HOST", "127.0.0.1"),
		DBPort:     getEnv("DB_PORT", "55432"),
		DBUser:     getEnv("DB_USER", "finpharm"),
		DBPassword: getEnv("DB_PASSWORD", "finpharm"),
		DBName:     getEnv("DB_NAME", "transaction_db"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
}

func (c Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost,
		c.DBPort,
		c.DBUser,
		c.DBPassword,
		c.DBName,
		c.DBSSLMode,
	)
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
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}