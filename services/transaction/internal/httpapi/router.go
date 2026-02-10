package httpapi

import (
	"finpharm-ai/services/transaction/internal/config"
	"finpharm-ai/services/transaction/internal/httpapi/handler"
	"finpharm-ai/services/transaction/internal/httpapi/middleware"

	"github.com/gin-gonic/gin"
)

func NewRouter(cfg config.Config) *gin.Engine {
	_ = cfg // belum dipakai di router, tapi sengaja disiapkan untuk Day berikutnya

	r := gin.New()
	r.Use(middleware.RequestID(), middleware.RequestLogger(), gin.Recovery())

	h := handler.NewHealthHandler()
	stock := handler.NewStockHandler()

	r.GET("/", h.Hello)
	r.GET("/health", h.Health)

	v1 := r.Group("/v1")
	{
		v1.POST("/stock/check", stock.CheckStock)
	}

	return r
}
