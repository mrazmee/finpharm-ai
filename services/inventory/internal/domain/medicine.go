package domain

import "context"

type Medicine struct {
	ID   string
	Name string
	Type string
}

type ListMedicinesQuery struct {
	Limit  int
	Offset int
}

type ListMedicinesResult struct {
	Items  []Medicine
	Limit  int
	Offset int
	Total  int
}

type MedicineRepository interface {
	List(ctx context.Context, q ListMedicinesQuery) (ListMedicinesResult, error)
	GetByID(ctx context.Context, id string) (Medicine, error)
}

type MedicineUsecase interface {
	ListMedicines(ctx context.Context, q ListMedicinesQuery) (ListMedicinesResult, error)
	GetMedicine(ctx context.Context, id string) (Medicine, error)
}