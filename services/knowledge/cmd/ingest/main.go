package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"

	"finpharm-ai/services/knowledge/internal/config"
	"finpharm-ai/services/knowledge/internal/ingest"

	_ "github.com/lib/pq"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})).
		With("service", "knowledge-ingest")
	slog.SetDefault(logger)

	cfg := config.Load()
	if err := cfg.ValidateForIngest(); err != nil {
		slog.Error("config_invalid", "error", err)
		os.Exit(1)
	}

	docs, err := ingest.LoadMarkdownDocuments(cfg.SourceDir)
	if err != nil {
		slog.Error("load_documents_error", "error", err)
		os.Exit(1)
	}
	if len(docs) == 0 {
		slog.Warn("no_documents_found", "source_dir", cfg.SourceDir)
		return
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

	store := ingest.NewStore(db)
	embedder := ingest.NewGeminiBatchEmbedder(cfg.GeminiAPIKey, cfg.EmbeddingModel, cfg.EmbeddingOutputDimension)

	ctx := context.Background()

	var ingestedDocuments int
	var ingestedChunks int

	for _, doc := range docs {
		existingChecksum, found, err := store.GetDocumentChecksum(ctx, doc.SourceKey)
		if err != nil {
			slog.Error("load_existing_checksum_error", "source_key", doc.SourceKey, "error", err)
			os.Exit(1)
		}
		if found && existingChecksum == doc.Checksum {
			slog.Info("document_skipped_unchanged", "source_key", doc.SourceKey)
			continue
		}

		chunks := ingest.ChunkDocument(doc, cfg.ChunkMaxChars, cfg.ChunkOverlapChars)
		if len(chunks) == 0 {
			slog.Warn("document_has_no_chunks", "source_key", doc.SourceKey)
			continue
		}

		slog.Info("document_chunked",
			"source_key", doc.SourceKey,
			"title", doc.Title,
			"chunks", len(chunks),
			"category", doc.Category,
		)

		if cfg.DryRun {
			continue
		}

		embeddedChunks := make([]ingest.EmbeddedChunk, 0, len(chunks))

		for start := 0; start < len(chunks); start += cfg.BatchSize {
			end := start + cfg.BatchSize
			if end > len(chunks) {
				end = len(chunks)
			}

			batch := chunks[start:end]
			inputs := make([]ingest.EmbeddingInput, 0, len(batch))
			for _, chunk := range batch {
				inputs = append(inputs, ingest.EmbeddingInput{
					Title: doc.Title,
					Text:  chunk.Content,
				})
			}

			vectors, err := embedder.Embed(ctx, inputs)
			if err != nil {
				slog.Error("embed_batch_error",
					"source_key", doc.SourceKey,
					"batch_start", start,
					"batch_end", end,
					"error", err,
				)
				os.Exit(1)
			}

			for i, chunk := range batch {
				embeddedChunks = append(embeddedChunks, ingest.EmbeddedChunk{
					Index:          chunk.Index,
					Heading:        chunk.Heading,
					Content:        chunk.Content,
					ContentLength:  chunk.ContentLength,
					TokenEstimate:  chunk.TokenEstimate,
					EmbeddingModel: cfg.EmbeddingModel,
					Embedding:      vectors[i],
				})
			}
		}

		_, err = store.UpsertDocumentAndReplaceChunks(ctx, ingest.StoredDocument{
			SourceKey:  doc.SourceKey,
			SourcePath: doc.SourcePath,
			Title:      doc.Title,
			Category:   doc.Category,
			DocType:    doc.DocType,
			Version:    doc.Version,
			Owner:      doc.Owner,
			Checksum:   doc.Checksum,
			Metadata:   doc.Metadata,
			Content:    doc.Body,
		}, embeddedChunks)
		if err != nil {
			slog.Error("persist_document_error", "source_key", doc.SourceKey, "error", err)
			os.Exit(1)
		}

		ingestedDocuments++
		ingestedChunks += len(embeddedChunks)

		slog.Info("document_ingested",
			"source_key", doc.SourceKey,
			"chunks", len(embeddedChunks),
		)
	}

	slog.Info("knowledge_ingestion_complete",
		"documents", ingestedDocuments,
		"chunks", ingestedChunks,
		"dry_run", cfg.DryRun,
	)
}