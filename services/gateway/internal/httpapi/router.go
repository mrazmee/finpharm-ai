package httpapi

import (
	"finpharm-ai/services/gateway/internal/config"
	"finpharm-ai/services/gateway/internal/httpapi/handler"
	"finpharm-ai/services/gateway/internal/httpapi/middleware"

	"github.com/gin-gonic/gin"
)

func NewRouter(cfg config.Config) *gin.Engine {
	r := gin.New()
	r.Use(middleware.RequestID(), middleware.RequestLogger(), gin.Recovery())

	h := handler.NewHealthHandler()
	proxy := handler.NewStockProxyHandler(cfg.TransactionBaseURL)
	debugProxy := handler.NewDebugProxyHandler(cfg.TransactionBaseURL)

	r.GET("/", h.Hello)
	r.GET("/health", h.Health)

	v1 := r.Group("/v1")
	{
		v1.POST("/stock/check", proxy.CheckStock)
		v1.GET("/debug/sleep", debugProxy.Sleep)
	}

	return r
}
