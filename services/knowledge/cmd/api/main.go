package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"finpharm-ai/internal/telemetry/audithttp"
	"finpharm-ai/internal/telemetry/tracehttp"
	"finpharm-ai/services/knowledge/internal/chat"
	"finpharm-ai/services/knowledge/internal/config"
	"finpharm-ai/services/knowledge/internal/httpapi"
	"finpharm-ai/services/knowledge/internal/httpapi/handler"
	"finpharm-ai/services/knowledge/internal/observability"

	_ "github.com/lib/pq"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})).
		With("service", "knowledge")
	slog.SetDefault(logger)

	cfg := config.Load()
	if err := cfg.ValidateForAPI(); err != nil {
		slog.Error("config_invalid", "error", err)
		os.Exit(1)
	}

	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		slog.Error("db_connect_error", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		slog.Error("db_ping_error", "error", err)
		os.Exit(1)
	}

	chatService := chat.NewService(db, cfg)
	chatHandler := handler.NewChatHandler(chatService)

	router := httpapi.NewRouter(chatHandler)

	baseMux := http.NewServeMux()
	baseMux.Handle("/metrics", observability.MetricsHandler())
	baseMux.Handle("/", router)

	appHandler := tracehttp.Handler("knowledge", audithttp.Handler("knowledge", baseMux))

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
			"metrics_path", "/metrics",
			"trace_header", tracehttp.HeaderTraceID,
			"embedding_model", cfg.EmbeddingModel,
			"answer_model", cfg.AnswerModel,
			"answer_min_top_score", cfg.AnswerMinTopScore,
			"answer_score_window", cfg.AnswerScoreWindow,
			"answer_max_chunks_per_document", cfg.AnswerMaxChunksPerDocument,
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