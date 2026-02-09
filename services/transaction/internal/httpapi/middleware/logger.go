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

		slog.Info("http_request",
			"request_id", rid,
			"method", c.Request.Method,
			"path", c.FullPath(),
			"status", c.Writer.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
			"from_gateway", c.GetHeader("X-From-Gateway"),
			"errors", len(c.Errors),
		)
	}
}
