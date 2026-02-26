package httpapi

import (
	"finpharm-ai/services/inventory/internal/config"
	"finpharm-ai/services/inventory/internal/httpapi/handler"
	"finpharm-ai/services/inventory/internal/httpapi/middleware"
	"finpharm-ai/services/inventory/internal/repository"
	"finpharm-ai/services/inventory/internal/usecase"

	"github.com/gin-gonic/gin"
)

func NewRouter(cfg config.Config) *gin.Engine {
	r := gin.New()
	r.Use(middleware.RequestID(), middleware.RequestLogger(), gin.Recovery())

	// DI
	repo := repository.NewStockMemoryRepo()
	uc := usecase.NewStockUsecase(repo)
	stock := handler.NewStockHandler(uc)
	health := handler.NewHealthHandler()

	r.GET("/", health.Hello)
	r.GET("/health", health.Health)

	v1 := r.Group("/v1")
	{
		v1.POST("/stock/check", stock.CheckStock)
	}

	return r
}