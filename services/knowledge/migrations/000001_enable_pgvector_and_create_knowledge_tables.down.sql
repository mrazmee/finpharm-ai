DROP INDEX IF EXISTS idx_knowledge_chunks_embedding_cosine;
DROP INDEX IF EXISTS idx_knowledge_chunks_document_id_chunk_index;
DROP TABLE IF EXISTS knowledge_chunks;

DROP INDEX IF EXISTS idx_knowledge_documents_category;
DROP TABLE IF EXISTS knowledge_documents;

DROP EXTENSION IF EXISTS vector;