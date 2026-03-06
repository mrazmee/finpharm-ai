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
	_ = cfg

	r := gin.New()
	r.Use(middleware.RequestID(), middleware.RequestLogger(), gin.Recovery())

	// DI
	stockRepo := repository.NewStockMemoryRepo()
	stockUC := usecase.NewStockUsecase(stockRepo)
	stockHandler := handler.NewStockHandler(stockUC)

	medRepo := repository.NewMedicineMemoryRepo()
	medUC := usecase.NewMedicineUsecase(medRepo)
	medHandler := handler.NewMedicineHandler(medUC)

	health := handler.NewHealthHandler()

	r.GET("/", health.Hello)
	r.GET("/health", health.Health)

	v1 := r.Group("/v1")
	{
		v1.POST("/stock/check", stockHandler.CheckStock)

		v1.GET("/medicines", medHandler.ListMedicines)
		v1.GET("/medicines/:id", medHandler.GetMedicine)
	}

	return r
}