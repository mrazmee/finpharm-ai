package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"finpharm-ai/services/transaction/internal/config"
	"finpharm-ai/services/transaction/internal/domain"
	"finpharm-ai/services/transaction/internal/httpapi"
	"finpharm-ai/services/transaction/internal/httpapi/handler"

	"github.com/gin-gonic/gin"
)

type fakeTransactionUsecase struct {
	capturedCreate domain.CreateTransactionRequest
	createResult   domain.Transaction
	createErr      error

	capturedList domain.ListTransactionsRequest
	listResult   domain.ListTransactionsResult
	listErr      error
}

func (f *fakeTransactionUsecase) CreateTransaction(ctx context.Context, req domain.CreateTransactionRequest) (domain.Transaction, error) {
	f.capturedCreate = req
	if f.createErr != nil {
		return domain.Transaction{}, f.createErr
	}
	return f.createResult, nil
}

func (f *fakeTransactionUsecase) ListTransactions(ctx context.Context, req domain.ListTransactionsRequest) (domain.ListTransactionsResult, error) {
	f.capturedList = req
	if f.listErr != nil {
		return domain.ListTransactionsResult{}, f.listErr
	}
	return f.listResult, nil
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
		t.Fatalf("failed to decode error response: %v, body=%s", err, body)
	}
	return resp
}

func TestCreateTransaction_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uc := &fakeTransactionUsecase{
		createResult: domain.Transaction{
			ID:     "TXN-20260312120000-AB12CD34",
			Status: domain.TransactionStatusPending,
			Items: []domain.TransactionItem{
				{MedicineID: "PARA500", Qty: 10},
				{MedicineID: "AMOX500", Qty: 2},
			},
			CreatedAt: time.Date(2026, 3, 12, 12, 0, 0, 0, time.UTC),
		},
	}

	txHandler := handler.NewTransactionHandler(uc)
	r := httpapi.NewRouter(config.Config{AppEnv: "local"}, nil, txHandler)

	body := `{"items":[{"medicine_id":"PARA500","qty":10},{"medicine_id":"AMOX500","qty":2}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/transactions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "tx-create-123")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", w.Code, w.Body.String())
	}
	if len(uc.capturedCreate.Items) != 2 {
		t.Fatalf("expected 2 items captured, got %d", len(uc.capturedCreate.Items))
	}
	if !strings.Contains(w.Body.String(), `"id":"TXN-20260312120000-AB12CD34"`) {
		t.Fatalf("expected response contains transaction id, body=%s", w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"request_id":"tx-create-123"`) {
		t.Fatalf("expected response contains request_id, body=%s", w.Body.String())
	}
}

func TestCreateTransaction_InsufficientStock(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uc := &fakeTransactionUsecase{
		createErr: &domain.InsufficientStockError{
			MedicineID:   "PARA500",
			RequestedQty: 10,
			AvailableQty: 3,
		},
	}

	txHandler := handler.NewTransactionHandler(uc)
	r := httpapi.NewRouter(config.Config{AppEnv: "local"}, nil, txHandler)

	body := `{"items":[{"medicine_id":"PARA500","qty":10}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/transactions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"code":"INSUFFICIENT_STOCK"`) {
		t.Fatalf("expected insufficient stock error code, body=%s", w.Body.String())
	}
}

