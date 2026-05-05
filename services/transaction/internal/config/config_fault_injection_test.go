package config

import "testing"

func TestValidate_RejectsFaultInjectionOutsideLocal(t *testing.T) {
	cfg := Load()
	cfg.AppEnv = "production"
	cfg.TxForceAuditApproved = true

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for TX_FORCE_AUDIT_APPROVED outside local")
	}

	cfg = Load()
	cfg.AppEnv = "production"
	cfg.TxForceDeductFailure = true

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for TX_FORCE_DEDUCT_FAILURE outside local")
	}
}

func TestValidate_AllowsFaultInjectionInLocal(t *testing.T) {
	cfg := Load()
	cfg.AppEnv = "local"
	cfg.TxForceAuditApproved = true
	cfg.TxForceDeductFailure = true

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected validation to pass in local, got %v", err)
	}
}