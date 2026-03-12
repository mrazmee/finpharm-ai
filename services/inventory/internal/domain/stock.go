package domain

import "context"

type StockCheckRequest struct {
	MedicineID string
	Qty        int
}

type StockCheckResult struct {
	MedicineID   string
	RequestedQty int
	AvailableQty int
	IsAvailable  bool
}

type StockRepository interface {
	GetAvailableQty(ctx context.Context, medicineID string) (int, error)
}

type StockUsecase interface {
	CheckStock(ctx context.Context, req StockCheckRequest) (StockCheckResult, error)
}
