package repository

import (
	"context"

	"finpharm-ai/services/inventory/internal/domain"
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
		return 0, &domain.NotFoundError{Resource: "medicine", Key: medicineID}
	}
	return qty, nil
}

func (r *StockMemoryRepo) DeductStock(ctx context.Context, medicineID string, qty int) (int, error) {
	_ = ctx

	available, ok := r.stock[medicineID]
	if !ok {
		return 0, &domain.NotFoundError{Resource: "medicine", Key: medicineID}
	}
	if available < qty {
		return 0, &domain.InsufficientStockError{
			MedicineID:   medicineID,
			RequestedQty: qty,
			AvailableQty: available,
		}
	}

	remaining := available - qty
	r.stock[medicineID] = remaining
	return remaining, nil
}