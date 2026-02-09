package handler

import (
	"finpharm-ai/services/gateway/internal/httpapi/middleware"

	"github.com/gin-gonic/gin"
)

type ErrorBody struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Details   any    `json:"details,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

func RespondError(c *gin.Context, status int, code, message string, details any) {
	ridVal, _ := c.Get(middleware.CtxKeyRequestID)
	rid, _ := ridVal.(string)

	c.JSON(status, ErrorResponse{
		Error: ErrorBody{
			Code:      code,
			Message:   message,
			Details:   details,
			RequestID: rid,
		},
	})
}
