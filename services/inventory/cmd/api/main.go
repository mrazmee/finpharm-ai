package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"finpharm-ai/services/inventory/internal/config"
	"finpharm-ai/services/inventory/internal/domain"
	"finpharm-ai/services/inventory/internal/httpapi"
	"finpharm-ai/services/inventory/internal/httpapi/handler"
	"finpharm-ai/services/inventory/internal/repository"
	"finpharm-ai/services/inventory/internal/usecase"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})).
		With("service", "inventory")
	slog.SetDefault(logger)

	cfg := config.Load()

	var (
		medicineRepo domain.MedicineRepository
		stockRepo    domain.StockRepository
		cleanup      = func() error { return nil }
	)

	switch cfg.StorageDriver {
	case "", "memory":
		slog.Info("inventory_storage_selected", "storage_driver", "memory")
		medicineRepo = repository.NewMedicineMemoryRepo()
		stockRepo = repository.NewStockMemoryRepo()

	case "postgres":
		db, err := repository.OpenPostgres(cfg.DBConnString())
		if err != nil {
			slog.Error("db_connect_error", "error", err, "db_name", cfg.DBName)
			os.Exit(1)
		}

		cleanup = db.Close

		slog.Info("inventory_storage_selected",
			"storage_driver", "postgres",
			"db_name", cfg.DBName,
			"db_host", cfg.DBHost,
			"db_port", cfg.DBPort,
		)

		medicineRepo = repository.NewMedicineSQLXRepo(db)
		stockRepo = repository.NewStockSQLXRepo(db)

	default:
		slog.Error("invalid_storage_driver", "storage_driver", cfg.StorageDriver)
		os.Exit(1)
	}

	defer func() {
		if err := cleanup(); err != nil {
			slog.Error("resource_cleanup_error", "error", err)
		}
	}()

	stockUC := usecase.NewStockUsecase(stockRepo)
	stockHandler := handler.NewStockHandler(stockUC)

	medUC := usecase.NewMedicineUsecase(medicineRepo)
	medHandler := handler.NewMedicineHandler(medUC)

	router := httpapi.NewRouter(cfg, stockHandler, medHandler)

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
			"storage_driver", cfg.StorageDriver,
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
		slog.Error("server_shutdown_error", "error", err)
		os.Exit(1)
	}

	slog.Info("server_stopped")
}