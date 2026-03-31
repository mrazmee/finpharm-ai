package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"finpharm-ai/services/ai-auditor/internal/config"
	"finpharm-ai/services/ai-auditor/internal/domain"
	"finpharm-ai/services/ai-auditor/internal/httpapi"
	"finpharm-ai/services/ai-auditor/internal/httpapi/handler"
	"finpharm-ai/services/ai-auditor/internal/provider"
	"finpharm-ai/services/ai-auditor/internal/usecase"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})).
		With("service", "ai-auditor")
	slog.SetDefault(logger)

	cfg := config.Load()

	var primaryProvider domain.AuditProvider
	fallbackProvider := provider.NewSafeFallbackProvider(cfg.AuditFailOpen)

	switch strings.ToLower(strings.TrimSpace(cfg.AuditProvider)) {
	case "mock":
		primaryProvider = provider.NewRuleBasedProvider()
		slog.Info("audit_provider_selected",
			"provider", "mock",
			"model", "rule-based-v1",
		)
	case "", "gemini":
		if strings.TrimSpace(cfg.GeminiAPIKey) == "" {
			slog.Warn("audit_provider_fallback_startup",
				"reason", "gemini api key missing",
				"provider", "fallback",
				"model", "safe-review-v1",
			)
		} else {
			primaryProvider = provider.NewGeminiProvider(cfg.GeminiAPIKey, cfg.GeminiModel, cfg.GeminiTimeout)
			slog.Info("audit_provider_selected",
				"provider", "gemini",
				"model", cfg.GeminiModel,
				"timeout_ms", int(cfg.GeminiTimeout.Milliseconds()),
			)
		}
	default:
		slog.Warn("audit_provider_unknown",
			"configured_provider", cfg.AuditProvider,
			"provider", "fallback",
			"model", "safe-review-v1",
		)
	}

	auditUC := usecase.NewAuditUsecase(primaryProvider, fallbackProvider)
	auditHandler := handler.NewAuditHandler(auditUC)
	router := httpapi.NewRouter(cfg, auditHandler)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	go func() {
		slog.Info("server_start",
			"port", cfg.Port,
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