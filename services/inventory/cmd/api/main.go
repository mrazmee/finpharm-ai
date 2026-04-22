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
	"finpharm-ai/services/inventory/internal/config"
	"finpharm-ai/services/inventory/internal/domain"
	"finpharm-ai/services/inventory/internal/httpapi"
	"finpharm-ai/services/inventory/internal/httpapi/handler"
	"finpharm-ai/services/inventory/internal/observability"
	"finpharm-ai/services/inventory/internal/repository"
	"finpharm-ai/services/inventory/internal/usecase"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})).
		With("service", "inventory")
	slog.SetDefault(logger)

	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		slog.Error("config_invalid", "error", err)
		os.Exit(1)
	}

	var medicineRepo domain.MedicineRepository
	var stockRepo domain.StockRepository
	var closeDB func()

	switch strings.ToLower(strings.TrimSpace(cfg.StorageDriver)) {
	case "memory":
		medicineRepo = repository.NewMedicineMemoryRepo()
		stockRepo = repository.NewStockMemoryRepo()
		closeDB = func() {}
	default:
		db, err := repository.OpenPostgres(cfg.DBConnString())
		if err != nil {
			slog.Error("db_connect_error", "error", err)
			os.Exit(1)
		}

		closeDB = func() {
			_ = db.Close()
		}

		medicineRepo = repository.NewMedicineSQLXRepo(db)
		stockRepo = repository.NewStockSQLXRepo(db)
	}

	defer closeDB()

	stockUC := usecase.NewStockUsecase(stockRepo)
	medicineUC := usecase.NewMedicineUsecase(medicineRepo)

	stockHandler := handler.NewStockHandler(stockUC)
	medicineHandler := handler.NewMedicineHandler(medicineUC)

	router := httpapi.NewRouter(cfg, stockHandler, medicineHandler)

	baseMux := http.NewServeMux()
	baseMux.Handle("/metrics", observability.MetricsHandler())
	baseMux.Handle("/", router)

	appHandler := tracehttp.Handler("inventory", audithttp.Handler("inventory", baseMux))

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
			"storage_driver", cfg.StorageDriver,
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