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

type StockDeductRequest struct {
	MedicineID string
	Qty        int
}

type StockDeductResult struct {
	MedicineID   string
	DeductedQty  int
	RemainingQty int
}

type StockRepository interface {
	GetAvailableQty(ctx context.Context, medicineID string) (int, error)
	DeductStock(ctx context.Context, medicineID string, qty int) (int, error)
}

type StockUsecase interface {
	CheckStock(ctx context.Context, req StockCheckRequest) (StockCheckResult, error)
	DeductStock(ctx context.Context, req StockDeductRequest) (StockDeductResult, error)
}