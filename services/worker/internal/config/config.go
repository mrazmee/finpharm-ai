package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppEnv          string
	WorkerName      string
	RabbitMQURL     string
	RabbitMQExchange string
	QueueName       string
	RoutingKey      string
	ConsumerTag     string
	PrefetchCount   int
	ShutdownTimeout time.Duration
}

func Load() Config {
	return Config{
		AppEnv:           getEnv("APP_ENV", "local"),
		WorkerName:       getEnv("WORKER_NAME", "notification-worker"),
		RabbitMQURL:      getEnv("RABBITMQ_URL", "amqp://finpharm:finpharm@localhost:5672/"),
		RabbitMQExchange: getEnv("RABBITMQ_EXCHANGE", "finpharm.events"),
		QueueName:        getEnv("RABBITMQ_TRANSACTION_APPROVED_QUEUE", "transaction.approved.queue"),
		RoutingKey:       getEnv("RABBITMQ_TRANSACTION_APPROVED_ROUTING_KEY", "transaction.approved"),
		ConsumerTag:      getEnv("RABBITMQ_CONSUMER_TAG", "worker.transaction.approved"),
		PrefetchCount:    getEnvInt("RABBITMQ_PREFETCH_COUNT", 10),
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