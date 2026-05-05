package domain

import "fmt"

type ValidationError struct {
	Field  string
	Reason string
}

func (e *ValidationError) Error() string {
	if e == nil {
		return "validation error"
	}
	if e.Field == "" {
		return "validation error"
	}
	if e.Reason == "" {
		return fmt.Sprintf("validation error on %s", e.Field)
	}
	return fmt.Sprintf("validation error on %s: %s", e.Field, e.Reason)
}

func IsValidation(err error) (*ValidationError, bool) {
	ve, ok := err.(*ValidationError)
	return ve, ok
}