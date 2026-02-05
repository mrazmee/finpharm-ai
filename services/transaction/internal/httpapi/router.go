package httpapi

import (
	"finpharm-ai/services/transaction/internal/httpapi/handler"

	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	h := handler.NewHealthHandler()

	r.GET("/", h.Hello)
	r.GET("/health", h.Health)

	return r
}
