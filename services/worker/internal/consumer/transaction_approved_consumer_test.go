package consumer

import (
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func TestDecodeTransactionApprovedEvent_OK(t *testing.T) {
	body := []byte(`{
		"event_name":"transaction.approved",
		"transaction_id":"TXN-20260406110000-AAAA1111",
		"idempotency_key":"idem-901",
		"status":"APPROVED",
		"items":[{"medicine_id":"PARA500","qty":1}]
	}`)

	event, err := decodeTransactionApprovedEvent(body)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if event.EventName != "transaction.approved" {
		t.Fatalf("expected event_name transaction.approved, got %q", event.EventName)
	}
	if event.TransactionID != "TXN-20260406110000-AAAA1111" {
		t.Fatalf("expected transaction id, got %q", event.TransactionID)
	}
}

func TestDecodeTransactionApprovedEvent_InvalidJSON(t *testing.T) {
	body := []byte(`{"event_name":`)

	_, err := decodeTransactionApprovedEvent(body)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDecodeTransactionApprovedEvent_MissingTransactionID(t *testing.T) {
	body := []byte(`{
		"event_name":"transaction.approved",
		"transaction_id":""
	}`)

	_, err := decodeTransactionApprovedEvent(body)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCurrentRetryCount_DefaultZero(t *testing.T) {
	got := currentRetryCount(nil)
	if got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestCurrentRetryCount_FromHeaders(t *testing.T) {
	got := currentRetryCount(amqp.Table{
		"x-retry-count": int32(2),
	})
	if got != 2 {
		t.Fatalf("expected 2, got %d", got)
	}
}

func TestWithRetryHeader(t *testing.T) {
	headers := withRetryHeader(amqp.Table{
		"foo": "bar",
	}, 3)

	if headers["foo"] != "bar" {
		t.Fatalf("expected existing header preserved, got %v", headers["foo"])
	}
	if headers["x-retry-count"] != 3 {
		t.Fatalf("expected retry header 3, got %v", headers["x-retry-count"])
	}
}