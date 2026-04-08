package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"finpharm-ai/services/worker/internal/config"
	"finpharm-ai/services/worker/internal/consumer"
	"finpharm-ai/services/worker/internal/observability"
	"finpharm-ai/services/worker/internal/processor"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})).
		With("service", "worker")
	slog.SetDefault(logger)

	cfg := config.Load()

	handler := processor.NewTransactionApprovedProcessor()

	consumerSvc, err := consumer.NewTransactionApprovedConsumer(cfg, handler)
	if err != nil {
		slog.Error("worker_init_failed", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := consumerSvc.Close(); err != nil {
			slog.Warn("worker_close_failed", "error", err)
		}
	}()

	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", observability.MetricsHandler())

	metricsSrv := &http.Server{
		Addr:    ":" + cfg.MetricsPort,
		Handler: metricsMux,
	}

	go func() {
		slog.Info("worker_metrics_server_start",
			"metrics_port", cfg.MetricsPort,
			"metrics_path", "/metrics",
		)
		if err := metricsSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("worker_metrics_server_error", "error", err)
			os.Exit(1)
		}
	}()

	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()

	runErrCh := make(chan error, 1)
	go func() {
		runErrCh <- consumerSvc.Run(workerCtx)
	}()

	slog.Info("worker_start",
		"worker_name", cfg.WorkerName,
		"queue", cfg.QueueName,
		"exchange", cfg.RabbitMQExchange,
		"routing_key", cfg.RoutingKey,
		"consumer_tag", cfg.ConsumerTag,
		"metrics_port", cfg.MetricsPort,
		"shutdown_timeout_ms", int(cfg.ShutdownTimeout.Milliseconds()),
	)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-stop:
		slog.Info("worker_shutdown_signal", "signal", sig.String())
		workerCancel()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer shutdownCancel()

		_ = metricsSrv.Shutdown(shutdownCtx)

		select {
		case err := <-runErrCh:
			if err != nil {
				slog.Error("worker_run_failed", "error", err)
				os.Exit(1)
			}
			slog.Info("worker_stopped")
		case <-shutdownCtx.Done():
			slog.Error("worker_shutdown_timeout",
				"shutdown_timeout_ms", int(cfg.ShutdownTimeout.Milliseconds()),
			)
			os.Exit(1)
		}

	case err := <-runErrCh:
		_ = metricsSrv.Shutdown(context.Background())
		if err != nil {
			slog.Error("worker_run_failed", "error", err)
			os.Exit(1)
		}
		slog.Info("worker_stopped")
	}
}