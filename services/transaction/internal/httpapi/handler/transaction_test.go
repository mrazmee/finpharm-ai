package handler_test

import (
	"context"
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
	captured domain.CreateTransactionRequest
	result   domain.Transaction
	err      error
}

func (f *fakeTransactionUsecase) CreateTransaction(ctx context.Context, req domain.CreateTransactionRequest) (domain.Transaction, error) {
	f.captured = req
	if f.err != nil {
		return domain.Transaction{}, f.err
	}
	return f.result, nil
}

func TestCreateTransaction_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uc := &fakeTransactionUsecase{
		result: domain.Transaction{
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
	if len(uc.captured.Items) != 2 {
		t.Fatalf("expected 2 items captured, got %d", len(uc.captured.Items))
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
		err: &domain.InsufficientStockError{
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
