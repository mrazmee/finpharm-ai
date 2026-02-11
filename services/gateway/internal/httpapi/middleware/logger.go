package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		ridVal, _ := c.Get(CtxKeyRequestID)
		rid, _ := ridVal.(string)

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		slog.Info("http_request",
			"request_id", rid,
			"method", c.Request.Method,
			"path", path,
			"status", c.Writer.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
			"client_ip", c.ClientIP(),
			"errors", len(c.Errors),
		)
	}
}
