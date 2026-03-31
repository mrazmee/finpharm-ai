package handler

import "github.com/gin-gonic/gin"

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler { return &HealthHandler{} }

func (h *HealthHandler) Hello(c *gin.Context) {
	c.String(200, "hello ai-auditor service")
}

func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(200, gin.H{"service": "ai-auditor", "status": "ok"})
}