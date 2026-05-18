package httpapi

import (
	"finpharm-ai/services/knowledge/internal/httpapi/handler"
	"finpharm-ai/services/knowledge/internal/httpapi/middleware"
	"finpharm-ai/services/knowledge/internal/observability"

	"github.com/gin-gonic/gin"
)

func NewRouter(chatHandler *handler.ChatHandler) *gin.Engine {
	r := gin.New()
	r.Use(
		middleware.RequestID(),
		middleware.RequestLogger(),
		observability.Middleware("knowledge"),
		gin.Recovery(),
	)

	health := handler.NewHealthHandler()

	r.GET("/", health.Hello)
	r.GET("/health", health.Health)

	v1 := r.Group("/v1")
	{
		chatGroup := v1.Group("/chat")
		chatGroup.POST("/sop", chatHandler.HandleSOPChat)
	}

	return r
}