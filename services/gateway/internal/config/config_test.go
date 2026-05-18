package config

import (
	"testing"
	"time"
)

func TestValidate_OK(t *testing.T) {
	cfg := Config{
		AppEnv:                "local",
		Port:                  "8080",
		InventoryBaseURL:      "http://localhost:8082",
		TransactionBaseURL:    "http://localhost:8081",
		KnowledgeBaseURL:      "http://localhost:8084",
		ReadTimeout:           5 * time.Second,
		WriteTimeout:          5 * time.Second,
		IdleTimeout:           30 * time.Second,
		ShutdownTimeout:       5 * time.Second,
		AuthEnabled:           true,
		JWTSecret:             "secret",
		JWTIssuer:             "issuer",
		JWTExpireMinutes:      60,
		RateLimitEnabled:      true,
		RateLimitGeneralLimit: 60,
		RateLimitAuthLimit:    20,
		RateLimitWindow:       60 * time.Second,
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid config, got error: %v", err)
	}
}

func TestValidate_MissingJWTSecretWhenAuthEnabled(t *testing.T) {
	cfg := Config{
		AppEnv:                "local",
		Port:                  "8080",
		InventoryBaseURL:      "http://localhost:8082",
		TransactionBaseURL:    "http://localhost:8081",
		KnowledgeBaseURL:      "http://localhost:8084",
		ReadTimeout:           5 * time.Second,
		WriteTimeout:          5 * time.Second,
		IdleTimeout:           30 * time.Second,
		ShutdownTimeout:       5 * time.Second,
		AuthEnabled:           true,
		JWTIssuer:             "issuer",
		JWTExpireMinutes:      60,
		RateLimitEnabled:      true,
		RateLimitGeneralLimit: 60,
		RateLimitAuthLimit:    20,
		RateLimitWindow:       60 * time.Second,
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestValidate_DefaultJWTSecretRejectedOutsideLocal(t *testing.T) {
	cfg := Config{
		AppEnv:                "prod",
		Port:                  "8080",
		InventoryBaseURL:      "http://localhost:8082",
		TransactionBaseURL:    "http://localhost:8081",
		KnowledgeBaseURL:      "http://localhost:8084",
		ReadTimeout:           5 * time.Second,
		WriteTimeout:          5 * time.Second,
		IdleTimeout:           30 * time.Second,
		ShutdownTimeout:       5 * time.Second,
		AuthEnabled:           true,
		JWTSecret:             "dev-secret-change-me",
		JWTIssuer:             "issuer",
		JWTExpireMinutes:      60,
		RateLimitEnabled:      true,
		RateLimitGeneralLimit: 60,
		RateLimitAuthLimit:    20,
		RateLimitWindow:       60 * time.Second,
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestValidate_InvalidInventoryBaseURL(t *testing.T) {
	cfg := Config{
		AppEnv:             "local",
		Port:               "8080",
		InventoryBaseURL:   "localhost:8082",
		TransactionBaseURL: "http://localhost:8081",
		KnowledgeBaseURL:   "http://localhost:8084",
		ReadTimeout:        5 * time.Second,
		WriteTimeout:       5 * time.Second,
		IdleTimeout:        30 * time.Second,
		ShutdownTimeout:    5 * time.Second,
		AuthEnabled:        false,
		RateLimitEnabled:   false,
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestValidate_InvalidKnowledgeBaseURL(t *testing.T) {
	cfg := Config{
		AppEnv:             "local",
		Port:               "8080",
		InventoryBaseURL:   "http://localhost:8082",
		TransactionBaseURL: "http://localhost:8081",
		KnowledgeBaseURL:   "localhost:8084",
		ReadTimeout:        5 * time.Second,
		WriteTimeout:       5 * time.Second,
		IdleTimeout:        30 * time.Second,
		ShutdownTimeout:    5 * time.Second,
		AuthEnabled:        false,
		RateLimitEnabled:   false,
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error, got nil")
	}
}