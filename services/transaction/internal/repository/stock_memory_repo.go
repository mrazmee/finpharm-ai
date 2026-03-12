package repository

import (
	"context"

	"finpharm-ai/services/transaction/internal/domain"
)

type StockMemoryRepo struct {
	stock map[string]int
}

func NewStockMemoryRepo() *StockMemoryRepo {
	return &StockMemoryRepo{
		stock: map[string]int{
			"AMOX500":     120,
			"PARA500":     80,
			"OBATKERAS-X": 5,
		},
	}
}

func (r *StockMemoryRepo) GetAvailableQty(ctx context.Context, medicineID string, requestedQty int) (int, error) {
	_ = ctx
	_ = requestedQty

	qty, ok := r.stock[medicineID]
	if !ok {
		return 0, &domain.NotFoundError{
			Resource: "medicine",
			Key:      medicineID,
		}
	}
	return qty, nil
}