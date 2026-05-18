# Day 48 — Retrieval / Similarity Search Foundation

## Objective
Membangun lapisan retrieval di atas knowledge base SOP yang sudah di-ingest pada Day 47, sehingga sistem dapat:
- menerima pertanyaan user
- mengubah pertanyaan menjadi embedding query
- mencari chunk SOP paling relevan di pgvector
- mengembalikan top-k hasil retrieval lengkap dengan metadata sumber

## Scope
Day 48 mencakup:
- validasi config khusus retrieval query
- query embedder untuk Gemini Embedding
- SQL similarity search ke `knowledge_chunks`
- join hasil dengan `knowledge_documents`
- command CLI untuk retrieval manual
- script PowerShell / shell untuk menjalankan retrieval
- dokumentasi retrieval foundation

Day 48 belum mencakup:
- answer generation berbasis LLM
- citation formatting final untuk UI
- HTTP endpoint chatbot
- reranking
- hybrid search keyword + vector

## Files Added / Changed
- `services/knowledge/internal/config/config.go`
- `services/knowledge/internal/config/config_test.go`
- `services/knowledge/internal/retrieval/query_embedder.go`
- `services/knowledge/internal/retrieval/store.go`
- `services/knowledge/internal/retrieval/store_test.go`
- `services/knowledge/cmd/query/main.go`
- `scripts/run-knowledge-query.ps1`
- `scripts/run-knowledge-query.sh`
- `docs/day48.md`

## Retrieval Flow
1. user query masuk melalui CLI
2. query di-embed dengan model embedding yang sama family-nya dengan ingestion
3. embedding query dipakai untuk similarity search ke `knowledge_chunks`
4. hasil diurutkan berdasarkan cosine similarity
5. top-k chunk dikembalikan bersama:
   - title
   - category
   - source key
   - heading
   - chunk index
   - content
   - score

## Commands
### Run tests
```powershell
go test ./services/knowledge/... -count=1 -v
```

### Run retrieval
```powershell
.\scripts
un-knowledge-query.ps1 -Query "apakah amoxicillin bisa dijual tanpa resep?"
```

### Optional parameters
```powershell
.\scripts
un-knowledge-query.ps1 -Query "bagaimana menangani pembelian obat keras yang mencurigakan?" -TopK 6 -MinScore 0.40
```

## Verification Examples
Pertanyaan yang bagus untuk uji retrieval:
- `apakah amoxicillin bisa dijual tanpa resep?`
- `kapan transaksi obat keras harus direview supervisor?`
- `apa edukasi minimal untuk paracetamol otc?`

## Expected Result
- pertanyaan tentang antibiotik harus mengembalikan chunk dari SOP `antibiotic-amoxicillin.md`
- pertanyaan tentang obat keras / red flags harus mengembalikan chunk dari SOP `controlled-medicine-review.md`
- pertanyaan tentang edukasi OTC harus mengembalikan chunk dari SOP `otc-paracetamol.md`

## Next Step
Day 49 akan fokus pada:
- answer synthesis berbasis top-k retrieval results
- citation rendering dari metadata source
- fondasi endpoint chatbot SOP