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

	listCalls    int
	capturedList domain.ListTransactionsRequest
	listResult   domain.ListTransactionsResult
	listErr      error
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

func (f *fakeTransactionRepo) List(ctx context.Context, req domain.ListTransactionsRequest) (domain.ListTransactionsResult, error) {
	f.listCalls++
	f.capturedList = req
	if f.listErr != nil {
		return domain.ListTransactionsResult{}, f.listErr
	}
	return f.listResult, nil
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

func TestListTransactions_DefaultPagination(t *testing.T) {
	repo := &fakeTransactionRepo{
		listResult: domain.ListTransactionsResult{
			Items: []domain.Transaction{
				{
					ID:     "TXN-20260313093000-AAAA1111",
					Status: domain.TransactionStatusPending,
					Items: []domain.TransactionItem{
						{MedicineID: "PARA500", Qty: 2},
					},
					CreatedAt: time.Date(2026, 3, 13, 9, 30, 0, 0, time.UTC),
				},
			},
			Limit:  10,
			Offset: 0,
			Total:  1,
		},
	}
	stockRepo := &fakeStockRepo{}

	uc := usecase.NewTransactionUsecase(repo, stockRepo)

	result, err := uc.ListTransactions(context.Background(), domain.ListTransactionsRequest{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.listCalls != 1 {
		t.Fatalf("expected repo.List called once, got %d", repo.listCalls)
	}
	if repo.capturedList.Limit != 10 {
		t.Fatalf("expected default limit 10, got %d", repo.capturedList.Limit)
	}
	if repo.capturedList.Offset != 0 {
		t.Fatalf("expected default offset 0, got %d", repo.capturedList.Offset)
	}
	if result.Total != 1 {
		t.Fatalf("expected total 1, got %d", result.Total)
	}
}

func TestListTransactions_WithFilter(t *testing.T) {
	repo := &fakeTransactionRepo{
		listResult: domain.ListTransactionsResult{
			Items:  []domain.Transaction{},
			Limit:  5,
			Offset: 10,
			Total:  0,
		},
	}
	stockRepo := &fakeStockRepo{}

	uc := usecase.NewTransactionUsecase(repo, stockRepo)

	_, err := uc.ListTransactions(context.Background(), domain.ListTransactionsRequest{
		Limit:  5,
		Offset: 10,
		Status: "pending",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.listCalls != 1 {
		t.Fatalf("expected repo.List called once, got %d", repo.listCalls)
	}
	if repo.capturedList.Limit != 5 {
		t.Fatalf("expected limit 5, got %d", repo.capturedList.Limit)
	}
	if repo.capturedList.Offset != 10 {
		t.Fatalf("expected offset 10, got %d", repo.capturedList.Offset)
	}
	if repo.capturedList.Status != "PENDING" {
		t.Fatalf("expected normalized status PENDING, got %q", repo.capturedList.Status)
	}
}

func TestListTransactions_InvalidLimit(t *testing.T) {
	repo := &fakeTransactionRepo{}
	stockRepo := &fakeStockRepo{}

	uc := usecase.NewTransactionUsecase(repo, stockRepo)

	_, err := uc.ListTransactions(context.Background(), domain.ListTransactionsRequest{
		Limit: -1,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	ve, ok := domain.IsValidation(err)
	if !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}
	if ve.Field != "limit" {
		t.Fatalf("expected field limit, got %s", ve.Field)
	}
	if repo.listCalls != 0 {
		t.Fatalf("expected repo.List not called, got %d", repo.listCalls)
	}
}

func TestListTransactions_InvalidOffset(t *testing.T) {
	repo := &fakeTransactionRepo{}
	stockRepo := &fakeStockRepo{}

	uc := usecase.NewTransactionUsecase(repo, stockRepo)

	_, err := uc.ListTransactions(context.Background(), domain.ListTransactionsRequest{
		Offset: -1,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	ve, ok := domain.IsValidation(err)
	if !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}
	if ve.Field != "offset" {
		t.Fatalf("expected field offset, got %s", ve.Field)
	}
	if repo.listCalls != 0 {
		t.Fatalf("expected repo.List not called, got %d", repo.listCalls)
	}
}