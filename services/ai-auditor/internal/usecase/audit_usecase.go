package usecase

import (
	"context"
	"fmt"
	"strings"

	"finpharm-ai/services/ai-auditor/internal/domain"
)

type AuditUsecase struct{}

func NewAuditUsecase() *AuditUsecase {
	return &AuditUsecase{}
}

func (u *AuditUsecase) AuditTransaction(ctx context.Context, req domain.AuditTransactionRequest) (domain.AuditTransactionResult, error) {
	_ = ctx

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

	highRisk := false
	highRiskReasons := make([]string, 0)

	for i, item := range req.Items {
		medicineID := strings.TrimSpace(item.MedicineID)
		if medicineID == "" {
			return domain.AuditTransactionResult{}, &domain.ValidationError{
				Field:  fmt.Sprintf("items[%d].medicine_id", i),
				Reason: "is required",
			}
		}
		if item.Qty <= 0 {
			return domain.AuditTransactionResult{}, &domain.ValidationError{
				Field:  fmt.Sprintf("items[%d].qty", i),
				Reason: "must be > 0",
			}
		}

		upperID := strings.ToUpper(medicineID)
		if strings.Contains(upperID, "OBATKERAS") {
			highRisk = true
			highRiskReasons = append(highRiskReasons, fmt.Sprintf("high-risk medicine detected: %s", medicineID))
		}
		if item.Qty >= 20 {
			highRisk = true
			highRiskReasons = append(highRiskReasons, fmt.Sprintf("high quantity detected: %s=%d", medicineID, item.Qty))
		}
	}

	if highRisk {
		return domain.AuditTransactionResult{
			Decision:  domain.AuditDecisionReview,
			RiskScore: 0.91,
			Reason:    strings.Join(highRiskReasons, "; "),
			Provider:  "mock",
			Model:     "rule-based-v1",
		}, nil
	}

	return domain.AuditTransactionResult{
		Decision:  domain.AuditDecisionApproved,
		RiskScore: 0.12,
		Reason:    "mock audit result: no suspicious pattern detected",
		Provider:  "mock",
		Model:     "rule-based-v1",
	}, nil
}