package processor_test

import (
	"strings"
	"testing"
	"time"

	"finpharm-ai/services/worker/internal/domain"
	"finpharm-ai/services/worker/internal/processor"
)

func TestBuildDummyOutput(t *testing.T) {
	event := domain.TransactionApprovedEvent{
		EventName:      "transaction.approved",
		TransactionID:  "TXN-20260406103000-AAAA1111",
		IdempotencyKey: "idem-801",
		Status:         "APPROVED",
		Items: []domain.TransactionApprovedEventItem{
			{MedicineID: "PARA500", Qty: 1},
			{MedicineID: "AMOX500", Qty: 2},
		},
		PublishedAt: time.Date(2026, 4, 6, 10, 30, 10, 0, time.UTC),
	}

	output := processor.BuildDummyOutput(event)

	if !strings.Contains(output.NotificationMessage, "TXN-20260406103000-AAAA1111") {
		t.Fatalf("expected transaction id in notification message, got %q", output.NotificationMessage)
	}
	if !strings.Contains(output.NotificationMessage, "PARA500 x1") {
		t.Fatalf("expected PARA500 item in notification message, got %q", output.NotificationMessage)
	}
	if !strings.Contains(output.NotificationMessage, "AMOX500 x2") {
		t.Fatalf("expected AMOX500 item in notification message, got %q", output.NotificationMessage)
	}
	if !strings.Contains(output.ReportLine, "transaction.approved") {
		t.Fatalf("expected event name in report line, got %q", output.ReportLine)
	}
	if !strings.Contains(output.ReportLine, "APPROVED") {
		t.Fatalf("expected status in report line, got %q", output.ReportLine)
	}
}