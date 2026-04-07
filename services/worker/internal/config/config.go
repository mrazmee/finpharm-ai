package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppEnv              string
	WorkerName          string
	RabbitMQURL         string
	RabbitMQExchange    string
	QueueName           string
	RoutingKey          string
	RetryQueueName      string
	RetryRoutingKey     string
	DLQName             string
	DLQRoutingKey       string
	ConsumerTag         string
	PrefetchCount       int
	MaxRetryCount       int
	RetryDelayMs        int
	ShutdownTimeout     time.Duration
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
		ShutdownTimeout:  time.Duration(getEnvInt("SHUTDOWN_TIMEOUT_MS", 7000)) * time.Millisecond,
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