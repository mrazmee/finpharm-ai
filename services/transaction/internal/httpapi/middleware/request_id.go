package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	HeaderRequestID = "X-Request-ID"
	CtxKeyRequestID = "request_id"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(HeaderRequestID)
		if rid == "" {
			rid = newRequestID()
		}

		c.Set(CtxKeyRequestID, rid)
		c.Writer.Header().Set(HeaderRequestID, rid)

		c.Next()
	}
}

func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// fallback kalau crypto/rand gagal (rare)
		return strconv.FormatInt(time.Now().UnixNano(), 16)
	}
	return hex.EncodeToString(b[:])
}
