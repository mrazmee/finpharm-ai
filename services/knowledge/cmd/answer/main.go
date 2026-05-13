package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"finpharm-ai/services/knowledge/internal/chat"
	"finpharm-ai/services/knowledge/internal/config"

	_ "github.com/lib/pq"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})).
		With("service", "knowledge-answer")
	slog.SetDefault(logger)

	query := flag.String("q", "", "user question for SOP answering")
	limit := flag.Int("k", 5, "number of top chunks to retrieve")
	minScore := flag.Float64("min-score", 0.45, "minimum similarity score to include")
	flag.Parse()

	cfg := config.Load()
	if err := cfg.ValidateForAnswer(); err != nil {
		slog.Error("config_invalid", "error", err)
		os.Exit(1)
	}

	if strings.TrimSpace(*query) == "" {
		slog.Error("query_invalid", "error", "flag -q is required")
		os.Exit(1)
	}
	if *limit <= 0 {
		slog.Error("limit_invalid", "error", "flag -k must be > 0")
		os.Exit(1)
	}
	if *minScore < 0 || *minScore > 1 {
		slog.Error("min_score_invalid", "error", "flag -min-score must be between 0 and 1")
		os.Exit(1)
	}

	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		slog.Error("db_open_error", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		slog.Error("db_ping_error", "error", err)
		os.Exit(1)
	}

	service := chat.NewService(db, cfg)

	result, err := service.Answer(context.Background(), chat.Request{
		Question: *query,
		TopK:     *limit,
		MinScore: *minScore,
	})
	if err != nil {
		slog.Error("knowledge_answer_error", "error", err)
		os.Exit(1)
	}

	fmt.Printf("Question: %s\n\n", result.Question)
	fmt.Println("Answer:")
	fmt.Println(result.Answer)

	if result.Fallback {
		return
	}

	fmt.Println()
	fmt.Println("Sources:")
	for _, item := range result.Sources {
		fmt.Printf("%s %s | %s | %s | score=%.4f\n",
			item.Ref,
			item.Title,
			item.Heading,
			item.SourceKey,
			item.Score,
		)
	}
}