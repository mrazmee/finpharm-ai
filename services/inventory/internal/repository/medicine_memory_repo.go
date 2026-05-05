package repository

import (
	"context"

	"finpharm-ai/services/inventory/internal/domain"
)

type MedicineMemoryRepo struct {
	items []domain.Medicine
}

func NewMedicineMemoryRepo() *MedicineMemoryRepo {
	return &MedicineMemoryRepo{
		items: []domain.Medicine{
			{ID: "PARA500", Name: "Paracetamol 500mg", Type: "OTC"},
			{ID: "AMOX500", Name: "Amoxicillin 500mg", Type: "ANTIBIOTIC"},
			{ID: "OBATKERAS-X", Name: "Obat Keras X", Type: "CONTROLLED"},
		},
	}
}

func (r *MedicineMemoryRepo) List(ctx context.Context, q domain.ListMedicinesQuery) (domain.ListMedicinesResult, error) {
	_ = ctx

	total := len(r.items)

	limit := q.Limit
	offset := q.Offset
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	if offset > total {
		offset = total
	}

	end := offset + limit
	if end > total {
		end = total
	}

	items := make([]domain.Medicine, 0, end-offset)
	for i := offset; i < end; i++ {
		items = append(items, r.items[i])
	}

	return domain.ListMedicinesResult{
		Items:  items,
		Limit:  limit,
		Offset: offset,
		Total:  total,
	}, nil
}

func (r *MedicineMemoryRepo) GetByID(ctx context.Context, id string) (domain.Medicine, error) {
	_ = ctx

	for _, m := range r.items {
		if m.ID == id {
			return m, nil
		}
	}
	return domain.Medicine{}, &domain.NotFoundError{
		Resource: "medicine",
		Key:      id,
	}
}
