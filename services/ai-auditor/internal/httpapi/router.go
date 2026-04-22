package httpapi

import (
	"finpharm-ai/services/ai-auditor/internal/config"
	"finpharm-ai/services/ai-auditor/internal/httpapi/handler"
	"finpharm-ai/services/ai-auditor/internal/httpapi/middleware"
	"finpharm-ai/services/ai-auditor/internal/observability"

	"github.com/gin-gonic/gin"
)

func NewRouter(cfg config.Config, auditHandler *handler.AuditHandler) *gin.Engine {
	_ = cfg

	r := gin.New()
	r.Use(middleware.RequestID(), middleware.RequestLogger(), observability.Middleware("ai-auditor"), gin.Recovery())

	health := handler.NewHealthHandler()

	r.GET("/", health.Hello)
	r.GET("/health", health.Health)

	v1 := r.Group("/v1")
	{
		if auditHandler != nil {
			v1.POST("/audit/transaction", auditHandler.AuditTransaction)
		}
	}

	return r
}