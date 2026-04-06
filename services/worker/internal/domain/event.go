package domain

import "time"

type TransactionApprovedEvent struct {
	EventName      string                         `json:"event_name"`
	TransactionID  string                         `json:"transaction_id"`
	IdempotencyKey string                         `json:"idempotency_key"`
	Status         string                         `json:"status"`
	Items          []TransactionApprovedEventItem `json:"items"`
	Audit          *TransactionApprovedEventAudit `json:"audit,omitempty"`
	CreatedAt      time.Time                      `json:"created_at"`
	PublishedAt    time.Time                      `json:"published_at"`
}

type TransactionApprovedEventItem struct {
	MedicineID string `json:"medicine_id"`
	Qty        int    `json:"qty"`
}

type TransactionApprovedEventAudit struct {
	Decision  string    `json:"decision"`
	RiskScore float64   `json:"risk_score"`
	Reason    string    `json:"reason"`
	Provider  string    `json:"provider"`
	Model     string    `json:"model"`
	AuditedAt time.Time `json:"audited_at"`
}