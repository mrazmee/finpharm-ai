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

type TransactionRepository interface {
	Create(ctx context.Context, tx Transaction) (Transaction, error)
}

type TransactionUsecase interface {
	CreateTransaction(ctx context.Context, req CreateTransactionRequest) (Transaction, error)
}
