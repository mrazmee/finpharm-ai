package consumer

import "testing"

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