package usecase_test

import (
	"context"
	"testing"

	"finpharm-ai/services/ai-auditor/internal/domain"
	"finpharm-ai/services/ai-auditor/internal/usecase"
)

func TestAuditTransaction_Approved(t *testing.T) {
	uc := usecase.NewAuditUsecase()

	result, err := uc.AuditTransaction(context.Background(), domain.AuditTransactionRequest{
		TransactionID: "TXN-20260322130000-AAAA1111",
		Items: []domain.AuditTransactionItem{
			{MedicineID: "PARA500", Qty: 2},
			{MedicineID: "AMOX500", Qty: 1},
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Decision != domain.AuditDecisionApproved {
		t.Fatalf("expected APPROVED, got %s", result.Decision)
	}
	if result.Provider != "mock" {
		t.Fatalf("expected provider mock, got %q", result.Provider)
	}
}

func TestAuditTransaction_ReviewForHighRiskMedicine(t *testing.T) {
	uc := usecase.NewAuditUsecase()

	result, err := uc.AuditTransaction(context.Background(), domain.AuditTransactionRequest{
		TransactionID: "TXN-20260322130000-BBBB2222",
		Items: []domain.AuditTransactionItem{
			{MedicineID: "OBATKERAS-X", Qty: 2},
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Decision != domain.AuditDecisionReview {
		t.Fatalf("expected REVIEW, got %s", result.Decision)
	}
	if result.RiskScore <= 0.5 {
		t.Fatalf("expected risk score > 0.5, got %f", result.RiskScore)
	}
}

func TestAuditTransaction_MissingTransactionID(t *testing.T) {
	uc := usecase.NewAuditUsecase()

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