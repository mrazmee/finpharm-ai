package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"finpharm-ai/services/ai-auditor/internal/domain"
	"finpharm-ai/services/ai-auditor/internal/httpapi/handler"
	"finpharm-ai/services/ai-auditor/internal/httpapi/middleware"

	"github.com/gin-gonic/gin"
)

type fakeAuditUsecase struct {
	result domain.AuditTransactionResult
	err    error
}

func (f *fakeAuditUsecase) AuditTransaction(ctx context.Context, req domain.AuditTransactionRequest) (domain.AuditTransactionResult, error) {
	if f.err != nil {
		return domain.AuditTransactionResult{}, f.err
	}
	return f.result, nil
}

type errorEnvelope struct {
	Error struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		RequestID string `json:"request_id"`
		Details   struct {
			Field  string `json:"field"`
			Reason string `json:"reason"`
		} `json:"details"`
	} `json:"error"`
}

func decodeErrorEnvelope(t *testing.T, body string) errorEnvelope {
	t.Helper()

	var resp errorEnvelope
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("failed to decode error response: %v body=%s", err, body)
	}
	return resp
}

func TestAuditTransaction_Approved(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uc := &fakeAuditUsecase{
		result: domain.AuditTransactionResult{
			Decision:  domain.AuditDecisionApproved,
			RiskScore: 0.12,
			Reason:    "mock audit result: no suspicious pattern detected",
			Provider:  "mock",
			Model:     "rule-based-v1",
		},
	}

	h := handler.NewAuditHandler(uc)

	r := gin.New()
	r.Use(middleware.RequestID(), middleware.RequestLogger(), gin.Recovery())
	r.POST("/v1/audit/transaction", h.AuditTransaction)

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/audit/transaction",
		strings.NewReader(`{"transaction_id":"TXN-1","items":[{"medicine_id":"PARA500","qty":2}]}`),
	)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"decision":"APPROVED"`) {
		t.Fatalf("expected decision APPROVED, body=%s", w.Body.String())
	}
}

func TestAuditTransaction_ValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uc := &fakeAuditUsecase{
		err: &domain.ValidationError{
			Field:  "transaction_id",
			Reason: "is required",
		},
	}

	h := handler.NewAuditHandler(uc)

	r := gin.New()
	r.Use(middleware.RequestID(), middleware.RequestLogger(), gin.Recovery())
	r.POST("/v1/audit/transaction", h.AuditTransaction)

	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/audit/transaction",
		strings.NewReader(`{"transaction_id":"","items":[{"medicine_id":"PARA500","qty":2}]}`),
	)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", w.Code, w.Body.String())
	}

	resp := decodeErrorEnvelope(t, w.Body.String())
	if resp.Error.Code != "VALIDATION_ERROR" {
		t.Fatalf("expected VALIDATION_ERROR, got %q body=%s", resp.Error.Code, w.Body.String())
	}
	if resp.Error.Details.Field != "transaction_id" {
		t.Fatalf("expected field transaction_id, got %q body=%s", resp.Error.Details.Field, w.Body.String())
	}
}