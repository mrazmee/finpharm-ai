package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Hello(c *gin.Context) {
	RespondOK(c, http.StatusOK, gin.H{
		"message": "knowledge service",
	})
}

func (h *HealthHandler) Health(c *gin.Context) {
	RespondOK(c, http.StatusOK, gin.H{
		"status":  "ok",
		"service": "knowledge",
	})
}