CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS knowledge_documents (
    id BIGSERIAL PRIMARY KEY,
    source_key TEXT NOT NULL UNIQUE,
    source_path TEXT NOT NULL,
    title TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT 'general',
    doc_type TEXT NOT NULL DEFAULT 'sop',
    version TEXT NOT NULL DEFAULT '1.0',
    owner TEXT NOT NULL DEFAULT 'operations',
    checksum TEXT NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_knowledge_documents_category
    ON knowledge_documents (category);

CREATE TABLE IF NOT EXISTS knowledge_chunks (
    id BIGSERIAL PRIMARY KEY,
    document_id BIGINT NOT NULL REFERENCES knowledge_documents(id) ON DELETE CASCADE,
    chunk_index INTEGER NOT NULL,
    heading TEXT NOT NULL DEFAULT '',
    content TEXT NOT NULL,
    content_length INTEGER NOT NULL,
    token_estimate INTEGER NOT NULL,
    embedding_model TEXT NOT NULL,
    embedding vector(768) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (document_id, chunk_index)
);

CREATE INDEX IF NOT EXISTS idx_knowledge_chunks_document_id_chunk_index
    ON knowledge_chunks (document_id, chunk_index);

CREATE INDEX IF NOT EXISTS idx_knowledge_chunks_embedding_cosine
    ON knowledge_chunks USING hnsw (embedding vector_cosine_ops);