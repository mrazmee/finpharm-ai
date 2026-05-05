package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"finpharm-ai/services/gateway/internal/config"
	"finpharm-ai/services/gateway/internal/httpapi/middleware"

	"github.com/gin-gonic/gin"
)

func testGatewayConfig() config.Config {
	return config.Config{
		AppEnv:           "local",
		AuthEnabled:      true,
		JWTSecret:        "test-secret",
		JWTIssuer:        "finpharm-gateway",
		JWTExpireMinutes: 60,
	}
}

func TestIssueToken_StaffAndSupervisor(t *testing.T) {
	cfg := testGatewayConfig()

	staffToken, _, err := middleware.IssueToken(cfg, "staff-001", "staff")
	if err != nil {
		t.Fatalf("expected no error for staff token, got %v", err)
	}
	if staffToken == "" {
		t.Fatal("expected non-empty staff token")
	}

	supervisorToken, _, err := middleware.IssueToken(cfg, "supervisor-001", "supervisor")
	if err != nil {
		t.Fatalf("expected no error for supervisor token, got %v", err)
	}
	if supervisorToken == "" {
		t.Fatal("expected non-empty supervisor token")
	}
}

func TestJWTAuth_UnauthorizedWhenMissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := testGatewayConfig()

	r := gin.New()
	r.GET("/protected", middleware.JWTAuth(cfg), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestJWTAuth_InjectsIdentityToContextAndHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := testGatewayConfig()
	token, _, err := middleware.IssueToken(cfg, "staff-001", "staff")
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	r := gin.New()
	r.GET("/protected", middleware.JWTAuth(cfg), func(c *gin.Context) {
		userID := c.GetString(middleware.CtxKeyUserID)
		role := c.GetString(middleware.CtxKeyUserRole)

		c.JSON(http.StatusOK, gin.H{
			"user_id":    userID,
			"role":       role,
			"x_user_id":  c.Request.Header.Get("X-User-ID"),
			"x_user_role": c.Request.Header.Get("X-User-Role"),
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}

	if body["user_id"] != "staff-001" {
		t.Fatalf("expected user_id staff-001, got %#v", body["user_id"])
	}
	if body["role"] != "staff" {
		t.Fatalf("expected role staff, got %#v", body["role"])
	}
	if body["x_user_id"] != "staff-001" {
		t.Fatalf("expected x_user_id staff-001, got %#v", body["x_user_id"])
	}
	if body["x_user_role"] != "staff" {
		t.Fatalf("expected x_user_role staff, got %#v", body["x_user_role"])
	}
}

func TestRequireRoles_StaffForbiddenForSupervisorRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := testGatewayConfig()
	token, _, err := middleware.IssueToken(cfg, "staff-001", "staff")
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	r := gin.New()
	r.GET(
		"/supervisor-only",
		middleware.JWTAuth(cfg),
		middleware.RequireRoles("supervisor"),
		func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"ok": true})
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/supervisor-only", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestRequireRoles_SupervisorAllowed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := testGatewayConfig()
	token, _, err := middleware.IssueToken(cfg, "supervisor-001", "supervisor")
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	r := gin.New()
	r.GET(
		"/supervisor-only",
		middleware.JWTAuth(cfg),
		middleware.RequireRoles("supervisor"),
		func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"ok": true})
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/supervisor-only", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}