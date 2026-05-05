package provider

import (
	"context"

	"finpharm-ai/services/ai-auditor/internal/domain"
)

type SafeFallbackProvider struct {
	failOpen bool
}

func NewSafeFallbackProvider(failOpen bool) *SafeFallbackProvider {
	return &SafeFallbackProvider{failOpen: failOpen}
}

func (p *SafeFallbackProvider) AuditTransaction(ctx context.Context, req domain.AuditTransactionRequest) (domain.AuditTransactionResult, error) {
	_ = ctx
	_ = req

	if p.failOpen {
		return domain.AuditTransactionResult{
			Decision:  domain.AuditDecisionApproved,
			RiskScore: 0.10,
			Reason:    "gemini unavailable; fallback approve mode enabled",
			Provider:  "fallback",
			Model:     "safe-approve-v1",
		}, nil
	}

	return domain.AuditTransactionResult{
		Decision:  domain.AuditDecisionReview,
		RiskScore: 0.55,
		Reason:    "gemini unavailable; fallback review required",
		Provider:  "fallback",
		Model:     "safe-review-v1",
	}, nil
}