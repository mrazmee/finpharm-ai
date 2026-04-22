package config

import (
	"testing"
	"time"
)

func TestValidateOK(t *testing.T) {
	cfg := Config{
		AppEnv:                             "local",
		Port:                               "8081",
		ReadTimeout:                        5 * time.Second,
		WriteTimeout:                       5 * time.Second,
		IdleTimeout:                        30 * time.Second,
		ShutdownTimeout:                    7 * time.Second,
		InventoryBaseURL:                   "http://localhost:8082",
		AIAuditorBaseURL:                   "http://localhost:8083",
		AIAuditorTimeout:                   5 * time.Second,
		RabbitMQURL:                        "amqp://guest:guest@localhost:5672/",
		RabbitMQExchange:                   "finpharm.events",
		RabbitMQTransactionApprovedQueue:   "transaction.approved.queue",
		RabbitMQTransactionApprovedRouting: "transaction.approved",
		DBHost:                             "127.0.0.1",
		DBPort:                             "55432",
		DBUser:                             "finpharm",
		DBPassword:                         "finpharm",
		DBName:                             "transaction_db",
		DBSSLMode:                          "disable",
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid config, got error: %v", err)
	}
}

func TestValidateRejectsInvalidInventoryBaseURL(t *testing.T) {
	cfg := Config{
		AppEnv:           "local",
		Port:             "8081",
		ReadTimeout:      5 * time.Second,
		WriteTimeout:     5 * time.Second,
		IdleTimeout:      30 * time.Second,
		ShutdownTimeout:  7 * time.Second,
		InventoryBaseURL: "localhost:8082",
		AIAuditorBaseURL: "http://localhost:8083",
		AIAuditorTimeout: 5 * time.Second,
		DBHost:           "127.0.0.1",
		DBPort:           "55432",
		DBUser:           "finpharm",
		DBPassword:       "finpharm",
		DBName:           "transaction_db",
		DBSSLMode:        "disable",
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestValidateRejectsInvalidRabbitMQURL(t *testing.T) {
	cfg := Config{
		AppEnv:                             "local",
		Port:                               "8081",
		ReadTimeout:                        5 * time.Second,
		WriteTimeout:                       5 * time.Second,
		IdleTimeout:                        30 * time.Second,
		ShutdownTimeout:                    7 * time.Second,
		InventoryBaseURL:                   "http://localhost:8082",
		AIAuditorBaseURL:                   "http://localhost:8083",
		AIAuditorTimeout:                   5 * time.Second,
		RabbitMQURL:                        "http://localhost:5672",
		RabbitMQExchange:                   "finpharm.events",
		RabbitMQTransactionApprovedQueue:   "transaction.approved.queue",
		RabbitMQTransactionApprovedRouting: "transaction.approved",
		DBHost:                             "127.0.0.1",
		DBPort:                             "55432",
		DBUser:                             "finpharm",
		DBPassword:                         "finpharm",
		DBName:                             "transaction_db",
		DBSSLMode:                          "disable",
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error, got nil")
	}
}