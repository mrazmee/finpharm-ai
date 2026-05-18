package httpapi

import (
	"net/http"
	"strconv"
	"time"

	"finpharm-ai/services/gateway/internal/config"
	"finpharm-ai/services/gateway/internal/httpapi/handler"
	"finpharm-ai/services/gateway/internal/httpapi/middleware"
	"finpharm-ai/services/gateway/internal/observability"

	"github.com/gin-gonic/gin"
)

func NewRouter(cfg config.Config) *gin.Engine {
	if !cfg.IsDebugEnabled() {
		gin.SetMode(gin.ReleaseMode)
	}

	inventoryHandler := handler.NewInventoryProxyHandler(cfg.InventoryBaseURL)
	stockHandler := handler.NewStockProxyHandler(cfg.TransactionBaseURL)
	transactionHandler := handler.NewTransactionProxyHandler(cfg.TransactionBaseURL)
	knowledgeHandler := handler.NewKnowledgeProxyHandler(cfg.KnowledgeBaseURL)
	authHandler := handler.NewAuthHandler(cfg)

	router := gin.New()

	router.Use(
		middleware.RequestID(),
		gin.Logger(),
		observability.Middleware("gateway"),
		gin.Recovery(),
	)

	if cfg.RateLimitEnabled {
		generalLimiter := middleware.NewInMemoryRateLimiter(cfg.RateLimitGeneralLimit, cfg.RateLimitWindow)
		authLimiter := middleware.NewInMemoryRateLimiter(cfg.RateLimitAuthLimit, cfg.RateLimitWindow)
		router.Use(middleware.RateLimitGin(generalLimiter, authLimiter))
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":     "ok",
			"service":    "gateway",
			"request_id": c.Writer.Header().Get("X-Request-Id"),
		})
	})

	v1 := router.Group("/v1")

	if cfg.IsDebugEnabled() {
		v1.POST("/auth/token", authHandler.IssueToken)
	}

	if cfg.AuthEnabled {
		protected := v1.Group("")
		protected.Use(middleware.JWTAuth(cfg))

		protected.GET("/medicines", middleware.RequireRoles("staff", "supervisor"), inventoryHandler.ListMedicines)
		protected.GET("/medicines/:id", middleware.RequireRoles("staff", "supervisor"), inventoryHandler.GetMedicine)
		protected.POST("/stock/check", middleware.RequireRoles("staff", "supervisor"), stockHandler.CheckStock)
		protected.POST("/transactions", middleware.RequireRoles("staff", "supervisor"), transactionHandler.CreateTransaction)
		protected.GET("/transactions", middleware.RequireRoles("supervisor"), transactionHandler.ListTransactions)
		protected.POST("/chat/sop", middleware.RequireRoles("staff", "supervisor"), knowledgeHandler.ChatSOP)

		if cfg.IsDebugEnabled() {
			protected.GET("/debug/sleep", middleware.RequireRoles("supervisor"), debugSleepHandler())
		}
	} else {
		v1.GET("/medicines", inventoryHandler.ListMedicines)
		v1.GET("/medicines/:id", inventoryHandler.GetMedicine)
		v1.POST("/stock/check", stockHandler.CheckStock)
		v1.POST("/transactions", transactionHandler.CreateTransaction)
		v1.GET("/transactions", transactionHandler.ListTransactions)
		v1.POST("/chat/sop", knowledgeHandler.ChatSOP)

		if cfg.IsDebugEnabled() {
			v1.GET("/debug/sleep", debugSleepHandler())
		}
	}

	return router
}

func debugSleepHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		sleepMs := 1

		if raw := c.Query("ms"); raw != "" {
			if parsed, err := strconv.Atoi(raw); err == nil && parsed >= 0 && parsed <= 5000 {
				sleepMs = parsed
			}
		}

		time.Sleep(time.Duration(sleepMs) * time.Millisecond)

		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"slept_ms": sleepMs,
			},
			"request_id": c.Writer.Header().Get("X-Request-Id"),
		})
	}
}