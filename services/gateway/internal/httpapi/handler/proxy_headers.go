package handler

import (
	"net/http"
	"strings"

	"finpharm-ai/internal/telemetry/tracehttp"
	"finpharm-ai/services/gateway/internal/httpapi/middleware"

	"github.com/gin-gonic/gin"
)

func setProxyForwardHeaders(c *gin.Context, req *http.Request) {
	if ridVal, ok := c.Get(middleware.CtxKeyRequestID); ok {
		if rid, _ := ridVal.(string); strings.TrimSpace(rid) != "" {
			req.Header.Set(middleware.HeaderRequestID, rid)
		}
	}

	if traceID := strings.TrimSpace(c.Request.Header.Get(tracehttp.HeaderTraceID)); traceID != "" {
		req.Header.Set(tracehttp.HeaderTraceID, traceID)
	}

	if userVal, ok := c.Get(middleware.CtxKeyUserID); ok {
		if userID, _ := userVal.(string); strings.TrimSpace(userID) != "" {
			req.Header.Set(tracehttp.HeaderUserID, userID)
		}
	} else if userID := strings.TrimSpace(c.Request.Header.Get(tracehttp.HeaderUserID)); userID != "" {
		req.Header.Set(tracehttp.HeaderUserID, userID)
	}

	if roleVal, ok := c.Get(middleware.CtxKeyUserRole); ok {
		if role, _ := roleVal.(string); strings.TrimSpace(role) != "" {
			req.Header.Set(tracehttp.HeaderUserRole, role)
		}
	} else if role := strings.TrimSpace(c.Request.Header.Get(tracehttp.HeaderUserRole)); role != "" {
		req.Header.Set(tracehttp.HeaderUserRole, role)
	}

	req.Header.Set("X-Caller-Service", "gateway")
}