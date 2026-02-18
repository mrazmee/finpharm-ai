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

func (r *StockMemoryRepo) GetAvailableQty(ctx context.Context, medicineID string) (int, error) {
	_ = ctx

	qty, ok := r.stock[medicineID]
	if !ok {
		return 0, domain.ErrMedicineNotFound
	}
	return qty, nil
}
