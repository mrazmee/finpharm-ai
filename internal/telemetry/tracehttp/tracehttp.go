package tracehttp

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
)

const (
	HeaderTraceID   = "X-Trace-Id"
	HeaderRequestID = "X-Request-Id"
	HeaderUserID    = "X-User-Id"
	HeaderUserRole  = "X-User-Role"
)

type contextKey string

const (
	traceIDContextKey  contextKey = "trace_id"
	userIDContextKey   contextKey = "user_id"
	userRoleContextKey contextKey = "user_role"
)

func Handler(service string, next http.Handler) http.Handler {
	_ = service

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := strings.TrimSpace(r.Header.Get(HeaderTraceID))
		if traceID == "" {
			traceID = strings.TrimSpace(r.Header.Get(HeaderRequestID))
		}
		if traceID == "" {
			traceID = generateTraceID()
		}

		userID := strings.TrimSpace(r.Header.Get(HeaderUserID))
		userRole := strings.TrimSpace(r.Header.Get(HeaderUserRole))

		r.Header.Set(HeaderTraceID, traceID)
		w.Header().Set(HeaderTraceID, traceID)

		ctx := context.WithValue(r.Context(), traceIDContextKey, traceID)
		ctx = context.WithValue(ctx, userIDContextKey, userID)
		ctx = context.WithValue(ctx, userRoleContextKey, userRole)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func TraceIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	v, _ := ctx.Value(traceIDContextKey).(string)
	return v
}

func UserIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	v, _ := ctx.Value(userIDContextKey).(string)
	return v
}

func UserRoleFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	v, _ := ctx.Value(userRoleContextKey).(string)
	return v
}

func generateTraceID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "trace-fallback"
	}
	return hex.EncodeToString(b[:])
}