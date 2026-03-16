package domain

import (
	"context"
	"time"
)

type TransactionStatus string

const (
	TransactionStatusPending TransactionStatus = "PENDING"
)

type Transaction struct {
	ID        string
	Status    TransactionStatus
	Items     []TransactionItem
	CreatedAt time.Time
	UpdatedAt time.Time
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
	Items []TransactionItemInput
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
	Create(ctx context.Context, tx Transaction) (Transaction, error)
	List(ctx context.Context, req ListTransactionsRequest) (ListTransactionsResult, error)
}

type TransactionUsecase interface {
	CreateTransaction(ctx context.Context, req CreateTransactionRequest) (Transaction, error)
	ListTransactions(ctx context.Context, req ListTransactionsRequest) (ListTransactionsResult, error)
}