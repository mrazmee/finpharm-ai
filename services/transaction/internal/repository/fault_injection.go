package repository

import (
	"context"
	"fmt"

	"finpharm-ai/services/transaction/internal/domain"
)

var _ domain.AIAuditorRepository = (*ForcedApprovedAIAuditorRepo)(nil)
var _ domain.StockRepository = (*ForcedDeductFailureStockRepo)(nil)

type ForcedApprovedAIAuditorRepo struct{}

func NewForcedApprovedAIAuditorRepo() *ForcedApprovedAIAuditorRepo {
	return &ForcedApprovedAIAuditorRepo{}
}

func (r *ForcedApprovedAIAuditorRepo) AuditTransaction(ctx context.Context, req domain.AuditTransactionRequest) (domain.AuditTransactionResult, error) {
	return domain.AuditTransactionResult{
		Decision:  domain.AuditDecisionApproved,
		RiskScore: 0.01,
		Reason:    "forced approved for local alert verification",
		Provider:  "local-fixture",
		Model:     "forced-approved",
	}, nil
}

type ForcedDeductFailureStockRepo struct {
	next domain.StockRepository
}

func NewForcedDeductFailureStockRepo(next domain.StockRepository) *ForcedDeductFailureStockRepo {
	return &ForcedDeductFailureStockRepo{next: next}
}

func (r *ForcedDeductFailureStockRepo) GetAvailableQty(ctx context.Context, medicineID string, requestedQty int) (int, error) {
	if r.next == nil {
		return 0, fmt.Errorf("forced deduct failure repo requires next stock repository")
	}
	return r.next.GetAvailableQty(ctx, medicineID, requestedQty)
}

func (r *ForcedDeductFailureStockRepo) DeductStock(ctx context.Context, medicineID string, qty int) error {
	return &domain.UpstreamError{
		Service: "inventory",
		Reason:  "forced deduct failure for local alert verification",
	}
}