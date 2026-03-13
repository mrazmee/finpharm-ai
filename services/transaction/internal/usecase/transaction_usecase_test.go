package usecase_test

import (
	"context"
	"testing"
	"time"

	"finpharm-ai/services/transaction/internal/domain"
	"finpharm-ai/services/transaction/internal/usecase"
)

type fakeTransactionRepo struct {
	createCalls int
	captured    domain.Transaction
	result      domain.Transaction
	err         error
}

func (f *fakeTransactionRepo) Create(ctx context.Context, tx domain.Transaction) (domain.Transaction, error) {
	f.createCalls++
	f.captured = tx
	if f.err != nil {
		return domain.Transaction{}, f.err
	}
	if f.result.ID != "" {
		return f.result, nil
	}

	tx.CreatedAt = time.Date(2026, 3, 12, 10, 0, 0, 0, time.UTC)
	tx.UpdatedAt = tx.CreatedAt
	return tx, nil
}

type fakeStockRepo struct {
	available map[string]int
	calls     int
}

func (f *fakeStockRepo) GetAvailableQty(ctx context.Context, medicineID string, requestedQty int) (int, error) {
	f.calls++
	qty, ok := f.available[medicineID]
	if !ok {
		return 0, &domain.NotFoundError{Resource: "medicine", Key: medicineID}
	}
	return qty, nil
}

func TestCreateTransaction_Success(t *testing.T) {
	repo := &fakeTransactionRepo{}
	stockRepo := &fakeStockRepo{
		available: map[string]int{
			"PARA500": 80,
			"AMOX500": 25,
		},
	}

	uc := usecase.NewTransactionUsecase(repo, stockRepo)

	result, err := uc.CreateTransaction(context.Background(), domain.CreateTransactionRequest{
		Items: []domain.TransactionItemInput{
			{MedicineID: "PARA500", Qty: 10},
			{MedicineID: "AMOX500", Qty: 2},
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.createCalls != 1 {
		t.Fatalf("expected repo.Create called once, got %d", repo.createCalls)
	}
	if stockRepo.calls != 2 {
		t.Fatalf("expected 2 stock checks, got %d", stockRepo.calls)
	}
	if result.Status != domain.TransactionStatusPending {
		t.Fatalf("expected status PENDING, got %s", result.Status)
	}
	if len(result.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result.Items))
	}
	if repo.captured.ID == "" {
		t.Fatal("expected generated transaction id, got empty")
	}
	if repo.captured.Items[0].MedicineID != "PARA500" {
		t.Fatalf("expected first item PARA500, got %s", repo.captured.Items[0].MedicineID)
	}
}

func TestCreateTransaction_InsufficientStock(t *testing.T) {
	repo := &fakeTransactionRepo{}
	stockRepo := &fakeStockRepo{
		available: map[string]int{
			"PARA500": 3,
		},
	}

	uc := usecase.NewTransactionUsecase(repo, stockRepo)

	_, err := uc.CreateTransaction(context.Background(), domain.CreateTransactionRequest{
		Items: []domain.TransactionItemInput{
			{MedicineID: "PARA500", Qty: 10},
		},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	insufficient, ok := domain.IsInsufficientStock(err)
	if !ok {
		t.Fatalf("expected InsufficientStockError, got %T", err)
	}
	if insufficient.MedicineID != "PARA500" {
		t.Fatalf("expected medicine PARA500, got %s", insufficient.MedicineID)
	}
	if repo.createCalls != 0 {
		t.Fatalf("expected repo.Create not called, got %d", repo.createCalls)
	}
}
