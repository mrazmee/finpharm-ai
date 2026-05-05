package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateEntry struct {
	count       int
	windowStart time.Time
}

type InMemoryRateLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	entries map[string]rateEntry
}

func NewInMemoryRateLimiter(limit int, window time.Duration) *InMemoryRateLimiter {
	if limit <= 0 {
		limit = 60
	}
	if window <= 0 {
		window = time.Minute
	}

	return &InMemoryRateLimiter{
		limit:   limit,
		window:  window,
		entries: make(map[string]rateEntry),
	}
}

func (l *InMemoryRateLimiter) Allow(key string) (bool, int, time.Duration) {
	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	entry, ok := l.entries[key]
	if !ok || now.Sub(entry.windowStart) >= l.window {
		l.entries[key] = rateEntry{
			count:       1,
			windowStart: now,
		}
		return true, l.limit - 1, l.window
	}

	if entry.count >= l.limit {
		retryAfter := l.window - now.Sub(entry.windowStart)
		if retryAfter < 0 {
			retryAfter = 0
		}
		return false, 0, retryAfter
	}

	entry.count++
	l.entries[key] = entry

	remaining := l.limit - entry.count
	if remaining < 0 {
		remaining = 0
	}

	retryAfter := l.window - now.Sub(entry.windowStart)
	if retryAfter < 0 {
		retryAfter = 0
	}

	return true, remaining, retryAfter
}

func RateLimitGin(generalLimiter, authLimiter *InMemoryRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		limiter := generalLimiter
		if isAuthPath(c.Request.URL.Path) && authLimiter != nil {
			limiter = authLimiter
		}

		if limiter == nil {
			c.Next()
			return
		}

		key := clientIPFromRequest(c.Request)
		allowed, remaining, retryAfter := limiter.Allow(key)

		c.Header("X-RateLimit-Limit", strconv.Itoa(limiter.limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Window-Seconds", strconv.Itoa(int(limiter.window.Seconds())))

		if !allowed {
			requestID := strings.TrimSpace(c.Writer.Header().Get("X-Request-Id"))
			if requestID == "" {
				requestID = strings.TrimSpace(c.GetHeader("X-Request-Id"))
			}
			if requestID == "" {
				requestID = generateRequestID()
			}

			c.Header("X-Request-Id", requestID)
			c.Header("Retry-After", strconv.Itoa(int(retryAfter.Seconds())+1))

			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{
					"code":    "RATE_LIMITED",
					"message": "too many requests",
					"details": gin.H{
						"retry_after_seconds": int(retryAfter.Seconds()) + 1,
					},
					"request_id": requestID,
				},
			})
			return
		}

		c.Next()
	}
}

func isAuthPath(path string) bool {
	return path == "/v1/auth/token"
}

func clientIPFromRequest(r *http.Request) string {
	if r == nil {
		return "unknown"
	}

	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return ip
			}
		}
	}

	if xrip := strings.TrimSpace(r.Header.Get("X-Real-Ip")); xrip != "" {
		return xrip
	}

	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && host != "" {
		return host
	}

	if ra := strings.TrimSpace(r.RemoteAddr); ra != "" {
		return ra
	}

	return "unknown"
}

func generateRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "rl-fallback-request-id"
	}
	return hex.EncodeToString(b[:])
}