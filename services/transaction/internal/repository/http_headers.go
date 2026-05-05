package repository

import (
	"context"
	"net/http"

	"finpharm-ai/internal/telemetry/tracehttp"
)

func applyCommonHeadersFromContext(ctx context.Context, req *http.Request) {
	if req == nil {
		return
	}

	if requestID := RequestIDFromContext(ctx); requestID != "" {
		req.Header.Set(tracehttp.HeaderRequestID, requestID)
	}

	if traceID := tracehttp.TraceIDFromContext(ctx); traceID != "" {
		req.Header.Set(tracehttp.HeaderTraceID, traceID)
	}

	if userID := tracehttp.UserIDFromContext(ctx); userID != "" {
		req.Header.Set(tracehttp.HeaderUserID, userID)
	}

	if role := tracehttp.UserRoleFromContext(ctx); role != "" {
		req.Header.Set(tracehttp.HeaderUserRole, role)
	}
}