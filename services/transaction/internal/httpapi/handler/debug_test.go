package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"finpharm-ai/services/transaction/internal/config"
	"finpharm-ai/services/transaction/internal/httpapi"
	"github.com/gin-gonic/gin"
)

func TestDebugSleep_EnabledInLocal(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := httpapi.NewRouter(config.Config{AppEnv: "local"})

	req := httptest.NewRequest(http.MethodGet, "/v1/debug/sleep?ms=1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestDebugSleep_DisabledInProd(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := httpapi.NewRouter(config.Config{AppEnv: "prod"})

	req := httptest.NewRequest(http.MethodGet, "/v1/debug/sleep?ms=1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d, body=%s", w.Code, w.Body.String())
	}
}
