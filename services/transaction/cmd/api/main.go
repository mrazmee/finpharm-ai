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
	"finpharm-ai/services/transaction/internal/repository"
	"finpharm-ai/services/transaction/internal/usecase"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})).
		With("service", "transaction")
	slog.SetDefault(logger)

	cfg := config.Load()

	httpClient := &http.Client{Timeout: 4 * time.Second}
	breaker := repository.NewCircuitBreaker(3, 5*time.Second)
	stockRepo := repository.NewStockHTTPRepo(cfg.InventoryBaseURL, httpClient, breaker)
	stockUC := usecase.NewStockUsecase(stockRepo)
	stockHandler := handler.NewStockHandler(stockUC)

	db, err := repository.OpenPostgres(cfg.DBConnString())
	if err != nil {
		slog.Error("db_connect_error", "error", err, "db_name", cfg.DBName)
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("resource_cleanup_error", "error", err)
		}
	}()

	slog.Info("transaction_persistence_selected",
		"driver", "postgres",
		"db_name", cfg.DBName,
		"db_host", cfg.DBHost,
		"db_port", cfg.DBPort,
	)

	transactionRepo := repository.NewTransactionSQLXRepo(db)
	transactionUC := usecase.NewTransactionUsecase(transactionRepo, stockRepo)
	transactionHandler := handler.NewTransactionHandler(transactionUC)

	router := httpapi.NewRouter(cfg, stockHandler, transactionHandler)

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
