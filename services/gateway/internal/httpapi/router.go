package httpapi

import (
	"finpharm-ai/services/gateway/internal/httpapi/handler"

	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	h := handler.NewHealthHandler()
	proxy := handler.NewStockProxyHandler()

	r.GET("/", h.Hello)
	r.GET("/health", h.Health)

	v1 := r.Group("/v1")
	{
		v1.POST("/stock/check", proxy.CheckStock)
	}

	return r
}
