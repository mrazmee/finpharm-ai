package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"finpharm-ai/services/gateway/internal/config"
	"finpharm-ai/services/gateway/internal/httpapi/handler"

	"github.com/gin-gonic/gin"
)

func TestIssueToken_LocalOK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{
		AppEnv:           "local",
		JWTSecret:        "test-secret",
		JWTIssuer:        "finpharm-gateway",
		JWTExpireMinutes: 60,
	}

	authHandler := handler.NewAuthHandler(cfg)

	r := gin.New()
	r.POST("/v1/auth/token", authHandler.IssueToken)

	body := []byte(`{"user_id":"staff-001","role":"staff"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/token", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	data, ok := resp["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data object, got %#v", resp["data"])
	}

	if data["access_token"] == "" {
		t.Fatalf("expected non-empty access_token, got %#v", data["access_token"])
	}
}

func TestIssueToken_ProdNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{
		AppEnv:           "prod",
		JWTSecret:        "test-secret",
		JWTIssuer:        "finpharm-gateway",
		JWTExpireMinutes: 60,
	}

	authHandler := handler.NewAuthHandler(cfg)

	r := gin.New()
	r.POST("/v1/auth/token", authHandler.IssueToken)

	body := []byte(`{"user_id":"staff-001","role":"staff"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/token", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", rec.Code, rec.Body.String())
	}
}