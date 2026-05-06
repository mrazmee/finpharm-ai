package config

import "testing"

func TestValidateForMigrate_OK(t *testing.T) {
	cfg := Load()
	if err := cfg.ValidateForMigrate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidateForIngest_RejectsOverlapTooLarge(t *testing.T) {
	cfg := Load()
	cfg.ChunkMaxChars = 500
	cfg.ChunkOverlapChars = 500

	if err := cfg.ValidateForIngest(); err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestValidateForIngest_RequiresAPIKeyWhenNotDryRun(t *testing.T) {
	cfg := Load()
	cfg.GeminiAPIKey = ""
	cfg.DryRun = false

	if err := cfg.ValidateForIngest(); err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestValidateForIngest_AllowsDryRunWithoutAPIKey(t *testing.T) {
	cfg := Load()
	cfg.GeminiAPIKey = ""
	cfg.DryRun = true

	if err := cfg.ValidateForIngest(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidateForQuery_RequiresAPIKey(t *testing.T) {
	cfg := Load()
	cfg.GeminiAPIKey = ""

	if err := cfg.ValidateForQuery(); err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestValidateForQuery_OK(t *testing.T) {
	cfg := Load()
	cfg.GeminiAPIKey = "dummy-key"

	if err := cfg.ValidateForQuery(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidateForAnswer_OK(t *testing.T) {
	cfg := Load()
	cfg.GeminiAPIKey = "dummy-key"

	if err := cfg.ValidateForAnswer(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidateForAnswer_RejectsBadTemperature(t *testing.T) {
	cfg := Load()
	cfg.GeminiAPIKey = "dummy-key"
	cfg.AnswerTemperature = 2.5

	if err := cfg.ValidateForAnswer(); err == nil {
		t.Fatal("expected validation error, got nil")
	}
}