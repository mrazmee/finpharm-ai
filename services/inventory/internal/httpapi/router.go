package httpapi

import (
	"finpharm-ai/services/inventory/internal/config"
	"finpharm-ai/services/inventory/internal/httpapi/handler"
	"finpharm-ai/services/inventory/internal/httpapi/middleware"
	"finpharm-ai/services/inventory/internal/observability"

	"github.com/gin-gonic/gin"
)

func NewRouter(
	cfg config.Config,
	stockHandler *handler.StockHandler,
	medHandler *handler.MedicineHandler,
) *gin.Engine {
	_ = cfg

	r := gin.New()
	r.Use(middleware.RequestID(), middleware.RequestLogger(), observability.Middleware("inventory"), gin.Recovery())

	health := handler.NewHealthHandler()

	r.GET("/", health.Hello)
	r.GET("/health", health.Health)

	v1 := r.Group("/v1")
	{
		v1.POST("/stock/check", stockHandler.CheckStock)
		v1.POST("/stock/deduct", stockHandler.DeductStock)

		v1.GET("/medicines", medHandler.ListMedicines)
		v1.GET("/medicines/:id", medHandler.GetMedicine)
	}

	return r
}