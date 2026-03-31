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
	Decision  AuditDecision `json:"decision"`
	RiskScore float64       `json:"risk_score"`
	Reason    string        `json:"reason"`
	Provider  string        `json:"provider"`
	Model     string        `json:"model"`
}

type AuditProvider interface {
	AuditTransaction(ctx context.Context, req AuditTransactionRequest) (AuditTransactionResult, error)
}

type AuditUsecase interface {
	AuditTransaction(ctx context.Context, req AuditTransactionRequest) (AuditTransactionResult, error)
}