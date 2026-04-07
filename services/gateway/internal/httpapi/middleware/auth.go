package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"finpharm-ai/services/gateway/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	CtxKeyUserID   = "user_id"
	CtxKeyUserRole = "user_role"
)

type GatewayClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func IssueToken(cfg config.Config, userID string, role string) (string, time.Time, error) {
	userID = strings.TrimSpace(userID)
	role = normalizeRole(role)

	if userID == "" {
		return "", time.Time{}, errors.New("user_id is required")
	}
	if !isAllowedRole(role) {
		return "", time.Time{}, errors.New("role must be staff or supervisor")
	}
	if strings.TrimSpace(cfg.JWTSecret) == "" {
		return "", time.Time{}, errors.New("jwt secret is required")
	}

	now := time.Now().UTC()
	expiresAt := now.Add(time.Duration(cfg.JWTExpireMinutes) * time.Minute)

	claims := GatewayClaims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    cfg.JWTIssuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return signed, expiresAt, nil
}

func JWTAuth(cfg config.Config) gin.HandlerFunc {
	if !cfg.AuthEnabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if authHeader == "" {
			abortAuth(c, http.StatusUnauthorized, "UNAUTHORIZED", "missing bearer token")
			return
		}

		const prefix = "Bearer "
		if !strings.HasPrefix(authHeader, prefix) {
			abortAuth(c, http.StatusUnauthorized, "UNAUTHORIZED", "invalid authorization header")
			return
		}

		tokenStr := strings.TrimSpace(strings.TrimPrefix(authHeader, prefix))
		if tokenStr == "" {
			abortAuth(c, http.StatusUnauthorized, "UNAUTHORIZED", "missing bearer token")
			return
		}

		claims := &GatewayClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(cfg.JWTSecret), nil
		}, jwt.WithIssuer(cfg.JWTIssuer))
		if err != nil || !token.Valid {
			abortAuth(c, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired token")
			return
		}

		userID := strings.TrimSpace(claims.Subject)
		role := normalizeRole(claims.Role)

		if userID == "" || !isAllowedRole(role) {
			abortAuth(c, http.StatusUnauthorized, "UNAUTHORIZED", "invalid token claims")
			return
		}

		c.Set(CtxKeyUserID, userID)
		c.Set(CtxKeyUserRole, role)

		c.Request.Header.Set("X-User-ID", userID)
		c.Request.Header.Set("X-User-Role", role)

		c.Next()
	}
}

func RequireRoles(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		role = normalizeRole(role)
		if role != "" {
			allowed[role] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		if len(allowed) == 0 {
			c.Next()
			return
		}

		roleValue, exists := c.Get(CtxKeyUserRole)
		if !exists {
			abortAuth(c, http.StatusUnauthorized, "UNAUTHORIZED", "missing authenticated role")
			return
		}

		role, _ := roleValue.(string)
		role = normalizeRole(role)

		if _, ok := allowed[role]; !ok {
			abortAuth(c, http.StatusForbidden, "FORBIDDEN", "insufficient role")
			return
		}

		c.Next()
	}
}

func normalizeRole(role string) string {
	return strings.ToLower(strings.TrimSpace(role))
}

func isAllowedRole(role string) bool {
	switch normalizeRole(role) {
	case "staff", "supervisor":
		return true
	default:
		return false
	}
}

func abortAuth(c *gin.Context, statusCode int, code string, message string) {
	requestID := c.Writer.Header().Get("X-Request-Id")
	c.AbortWithStatusJSON(statusCode, gin.H{
		"error": gin.H{
			"code":       code,
			"message":    message,
			"request_id": requestID,
		},
	})
}