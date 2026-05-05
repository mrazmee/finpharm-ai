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
	AppEnv           string
	WorkerName       string
	RabbitMQURL      string
	RabbitMQExchange string
	QueueName        string
	RoutingKey       string
	RetryQueueName   string
	RetryRoutingKey  string
	DLQName          string
	DLQRoutingKey    string
	ConsumerTag      string
	PrefetchCount    int
	MaxRetryCount    int
	RetryDelayMs     int
	MetricsPort      string
	ShutdownTimeout  time.Duration
}

func Load() Config {
	return Config{
		AppEnv:           getEnv("APP_ENV", "local"),
		WorkerName:       getEnv("WORKER_NAME", "notification-worker"),
		RabbitMQURL:      getEnv("RABBITMQ_URL", "amqp://finpharm:finpharm@localhost:5672/"),
		RabbitMQExchange: getEnv("RABBITMQ_EXCHANGE", "finpharm.events"),
		QueueName:        getEnv("RABBITMQ_TRANSACTION_APPROVED_QUEUE", "transaction.approved.queue"),
		RoutingKey:       getEnv("RABBITMQ_TRANSACTION_APPROVED_ROUTING_KEY", "transaction.approved"),
		RetryQueueName:   getEnv("RABBITMQ_TRANSACTION_APPROVED_RETRY_QUEUE", "transaction.approved.retry.queue"),
		RetryRoutingKey:  getEnv("RABBITMQ_TRANSACTION_APPROVED_RETRY_ROUTING_KEY", "transaction.approved.retry"),
		DLQName:          getEnv("RABBITMQ_TRANSACTION_APPROVED_DLQ", "transaction.approved.dlq"),
		DLQRoutingKey:    getEnv("RABBITMQ_TRANSACTION_APPROVED_DLQ_ROUTING_KEY", "transaction.approved.dlq"),
		ConsumerTag:      getEnv("RABBITMQ_CONSUMER_TAG", "worker.transaction.approved"),
		PrefetchCount:    getEnvInt("RABBITMQ_PREFETCH_COUNT", 10),
		MaxRetryCount:    getEnvInt("RABBITMQ_MAX_RETRY_COUNT", 3),
		RetryDelayMs:     getEnvInt("RABBITMQ_RETRY_DELAY_MS", 5000),
		MetricsPort:      getEnv("WORKER_METRICS_PORT", "9094"),
		ShutdownTimeout:  time.Duration(getEnvInt("SHUTDOWN_TIMEOUT_MS", 7000)) * time.Millisecond,
	}
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.WorkerName) == "" {
		return errConfig("WORKER_NAME is required")
	}
	if strings.TrimSpace(c.RabbitMQURL) == "" {
		return errConfig("RABBITMQ_URL is required")
	}
	if err := validateAMQPURL(c.RabbitMQURL, "RABBITMQ_URL"); err != nil {
		return err
	}
	if strings.TrimSpace(c.RabbitMQExchange) == "" {
		return errConfig("RABBITMQ_EXCHANGE is required")
	}
	if strings.TrimSpace(c.QueueName) == "" {
		return errConfig("RABBITMQ_TRANSACTION_APPROVED_QUEUE is required")
	}
	if strings.TrimSpace(c.RoutingKey) == "" {
		return errConfig("RABBITMQ_TRANSACTION_APPROVED_ROUTING_KEY is required")
	}
	if strings.TrimSpace(c.RetryQueueName) == "" {
		return errConfig("RABBITMQ_TRANSACTION_APPROVED_RETRY_QUEUE is required")
	}
	if strings.TrimSpace(c.RetryRoutingKey) == "" {
		return errConfig("RABBITMQ_TRANSACTION_APPROVED_RETRY_ROUTING_KEY is required")
	}
	if strings.TrimSpace(c.DLQName) == "" {
		return errConfig("RABBITMQ_TRANSACTION_APPROVED_DLQ is required")
	}
	if strings.TrimSpace(c.DLQRoutingKey) == "" {
		return errConfig("RABBITMQ_TRANSACTION_APPROVED_DLQ_ROUTING_KEY is required")
	}
	if strings.TrimSpace(c.ConsumerTag) == "" {
		return errConfig("RABBITMQ_CONSUMER_TAG is required")
	}
	if c.PrefetchCount <= 0 {
		return errConfig("RABBITMQ_PREFETCH_COUNT must be > 0")
	}
	if c.MaxRetryCount < 0 {
		return errConfig("RABBITMQ_MAX_RETRY_COUNT must be >= 0")
	}
	if c.RetryDelayMs <= 0 {
		return errConfig("RABBITMQ_RETRY_DELAY_MS must be > 0")
	}
	if strings.TrimSpace(c.MetricsPort) == "" || !isPositiveIntegerString(c.MetricsPort) {
		return errConfig("WORKER_METRICS_PORT must be a positive integer")
	}
	if c.ShutdownTimeout <= 0 {
		return errConfig("SHUTDOWN_TIMEOUT_MS must be > 0")
	}
	return nil
}

func errConfig(msg string) error {
	return fmt.Errorf("config validation error: %s", msg)
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
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}