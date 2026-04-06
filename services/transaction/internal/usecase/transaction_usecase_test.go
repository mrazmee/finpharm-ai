package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"finpharm-ai/services/transaction/internal/domain"
	"finpharm-ai/services/transaction/internal/usecase"
)

type fakeTransactionRepo struct {
	createCalls  int
	captured     domain.Transaction
	createResult domain.CreateTransactionResult
	err          error

	updateStatusCalls int
	updatedID         string
	updatedStatus     domain.TransactionStatus
	updateErr         error

	updateAuditCalls int
	auditedID        string
	auditValue       domain.TransactionAudit
	updateAuditErr   error

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

	tx.CreatedAt = time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)
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

func (f *fakeTransactionRepo) UpdateAudit(ctx context.Context, transactionID string, audit domain.TransactionAudit) error {
	f.updateAuditCalls++
	f.auditedID = transactionID
	f.auditValue = audit
	return f.updateAuditErr
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

type fakeAuditRepo struct {
	result domain.AuditTransactionResult
	err    error
	calls  int
}

func (f *fakeAuditRepo) AuditTransaction(ctx context.Context, req domain.AuditTransactionRequest) (domain.AuditTransactionResult, error) {
	f.calls++
	if f.err != nil {
		return domain.AuditTransactionResult{}, f.err
	}
	return f.result, nil
}

type fakePublisher struct {
	publishCalls int
	event        domain.TransactionApprovedEvent
	err          error
}

func (f *fakePublisher) PublishTransactionApproved(ctx context.Context, event domain.TransactionApprovedEvent) error {
	f.publishCalls++
	f.event = event
	return f.err
}

func TestCreateTransaction_Success_PublishesApprovedEvent(t *testing.T) {
	repo := &fakeTransactionRepo{}
	stockRepo := &fakeStockRepo{
		available: map[string]int{
			"PARA500": 80,
			"AMOX500": 25,
		},
		deductErrFor: map[string]error{},
	}
	auditRepo := &fakeAuditRepo{
		result: domain.AuditTransactionResult{
			Decision:  domain.AuditDecisionApproved,
			RiskScore: 0.05,
			Reason:    "safe transaction",
			Provider:  "gemini",
			Model:     "gemini-2.5-flash",
		},
	}
	publisher := &fakePublisher{}

	uc := usecase.NewTransactionUsecase(repo, stockRepo, auditRepo, publisher)

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
	if repo.updateStatusCalls != 1 || repo.updatedStatus != domain.TransactionStatusApproved {
		t.Fatalf("expected final status APPROVED, got updateCalls=%d status=%s", repo.updateStatusCalls, repo.updatedStatus)
	}
	if publisher.publishCalls != 1 {
		t.Fatalf("expected publish called once, got %d", publisher.publishCalls)
	}
	if publisher.event.EventName != "transaction.approved" {
		t.Fatalf("expected event name transaction.approved, got %q", publisher.event.EventName)
	}
	if publisher.event.TransactionID == "" {
		t.Fatal("expected transaction id in event")
	}
	if publisher.event.Audit == nil {
		t.Fatal("expected audit metadata in event")
	}
	if result.Transaction.Status != domain.TransactionStatusApproved {
		t.Fatalf("expected status APPROVED, got %s", result.Transaction.Status)
	}
}

func TestCreateTransaction_PublishFailure_DoesNotFailApprovedTransaction(t *testing.T) {
	repo := &fakeTransactionRepo{}
	stockRepo := &fakeStockRepo{
		available: map[string]int{
			"PARA500": 10,
		},
		deductErrFor: map[string]error{},
	}
	auditRepo := &fakeAuditRepo{
		result: domain.AuditTransactionResult{
			Decision:  domain.AuditDecisionApproved,
			RiskScore: 0.05,
			Reason:    "safe transaction",
			Provider:  "gemini",
			Model:     "gemini-2.5-flash",
		},
	}
	publisher := &fakePublisher{
		err: errors.New("rabbitmq unavailable"),
	}

	uc := usecase.NewTransactionUsecase(repo, stockRepo, auditRepo, publisher)

	result, err := uc.CreateTransaction(context.Background(), domain.CreateTransactionRequest{
		IdempotencyKey: "idem-002",
		Items: []domain.TransactionItemInput{
			{MedicineID: "PARA500", Qty: 1},
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Transaction.Status != domain.TransactionStatusApproved {
		t.Fatalf("expected status APPROVED, got %s", result.Transaction.Status)
	}
	if publisher.publishCalls != 1 {
		t.Fatalf("expected publish called once, got %d", publisher.publishCalls)
	}
}

func TestCreateTransaction_ReplayByIdempotencyKey(t *testing.T) {
	existing := domain.Transaction{
		ID:             "TXN-20260401100000-AAAA1111",
		IdempotencyKey: "idem-001",
		Status:         domain.TransactionStatusApproved,
		Items: []domain.TransactionItem{
			{MedicineID: "PARA500", Qty: 2},
		},
		Audit: &domain.TransactionAudit{
			Decision:  domain.AuditDecisionApproved,
			RiskScore: 0.05,
			Reason:    "safe transaction",
			Provider:  "gemini",
			Model:     "gemini-2.5-flash",
			AuditedAt: time.Date(2026, 4, 1, 10, 0, 1, 0, time.UTC),
		},
		CreatedAt: time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC),
	}

	repo := &fakeTransactionRepo{
		existing: existing,
		found:    true,
	}
	stockRepo := &fakeStockRepo{
		available:    map[string]int{"PARA500": 80},
		deductErrFor: map[string]error{},
	}
	auditRepo := &fakeAuditRepo{}
	publisher := &fakePublisher{}

	uc := usecase.NewTransactionUsecase(repo, stockRepo, auditRepo, publisher)

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
	if publisher.publishCalls != 0 {
		t.Fatalf("expected no publish on replay, got %d", publisher.publishCalls)
	}
}

func TestCreateTransaction_MissingIdempotencyKey(t *testing.T) {
	repo := &fakeTransactionRepo{}
	stockRepo := &fakeStockRepo{
		available:    map[string]int{},
		deductErrFor: map[string]error{},
	}
	auditRepo := &fakeAuditRepo{}
	publisher := &fakePublisher{}

	uc := usecase.NewTransactionUsecase(repo, stockRepo, auditRepo, publisher)

	_, err := uc.CreateTransaction(context.Background(), domain.CreateTransactionRequest{
		Items: []domain.TransactionItemInput{
			{MedicineID: "PARA500", Qty: 1},
		},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	ve, ok := domain.IsValidation(err)
	if !ok || ve.Field != "idempotency_key" {
		t.Fatalf("expected ValidationError on idempotency_key, got %v", err)
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
	auditRepo := &fakeAuditRepo{}
	publisher := &fakePublisher{}

	uc := usecase.NewTransactionUsecase(repo, stockRepo, auditRepo, publisher)

	_, err := uc.CreateTransaction(context.Background(), domain.CreateTransactionRequest{
		IdempotencyKey: "idem-003",
		Items: []domain.TransactionItemInput{
			{MedicineID: "PARA500", Qty: 10},
		},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := domain.IsInsufficientStock(err); !ok {
		t.Fatalf("expected InsufficientStockError, got %T", err)
	}
	if publisher.publishCalls != 0 {
		t.Fatalf("expected no publish, got %d", publisher.publishCalls)
	}
}

func TestCreateTransaction_AuditReviewMarksPendingReview_WithoutPublish(t *testing.T) {
	repo := &fakeTransactionRepo{}
	stockRepo := &fakeStockRepo{
		available: map[string]int{
			"PARA500": 10,
		},
		deductErrFor: map[string]error{},
	}
	auditRepo := &fakeAuditRepo{
		result: domain.AuditTransactionResult{
			Decision:  domain.AuditDecisionReview,
			RiskScore: 0.60,
			Reason:    "requires pharmacist review",
			Provider:  "gemini",
			Model:     "gemini-2.5-flash",
		},
	}
	publisher := &fakePublisher{}

	uc := usecase.NewTransactionUsecase(repo, stockRepo, auditRepo, publisher)

	result, err := uc.CreateTransaction(context.Background(), domain.CreateTransactionRequest{
		IdempotencyKey: "idem-004",
		Items: []domain.TransactionItemInput{
			{MedicineID: "PARA500", Qty: 2},
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Transaction.Status != domain.TransactionStatusPendingReview {
		t.Fatalf("expected PENDING_REVIEW, got %s", result.Transaction.Status)
	}
	if publisher.publishCalls != 0 {
		t.Fatalf("expected no publish for pending review, got %d", publisher.publishCalls)
	}
}

func TestCreateTransaction_HighRiskReviewMarksFlagged_WithoutPublish(t *testing.T) {
	repo := &fakeTransactionRepo{}
	stockRepo := &fakeStockRepo{
		available: map[string]int{
			"OBATKERAS-X": 5,
		},
		deductErrFor: map[string]error{},
	}
	auditRepo := &fakeAuditRepo{
		result: domain.AuditTransactionResult{
			Decision:  domain.AuditDecisionReview,
			RiskScore: 0.91,
			Reason:    "high-risk medicine detected",
			Provider:  "mock",
			Model:     "rule-based-v1",
		},
	}
	publisher := &fakePublisher{}

	uc := usecase.NewTransactionUsecase(repo, stockRepo, auditRepo, publisher)

	result, err := uc.CreateTransaction(context.Background(), domain.CreateTransactionRequest{
		IdempotencyKey: "idem-005",
		Items: []domain.TransactionItemInput{
			{MedicineID: "OBATKERAS-X", Qty: 2},
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Transaction.Status != domain.TransactionStatusFlagged {
		t.Fatalf("expected FLAGGED, got %s", result.Transaction.Status)
	}
	if publisher.publishCalls != 0 {
		t.Fatalf("expected no publish for flagged transaction, got %d", publisher.publishCalls)
	}
}

func TestCreateTransaction_AuditErrorMarksPendingReview_WithoutPublish(t *testing.T) {
	repo := &fakeTransactionRepo{}
	stockRepo := &fakeStockRepo{
		available: map[string]int{
			"PARA500": 10,
		},
		deductErrFor: map[string]error{},
	}
	auditRepo := &fakeAuditRepo{
		err: &domain.UpstreamError{
			Service: "ai-auditor",
			Reason:  "connection refused",
		},
	}
	publisher := &fakePublisher{}

	uc := usecase.NewTransactionUsecase(repo, stockRepo, auditRepo, publisher)

	result, err := uc.CreateTransaction(context.Background(), domain.CreateTransactionRequest{
		IdempotencyKey: "idem-006",
		Items: []domain.TransactionItemInput{
			{MedicineID: "PARA500", Qty: 2},
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Transaction.Status != domain.TransactionStatusPendingReview {
		t.Fatalf("expected PENDING_REVIEW, got %s", result.Transaction.Status)
	}
	if publisher.publishCalls != 0 {
		t.Fatalf("expected no publish when ai auditor fails, got %d", publisher.publishCalls)
	}
}

func TestCreateTransaction_DeductFailureMarksFailed_WithoutPublish(t *testing.T) {
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
	auditRepo := &fakeAuditRepo{
		result: domain.AuditTransactionResult{
			Decision:  domain.AuditDecisionApproved,
			RiskScore: 0.05,
			Reason:    "safe transaction",
			Provider:  "gemini",
			Model:     "gemini-2.5-flash",
		},
	}
	publisher := &fakePublisher{}

	uc := usecase.NewTransactionUsecase(repo, stockRepo, auditRepo, publisher)

	result, err := uc.CreateTransaction(context.Background(), domain.CreateTransactionRequest{
		IdempotencyKey: "idem-007",
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
	if result.Transaction.Status != domain.TransactionStatusFailed {
		t.Fatalf("expected FAILED, got %s", result.Transaction.Status)
	}
	if publisher.publishCalls != 0 {
		t.Fatalf("expected no publish on failed transaction, got %d", publisher.publishCalls)
	}
}

func TestListTransactions_DefaultPagination(t *testing.T) {
	repo := &fakeTransactionRepo{
		listResult: domain.ListTransactionsResult{
			Items: []domain.Transaction{
				{
					ID:     "TXN-20260401093000-AAAA1111",
					Status: domain.TransactionStatusApproved,
					Items: []domain.TransactionItem{
						{MedicineID: "PARA500", Qty: 2},
					},
					Audit: &domain.TransactionAudit{
						Decision:  domain.AuditDecisionApproved,
						RiskScore: 0.05,
						Reason:    "safe transaction",
						Provider:  "gemini",
						Model:     "gemini-2.5-flash",
						AuditedAt: time.Date(2026, 4, 1, 9, 30, 5, 0, time.UTC),
					},
					CreatedAt: time.Date(2026, 4, 1, 9, 30, 0, 0, time.UTC),
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
	auditRepo := &fakeAuditRepo{}
	publisher := &fakePublisher{}

	uc := usecase.NewTransactionUsecase(repo, stockRepo, auditRepo, publisher)

	result, err := uc.ListTransactions(context.Background(), domain.ListTransactionsRequest{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected total 1, got %d", result.Total)
	}
	if result.Items[0].Audit == nil {
		t.Fatal("expected list result to include audit metadata")
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
	auditRepo := &fakeAuditRepo{}
	publisher := &fakePublisher{}

	uc := usecase.NewTransactionUsecase(repo, stockRepo, auditRepo, publisher)

	_, err := uc.ListTransactions(context.Background(), domain.ListTransactionsRequest{
		Limit:  5,
		Offset: 10,
		Status: "pending_review",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.capturedList.Status != "PENDING_REVIEW" {
		t.Fatalf("expected normalized status PENDING_REVIEW, got %q", repo.capturedList.Status)
	}
}

func TestListTransactions_InvalidLimit(t *testing.T) {
	repo := &fakeTransactionRepo{}
	stockRepo := &fakeStockRepo{
		available:    map[string]int{},
		deductErrFor: map[string]error{},
	}
	auditRepo := &fakeAuditRepo{}
	publisher := &fakePublisher{}

	uc := usecase.NewTransactionUsecase(repo, stockRepo, auditRepo, publisher)

	_, err := uc.ListTransactions(context.Background(), domain.ListTransactionsRequest{
		Limit: -1,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	ve, ok := domain.IsValidation(err)
	if !ok || ve.Field != "limit" {
		t.Fatalf("expected ValidationError on limit, got %v", err)
	}
}

func TestListTransactions_InvalidOffset(t *testing.T) {
	repo := &fakeTransactionRepo{}
	stockRepo := &fakeStockRepo{
		available:    map[string]int{},
		deductErrFor: map[string]error{},
	}
	auditRepo := &fakeAuditRepo{}
	publisher := &fakePublisher{}

	uc := usecase.NewTransactionUsecase(repo, stockRepo, auditRepo, publisher)

	_, err := uc.ListTransactions(context.Background(), domain.ListTransactionsRequest{
		Offset: -1,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	ve, ok := domain.IsValidation(err)
	if !ok || ve.Field != "offset" {
		t.Fatalf("expected ValidationError on offset, got %v", err)
	}
}