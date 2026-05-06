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
	"finpharm-ai/services/knowledge/internal/synthesis"

	_ "github.com/lib/pq"
)

const insufficientContextMessage = "Saya belum menemukan dasar SOP yang cukup untuk menjawab pertanyaan ini."

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

	ctx := context.Background()

	queryEmbedder := retrieval.NewQueryEmbedder(cfg.GeminiAPIKey, cfg.EmbeddingModel, cfg.EmbeddingOutputDimension)
	queryEmbedding, err := queryEmbedder.EmbedQuery(ctx, *query)
	if err != nil {
		slog.Error("query_embedding_error", "error", err)
		os.Exit(1)
	}

	searchStore := retrieval.NewStore(db)
	results, err := searchStore.Search(ctx, queryEmbedding, *limit, *minScore)
	if err != nil {
		slog.Error("knowledge_search_error", "error", err)
		os.Exit(1)
	}

	slog.Info("knowledge_answer_retrieval_complete",
		"query", *query,
		"results", len(results),
		"top_k", *limit,
		"min_score", *minScore,
	)

	if len(results) == 0 {
		fmt.Printf("Question: %s\n\n", strings.TrimSpace(*query))
		fmt.Println("Answer:")
		fmt.Println(insufficientContextMessage)
		return
	}

	snippets := make([]synthesis.SourceSnippet, 0, len(results))
	for i, item := range results {
		ref := fmt.Sprintf("[S%d]", i+1)
		snippets = append(snippets, synthesis.SourceSnippet{
			Ref:       ref,
			Title:     item.Title,
			Category:  item.Category,
			SourceKey: item.SourceKey,
			Heading:   item.Heading,
			Content:   item.Content,
			Score:     item.Score,
		})
	}

	prompt := synthesis.BuildGroundedAnswerPrompt(*query, snippets)

	generator := synthesis.NewGenerator(
		cfg.GeminiAPIKey,
		cfg.AnswerModel,
		cfg.AnswerTemperature,
		cfg.AnswerMaxOutputTokens,
	)

	answer, err := generator.Generate(ctx, prompt)
	if err != nil {
		slog.Error("answer_generation_error", "error", err)
		os.Exit(1)
	}

	slog.Info("knowledge_answer_complete",
		"query", *query,
		"answer_chars", len(answer),
	)

	fmt.Printf("Question: %s\n\n", strings.TrimSpace(*query))
	fmt.Println("Answer:")
	fmt.Println(strings.TrimSpace(answer))
	fmt.Println()
	fmt.Println("Sources:")
	for i, item := range results {
		fmt.Printf("[S%d] %s | %s | %s | score=%.4f\n",
			i+1,
			item.Title,
			item.Heading,
			item.SourceKey,
			item.Score,
		)
	}
}