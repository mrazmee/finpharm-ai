package usecase

import (
	"context"
	"strings"

	"finpharm-ai/services/inventory/internal/domain"
)

type MedicineUsecase struct {
	repo domain.MedicineRepository
}

func NewMedicineUsecase(repo domain.MedicineRepository) *MedicineUsecase {
	return &MedicineUsecase{repo: repo}
}

func (u *MedicineUsecase) ListMedicines(ctx context.Context, q domain.ListMedicinesQuery) (domain.ListMedicinesResult, error) {
	return u.repo.List(ctx, q)
}

func (u *MedicineUsecase) GetMedicine(ctx context.Context, id string) (domain.Medicine, error) {
	if strings.TrimSpace(id) == "" {
		return domain.Medicine{}, &domain.ValidationError{
			Field:  "id",
			Reason: "is required",
		}
	}
	return u.repo.GetByID(ctx, id)
}