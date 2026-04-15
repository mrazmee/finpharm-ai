package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimitHandler_AllowsUntilLimitThenBlocks(t *testing.T) {
	generalLimiter := NewInMemoryRateLimiter(2, time.Minute)

	h := RateLimitHandler(generalLimiter, nil, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req1 := httptest.NewRequest(http.MethodGet, "/v1/medicines", nil)
	req1.RemoteAddr = "192.0.2.10:12345"
	rec1 := httptest.NewRecorder()
	h.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusOK {
		t.Fatalf("expected first request 200, got %d", rec1.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/v1/medicines", nil)
	req2.RemoteAddr = "192.0.2.10:12345"
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("expected second request 200, got %d", rec2.Code)
	}

	req3 := httptest.NewRequest(http.MethodGet, "/v1/medicines", nil)
	req3.RemoteAddr = "192.0.2.10:12345"
	rec3 := httptest.NewRecorder()
	h.ServeHTTP(rec3, req3)

	if rec3.Code != http.StatusTooManyRequests {
		t.Fatalf("expected third request 429, got %d", rec3.Code)
	}
}

func TestRateLimitHandler_UsesAuthLimiterForAuthPath(t *testing.T) {
	generalLimiter := NewInMemoryRateLimiter(10, time.Minute)
	authLimiter := NewInMemoryRateLimiter(1, time.Minute)

	h := RateLimitHandler(generalLimiter, authLimiter, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req1 := httptest.NewRequest(http.MethodPost, "/v1/auth/token", nil)
	req1.RemoteAddr = "192.0.2.20:12345"
	rec1 := httptest.NewRecorder()
	h.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusOK {
		t.Fatalf("expected first auth request 200, got %d", rec1.Code)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/v1/auth/token", nil)
	req2.RemoteAddr = "192.0.2.20:12345"
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected second auth request 429, got %d", rec2.Code)
	}
}