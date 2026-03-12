package handler

import (
	"finpharm-ai/services/inventory/internal/httpapi/middleware"

	"github.com/gin-gonic/gin"
)

type ErrorBody struct {
	Code      string      `json:"code"`
	Message   string      `json:"message"`
	Details   interface{} `json:"details,omitempty"`
	RequestID string      `json:"request_id"`
}

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

func RespondError(c *gin.Context, status int, code, message string, details interface{}) {
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

type SuccessResponse struct {
	Data      interface{} `json:"data"`
	RequestID string      `json:"request_id"`
}

func RespondOK(c *gin.Context, status int, data interface{}) {
	ridVal, _ := c.Get(middleware.CtxKeyRequestID)
	rid, _ := ridVal.(string)

	c.JSON(status, SuccessResponse{
		Data:      data,
		RequestID: rid,
	})
}
