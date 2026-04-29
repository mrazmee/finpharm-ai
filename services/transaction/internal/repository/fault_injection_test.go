package repository

import (
	"context"
	"testing"

	"finpharm-ai/services/transaction/internal/domain"
)

type fakeStockRepo struct {
	availableQty int
}

func (f *fakeStockRepo) GetAvailableQty(ctx context.Context, medicineID string, requestedQty int) (int, error) {
	return f.availableQty, nil
}

func (f *fakeStockRepo) DeductStock(ctx context.Context, medicineID string, qty int) error {
	return nil
}

func TestForcedApprovedAIAuditorRepo_ReturnsApproved(t *testing.T) {
	repo := NewForcedApprovedAIAuditorRepo()

	result, err := repo.AuditTransaction(context.Background(), domain.AuditTransactionRequest{
		TransactionID: "TXN-TEST-001",
		Items: []domain.AuditTransactionItem{
			{MedicineID: "PARA500", Qty: 1},
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Decision != domain.AuditDecisionApproved {
		t.Fatalf("expected decision APPROVED, got %s", result.Decision)
	}
	if result.Provider != "local-fixture" {
		t.Fatalf("expected provider local-fixture, got %s", result.Provider)
	}
}

func TestForcedDeductFailureStockRepo_DelegatesCheckAndFailsDeduct(t *testing.T) {
	repo := NewForcedDeductFailureStockRepo(&fakeStockRepo{availableQty: 10})

	qty, err := repo.GetAvailableQty(context.Background(), "PARA500", 1)
	if err != nil {
		t.Fatalf("expected no error from GetAvailableQty, got %v", err)
	}
	if qty != 10 {
		t.Fatalf("expected available qty 10, got %d", qty)
	}

	err = repo.DeductStock(context.Background(), "PARA500", 1)
	if err == nil {
		t.Fatal("expected deduct failure error, got nil")
	}

	upstreamErr, ok := domain.IsUpstream(err)
	if !ok {
		t.Fatalf("expected upstream error, got %T", err)
	}
	if upstreamErr.Service != "inventory" {
		t.Fatalf("expected upstream service inventory, got %s", upstreamErr.Service)
	}
}