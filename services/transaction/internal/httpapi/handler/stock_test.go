package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"finpharm-ai/services/transaction/internal/config"
	"finpharm-ai/services/transaction/internal/httpapi"
	"finpharm-ai/services/transaction/internal/httpapi/handler"
	"finpharm-ai/services/transaction/internal/repository"
	"finpharm-ai/services/transaction/internal/usecase"

	"github.com/gin-gonic/gin"
)

func newStockCheckRouter(cfg config.Config) *gin.Engine {
	httpClient := &http.Client{Timeout: 4 * time.Second}
	breaker := repository.NewCircuitBreaker(3, 5*time.Second)
	stockRepo := repository.NewStockHTTPRepo(cfg.InventoryBaseURL, httpClient, breaker)
	stockUC := usecase.NewStockUsecase(stockRepo)
	stockHandler := handler.NewStockHandler(stockUC)

	return httpapi.NewRouter(cfg, stockHandler, nil)
}

func TestCheckStock_ForwardsOriginalQtyToInventory(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var capturedQty int32

	inv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			MedicineID string `json:"medicine_id"`
			Qty        int    `json:"qty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		atomic.StoreInt32(&capturedQty, int32(req.Qty))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"medicine_id":"PARA500","requested_qty":7,"available_qty":80,"is_available":true},"request_id":"inv"}`))
	}))
	defer inv.Close()

	r := newStockCheckRouter(config.Config{
		AppEnv:           "local",
		InventoryBaseURL: inv.URL,
	})

	body := []byte(`{"medicine_id":"PARA500","qty":7}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/stock/check", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "t-req-qty-forward")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	if got := atomic.LoadInt32(&capturedQty); got != 7 {
		t.Fatalf("expected forwarded qty=7, got %d", got)
	}
}

func TestCheckStock_RetryOnceOnUpstreamError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var calls int32

	inv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)

		if r.Header.Get("X-Request-ID") == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if r.Header.Get("X-Caller-Service") != "transaction" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if atomic.LoadInt32(&calls) == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"medicine_id":"PARA500","requested_qty":1,"available_qty":80,"is_available":true},"request_id":"inv"}`))
	}))
	defer inv.Close()

	r := newStockCheckRouter(config.Config{
		AppEnv:           "local",
		InventoryBaseURL: inv.URL,
	})

	body := []byte(`{"medicine_id":"PARA500","qty":1}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/stock/check", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "t-req-123")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("expected 2 calls due to retry, got %d", calls)
	}
}

func TestCheckStock_CircuitBreakerOpen_FailFast502(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var calls int32

	inv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer inv.Close()

	r := newStockCheckRouter(config.Config{
		AppEnv:           "local",
		InventoryBaseURL: inv.URL,
	})

	body := []byte(`{"medicine_id":"PARA500","qty":1}`)

	for i := 0; i < 4; i++ {
		req := httptest.NewRequest(http.MethodPost, "/v1/stock/check", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Request-ID", "t-req-123")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}

	before := atomic.LoadInt32(&calls)

	req := httptest.NewRequest(http.MethodPost, "/v1/stock/check", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-ID", "t-req-123")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d body=%s", w.Code, w.Body.String())
	}

	after := atomic.LoadInt32(&calls)
	if after != before {
		t.Fatalf("expected no additional upstream calls when breaker open, before=%d after=%d", before, after)
	}
}
