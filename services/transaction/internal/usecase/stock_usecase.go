package usecase

import (
	"context"
	"strings"

	"finpharm-ai/services/transaction/internal/domain"
)

type StockUsecase struct {
	repo domain.StockRepository
}

func NewStockUsecase(repo domain.StockRepository) *StockUsecase {
	return &StockUsecase{repo: repo}
}

func (u *StockUsecase) CheckStock(ctx context.Context, req domain.StockCheckRequest) (domain.StockCheckResult, error) {
	if strings.TrimSpace(req.MedicineID) == "" {
		return domain.StockCheckResult{}, domain.ErrValidation
	}
	if req.Qty <= 0 {
		return domain.StockCheckResult{}, domain.ErrValidation
	}

	available, err := u.repo.GetAvailableQty(ctx, req.MedicineID)
	if err != nil {
		return domain.StockCheckResult{}, err // could be ErrMedicineNotFound or internal error
	}

	return domain.StockCheckResult{
		MedicineID:   req.MedicineID,
		RequestedQty: req.Qty,
		AvailableQty: available,
		IsAvailable:  available >= req.Qty,
	}, nil
}
