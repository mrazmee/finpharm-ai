package httpapi

import (
	"finpharm-ai/services/transaction/internal/config"
	"finpharm-ai/services/transaction/internal/httpapi/handler"
	"finpharm-ai/services/transaction/internal/httpapi/middleware"
	"finpharm-ai/services/transaction/internal/observability"

	"github.com/gin-gonic/gin"
)

func NewRouter(cfg config.Config, stockHandler *handler.StockHandler, transactionHandler *handler.TransactionHandler) *gin.Engine {
	r := gin.New()
	r.Use(middleware.RequestID(), middleware.RequestLogger(), observability.Middleware("transaction"), gin.Recovery())

	health := handler.NewHealthHandler()

	r.GET("/", health.Hello)
	r.GET("/health", health.Health)

	v1 := r.Group("/v1")
	{
		if stockHandler != nil {
			v1.POST("/stock/check", stockHandler.CheckStock)
		}
		if transactionHandler != nil {
			v1.POST("/transactions", transactionHandler.CreateTransaction)
			v1.GET("/transactions", transactionHandler.ListTransactions)
		}

		if cfg.IsDebugEnabled() {
			debug := handler.NewDebugHandler()
			v1.GET("/debug/sleep", debug.Sleep)
		}
	}

	return r
}