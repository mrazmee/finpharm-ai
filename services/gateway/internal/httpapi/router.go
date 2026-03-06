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
	txProxy := handler.NewStockProxyHandler(cfg.TransactionBaseURL)
	invProxy := handler.NewInventoryProxyHandler(cfg.InventoryBaseURL)

	r.GET("/", h.Hello)
	r.GET("/health", h.Health)

	v1 := r.Group("/v1")
	{
		// Transaction routes
		v1.POST("/stock/check", txProxy.CheckStock)

		// Inventory routes (NEW)
		v1.GET("/medicines", invProxy.ListMedicines)
		v1.GET("/medicines/:id", invProxy.GetMedicine)

		// Debug route (local/dev only)
		if cfg.IsDebugEnabled() {
			debugProxy := handler.NewDebugProxyHandler(cfg.TransactionBaseURL)
			v1.GET("/debug/sleep", debugProxy.Sleep)
		}
	}

	return r
}