package handler_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"finpharm-ai/services/transaction/internal/config"
	"finpharm-ai/services/transaction/internal/httpapi"
	"github.com/gin-gonic/gin"
)

func TestCheckStock_RetryOnceOnUpstreamError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var calls int32

	// First call fails (500), second call succeeds
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

	r := httpapi.NewRouter(config.Config{
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

	// Always fail (500)
	inv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer inv.Close()

	r := httpapi.NewRouter(config.Config{
		AppEnv:           "local",
		InventoryBaseURL: inv.URL,
	})

	body := []byte(`{"medicine_id":"PARA500","qty":1}`)

	// Do enough requests to trip breaker (threshold=3). Each request retries once, so calls accumulate fast.
	for i := 0; i < 4; i++ {
		req := httptest.NewRequest(http.MethodPost, "/v1/stock/check", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Request-ID", "t-req-123")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}

	// After breaker opens, next request should fail fast without hitting inventory
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