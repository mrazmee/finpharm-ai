package usecase

import (
	"context"
	"strings"

	"finpharm-ai/services/inventory/internal/domain"
)

type StockUsecase struct {
	repo domain.StockRepository
}

func NewStockUsecase(repo domain.StockRepository) *StockUsecase {
	return &StockUsecase{repo: repo}
}

func (u *StockUsecase) CheckStock(ctx context.Context, req domain.StockCheckRequest) (domain.StockCheckResult, error) {
	if strings.TrimSpace(req.MedicineID) == "" {
		return domain.StockCheckResult{}, &domain.ValidationError{
			Field:  "medicine_id",
			Reason: "is required",
		}
	}
	if req.Qty <= 0 {
		return domain.StockCheckResult{}, &domain.ValidationError{
			Field:  "qty",
			Reason: "must be > 0",
		}
	}

	available, err := u.repo.GetAvailableQty(ctx, req.MedicineID)
	if err != nil {
		return domain.StockCheckResult{}, err
	}

	return domain.StockCheckResult{
		MedicineID:   req.MedicineID,
		RequestedQty: req.Qty,
		AvailableQty: available,
		IsAvailable:  available >= req.Qty,
	}, nil
}

func (u *StockUsecase) DeductStock(ctx context.Context, req domain.StockDeductRequest) (domain.StockDeductResult, error) {
	if strings.TrimSpace(req.MedicineID) == "" {
		return domain.StockDeductResult{}, &domain.ValidationError{
			Field:  "medicine_id",
			Reason: "is required",
		}
	}
	if req.Qty <= 0 {
		return domain.StockDeductResult{}, &domain.ValidationError{
			Field:  "qty",
			Reason: "must be > 0",
		}
	}

	remaining, err := u.repo.DeductStock(ctx, req.MedicineID, req.Qty)
	if err != nil {
		return domain.StockDeductResult{}, err
	}

	return domain.StockDeductResult{
		MedicineID:   req.MedicineID,
		DeductedQty:  req.Qty,
		RemainingQty: remaining,
	}, nil
}