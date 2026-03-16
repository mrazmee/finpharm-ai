package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"finpharm-ai/services/inventory/internal/domain"
	"finpharm-ai/services/inventory/internal/httpapi/handler"

	"github.com/gin-gonic/gin"
)

type fakeStockUsecase struct {
	checkResult domain.StockCheckResult
	checkErr    error

	deductResult domain.StockDeductResult
	deductErr    error
}

func (f *fakeStockUsecase) CheckStock(ctx context.Context, req domain.StockCheckRequest) (domain.StockCheckResult, error) {
	if f.checkErr != nil {
		return domain.StockCheckResult{}, f.checkErr
	}
	return f.checkResult, nil
}

func (f *fakeStockUsecase) DeductStock(ctx context.Context, req domain.StockDeductRequest) (domain.StockDeductResult, error) {
	if f.deductErr != nil {
		return domain.StockDeductResult{}, f.deductErr
	}
	return f.deductResult, nil
}

type errorEnvelope struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Details struct {
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
		t.Fatalf("failed to decode error response: %v body=%s", err, body)
	}
	return resp
}

func TestDeductStock_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uc := &fakeStockUsecase{
		deductResult: domain.StockDeductResult{
			MedicineID:   "PARA500",
			DeductedQty:  2,
			RemainingQty: 78,
		},
	}

	h := handler.NewStockHandler(uc)

	r := gin.New()
	r.POST("/v1/stock/deduct", h.DeductStock)

	req := httptest.NewRequest(http.MethodPost, "/v1/stock/deduct", strings.NewReader(`{"medicine_id":"PARA500","qty":2}`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"remaining_qty":78`) {
		t.Fatalf("expected remaining_qty 78, body=%s", w.Body.String())
	}
}

func TestDeductStock_InsufficientStock(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uc := &fakeStockUsecase{
		deductErr: &domain.InsufficientStockError{
			MedicineID:   "PARA500",
			RequestedQty: 10,
			AvailableQty: 3,
		},
	}

	h := handler.NewStockHandler(uc)

	r := gin.New()
	r.POST("/v1/stock/deduct", h.DeductStock)

	req := httptest.NewRequest(http.MethodPost, "/v1/stock/deduct", strings.NewReader(`{"medicine_id":"PARA500","qty":10}`))
	req.Header.Set("Content-Type", "application/json")

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