package handler_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"finpharm-ai/services/transaction/internal/config"
	"finpharm-ai/services/transaction/internal/httpapi"
	"github.com/gin-gonic/gin"
)

func TestCheckStock_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := httpapi.NewRouter(config.Config{AppEnv: "local"})

	body := []byte(`{"medicine_id":"PARA500","qty":10}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/stock/check", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestCheckStock_InvalidBody_HasRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := httpapi.NewRouter(config.Config{AppEnv: "local"})

	body := []byte(`{"medicine_id":"","qty":0}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/stock/check", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d, body=%s", w.Code, w.Body.String())
	}

	// Pastikan error response mengandung request_id
	respBody := w.Body.String()
	if !contains(respBody, `"request_id"`) {
		t.Fatalf("expected response body to contain request_id, got: %s", respBody)
	}
}

// helper minimal (tanpa strings.Contains biar kamu lihat pattern)
func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}