func TestListTransactions_DefaultPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uc := &fakeTransactionUsecase{
		listResult: domain.ListTransactionsResult{
			Items: []domain.Transaction{
				{
					ID:     "TXN-20260313093000-AAAA1111",
					Status: domain.TransactionStatusPending,
					Items: []domain.TransactionItem{
						{MedicineID: "PARA500", Qty: 2},
					},
					CreatedAt: time.Date(2026, 3, 13, 9, 30, 0, 0, time.UTC),
				},
			},
			Limit:  10,
			Offset: 0,
			Total:  1,
		},
	}

	txHandler := handler.NewTransactionHandler(uc)
	r := httpapi.NewRouter(config.Config{AppEnv: "local"}, nil, txHandler)

	req := httptest.NewRequest(http.MethodGet, "/v1/transactions", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	if uc.capturedList.Limit != 0 {
		t.Fatalf("expected raw handler request limit 0 when absent, got %d", uc.capturedList.Limit)
	}
	if !strings.Contains(w.Body.String(), `"limit":10`) {
		t.Fatalf("expected response limit 10, body=%s", w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"total":1`) {
		t.Fatalf("expected response total 1, body=%s", w.Body.String())
	}
}

func TestListTransactions_WithQueryParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uc := &fakeTransactionUsecase{
		listResult: domain.ListTransactionsResult{
			Items:  []domain.Transaction{},
			Limit:  5,
			Offset: 10,
			Total:  0,
		},
	}

	txHandler := handler.NewTransactionHandler(uc)
	r := httpapi.NewRouter(config.Config{AppEnv: "local"}, nil, txHandler)

	req := httptest.NewRequest(http.MethodGet, "/v1/transactions?limit=5&offset=10&status=pending", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	if uc.capturedList.Limit != 5 {
		t.Fatalf("expected captured limit 5, got %d", uc.capturedList.Limit)
	}
	if uc.capturedList.Offset != 10 {
		t.Fatalf("expected captured offset 10, got %d", uc.capturedList.Offset)
	}
	if uc.capturedList.Status != "pending" {
		t.Fatalf("expected raw captured status pending, got %q", uc.capturedList.Status)
	}
}

func TestListTransactions_InvalidLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uc := &fakeTransactionUsecase{}
	txHandler := handler.NewTransactionHandler(uc)
	r := httpapi.NewRouter(config.Config{AppEnv: "local"}, nil, txHandler)

	req := httptest.NewRequest(http.MethodGet, "/v1/transactions?limit=abc", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", w.Code, w.Body.String())
	}

	resp := decodeErrorEnvelope(t, w.Body.String())
	if resp.Error.Code != "VALIDATION_ERROR" {
		t.Fatalf("expected validation error code, got %q body=%s", resp.Error.Code, w.Body.String())
	}
	if resp.Error.Details.Field != "limit" {
		t.Fatalf("expected field limit, got %q body=%s", resp.Error.Details.Field, w.Body.String())
	}
	if resp.Error.Details.Reason != "must be an integer" {
		t.Fatalf("expected reason must be an integer, got %q body=%s", resp.Error.Details.Reason, w.Body.String())
	}
}

func TestListTransactions_ZeroLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uc := &fakeTransactionUsecase{}
	txHandler := handler.NewTransactionHandler(uc)
	r := httpapi.NewRouter(config.Config{AppEnv: "local"}, nil, txHandler)

	req := httptest.NewRequest(http.MethodGet, "/v1/transactions?limit=0", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", w.Code, w.Body.String())
	}

	resp := decodeErrorEnvelope(t, w.Body.String())
	if resp.Error.Details.Field != "limit" {
		t.Fatalf("expected validation field limit, got %q body=%s", resp.Error.Details.Field, w.Body.String())
	}
	if resp.Error.Details.Reason != "must be > 0" {
		t.Fatalf("expected reason must be > 0, got %q body=%s", resp.Error.Details.Reason, w.Body.String())
	}
}

func TestListTransactions_LimitTooLarge(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uc := &fakeTransactionUsecase{}
	txHandler := handler.NewTransactionHandler(uc)
	r := httpapi.NewRouter(config.Config{AppEnv: "local"}, nil, txHandler)

	req := httptest.NewRequest(http.MethodGet, "/v1/transactions?limit=101", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", w.Code, w.Body.String())
	}

	resp := decodeErrorEnvelope(t, w.Body.String())
	if resp.Error.Details.Field != "limit" {
		t.Fatalf("expected validation field limit, got %q body=%s", resp.Error.Details.Field, w.Body.String())
	}
	if resp.Error.Details.Reason != "must be <= 100" {
		t.Fatalf("expected reason must be <= 100, got %q body=%s", resp.Error.Details.Reason, w.Body.String())
	}
}