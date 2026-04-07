package handler

import (
	"net/http"
	"strings"

	"finpharm-ai/services/gateway/internal/config"
	"finpharm-ai/services/gateway/internal/httpapi/middleware"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	cfg config.Config
}

func NewAuthHandler(cfg config.Config) *AuthHandler {
	return &AuthHandler{cfg: cfg}
}

type IssueTokenRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

type IssueTokenResponse struct {
	AccessToken string            `json:"access_token"`
	TokenType   string            `json:"token_type"`
	ExpiresAt   string            `json:"expires_at"`
	User        AuthenticatedUser `json:"user"`
}

type AuthenticatedUser struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

func (h *AuthHandler) IssueToken(c *gin.Context) {
	if !h.cfg.IsDebugEnabled() {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":       "NOT_FOUND",
				"message":    "route not found",
				"request_id": c.Writer.Header().Get("X-Request-Id"),
			},
		})
		return
	}

	var req IssueTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":       "VALIDATION_ERROR",
				"message":    "invalid request body",
				"request_id": c.Writer.Header().Get("X-Request-Id"),
			},
		})
		return
	}

	req.UserID = strings.TrimSpace(req.UserID)
	req.Role = strings.ToLower(strings.TrimSpace(req.Role))

	token, expiresAt, err := middleware.IssueToken(h.cfg, req.UserID, req.Role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":       "VALIDATION_ERROR",
				"message":    err.Error(),
				"request_id": c.Writer.Header().Get("X-Request-Id"),
			},
		})
		return
	}

	user := AuthenticatedUser(req)

	c.JSON(http.StatusOK, gin.H{
		"data": IssueTokenResponse{
			AccessToken: token,
			TokenType:   "Bearer",
			ExpiresAt:   expiresAt.UTC().Format("2006-01-02T15:04:05Z"),
			User:        user,
		},
		"request_id": c.Writer.Header().Get("X-Request-Id"),
	})
}