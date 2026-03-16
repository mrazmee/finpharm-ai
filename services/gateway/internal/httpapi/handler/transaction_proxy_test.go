package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"finpharm-ai/services/gateway/internal/httpapi/handler"
	"finpharm-ai/services/gateway/internal/httpapi/middleware"

	"github.com/gin-gonic/gin"
)

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

	resp := decodeErrorEnvelope(t, w.Body.String())
	if resp.Error.Code != "VALIDATION_ERROR" {
		t.Fatalf("expected validation error response, got %q body=%s", resp.Error.Code, w.Body.String())
	}
}

func TestGatewayListTransactions_ProxyOK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var gotRequestID string
	var gotCallerService string
	var gotRawQuery string

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/transactions" || r.Method != http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		gotRequestID = r.Header.Get(middleware.HeaderRequestID)
		gotCallerService = r.Header.Get("X-Caller-Service")
		gotRawQuery = r.URL.RawQuery

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"items":[],"limit":5,"offset":10,"total":0},"request_id":"x"}`))
	}))
	defer upstream.Close()

	r := gin.New()
	r.Use(middleware.RequestID(), middleware.RequestLogger(), gin.Recovery())

	proxy := handler.NewTransactionProxyHandler(upstream.URL)
	r.GET("/v1/transactions", proxy.ListTransactions)

	req := httptest.NewRequest(http.MethodGet, "/v1/transactions?limit=5&offset=10&status=pending", nil)
	req.Header.Set(middleware.HeaderRequestID, "gw-list-123")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", w.Code, w.Body.String())
	}
	if gotRequestID != "gw-list-123" {
		t.Fatalf("expected propagated request id gw-list-123, got %q", gotRequestID)
	}
	if gotCallerService != "gateway" {
		t.Fatalf("expected caller service gateway, got %q", gotCallerService)
	}
	if gotRawQuery != "limit=5&offset=10&status=pending" {
		t.Fatalf("expected raw query preserved, got %q", gotRawQuery)
	}
}

func TestGatewayListTransactions_InvalidLimit(t *testing.T) {
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
	r.GET("/v1/transactions", proxy.ListTransactions)

	req := httptest.NewRequest(http.MethodGet, "/v1/transactions?limit=abc", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d, body=%s", w.Code, w.Body.String())
	}
	if upstreamCalled {
		t.Fatalf("expected gateway validation to fail fast before calling upstream")
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

func TestGatewayListTransactions_ZeroLimit(t *testing.T) {
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
	r.GET("/v1/transactions", proxy.ListTransactions)

	req := httptest.NewRequest(http.MethodGet, "/v1/transactions?limit=0", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d, body=%s", w.Code, w.Body.String())
	}
	if upstreamCalled {
		t.Fatalf("expected gateway validation to fail fast before calling upstream")
	}

	resp := decodeErrorEnvelope(t, w.Body.String())
	if resp.Error.Details.Field != "limit" {
		t.Fatalf("expected validation error on limit, got %q body=%s", resp.Error.Details.Field, w.Body.String())
	}
	if resp.Error.Details.Reason != "must be > 0" {
		t.Fatalf("expected reason must be > 0, got %q body=%s", resp.Error.Details.Reason, w.Body.String())
	}
}

func TestGatewayListTransactions_LimitTooLarge(t *testing.T) {
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
	r.GET("/v1/transactions", proxy.ListTransactions)

	req := httptest.NewRequest(http.MethodGet, "/v1/transactions?limit=101", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d, body=%s", w.Code, w.Body.String())
	}
	if upstreamCalled {
		t.Fatalf("expected gateway validation to fail fast before calling upstream")
	}

	resp := decodeErrorEnvelope(t, w.Body.String())
	if resp.Error.Details.Field != "limit" {
		t.Fatalf("expected validation error on limit, got %q body=%s", resp.Error.Details.Field, w.Body.String())
	}
	if resp.Error.Details.Reason != "must be <= 100" {
		t.Fatalf("expected reason must be <= 100, got %q body=%s", resp.Error.Details.Reason, w.Body.String())
	}
}