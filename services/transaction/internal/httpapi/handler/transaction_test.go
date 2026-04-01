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
	createResult   domain.CreateTransactionResult
	createErr      error

	capturedList domain.ListTransactionsRequest
	listResult   domain.ListTransactionsResult
	listErr      error
}

func (f *fakeTransactionUsecase) CreateTransaction(ctx context.Context, req domain.CreateTransactionRequest) (domain.CreateTransactionResult, error) {
	f.capturedCreate = req
	if f.createErr != nil {
		return domain.CreateTransactionResult{}, f.createErr
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
			Field        string `json:"field"`
			Reason       string `json:"reason"`
			MedicineID   string `json:"medicine_id"`
			RequestedQty int    `json:"requested_qty"`
			AvailableQty int    `json:"available_qty"`
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
		createResult: domain.CreateTransactionResult{
			Transaction: domain.Transaction{
				ID:             "TXN-20260331120000-AB12CD34",
				IdempotencyKey: "idem-123",
				Status:         domain.TransactionStatusApproved,
				Items: []domain.TransactionItem{
					{MedicineID: "PARA500", Qty: 10},
					{MedicineID: "AMOX500", Qty: 2},
				},
				CreatedAt: time.Date(2026, 3, 31, 12, 0, 0, 0, time.UTC),
			},
			IsReplay: false,
		},
	}

	txHandler := handler.NewTransactionHandler(uc)
	r := httpapi.NewRouter(config.Config{AppEnv: "local"}, nil, txHandler)

	body := `{"items":[{"medicine_id":"PARA500","qty":10},{"medicine_id":"AMOX500","qty":2}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/transactions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "tx-create-123")
	req.Header.Set("Idempotency-Key", "idem-123")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", w.Code, w.Body.String())
	}
	if uc.capturedCreate.IdempotencyKey != "idem-123" {
		t.Fatalf("expected captured idempotency key idem-123, got %q", uc.capturedCreate.IdempotencyKey)
	}
	if got := w.Header().Get("Idempotency-Key"); got != "idem-123" {
		t.Fatalf("expected response header Idempotency-Key idem-123, got %q", got)
	}
	if !strings.Contains(w.Body.String(), `"status":"APPROVED"`) {
		t.Fatalf("expected status APPROVED, body=%s", w.Body.String())
	}
}

func TestCreateTransaction_PendingReviewStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uc := &fakeTransactionUsecase{
		createResult: domain.CreateTransactionResult{
			Transaction: domain.Transaction{
				ID:             "TXN-20260331121000-AB12CD34",
				IdempotencyKey: "idem-124",
				Status:         domain.TransactionStatusPendingReview,
				Items: []domain.TransactionItem{
					{MedicineID: "PARA500", Qty: 2},
				},
				CreatedAt: time.Date(2026, 3, 31, 12, 10, 0, 0, time.UTC),
			},
			IsReplay: false,
		},
	}

	txHandler := handler.NewTransactionHandler(uc)
	r := httpapi.NewRouter(config.Config{AppEnv: "local"}, nil, txHandler)

	body := `{"items":[{"medicine_id":"PARA500","qty":2}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/transactions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "idem-124")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"status":"PENDING_REVIEW"`) {
		t.Fatalf("expected status PENDING_REVIEW, body=%s", w.Body.String())
	}
}

func TestCreateTransaction_FlaggedStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uc := &fakeTransactionUsecase{
		createResult: domain.CreateTransactionResult{
			Transaction: domain.Transaction{
				ID:             "TXN-20260331122000-AB12CD34",
				IdempotencyKey: "idem-125",
				Status:         domain.TransactionStatusFlagged,
				Items: []domain.TransactionItem{
					{MedicineID: "OBATKERAS-X", Qty: 2},
				},
				CreatedAt: time.Date(2026, 3, 31, 12, 20, 0, 0, time.UTC),
			},
			IsReplay: false,
		},
	}

	txHandler := handler.NewTransactionHandler(uc)
	r := httpapi.NewRouter(config.Config{AppEnv: "local"}, nil, txHandler)

	body := `{"items":[{"medicine_id":"OBATKERAS-X","qty":2}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/transactions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "idem-125")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"status":"FLAGGED"`) {
		t.Fatalf("expected status FLAGGED, body=%s", w.Body.String())
	}
}

func TestCreateTransaction_ReplayReturns200(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uc := &fakeTransactionUsecase{
		createResult: domain.CreateTransactionResult{
			Transaction: domain.Transaction{
				ID:             "TXN-20260331120000-AB12CD34",
				IdempotencyKey: "idem-123",
				Status:         domain.TransactionStatusApproved,
				Items: []domain.TransactionItem{
					{MedicineID: "PARA500", Qty: 10},
				},
				CreatedAt: time.Date(2026, 3, 31, 12, 0, 0, 0, time.UTC),
			},
			IsReplay: true,
		},
	}

	txHandler := handler.NewTransactionHandler(uc)
	r := httpapi.NewRouter(config.Config{AppEnv: "local"}, nil, txHandler)

	body := `{"items":[{"medicine_id":"PARA500","qty":10}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/transactions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "idem-123")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for replay, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestCreateTransaction_MissingIdempotencyKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uc := &fakeTransactionUsecase{}
	txHandler := handler.NewTransactionHandler(uc)
	r := httpapi.NewRouter(config.Config{AppEnv: "local"}, nil, txHandler)

	body := `{"items":[{"medicine_id":"PARA500","qty":10}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/transactions", strings.NewReader(body))
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
	if resp.Error.Details.Field != "header.Idempotency-Key" {
		t.Fatalf("expected field header.Idempotency-Key, got %q body=%s", resp.Error.Details.Field, w.Body.String())
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
	req.Header.Set("Idempotency-Key", "idem-999")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d body=%s", w.Code, w.Body.String())
	}

	resp := decodeErrorEnvelope(t, w.Body.String())
	if resp.Error.Code != "INSUFFICIENT_STOCK" {
		t.Fatalf("expected INSUFFICIENT_STOCK, got %q body=%s", resp.Error.Code, w.Body.String())
	}
	if resp.Error.Details.AvailableQty != 3 {
		t.Fatalf("expected available_qty 3, got %d body=%s", resp.Error.Details.AvailableQty, w.Body.String())
	}
}

func TestListTransactions_DefaultPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uc := &fakeTransactionUsecase{
		listResult: domain.ListTransactionsResult{
			Items: []domain.Transaction{
				{
					ID:     "TXN-20260331093000-AAAA1111",
					Status: domain.TransactionStatusApproved,
					Items: []domain.TransactionItem{
						{MedicineID: "PARA500", Qty: 2},
					},
					CreatedAt: time.Date(2026, 3, 31, 9, 30, 0, 0, time.UTC),
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
	if !strings.Contains(w.Body.String(), `"status":"APPROVED"`) {
		t.Fatalf("expected APPROVED in list response, body=%s", w.Body.String())
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

	req := httptest.NewRequest(http.MethodGet, "/v1/transactions?limit=5&offset=10&status=pending_review", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	if uc.capturedList.Status != "pending_review" {
		t.Fatalf("expected raw captured status pending_review, got %q", uc.capturedList.Status)
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
}