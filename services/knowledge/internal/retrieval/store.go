package retrieval

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type SearchResult struct {
	DocumentID    int64
	Title         string
	Category      string
	SourceKey     string
	ChunkIndex    int
	Heading       string
	Content       string
	TokenEstimate int
	Score         float64
}

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Search(ctx context.Context, queryEmbedding []float32, limit int, minScore float64) ([]SearchResult, error) {
	if len(queryEmbedding) == 0 {
		return nil, fmt.Errorf("query embedding is required")
	}
	if limit <= 0 {
		return nil, fmt.Errorf("limit must be > 0")
	}
	if s.db == nil {
		return nil, fmt.Errorf("database handle is required")
	}

	vectorLiteral := vectorLiteral(queryEmbedding)

	rows, err := s.db.QueryContext(ctx, `
		SELECT
			kd.id,
			kd.title,
			kd.category,
			kd.source_key,
			kc.chunk_index,
			kc.heading,
			kc.content,
			kc.token_estimate,
			1 - (kc.embedding <=> $1::vector) AS score
		FROM knowledge_chunks kc
		INNER JOIN knowledge_documents kd ON kd.id = kc.document_id
		WHERE 1 - (kc.embedding <=> $1::vector) >= $2
		ORDER BY kc.embedding <=> $1::vector ASC, kd.id ASC, kc.chunk_index ASC
		LIMIT $3
	`, vectorLiteral, minScore, limit)
	if err != nil {
		return nil, fmt.Errorf("search knowledge chunks: %w", err)
	}
	defer rows.Close()

	results := make([]SearchResult, 0, limit)
	for rows.Next() {
		var item SearchResult
		if err := rows.Scan(
			&item.DocumentID,
			&item.Title,
			&item.Category,
			&item.SourceKey,
			&item.ChunkIndex,
			&item.Heading,
			&item.Content,
			&item.TokenEstimate,
			&item.Score,
		); err != nil {
			return nil, fmt.Errorf("scan knowledge search result: %w", err)
		}
		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate knowledge search results: %w", err)
	}

	return results, nil
}

func vectorLiteral(values []float32) string {
	if len(values) == 0 {
		return "[]"
	}

	parts := make([]string, 0, len(values))
	for _, v := range values {
		parts = append(parts, fmt.Sprintf("%f", v))
	}
	return "[" + strings.Join(parts, ",") + "]"
}