package domain

import "errors"

var (
	ErrValidation        = errors.New("validation error")
	ErrMedicineNotFound  = errors.New("medicine not found")
)
