package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"finpharm-ai/services/worker/internal/config"
	"finpharm-ai/services/worker/internal/consumer"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		slog.Info("worker_shutdown_signal")
		cancel()
	}()

	slog.Info("worker_start",
		"worker_name", cfg.WorkerName,
		"queue", cfg.QueueName,
		"exchange", cfg.RabbitMQExchange,
		"routing_key", cfg.RoutingKey,
		"consumer_tag", cfg.ConsumerTag,
	)

	if err := consumerSvc.Run(ctx); err != nil {
		slog.Error("worker_run_failed", "error", err)
		os.Exit(1)
	}

	slog.Info("worker_stopped")
}