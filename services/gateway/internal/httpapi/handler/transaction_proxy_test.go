package handler_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"finpharm-ai/services/gateway/internal/httpapi/handler"
	"finpharm-ai/services/gateway/internal/httpapi/middleware"

	"github.com/gin-gonic/gin"
)

func TestGatewayCreateTransaction_ProxyOK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var gotRequestID string
	var gotCallerService string

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/transactions" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		gotRequestID = r.Header.Get(middleware.HeaderRequestID)
		gotCallerService = r.Header.Get("X-Caller-Service")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"data":{"id":"TXN-20260313093000-ABCD1234","status":"PENDING","items":[{"medicine_id":"PARA500","qty":2}],"created_at":"2026-03-13T09:30:00Z"},"request_id":"x"}`))
	}))
	defer upstream.Close()

	r := gin.New()
	r.Use(middleware.RequestID(), middleware.RequestLogger(), gin.Recovery())

	proxy := handler.NewTransactionProxyHandler(upstream.URL)
	r.POST("/v1/transactions", proxy.CreateTransaction)

	body := []byte(`{"items":[{"medicine_id":"PARA500","qty":2}]}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/transactions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(middleware.HeaderRequestID, "gw-tx-123")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d, body=%s", w.Code, w.Body.String())
	}
	if gotRequestID != "gw-tx-123" {
		t.Fatalf("expected propagated request id gw-tx-123, got %q", gotRequestID)
	}
	if gotCallerService != "gateway" {
		t.Fatalf("expected caller service gateway, got %q", gotCallerService)
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"status":"PENDING"`)) {
		t.Fatalf("expected response to contain transaction status, got %s", w.Body.String())
	}
}

func TestGatewayCreateTransaction_ValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	upstreamCalled := false
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamCalled = true
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer upstream.Close()

	r := gin.New()
	r.Use(middleware.RequestID(), middleware.RequestLogger(), gin.Recovery())

	proxy := handler.NewTransactionProxyHandler(upstream.URL)
	r.POST("/v1/transactions", proxy.CreateTransaction)

	body := []byte(`{"items":[{"medicine_id":"","qty":0}]}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/transactions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d, body=%s", w.Code, w.Body.String())
	}
	if upstreamCalled {
		t.Fatalf("expected gateway validation to fail fast before calling upstream")
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"code":"VALIDATION_ERROR"`)) {
		t.Fatalf("expected validation error response, got %s", w.Body.String())
	}
}