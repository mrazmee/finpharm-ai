package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"finpharm-ai/internal/telemetry/audithttp"
	"finpharm-ai/internal/telemetry/tracehttp"
	"finpharm-ai/services/ai-auditor/internal/config"
	"finpharm-ai/services/ai-auditor/internal/domain"
	"finpharm-ai/services/ai-auditor/internal/httpapi"
	"finpharm-ai/services/ai-auditor/internal/httpapi/handler"
	"finpharm-ai/services/ai-auditor/internal/observability"
	"finpharm-ai/services/ai-auditor/internal/provider"
	"finpharm-ai/services/ai-auditor/internal/usecase"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})).
		With("service", "ai-auditor")
	slog.SetDefault(logger)

	cfg := config.Load()

	fallbackProvider := provider.NewSafeFallbackProvider(cfg.AuditFailOpen)

	var primaryProvider domain.AuditProvider = fallbackProvider

	switch strings.ToLower(strings.TrimSpace(cfg.AuditProvider)) {
	case "gemini":
		primaryProvider = provider.NewGeminiProvider(
			cfg.GeminiAPIKey,
			cfg.GeminiModel,
			cfg.GeminiTimeout,
		)
	case "mock", "rule-based", "rule_based":
		primaryProvider = provider.NewRuleBasedProvider()
	case "fallback":
		primaryProvider = fallbackProvider
	default:
		primaryProvider = fallbackProvider
	}

	auditUC := usecase.NewAuditUsecase(primaryProvider, fallbackProvider)
	auditHandler := handler.NewAuditHandler(auditUC)

	router := httpapi.NewRouter(cfg, auditHandler)

	baseMux := http.NewServeMux()
	baseMux.Handle("/metrics", observability.MetricsHandler())
	baseMux.Handle("/", observability.InstrumentHandler("ai-auditor", router))

	appHandler := tracehttp.Handler("ai-auditor", audithttp.Handler("ai-auditor", baseMux))

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
			"audit_provider", cfg.AuditProvider,
			"gemini_model", cfg.GeminiModel,
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
		slog.Error("server_shutdown_error", "error", err, "shutdown_timeout_ms", int(cfg.ShutdownTimeout.Milliseconds()))
		os.Exit(1)
	}

	slog.Info("server_stopped")
}