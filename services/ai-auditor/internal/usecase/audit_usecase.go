package usecase

import (
	"context"
	"log/slog"
	"strings"

	"finpharm-ai/services/ai-auditor/internal/domain"
	"finpharm-ai/services/ai-auditor/internal/observability"
)

type AuditUsecase struct {
	primary  domain.AuditProvider
	fallback domain.AuditProvider
}

func NewAuditUsecase(primary, fallback domain.AuditProvider) *AuditUsecase {
	return &AuditUsecase{
		primary:  primary,
		fallback: fallback,
	}
}

func (u *AuditUsecase) AuditTransaction(ctx context.Context, req domain.AuditTransactionRequest) (domain.AuditTransactionResult, error) {
	if strings.TrimSpace(req.TransactionID) == "" {
		return domain.AuditTransactionResult{}, &domain.ValidationError{
			Field:  "transaction_id",
			Reason: "is required",
		}
	}
	if len(req.Items) == 0 {
		return domain.AuditTransactionResult{}, &domain.ValidationError{
			Field:  "items",
			Reason: "must contain at least 1 item",
		}
	}

	for i, item := range req.Items {
		if strings.TrimSpace(item.MedicineID) == "" {
			return domain.AuditTransactionResult{}, &domain.ValidationError{
				Field:  "items[" + itoa(i) + "].medicine_id",
				Reason: "is required",
			}
		}
		if item.Qty <= 0 {
			return domain.AuditTransactionResult{}, &domain.ValidationError{
				Field:  "items[" + itoa(i) + "].qty",
				Reason: "must be > 0",
			}
		}
	}

	if u.primary == nil {
		observability.IncFallback("primary_missing")

		result, err := u.fallback.AuditTransaction(ctx, req)
		if err == nil {
			observability.ObserveAuditDecision(string(result.Decision), result.Provider, result.Model)
		}
		return result, err
	}

	result, err := u.primary.AuditTransaction(ctx, req)
	if err == nil {
		observability.ObserveAuditDecision(string(result.Decision), result.Provider, result.Model)
		return result, nil
	}

	slog.Warn("audit_provider_fallback",
		"transaction_id", req.TransactionID,
		"reason", err.Error(),
	)

	observability.IncFallback("primary_failed")

	fallbackResult, fallbackErr := u.fallback.AuditTransaction(ctx, req)
	if fallbackErr == nil {
		observability.ObserveAuditDecision(string(fallbackResult.Decision), fallbackResult.Provider, fallbackResult.Model)
	}
	return fallbackResult, fallbackErr
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	sign := ""
	if i < 0 {
		sign = "-"
		i = -i
	}
	var digits [20]byte
	n := len(digits)
	for i > 0 {
		n--
		digits[n] = byte('0' + (i % 10))
		i /= 10
	}
	return sign + string(digits[n:])
}