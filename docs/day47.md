# Day 47 — RAG Ingestion SOP Foundation

## Objective
Membangun fondasi ingestion RAG untuk chatbot SOP farmasi dengan memanfaatkan PostgreSQL + pgvector yang sudah ada, dokumen SOP sintetis berformat markdown, proses chunking, embedding, dan penyimpanan chunk ke database.

## Scope
Day 47 mencakup:
- mengganti image Postgres agar mendukung pgvector
- menambahkan migration knowledge documents dan knowledge chunks
- membuat config untuk komponen knowledge ingestion
- membuat loader markdown + front matter parser
- membuat chunker dokumen
- membuat Gemini embeddings batch client
- membuat persistence ke Postgres
- membuat command migrate dan ingest
- menambahkan 3 dokumen SOP sintetis

Day 47 belum mencakup:
- query similarity search runtime
- chatbot retrieval endpoint
- citation rendering
- answer generation

## Commands
### 1. Recreate postgres with pgvector image
```powershell
docker compose up -d postgres --force-recreate
```

### 2. Run knowledge migrations
```powershell
.\scripts
un-knowledge-migrate.ps1
```

### 3. Run knowledge ingestion
```powershell
.\scripts
un-knowledge-ingest.ps1
```

## Required Environment
Defaults:
- database host: `127.0.0.1`
- database port: `55432`
- database name: `postgres`
- source dir: `./knowledge/sop`
- embedding model: `models/gemini-embedding-001`
- embedding dimension: `768`

Required for real ingestion:
- `GEMINI_API_KEY` or `GOOGLE_API_KEY`

Optional:
- `KNOWLEDGE_DRY_RUN=true` untuk hanya melihat hasil chunking tanpa memanggil embedding API

## Verification
### Config / unit test
```powershell
go test ./services/knowledge/... -count=1 -v
```

### Check ingestion logs
Harus terlihat:
- document chunked
- document ingested
- knowledge ingestion complete

### Check database
Contoh query:
```sql
SELECT COUNT(*) FROM knowledge_documents;
SELECT COUNT(*) FROM knowledge_chunks;
SELECT title, category, checksum FROM knowledge_documents ORDER BY id;
```

## Next Step
Day 48 akan fokus pada:
- similarity search query di pgvector
- retrieval helper
- top-k chunk lookup
- fondasi citation source tracking untuk chatbot SOP