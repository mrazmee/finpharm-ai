package usecase

import (
	"context"
	"errors"

	"finpharm-ai/services/transaction/internal/domain"
	"finpharm-ai/services/transaction/internal/repository"
)

type StockUsecase struct {
	repo domain.StockRepository
}

func NewStockUsecase(repo domain.StockRepository) *StockUsecase {
	return &StockUsecase{repo: repo}
}

func (u *StockUsecase) CheckStock(ctx context.Context, req domain.StockCheckRequest) (domain.StockCheckResult, error) {
	if req.MedicineID == "" {
		return domain.StockCheckResult{}, errors.New("medicine_id is required")
	}
	if req.Qty <= 0 {
		return domain.StockCheckResult{}, errors.New("qty must be > 0")
	}

	available, err := u.repo.GetAvailableQty(ctx, req.MedicineID)
	if err != nil {
		if errors.Is(err, repository.ErrMedicineNotFound) {
			// medicine not found => available 0, not available
			return domain.StockCheckResult{
				MedicineID:   req.MedicineID,
				RequestedQty: req.Qty,
				AvailableQty: 0,
				IsAvailable:  false,
			}, nil
		}
		return domain.StockCheckResult{}, err
	}

	return domain.StockCheckResult{
		MedicineID:   req.MedicineID,
		RequestedQty: req.Qty,
		AvailableQty: available,
		IsAvailable:  available >= req.Qty,
	}, nil
}
