package audithttp

import (
	"log/slog"
	"net/http"
	"strings"

	"finpharm-ai/internal/telemetry/tracehttp"
)

type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
	r.wroteHeader = true
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *statusRecorder) Write(p []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	return r.ResponseWriter.Write(p)
}

func Handler(service string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(rec, r)

		if !shouldAudit(r.Method, r.URL.Path) {
			return
		}

		requestID := strings.TrimSpace(rec.Header().Get(tracehttp.HeaderRequestID))
		if requestID == "" {
			requestID = strings.TrimSpace(r.Header.Get(tracehttp.HeaderRequestID))
		}

		traceID := strings.TrimSpace(rec.Header().Get(tracehttp.HeaderTraceID))
		if traceID == "" {
			traceID = strings.TrimSpace(r.Header.Get(tracehttp.HeaderTraceID))
		}

		userID := strings.TrimSpace(r.Header.Get(tracehttp.HeaderUserID))
		role := strings.TrimSpace(r.Header.Get(tracehttp.HeaderUserRole))

		slog.Info("audit_http_event",
			"audited_service", service,
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"request_id", requestID,
			"trace_id", traceID,
			"user_id", userID,
			"role", role,
			"remote_addr", r.RemoteAddr,
		)
	})
}

func shouldAudit(method string, path string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		if path == "/metrics" || path == "/health" {
			return false
		}
		return true
	case http.MethodGet:
		switch path {
		case "/v1/transactions", "/v1/debug/sleep":
			return true
		default:
			return false
		}
	default:
		return false
	}
}