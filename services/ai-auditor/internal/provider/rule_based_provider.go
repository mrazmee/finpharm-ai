package provider

import (
	"context"
	"fmt"
	"strings"

	"finpharm-ai/services/ai-auditor/internal/domain"
)

type RuleBasedProvider struct{}

func NewRuleBasedProvider() *RuleBasedProvider {
	return &RuleBasedProvider{}
}

func (p *RuleBasedProvider) AuditTransaction(ctx context.Context, req domain.AuditTransactionRequest) (domain.AuditTransactionResult, error) {
	_ = ctx

	highRisk := false
	reasons := make([]string, 0)

	for _, item := range req.Items {
		upperID := strings.ToUpper(strings.TrimSpace(item.MedicineID))
		if strings.Contains(upperID, "OBATKERAS") {
			highRisk = true
			reasons = append(reasons, fmt.Sprintf("high-risk medicine detected: %s", item.MedicineID))
		}
		if item.Qty >= 20 {
			highRisk = true
			reasons = append(reasons, fmt.Sprintf("high quantity detected: %s=%d", item.MedicineID, item.Qty))
		}
	}

	if highRisk {
		return domain.AuditTransactionResult{
			Decision:  domain.AuditDecisionReview,
			RiskScore: 0.91,
			Reason:    strings.Join(reasons, "; "),
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