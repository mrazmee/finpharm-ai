package ingest

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

type StoredDocument struct {
	SourceKey  string
	SourcePath string
	Title      string
	Category   string
	DocType    string
	Version    string
	Owner      string
	Checksum   string
	Metadata   map[string]string
	Content    string
}

type EmbeddedChunk struct {
	Index          int
	Heading        string
	Content        string
	ContentLength  int
	TokenEstimate  int
	EmbeddingModel string
	Embedding      []float32
}

func (s *Store) GetDocumentChecksum(ctx context.Context, sourceKey string) (string, bool, error) {
	var checksum string
	err := s.db.QueryRowContext(ctx, `
		SELECT checksum
		FROM knowledge_documents
		WHERE source_key = $1
	`, sourceKey).Scan(&checksum)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return checksum, true, nil
}

func (s *Store) UpsertDocumentAndReplaceChunks(ctx context.Context, doc StoredDocument, chunks []EmbeddedChunk) (int64, error) {
	metaJSON, err := json.Marshal(doc.Metadata)
	if err != nil {
		return 0, fmt.Errorf("marshal metadata: %w", err)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var documentID int64
	err = tx.QueryRowContext(ctx, `
		INSERT INTO knowledge_documents (
			source_key, source_path, title, category, doc_type, version, owner, checksum, metadata, content, updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,NOW())
		ON CONFLICT (source_key)
		DO UPDATE SET
			source_path = EXCLUDED.source_path,
			title = EXCLUDED.title,
			category = EXCLUDED.category,
			doc_type = EXCLUDED.doc_type,
			version = EXCLUDED.version,
			owner = EXCLUDED.owner,
			checksum = EXCLUDED.checksum,
			metadata = EXCLUDED.metadata,
			content = EXCLUDED.content,
			updated_at = NOW()
		RETURNING id
	`,
		doc.SourceKey,
		doc.SourcePath,
		doc.Title,
		doc.Category,
		doc.DocType,
		doc.Version,
		doc.Owner,
		doc.Checksum,
		metaJSON,
		doc.Content,
	).Scan(&documentID)
	if err != nil {
		return 0, fmt.Errorf("upsert document: %w", err)
	}

	if _, err = tx.ExecContext(ctx, `
		DELETE FROM knowledge_chunks
		WHERE document_id = $1
	`, documentID); err != nil {
		return 0, fmt.Errorf("delete old chunks: %w", err)
	}

	for _, chunk := range chunks {
		vectorLiteral := vectorLiteral(chunk.Embedding)

		if _, err = tx.ExecContext(ctx, `
			INSERT INTO knowledge_chunks (
				document_id,
				chunk_index,
				heading,
				content,
				content_length,
				token_estimate,
				embedding_model,
				embedding
			)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8::vector)
		`,
			documentID,
			chunk.Index,
			chunk.Heading,
			chunk.Content,
			chunk.ContentLength,
			chunk.TokenEstimate,
			chunk.EmbeddingModel,
			vectorLiteral,
		); err != nil {
			return 0, fmt.Errorf("insert chunk %d: %w", chunk.Index, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit knowledge transaction: %w", err)
	}

	return documentID, nil
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