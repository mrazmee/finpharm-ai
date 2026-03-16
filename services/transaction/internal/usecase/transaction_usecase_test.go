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
	createResult domain.CreateTransactionResult
	err         error

	updateStatusCalls int
	updatedID         string
	updatedStatus     domain.TransactionStatus
	updateErr         error

	getByKeyCalls int
	capturedKey   string
	existing      domain.Transaction
	found         bool
	getErr        error

	listCalls    int
	capturedList domain.ListTransactionsRequest
	listResult   domain.ListTransactionsResult
	listErr      error
}

func (f *fakeTransactionRepo) Create(ctx context.Context, tx domain.Transaction) (domain.CreateTransactionResult, error) {
	f.createCalls++
	f.captured = tx
	if f.err != nil {
		return domain.CreateTransactionResult{}, f.err
	}
	if f.createResult.Transaction.ID != "" {
		return f.createResult, nil
	}

	tx.CreatedAt = time.Date(2026, 3, 16, 10, 0, 0, 0, time.UTC)
	tx.UpdatedAt = tx.CreatedAt
	return domain.CreateTransactionResult{
		Transaction: tx,
		IsReplay:    false,
	}, nil
}

func (f *fakeTransactionRepo) UpdateStatus(ctx context.Context, transactionID string, status domain.TransactionStatus) error {
	f.updateStatusCalls++
	f.updatedID = transactionID
	f.updatedStatus = status
	return f.updateErr
}

func (f *fakeTransactionRepo) GetByIdempotencyKey(ctx context.Context, key string) (domain.Transaction, bool, error) {
	f.getByKeyCalls++
	f.capturedKey = key
	if f.getErr != nil {
		return domain.Transaction{}, false, f.getErr
	}
	return f.existing, f.found, nil
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
	available    map[string]int
	getCalls     int
	deductCalls  int
	deductErrFor map[string]error
}

func (f *fakeStockRepo) GetAvailableQty(ctx context.Context, medicineID string, requestedQty int) (int, error) {
	f.getCalls++
	qty, ok := f.available[medicineID]
	if !ok {
		return 0, &domain.NotFoundError{Resource: "medicine", Key: medicineID}
	}
	return qty, nil
}

func (f *fakeStockRepo) DeductStock(ctx context.Context, medicineID string, qty int) error {
	f.deductCalls++
	if err, ok := f.deductErrFor[medicineID]; ok {
		return err
	}

	available, ok := f.available[medicineID]
	if !ok {
		return &domain.NotFoundError{Resource: "medicine", Key: medicineID}
	}
	if available < qty {
		return &domain.InsufficientStockError{
			MedicineID:   medicineID,
			RequestedQty: qty,
			AvailableQty: available,
		}
	}

	f.available[medicineID] = available - qty
	return nil
}

