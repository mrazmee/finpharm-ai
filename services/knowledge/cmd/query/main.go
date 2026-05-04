package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"finpharm-ai/services/knowledge/internal/config"
	"finpharm-ai/services/knowledge/internal/retrieval"

	_ "github.com/lib/pq"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})).
		With("service", "knowledge-query")
	slog.SetDefault(logger)

	query := flag.String("q", "", "user query for SOP retrieval")
	limit := flag.Int("k", 5, "number of top chunks to retrieve")
	minScore := flag.Float64("min-score", 0.45, "minimum similarity score to include")
	flag.Parse()

	cfg := config.Load()
	if err := cfg.ValidateForQuery(); err != nil {
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

	embedder := retrieval.NewQueryEmbedder(cfg.GeminiAPIKey, cfg.EmbeddingModel, cfg.EmbeddingOutputDimension)

	ctx := context.Background()
	queryEmbedding, err := embedder.EmbedQuery(ctx, *query)
	if err != nil {
		slog.Error("query_embedding_error", "error", err)
		os.Exit(1)
	}

	store := retrieval.NewStore(db)
	results, err := store.Search(ctx, queryEmbedding, *limit, *minScore)
	if err != nil {
		slog.Error("knowledge_search_error", "error", err)
		os.Exit(1)
	}

	slog.Info("knowledge_search_complete",
		"query", *query,
		"results", len(results),
		"top_k", *limit,
		"min_score", *minScore,
	)

	if len(results) == 0 {
		fmt.Println("No matching SOP chunks found.")
		return
	}

	fmt.Printf("Query: %s\n\n", strings.TrimSpace(*query))

	for i, item := range results {
		fmt.Printf("[%d] score=%.4f | category=%s | source=%s\n", i+1, item.Score, item.Category, item.SourceKey)
		fmt.Printf("title   : %s\n", item.Title)
		fmt.Printf("heading : %s\n", item.Heading)
		fmt.Printf("chunk   : %d\n", item.ChunkIndex)
		fmt.Printf("content : %s\n\n", strings.TrimSpace(item.Content))
	}
}