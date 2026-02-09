package main

import (
	"log/slog"
	"os"

	"finpharm-ai/services/transaction/internal/httpapi"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})).With("service", "transaction")
	slog.SetDefault(logger)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	router := httpapi.NewRouter()

	if err := router.Run(":" + port); err != nil {
		slog.Error("failed to start server", "error", err)
		os.Exit(1)
	}
}
