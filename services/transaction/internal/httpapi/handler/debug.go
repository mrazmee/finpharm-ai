package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type DebugHandler struct{}

func NewDebugHandler() *DebugHandler {
	return &DebugHandler{}
}

// GET /v1/debug/sleep?ms=3000
func (h *DebugHandler) Sleep(c *gin.Context) {
	msStr := c.Query("ms")
	if msStr == "" {
		msStr = "1000"
	}

	ms, err := strconv.Atoi(msStr)
	if err != nil || ms < 0 || ms > 60000 {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid ms parameter (0..60000)", msStr)
		return
	}

	time.Sleep(time.Duration(ms) * time.Millisecond)

	c.JSON(http.StatusOK, gin.H{
		"slept_ms": ms,
	})
}
