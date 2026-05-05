package httpapi_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"finpharm-ai/services/gateway/internal/config"
	"finpharm-ai/services/gateway/internal/httpapi"
	"finpharm-ai/services/gateway/internal/httpapi/middleware"

	"github.com/gin-gonic/gin"
)

func newGatewayRouterTestConfig(transactionBaseURL string, inventoryBaseURL string) config.Config {
	return config.Config{
		AppEnv:             "local",
		Port:               "8080",
		TransactionBaseURL: transactionBaseURL,
		InventoryBaseURL:   inventoryBaseURL,
		AuthEnabled:        true,
		JWTSecret:          "test-secret",
		JWTIssuer:          "finpharm-gateway",
		JWTExpireMinutes:   60,
	}
}

func newStubServer(status int, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	}))
}

func TestRouterAuth_ListTransactions_UnauthorizedWithoutToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	txServer := newStubServer(http.StatusOK, `{"data":{"items":[],"limit":10,"offset":0,"total":0},"request_id":"tx-req"}`)
	defer txServer.Close()

	invServer := newStubServer(http.StatusOK, `{"data":[],"request_id":"inv-req"}`)
	defer invServer.Close()

	cfg := newGatewayRouterTestConfig(txServer.URL, invServer.URL)
	router := httpapi.NewRouter(cfg)

	req := httptest.NewRequest(http.MethodGet, "/v1/transactions", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRouterAuth_ListTransactions_UnauthorizedWithInvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	txServer := newStubServer(http.StatusOK, `{"data":{"items":[],"limit":10,"offset":0,"total":0},"request_id":"tx-req"}`)
	defer txServer.Close()

	invServer := newStubServer(http.StatusOK, `{"data":[],"request_id":"inv-req"}`)
	defer invServer.Close()

	cfg := newGatewayRouterTestConfig(txServer.URL, invServer.URL)
	router := httpapi.NewRouter(cfg)

	req := httptest.NewRequest(http.MethodGet, "/v1/transactions", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.value")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRouterAuth_DebugSleep_StaffForbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)

	txServer := newStubServer(http.StatusOK, `{"data":{"items":[],"limit":10,"offset":0,"total":0},"request_id":"tx-req"}`)
	defer txServer.Close()

	invServer := newStubServer(http.StatusOK, `{"data":[],"request_id":"inv-req"}`)
	defer invServer.Close()

	cfg := newGatewayRouterTestConfig(txServer.URL, invServer.URL)
	router := httpapi.NewRouter(cfg)

	token, _, err := middleware.IssueToken(cfg, "staff-001", "staff")
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/debug/sleep?ms=1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRouterAuth_DebugSleep_SupervisorAllowed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	txServer := newStubServer(http.StatusOK, `{"data":{"items":[],"limit":10,"offset":0,"total":0},"request_id":"tx-req"}`)
	defer txServer.Close()

	invServer := newStubServer(http.StatusOK, `{"data":[],"request_id":"inv-req"}`)
	defer invServer.Close()

	cfg := newGatewayRouterTestConfig(txServer.URL, invServer.URL)
	router := httpapi.NewRouter(cfg)

	token, _, err := middleware.IssueToken(cfg, "supervisor-001", "supervisor")
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/debug/sleep?ms=1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRouterAuth_ListTransactions_SupervisorAllowed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	txServer := newStubServer(http.StatusOK, `{"data":{"items":[],"limit":10,"offset":0,"total":0},"request_id":"tx-req"}`)
	defer txServer.Close()

	invServer := newStubServer(http.StatusOK, `{"data":[],"request_id":"inv-req"}`)
	defer invServer.Close()

	cfg := newGatewayRouterTestConfig(txServer.URL, invServer.URL)
	router := httpapi.NewRouter(cfg)

	token, _, err := middleware.IssueToken(cfg, "supervisor-001", "supervisor")
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/transactions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}