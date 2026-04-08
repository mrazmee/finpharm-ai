package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"finpharm-ai/services/transaction/internal/config"
	"finpharm-ai/services/transaction/internal/httpapi"
	"finpharm-ai/services/transaction/internal/httpapi/handler"
	"finpharm-ai/services/transaction/internal/observability"
	"finpharm-ai/services/transaction/internal/repository"
	"finpharm-ai/services/transaction/internal/usecase"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})).
		With("service", "transaction")
	slog.SetDefault(logger)

	cfg := config.Load()

	db, err := sqlx.Connect("postgres", cfg.DSN())
	if err != nil {
		slog.Error("db_connect_error", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	txRepo := repository.NewTransactionSQLXRepo(db)

	stockHTTPClient := &http.Client{Timeout: 4 * time.Second}
	stockBreaker := repository.NewCircuitBreaker(3, 5*time.Second)
	stockRepo := repository.NewStockHTTPRepo(cfg.InventoryBaseURL, stockHTTPClient, stockBreaker)

	aiAuditorClient := &http.Client{Timeout: cfg.AIAuditorTimeout + time.Second}
	aiAuditorRepo := repository.NewAIAuditorHTTPRepo(cfg.AIAuditorBaseURL, aiAuditorClient, cfg.AIAuditorTimeout)

	var eventPublisher *repository.RabbitMQTransactionEventPublisher
	if cfg.RabbitMQURL != "" {
		publisher, err := repository.NewRabbitMQTransactionEventPublisher(
			cfg.RabbitMQURL,
			cfg.RabbitMQExchange,
			cfg.RabbitMQTransactionApprovedQueue,
			cfg.RabbitMQTransactionApprovedRouting,
		)
		if err != nil {
			slog.Warn("rabbitmq_publisher_init_failed",
				"error", err,
				"rabbitmq_url", cfg.RabbitMQURL,
			)
		} else {
			eventPublisher = publisher
			defer func() {
				if err := eventPublisher.Close(); err != nil {
					slog.Warn("rabbitmq_publisher_close_failed", "error", err)
				}
			}()
		}
	}

	stockUC := usecase.NewStockUsecase(stockRepo)
	txUC := usecase.NewTransactionUsecase(txRepo, stockRepo, aiAuditorRepo, eventPublisher)

	stockHandler := handler.NewStockHandler(stockUC)
	txHandler := handler.NewTransactionHandler(txUC)

	router := httpapi.NewRouter(cfg, stockHandler, txHandler)

	mux := http.NewServeMux()
	mux.Handle("/metrics", observability.MetricsHandler())
	mux.Handle("/", observability.InstrumentHandler("transaction", router))

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	go func() {
		slog.Info("server_start",
			"port", cfg.Port,
			"inventory_base_url", cfg.InventoryBaseURL,
			"ai_auditor_base_url", cfg.AIAuditorBaseURL,
			"ai_auditor_timeout_ms", int(cfg.AIAuditorTimeout.Milliseconds()),
			"rabbitmq_exchange", cfg.RabbitMQExchange,
			"rabbitmq_transaction_approved_queue", cfg.RabbitMQTransactionApprovedQueue,
			"rabbitmq_transaction_approved_routing_key", cfg.RabbitMQTransactionApprovedRouting,
			"metrics_path", "/metrics",
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