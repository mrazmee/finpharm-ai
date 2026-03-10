package repository

import "context"

type ctxKey string

const requestIDKey ctxKey = "request_id"

func WithRequestID(ctx context.Context, requestID string) context.Context {
	if requestID == "" {
		return ctx
	}
	return context.WithValue(ctx, requestIDKey, requestID)
}

func RequestIDFromContext(ctx context.Context) string {
	v := ctx.Value(requestIDKey)
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}