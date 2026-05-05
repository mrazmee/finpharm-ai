package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestRateLimitGin_AllowsUntilLimitThenBlocks(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RateLimitGin(NewInMemoryRateLimiter(1, time.Minute), nil))
	router.POST("/v1/transactions", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req1 := httptest.NewRequest(http.MethodPost, "/v1/transactions", nil)
	req1.RemoteAddr = "192.0.2.10:1234"
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("expected first request status 200, got %d", w1.Code)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/v1/transactions", nil)
	req2.RemoteAddr = "192.0.2.10:1234"
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected second request status 429, got %d", w2.Code)
	}

	if got := w2.Header().Get("X-RateLimit-Limit"); got != "1" {
		t.Fatalf("expected X-RateLimit-Limit=1, got %q", got)
	}

	if got := w2.Header().Get("Retry-After"); got == "" {
		t.Fatal("expected Retry-After header to be set")
	}

	var body struct {
		Error struct {
			Code      string `json:"code"`
			Message   string `json:"message"`
			RequestID string `json:"request_id"`
		} `json:"error"`
	}
	if err := json.Unmarshal(w2.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if body.Error.Code != "RATE_LIMITED" {
		t.Fatalf("expected error code RATE_LIMITED, got %q", body.Error.Code)
	}
	if body.Error.RequestID == "" {
		t.Fatal("expected request_id to be present in error body")
	}
}

func TestRateLimitGin_UsesAuthLimiterForAuthPath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	generalLimiter := NewInMemoryRateLimiter(100, time.Minute)
	authLimiter := NewInMemoryRateLimiter(1, time.Minute)

	router := gin.New()
	router.Use(RateLimitGin(generalLimiter, authLimiter))

	router.POST("/v1/auth/token", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	router.POST("/v1/transactions", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req1 := httptest.NewRequest(http.MethodPost, "/v1/auth/token", nil)
	req1.RemoteAddr = "192.0.2.20:1234"
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("expected first auth request status 200, got %d", w1.Code)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/v1/auth/token", nil)
	req2.RemoteAddr = "192.0.2.20:1234"
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected second auth request status 429, got %d", w2.Code)
	}

	req3 := httptest.NewRequest(http.MethodPost, "/v1/transactions", nil)
	req3.RemoteAddr = "192.0.2.20:1234"
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)

	if w3.Code != http.StatusOK {
		t.Fatalf("expected non-auth request to use general limiter and pass, got %d", w3.Code)
	}
}