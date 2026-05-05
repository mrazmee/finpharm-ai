package config

import (
	"testing"
	"time"
)

func TestValidateMemoryConfigOK(t *testing.T) {
	cfg := Config{
		AppEnv:          "local",
		Port:            "8082",
		StorageDriver:   "memory",
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    5 * time.Second,
		IdleTimeout:     30 * time.Second,
		ShutdownTimeout: 7 * time.Second,
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid config, got error: %v", err)
	}
}

func TestValidatePostgresConfigRequiresDBFields(t *testing.T) {
	cfg := Config{
		AppEnv:          "local",
		Port:            "8082",
		StorageDriver:   "postgres",
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    5 * time.Second,
		IdleTimeout:     30 * time.Second,
		ShutdownTimeout: 7 * time.Second,
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestValidateRejectsUnknownStorageDriver(t *testing.T) {
	cfg := Config{
		AppEnv:          "local",
		Port:            "8082",
		StorageDriver:   "sqlite",
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    5 * time.Second,
		IdleTimeout:     30 * time.Second,
		ShutdownTimeout: 7 * time.Second,
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error, got nil")
	}
}