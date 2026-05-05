package config

import (
	"testing"
	"time"
)

func TestValidateOK(t *testing.T) {
	cfg := Config{
		AppEnv:           "local",
		WorkerName:       "notification-worker",
		RabbitMQURL:      "amqp://guest:guest@localhost:5672/",
		RabbitMQExchange: "finpharm.events",
		QueueName:        "transaction.approved.queue",
		RoutingKey:       "transaction.approved",
		RetryQueueName:   "transaction.approved.retry.queue",
		RetryRoutingKey:  "transaction.approved.retry",
		DLQName:          "transaction.approved.dlq",
		DLQRoutingKey:    "transaction.approved.dlq",
		ConsumerTag:      "worker.transaction.approved",
		PrefetchCount:    10,
		MaxRetryCount:    3,
		RetryDelayMs:     5000,
		MetricsPort:      "9094",
		ShutdownTimeout:  7 * time.Second,
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid config, got error: %v", err)
	}
}

func TestValidateRejectsInvalidAMQPURL(t *testing.T) {
	cfg := Config{
		AppEnv:           "local",
		WorkerName:       "notification-worker",
		RabbitMQURL:      "http://localhost:5672",
		RabbitMQExchange: "finpharm.events",
		QueueName:        "transaction.approved.queue",
		RoutingKey:       "transaction.approved",
		RetryQueueName:   "transaction.approved.retry.queue",
		RetryRoutingKey:  "transaction.approved.retry",
		DLQName:          "transaction.approved.dlq",
		DLQRoutingKey:    "transaction.approved.dlq",
		ConsumerTag:      "worker.transaction.approved",
		PrefetchCount:    10,
		MaxRetryCount:    3,
		RetryDelayMs:     5000,
		MetricsPort:      "9094",
		ShutdownTimeout:  7 * time.Second,
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error, got nil")
	}
}