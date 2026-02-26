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

type UpstreamError struct {
	Service string
	Reason  string
}

func (e UpstreamError) Error() string {
	if e.Service == "" {
		return "upstream error"
	}
	if e.Reason == "" {
		return fmt.Sprintf("upstream error: %s", e.Service)
	}
	return fmt.Sprintf("upstream error: %s (%s)", e.Service, e.Reason)
}

func IsValidation(err error) (*ValidationError, bool) {
	ve, ok := err.(*ValidationError)
	return ve, ok
}

func IsNotFound(err error) (*NotFoundError, bool) {
	nf, ok := err.(*NotFoundError)
	return nf, ok
}

func IsUpstream(err error) (*UpstreamError, bool) {
	ue, ok := err.(*UpstreamError)
	return ue, ok
}