func TestCreateTransaction_Success(t *testing.T) {
	repo := &fakeTransactionRepo{}
	stockRepo := &fakeStockRepo{
		available: map[string]int{
			"PARA500": 80,
			"AMOX500": 25,
		},
		deductErrFor: map[string]error{},
	}

	uc := usecase.NewTransactionUsecase(repo, stockRepo)

	result, err := uc.CreateTransaction(context.Background(), domain.CreateTransactionRequest{
		IdempotencyKey: "idem-001",
		Items: []domain.TransactionItemInput{
			{MedicineID: "PARA500", Qty: 10},
			{MedicineID: "AMOX500", Qty: 2},
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.getByKeyCalls != 1 {
		t.Fatalf("expected repo.GetByIdempotencyKey called once, got %d", repo.getByKeyCalls)
	}
	if repo.createCalls != 1 {
		t.Fatalf("expected repo.Create called once, got %d", repo.createCalls)
	}
	if repo.updateStatusCalls != 1 {
		t.Fatalf("expected repo.UpdateStatus called once, got %d", repo.updateStatusCalls)
	}
	if repo.updatedStatus != domain.TransactionStatusApproved {
		t.Fatalf("expected final status APPROVED, got %s", repo.updatedStatus)
	}
	if stockRepo.getCalls != 2 {
		t.Fatalf("expected 2 stock checks, got %d", stockRepo.getCalls)
	}
	if stockRepo.deductCalls != 2 {
		t.Fatalf("expected 2 stock deduct calls, got %d", stockRepo.deductCalls)
	}
	if result.IsReplay {
		t.Fatal("expected first request not replay")
	}
	if result.Transaction.Status != domain.TransactionStatusApproved {
		t.Fatalf("expected status APPROVED, got %s", result.Transaction.Status)
	}
	if repo.captured.IdempotencyKey != "idem-001" {
		t.Fatalf("expected idempotency key idem-001, got %q", repo.captured.IdempotencyKey)
	}
}

func TestCreateTransaction_ReplayByIdempotencyKey(t *testing.T) {
	existing := domain.Transaction{
		ID:             "TXN-20260316100000-AAAA1111",
		IdempotencyKey: "idem-001",
		Status:         domain.TransactionStatusApproved,
		Items: []domain.TransactionItem{
			{MedicineID: "PARA500", Qty: 2},
		},
		CreatedAt: time.Date(2026, 3, 16, 10, 0, 0, 0, time.UTC),
	}

	repo := &fakeTransactionRepo{
		existing: existing,
		found:    true,
	}
	stockRepo := &fakeStockRepo{
		available:    map[string]int{"PARA500": 80},
		deductErrFor: map[string]error{},
	}

	uc := usecase.NewTransactionUsecase(repo, stockRepo)

	result, err := uc.CreateTransaction(context.Background(), domain.CreateTransactionRequest{
		IdempotencyKey: "idem-001",
		Items: []domain.TransactionItemInput{
			{MedicineID: "PARA500", Qty: 2},
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.IsReplay {
		t.Fatal("expected replay result")
	}
	if result.Transaction.ID != existing.ID {
		t.Fatalf("expected existing transaction id %q, got %q", existing.ID, result.Transaction.ID)
	}
	if repo.createCalls != 0 {
		t.Fatalf("expected repo.Create not called, got %d", repo.createCalls)
	}
	if repo.updateStatusCalls != 0 {
		t.Fatalf("expected repo.UpdateStatus not called, got %d", repo.updateStatusCalls)
	}
	if stockRepo.getCalls != 0 || stockRepo.deductCalls != 0 {
		t.Fatalf("expected stock repo not called on replay, got checks=%d deduct=%d", stockRepo.getCalls, stockRepo.deductCalls)
	}
}

func TestCreateTransaction_MissingIdempotencyKey(t *testing.T) {
	repo := &fakeTransactionRepo{}
	stockRepo := &fakeStockRepo{
		available:    map[string]int{},
		deductErrFor: map[string]error{},
	}

	uc := usecase.NewTransactionUsecase(repo, stockRepo)

	_, err := uc.CreateTransaction(context.Background(), domain.CreateTransactionRequest{
		Items: []domain.TransactionItemInput{
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
	if ve.Field != "idempotency_key" {
		t.Fatalf("expected field idempotency_key, got %s", ve.Field)
	}
	if repo.createCalls != 0 {
		t.Fatalf("expected repo.Create not called, got %d", repo.createCalls)
	}
}

func TestCreateTransaction_InsufficientStockBeforeCreate(t *testing.T) {
	repo := &fakeTransactionRepo{}
	stockRepo := &fakeStockRepo{
		available: map[string]int{
			"PARA500": 3,
		},
		deductErrFor: map[string]error{},
	}

	uc := usecase.NewTransactionUsecase(repo, stockRepo)

	_, err := uc.CreateTransaction(context.Background(), domain.CreateTransactionRequest{
		IdempotencyKey: "idem-002",
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

func TestCreateTransaction_DeductFailureMarksFailed(t *testing.T) {
	repo := &fakeTransactionRepo{}
	stockRepo := &fakeStockRepo{
		available: map[string]int{
			"PARA500": 10,
			"AMOX500": 10,
		},
		deductErrFor: map[string]error{
			"AMOX500": &domain.InsufficientStockError{
				MedicineID:   "AMOX500",
				RequestedQty: 2,
				AvailableQty: 1,
			},
		},
	}

	uc := usecase.NewTransactionUsecase(repo, stockRepo)

	result, err := uc.CreateTransaction(context.Background(), domain.CreateTransactionRequest{
		IdempotencyKey: "idem-003",
		Items: []domain.TransactionItemInput{
			{MedicineID: "PARA500", Qty: 2},
			{MedicineID: "AMOX500", Qty: 2},
		},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if _, ok := domain.IsInsufficientStock(err); !ok {
		t.Fatalf("expected InsufficientStockError, got %T", err)
	}
	if repo.createCalls != 1 {
		t.Fatalf("expected repo.Create called once, got %d", repo.createCalls)
	}
	if repo.updateStatusCalls != 1 {
		t.Fatalf("expected repo.UpdateStatus called once, got %d", repo.updateStatusCalls)
	}
	if repo.updatedStatus != domain.TransactionStatusFailed {
		t.Fatalf("expected FAILED status, got %s", repo.updatedStatus)
	}
	if result.Transaction.Status != domain.TransactionStatusFailed {
		t.Fatalf("expected returned transaction status FAILED, got %s", result.Transaction.Status)
	}
}

func TestListTransactions_DefaultPagination(t *testing.T) {
	repo := &fakeTransactionRepo{
		listResult: domain.ListTransactionsResult{
			Items: []domain.Transaction{
				{
					ID:     "TXN-20260316093000-AAAA1111",
					Status: domain.TransactionStatusApproved,
					Items: []domain.TransactionItem{
						{MedicineID: "PARA500", Qty: 2},
					},
					CreatedAt: time.Date(2026, 3, 16, 9, 30, 0, 0, time.UTC),
				},
			},
			Limit:  10,
			Offset: 0,
			Total:  1,
		},
	}
	stockRepo := &fakeStockRepo{
		available:    map[string]int{},
		deductErrFor: map[string]error{},
	}

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
	stockRepo := &fakeStockRepo{
		available:    map[string]int{},
		deductErrFor: map[string]error{},
	}

	uc := usecase.NewTransactionUsecase(repo, stockRepo)

	_, err := uc.ListTransactions(context.Background(), domain.ListTransactionsRequest{
		Limit:  5,
		Offset: 10,
		Status: "approved",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.listCalls != 1 {
		t.Fatalf("expected repo.List called once, got %d", repo.listCalls)
	}
	if repo.capturedList.Status != "APPROVED" {
		t.Fatalf("expected normalized status APPROVED, got %q", repo.capturedList.Status)
	}
}

func TestListTransactions_InvalidLimit(t *testing.T) {
	repo := &fakeTransactionRepo{}
	stockRepo := &fakeStockRepo{
		available:    map[string]int{},
		deductErrFor: map[string]error{},
	}

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
	stockRepo := &fakeStockRepo{
		available:    map[string]int{},
		deductErrFor: map[string]error{},
	}

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