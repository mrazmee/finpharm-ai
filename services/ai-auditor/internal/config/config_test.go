package config

import (
	"testing"
	"time"
)

func TestValidateFallbackConfigOK(t *testing.T) {
	cfg := Config{
		AppEnv:          "local",
		Port:            "8083",
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    5 * time.Second,
		IdleTimeout:     30 * time.Second,
		ShutdownTimeout: 7 * time.Second,
		AuditProvider:   "fallback",
		GeminiTimeout:   3 * time.Second,
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid config, got error: %v", err)
	}
}

func TestValidateRejectsUnknownProvider(t *testing.T) {
	cfg := Config{
		AppEnv:          "local",
		Port:            "8083",
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    5 * time.Second,
		IdleTimeout:     30 * time.Second,
		ShutdownTimeout: 7 * time.Second,
		AuditProvider:   "other",
		GeminiTimeout:   3 * time.Second,
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestValidateGeminiRequiresAPIKeyOutsideLocal(t *testing.T) {
	cfg := Config{
		AppEnv:          "prod",
		Port:            "8083",
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    5 * time.Second,
		IdleTimeout:     30 * time.Second,
		ShutdownTimeout: 7 * time.Second,
		AuditProvider:   "gemini",
		GeminiModel:     "gemini-2.5-flash",
		GeminiTimeout:   3 * time.Second,
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error, got nil")
	}
}