package httpapi

import (
	"net/http"
	"time"

	"finpharm-ai/services/transaction/internal/config"
	"finpharm-ai/services/transaction/internal/httpapi/handler"
	"finpharm-ai/services/transaction/internal/httpapi/middleware"
	"finpharm-ai/services/transaction/internal/repository"
	"finpharm-ai/services/transaction/internal/usecase"

	"github.com/gin-gonic/gin"
)

func NewRouter(cfg config.Config) *gin.Engine {
	r := gin.New()
	r.Use(middleware.RequestID(), middleware.RequestLogger(), gin.Recovery())

	// shared dependencies
	httpClient := &http.Client{Timeout: 4 * time.Second}
	breaker := repository.NewCircuitBreaker(3, 5*time.Second)
	stockRepo := repository.NewStockHTTPRepo(cfg.InventoryBaseURL, httpClient, breaker)
	stockUC := usecase.NewStockUsecase(stockRepo)
	stockHandler := handler.NewStockHandler(stockUC)

	health := handler.NewHealthHandler()

	r.GET("/", health.Hello)
	r.GET("/health", health.Health)

	v1 := r.Group("/v1")
	{
		v1.POST("/stock/check", stockHandler.CheckStock)

		if cfg.IsDebugEnabled() {
			debug := handler.NewDebugHandler()
			v1.GET("/debug/sleep", debug.Sleep)
		}
	}

	return r
}