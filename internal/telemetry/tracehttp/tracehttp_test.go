package tracehttp_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"finpharm-ai/internal/telemetry/tracehttp"
)

func TestTraceHandler_PreservesIncomingTraceID(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get(tracehttp.HeaderTraceID)
		if got != "trace-123" {
			t.Fatalf("expected trace-123, got %q", got)
		}
		if ctxTrace := tracehttp.TraceIDFromContext(r.Context()); ctxTrace != "trace-123" {
			t.Fatalf("expected context trace trace-123, got %q", ctxTrace)
		}
		w.WriteHeader(http.StatusOK)
	})

	handler := tracehttp.Handler("test-service", next)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(tracehttp.HeaderTraceID, "trace-123")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get(tracehttp.HeaderTraceID) != "trace-123" {
		t.Fatalf("expected response trace trace-123, got %q", rec.Header().Get(tracehttp.HeaderTraceID))
	}
}

func TestTraceHandler_FallsBackToRequestID(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get(tracehttp.HeaderTraceID)
		if got != "req-456" {
			t.Fatalf("expected req-456, got %q", got)
		}
		w.WriteHeader(http.StatusOK)
	})

	handler := tracehttp.Handler("test-service", next)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(tracehttp.HeaderRequestID, "req-456")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get(tracehttp.HeaderTraceID) != "req-456" {
		t.Fatalf("expected response trace req-456, got %q", rec.Header().Get(tracehttp.HeaderTraceID))
	}
}

func TestTraceHandler_GeneratesTraceID(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get(tracehttp.HeaderTraceID)
		if got == "" {
			t.Fatal("expected generated trace id, got empty")
		}
		w.WriteHeader(http.StatusOK)
	})

	handler := tracehttp.Handler("test-service", next)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Header().Get(tracehttp.HeaderTraceID) == "" {
		t.Fatal("expected generated trace id in response header, got empty")
	}
}

func TestTraceHandler_StoresUserHeadersInContext(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := tracehttp.UserIDFromContext(r.Context()); got != "staff-001" {
			t.Fatalf("expected user_id staff-001, got %q", got)
		}
		if got := tracehttp.UserRoleFromContext(r.Context()); got != "staff" {
			t.Fatalf("expected user_role staff, got %q", got)
		}
		w.WriteHeader(http.StatusOK)
	})

	handler := tracehttp.Handler("test-service", next)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(tracehttp.HeaderUserID, "staff-001")
	req.Header.Set(tracehttp.HeaderUserRole, "staff")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
}