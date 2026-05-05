package domain

import "fmt"

type ValidationError struct {
	Field  string
	Reason string
}

func (e ValidationError) Error() string {
	if e.Field == "" {
		return "validation error"
	}
	if e.Reason == "" {
		return fmt.Sprintf("validation error on %s", e.Field)
	}
	return fmt.Sprintf("validation error on %s: %s", e.Field, e.Reason)
}

type NotFoundError struct {
	Resource string
	Key      string
}

func (e NotFoundError) Error() string {
	if e.Resource == "" {
		return "not found"
	}
	if e.Key == "" {
		return fmt.Sprintf("%s not found", e.Resource)
	}
	return fmt.Sprintf("%s not found: %s", e.Resource, e.Key)
}

type InsufficientStockError struct {
	MedicineID   string
	RequestedQty int
	AvailableQty int
}

func (e InsufficientStockError) Error() string {
	return fmt.Sprintf(
		"insufficient stock for %s: requested=%d available=%d",
		e.MedicineID,
		e.RequestedQty,
		e.AvailableQty,
	)
}

func IsValidation(err error) (*ValidationError, bool) {
	ve, ok := err.(*ValidationError)
	return ve, ok
}

func IsNotFound(err error) (*NotFoundError, bool) {
	nf, ok := err.(*NotFoundError)
	return nf, ok
}

func IsInsufficientStock(err error) (*InsufficientStockError, bool) {
	ie, ok := err.(*InsufficientStockError)
	return ie, ok
}