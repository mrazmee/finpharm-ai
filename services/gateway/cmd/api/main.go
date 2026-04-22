package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"finpharm-ai/internal/telemetry/audithttp"
	"finpharm-ai/internal/telemetry/tracehttp"
	"finpharm-ai/services/gateway/internal/config"
	"finpharm-ai/services/gateway/internal/httpapi"
	gwmiddleware "finpharm-ai/services/gateway/internal/httpapi/middleware"
	"finpharm-ai/services/gateway/internal/observability"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})).
		With("service", "gateway")
	slog.SetDefault(logger)

	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		slog.Error("config_invalid", "error", err)
		os.Exit(1)
	}

	router := httpapi.NewRouter(cfg)

	var appRouter http.Handler = router
	if cfg.RateLimitEnabled {
		generalLimiter := gwmiddleware.NewInMemoryRateLimiter(cfg.RateLimitGeneralLimit, cfg.RateLimitWindow)
		authLimiter := gwmiddleware.NewInMemoryRateLimiter(cfg.RateLimitAuthLimit, cfg.RateLimitWindow)
		appRouter = gwmiddleware.RateLimitHandler(generalLimiter, authLimiter, router)
	}

	baseMux := http.NewServeMux()
	baseMux.Handle("/metrics", observability.MetricsHandler())
	baseMux.Handle("/", appRouter)

	appHandler := tracehttp.Handler("gateway", audithttp.Handler("gateway", baseMux))

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      appHandler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	go func() {
		slog.Info("server_start",
			"port", cfg.Port,
			"inventory_base_url", cfg.InventoryBaseURL,
			"transaction_base_url", cfg.TransactionBaseURL,
			"auth_enabled", cfg.AuthEnabled,
			"jwt_issuer", cfg.JWTIssuer,
			"jwt_expire_minutes", cfg.JWTExpireMinutes,
			"rate_limit_enabled", cfg.RateLimitEnabled,
			"rate_limit_general_limit", cfg.RateLimitGeneralLimit,
			"rate_limit_auth_limit", cfg.RateLimitAuthLimit,
			"rate_limit_window_seconds", int(cfg.RateLimitWindow.Seconds()),
			"metrics_path", "/metrics",
			"trace_header", tracehttp.HeaderTraceID,
			"read_timeout_ms", int(cfg.ReadTimeout.Milliseconds()),
			"write_timeout_ms", int(cfg.WriteTimeout.Milliseconds()),
			"idle_timeout_ms", int(cfg.IdleTimeout.Milliseconds()),
		)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server_error", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	slog.Info("server_shutdown_signal")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server_shutdown_error",
			"error", err,
			"shutdown_timeout_ms", int(cfg.ShutdownTimeout.Milliseconds()),
		)
		os.Exit(1)
	}

	slog.Info("server_stopped")
}