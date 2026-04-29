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

	TxForceAuditApproved bool
	TxForceDeductFailure bool
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

		TxForceAuditApproved: getEnvBool("TX_FORCE_AUDIT_APPROVED", false),
		TxForceDeductFailure: getEnvBool("TX_FORCE_DEDUCT_FAILURE", false),
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
	if c.AIAuditorTimeout <= 0 {
		return errConfig("AI_AUDITOR_TIMEOUT_MS must be > 0")
	}

	if strings.TrimSpace(c.InventoryBaseURL) == "" {
		return errConfig("INVENTORY_BASE_URL is required")
	}
	if err := validateBaseURL(c.InventoryBaseURL, "INVENTORY_BASE_URL"); err != nil {
		return err
	}
	if strings.TrimSpace(c.AIAuditorBaseURL) == "" {
		return errConfig("AI_AUDITOR_BASE_URL is required")
	}
	if err := validateBaseURL(c.AIAuditorBaseURL, "AI_AUDITOR_BASE_URL"); err != nil {
		return err
	}

	if strings.TrimSpace(c.DBHost) == "" {
		return errConfig("DB_HOST is required")
	}
	if strings.TrimSpace(c.DBPort) == "" || !isPositiveIntegerString(c.DBPort) {
		return errConfig("DB_PORT must be a positive integer")
	}
	if strings.TrimSpace(c.DBUser) == "" {
		return errConfig("DB_USER is required")
	}
	if strings.TrimSpace(c.DBName) == "" {
		return errConfig("DB_NAME is required")
	}
	if strings.TrimSpace(c.DBSSLMode) == "" {
		return errConfig("DB_SSLMODE is required")
	}

	if strings.TrimSpace(c.RabbitMQURL) != "" {
		if err := validateAMQPURL(c.RabbitMQURL, "RABBITMQ_URL"); err != nil {
			return err
		}
		if strings.TrimSpace(c.RabbitMQExchange) == "" {
			return errConfig("RABBITMQ_EXCHANGE is required when RABBITMQ_URL is set")
		}
		if strings.TrimSpace(c.RabbitMQTransactionApprovedQueue) == "" {
			return errConfig("RABBITMQ_TRANSACTION_APPROVED_QUEUE is required when RABBITMQ_URL is set")
		}
		if strings.TrimSpace(c.RabbitMQTransactionApprovedRouting) == "" {
			return errConfig("RABBITMQ_TRANSACTION_APPROVED_ROUTING_KEY is required when RABBITMQ_URL is set")
		}
	}

	if (c.TxForceAuditApproved || c.TxForceDeductFailure) && !c.IsLocalEnv() {
		return errConfig("TX_FORCE_AUDIT_APPROVED and TX_FORCE_DEDUCT_FAILURE are allowed only in local/dev environment")
	}

	return nil
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
	return c.IsLocalEnv()
}

func (c Config) IsLocalEnv() bool {
	env := strings.ToLower(strings.TrimSpace(c.AppEnv))
	switch env {
	case "", "local", "dev", "development":
		return true
	default:
		return false
	}
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

func validateAMQPURL(raw string, field string) error {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return errConfig(field + " must be a valid absolute AMQP URL")
	}
	if parsed.Scheme != "amqp" && parsed.Scheme != "amqps" {
		return errConfig(field + " must use amqp or amqps scheme")
	}
	return nil
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