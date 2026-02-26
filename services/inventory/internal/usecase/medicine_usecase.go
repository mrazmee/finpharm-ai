package usecase

import (
	"context"

	"finpharm-ai/services/inventory/internal/domain"
)

type MedicineUsecase struct {
	repo domain.MedicineRepository
}

func NewMedicineUsecase(repo domain.MedicineRepository) *MedicineUsecase {
	return &MedicineUsecase{repo: repo}
}

func (u *MedicineUsecase) ListMedicines(ctx context.Context, q domain.ListMedicinesQuery) (domain.ListMedicinesResult, error) {
	// business rules for pagination could be placed here later
	return u.repo.List(ctx, q)
}