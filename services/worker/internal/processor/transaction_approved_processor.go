package processor

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"finpharm-ai/services/worker/internal/domain"
)

type DummyOutput struct {
	NotificationMessage string
	ReportLine          string
}

type TransactionApprovedProcessor struct{}

func NewTransactionApprovedProcessor() *TransactionApprovedProcessor {
	return &TransactionApprovedProcessor{}
}

func BuildDummyOutput(event domain.TransactionApprovedEvent) DummyOutput {
	itemParts := make([]string, 0, len(event.Items))
	for _, item := range event.Items {
		itemParts = append(itemParts, fmt.Sprintf("%s x%d", item.MedicineID, item.Qty))
	}

	notification := fmt.Sprintf(
		"[DUMMY NOTIFICATION] approved transaction %s processed with items: %s",
		event.TransactionID,
		strings.Join(itemParts, ", "),
	)

	report := fmt.Sprintf(
		"[DUMMY REPORT] transaction=%s status=%s event=%s published_at=%s",
		event.TransactionID,
		event.Status,
		event.EventName,
		event.PublishedAt.UTC().Format("2006-01-02T15:04:05Z"),
	)

	return DummyOutput{
		NotificationMessage: notification,
		ReportLine:          report,
	}
}

func (p *TransactionApprovedProcessor) HandleTransactionApproved(ctx context.Context, event domain.TransactionApprovedEvent) error {
	output := BuildDummyOutput(event)

	auditProvider := ""
	auditModel := ""
	if event.Audit != nil {
		auditProvider = event.Audit.Provider
		auditModel = event.Audit.Model
	}

	slog.Info("worker_notification_sent",
		"transaction_id", event.TransactionID,
		"message", output.NotificationMessage,
		"audit_provider", auditProvider,
		"audit_model", auditModel,
	)

	slog.Info("worker_report_generated",
		"transaction_id", event.TransactionID,
		"report", output.ReportLine,
	)

	return nil
}