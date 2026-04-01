package domain

import (
	"context"
	"time"
)

type TransactionStatus string

const (
	TransactionStatusPending       TransactionStatus = "PENDING"
	TransactionStatusPendingReview TransactionStatus = "PENDING_REVIEW"
	TransactionStatusApproved      TransactionStatus = "APPROVED"
	TransactionStatusFlagged       TransactionStatus = "FLAGGED"
	TransactionStatusFailed        TransactionStatus = "FAILED"
)

type TransactionAudit struct {
	Decision  AuditDecision
	RiskScore float64
	Reason    string
	Provider  string
	Model     string
	AuditedAt time.Time
}

type Transaction struct {
	ID             string
	IdempotencyKey string
	Status         TransactionStatus
	Items          []TransactionItem
	Audit          *TransactionAudit
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type TransactionItem struct {
	ID            int64
	TransactionID string
	MedicineID    string
	Qty           int
	CreatedAt     time.Time
}

type TransactionItemInput struct {
	MedicineID string
	Qty        int
}

type CreateTransactionRequest struct {
	IdempotencyKey string
	Items          []TransactionItemInput
}

type CreateTransactionResult struct {
	Transaction Transaction
	IsReplay    bool
}

type ListTransactionsRequest struct {
	Limit  int
	Offset int
	Status string
}

type ListTransactionsResult struct {
	Items  []Transaction
	Limit  int
	Offset int
	Total  int
}

type TransactionRepository interface {
	Create(ctx context.Context, tx Transaction) (CreateTransactionResult, error)
	UpdateStatus(ctx context.Context, transactionID string, status TransactionStatus) error
	UpdateAudit(ctx context.Context, transactionID string, audit TransactionAudit) error
	GetByIdempotencyKey(ctx context.Context, key string) (Transaction, bool, error)
	List(ctx context.Context, req ListTransactionsRequest) (ListTransactionsResult, error)
}

type TransactionUsecase interface {
	CreateTransaction(ctx context.Context, req CreateTransactionRequest) (CreateTransactionResult, error)
	ListTransactions(ctx context.Context, req ListTransactionsRequest) (ListTransactionsResult, error)
}