package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"finpharm-ai/services/gateway/internal/config"
	"finpharm-ai/services/gateway/internal/httpapi"
	"github.com/gin-gonic/gin"
)

func TestGatewayDebugSleep_EnabledInLocal(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Upstream fake transaction debug endpoint
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/debug/sleep" || r.Method != http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"slept_ms":1}`))
	}))
	defer upstream.Close()

	cfg := config.Config{
		AppEnv:             "local",
		Port:               "0",
		TransactionBaseURL: upstream.URL,
	}

	r := httpapi.NewRouter(cfg)

	req := httptest.NewRequest(http.MethodGet, "/v1/debug/sleep?ms=1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestGatewayDebugSleep_DisabledInProd(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// upstream tidak penting karena route gateway harus 404 duluan
	cfg := config.Config{
		AppEnv:             "prod",
		Port:               "0",
		TransactionBaseURL: "http://example.com",
	}

	r := httpapi.NewRouter(cfg)

	req := httptest.NewRequest(http.MethodGet, "/v1/debug/sleep?ms=1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d, body=%s", w.Code, w.Body.String())
	}
}
