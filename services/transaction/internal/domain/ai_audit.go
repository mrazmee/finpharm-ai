package domain

import "context"

type AuditDecision string

const (
	AuditDecisionApproved AuditDecision = "APPROVED"
	AuditDecisionReview   AuditDecision = "REVIEW"
)

type AuditTransactionItem struct {
	MedicineID string
	Qty        int
}

type AuditTransactionRequest struct {
	TransactionID string
	Items         []AuditTransactionItem
}

type AuditTransactionResult struct {
	Decision  AuditDecision
	RiskScore float64
	Reason    string
	Provider  string
	Model     string
}

type AIAuditorRepository interface {
	AuditTransaction(ctx context.Context, req AuditTransactionRequest) (AuditTransactionResult, error)
}