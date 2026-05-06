# Day 49 — Answer Synthesis + Citation-Ready Response

## Objective
Membangun lapisan jawaban di atas retrieval SOP sehingga sistem tidak hanya menemukan chunk yang relevan, tetapi juga:
- menyusun jawaban natural language
- tetap grounded pada source retrieval
- memberi citation inline
- jujur saat konteks SOP tidak cukup

## Scope
Day 49 mencakup:
- validasi config untuk answer synthesis
- prompt builder grounded
- generator Gemini untuk jawaban natural language
- command CLI answer
- script PowerShell / shell untuk menjalankan answer flow
- output dengan:
  - question
  - synthesized answer
  - source legend

Day 49 belum mencakup:
- HTTP chatbot endpoint
- memory / session conversation
- retrieval diversity tuning
- negative test tuning
- reranking

## Flow
1. user question masuk
2. query di-embed
3. retrieval mengambil top-k chunk
4. source snippets disusun menjadi grounded prompt
5. model menjawab hanya dari source yang diberikan
6. sistem menampilkan jawaban dan daftar sumber `[S1]`, `[S2]`, dst.

## Guardrails
Prompt synthesis memaksa model untuk:
- hanya memakai sumber yang diberikan
- tidak membuat aturan di luar SOP
- menulis citation inline
- menjawab jujur bila konteks tidak cukup

Fallback eksplisit:
`Saya belum menemukan dasar SOP yang cukup untuk menjawab pertanyaan ini.`

## Files Added / Changed
- `services/knowledge/internal/config/config.go`
- `services/knowledge/internal/config/config_test.go`
- `services/knowledge/internal/synthesis/prompt.go`
- `services/knowledge/internal/synthesis/prompt_test.go`
- `services/knowledge/internal/synthesis/generator.go`
- `services/knowledge/cmd/answer/main.go`
- `scripts/run-knowledge-answer.ps1`
- `scripts/run-knowledge-answer.sh`
- `docs/day49.md`

## Commands

### Run tests
```powershell
go test ./services/knowledge/... -count=1 -v

.\scripts\run-knowledge-answer.ps1 -Query "apakah amoxicillin bisa dijual tanpa resep?"

.\scripts\run-knowledge-answer.ps1 -Query "kapan transaksi obat keras harus direview supervisor?"
.\scripts\run-knowledge-answer.ps1 -Query "apa edukasi minimal untuk paracetamol otc?"
```

## Expected Result
- jawaban sudah natural language
- tetap grounded ke source SOP
- punya inline citation seperti `[S1]`
- menampilkan legend source di bawah jawaban

## Next Step
Day 50 akan fokus pada:
- retrieval quality improvements
- diversity / max chunks per document
- negative query handling
- threshold tuning