package handler_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"finpharm-ai/services/inventory/internal/config"
	"finpharm-ai/services/inventory/internal/httpapi"
	"github.com/gin-gonic/gin"
)

func TestListMedicines_DefaultPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := httpapi.NewRouter(config.Config{AppEnv: "local"})

	req := httptest.NewRequest(http.MethodGet, "/v1/medicines", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", w.Code, w.Body.String())
	}

	body := w.Body.String()
	if !strings.Contains(body, `"data"`) || !strings.Contains(body, `"request_id"`) {
		t.Fatalf("expected envelope data+request_id, got %s", body)
	}
}

func TestListMedicines_WithLimitOffset(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := httpapi.NewRouter(config.Config{AppEnv: "local"})

	req := httptest.NewRequest(http.MethodGet, "/v1/medicines?limit=1&offset=1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", w.Code, w.Body.String())
	}
}