package usecase_test

import (
	"context"
	"errors"
	"testing"

	"finpharm-ai/services/ai-auditor/internal/domain"
	"finpharm-ai/services/ai-auditor/internal/usecase"
)

type fakeProvider struct {
	result domain.AuditTransactionResult
	err    error
	calls  int
}

func (f *fakeProvider) AuditTransaction(ctx context.Context, req domain.AuditTransactionRequest) (domain.AuditTransactionResult, error) {
	f.calls++
	if f.err != nil {
		return domain.AuditTransactionResult{}, f.err
	}
	return f.result, nil
}

func TestAuditTransaction_UsesPrimaryProvider(t *testing.T) {
	primary := &fakeProvider{
		result: domain.AuditTransactionResult{
			Decision:  domain.AuditDecisionApproved,
			RiskScore: 0.12,
			Reason:    "gemini result",
			Provider:  "gemini",
			Model:     "gemini-2.5-flash",
		},
	}
	fallback := &fakeProvider{
		result: domain.AuditTransactionResult{
			Decision:  domain.AuditDecisionReview,
			RiskScore: 0.55,
			Reason:    "fallback review",
			Provider:  "fallback",
			Model:     "safe-review-v1",
		},
	}

	uc := usecase.NewAuditUsecase(primary, fallback)

	result, err := uc.AuditTransaction(context.Background(), domain.AuditTransactionRequest{
		TransactionID: "TXN-1",
		Items: []domain.AuditTransactionItem{
			{MedicineID: "PARA500", Qty: 2},
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Provider != "gemini" {
		t.Fatalf("expected provider gemini, got %q", result.Provider)
	}
	if primary.calls != 1 {
		t.Fatalf("expected primary called once, got %d", primary.calls)
	}
	if fallback.calls != 0 {
		t.Fatalf("expected fallback not called, got %d", fallback.calls)
	}
}

func TestAuditTransaction_FallsBackWhenPrimaryFails(t *testing.T) {
	primary := &fakeProvider{err: errors.New("gemini timeout")}
	fallback := &fakeProvider{
		result: domain.AuditTransactionResult{
			Decision:  domain.AuditDecisionReview,
			RiskScore: 0.55,
			Reason:    "gemini unavailable; fallback review required",
			Provider:  "fallback",
			Model:     "safe-review-v1",
		},
	}

	uc := usecase.NewAuditUsecase(primary, fallback)

	result, err := uc.AuditTransaction(context.Background(), domain.AuditTransactionRequest{
		TransactionID: "TXN-2",
		Items: []domain.AuditTransactionItem{
			{MedicineID: "PARA500", Qty: 2},
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Provider != "fallback" {
		t.Fatalf("expected provider fallback, got %q", result.Provider)
	}
	if result.Decision != domain.AuditDecisionReview {
		t.Fatalf("expected REVIEW, got %s", result.Decision)
	}
	if primary.calls != 1 || fallback.calls != 1 {
		t.Fatalf("expected both providers called once, got primary=%d fallback=%d", primary.calls, fallback.calls)
	}
}

func TestAuditTransaction_MissingTransactionID(t *testing.T) {
	primary := &fakeProvider{}
	fallback := &fakeProvider{}
	uc := usecase.NewAuditUsecase(primary, fallback)

	_, err := uc.AuditTransaction(context.Background(), domain.AuditTransactionRequest{
		Items: []domain.AuditTransactionItem{
			{MedicineID: "PARA500", Qty: 1},
		},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	ve, ok := domain.IsValidation(err)
	if !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}
	if ve.Field != "transaction_id" {
		t.Fatalf("expected field transaction_id, got %s", ve.Field)
	}
}