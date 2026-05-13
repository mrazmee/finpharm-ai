# Day 50 — Retrieval Quality Improvement

## Objective
Meningkatkan kualitas retrieval-answer flow agar lebih production-minded dengan fokus pada:
- diversity hasil retrieval
- confidence gating
- score-window filtering
- citation normalization
- cited-sources-only output
- clean fallback untuk pertanyaan di luar SOP

## Why this day exists
Setelah Day 49, answer flow sudah berjalan, tetapi masih ada beberapa gap kualitas:
- hasil retrieval bisa terlalu didominasi satu dokumen
- source legend kadang terlalu banyak
- citation dari model bisa tidak rapi
- query di luar domain masih perlu fallback yang lebih bersih
- answer flow perlu lebih disiplin sebelum dianggap siap sebagai backend portfolio yang kuat

Day 50 menutup gap tersebut tanpa menambah scope liar.

---

## Scope
Day 50 mencakup:
- config tambahan untuk quality control answer flow
- filter hasil retrieval berdasarkan top-score window
- pembatasan jumlah chunk per dokumen
- normalisasi citation model output
- ekstraksi citation refs dari jawaban final
- menampilkan hanya source yang benar-benar dipakai
- fallback langsung bila top retrieval score terlalu lemah
- tidak menampilkan source legend saat jawaban memang “konteks tidak cukup”

---

## What changed

### 1. Confidence gate
Answer flow sekarang tidak langsung memanggil model generatif untuk semua query.

Jika:
- hasil retrieval kosong, atau
- score tertinggi berada di bawah threshold minimum

maka sistem langsung menjawab:

`Saya belum menemukan dasar SOP yang cukup untuk menjawab pertanyaan ini.`

Tujuan:
- mengurangi jawaban terlalu percaya diri
- menjaga sistem tetap jujur untuk pertanyaan di luar SOP

### 2. Score-window filtering
Sebelum melakukan diversification, hasil retrieval sekarang lebih dulu difilter dengan **score window** dari top result.

Artinya:
- chunk yang terlalu jauh kualitasnya dari hasil teratas tidak ikut dibawa ke prompt
- context jadi lebih relevan
- peluang context tercampur dokumen lain yang kurang kuat jadi lebih kecil

### 3. Diversity per document
Setelah score-window filtering, sistem menerapkan pembatas:
- maksimal sejumlah chunk per dokumen

Tujuan:
- mengurangi dominasi satu dokumen
- mencegah prompt terlalu repetitif
- membuat context lebih hemat dan lebih terkontrol

### 4. Citation normalization
Model generatif kadang dapat menghasilkan citation yang tidak lengkap, misalnya:
- `[S2`

Day 50 menambahkan normalisasi citation sehingga pola seperti itu dibetulkan menjadi:
- `[S2]`

Tujuan:
- menjaga output tetap rapi
- membantu proses ekstraksi citation refs
- mengurangi warning akibat citation setengah jadi

### 5. Cited-sources-only output
Source legend di bawah jawaban sekarang tidak lagi menampilkan semua hasil retrieval mentah.

Yang ditampilkan adalah:
- hanya source yang benar-benar muncul di jawaban final melalui citation seperti `[S1]`, `[S2]`, dst.

Tujuan:
- output lebih bersih
- source legend benar-benar selaras dengan isi jawaban
- lebih siap untuk dipakai frontend di tahap berikutnya

### 6. Clean fallback
Jika jawaban final adalah:

`Saya belum menemukan dasar SOP yang cukup untuk menjawab pertanyaan ini.`

maka sistem:
- tidak menampilkan source legend
- tidak menampilkan source yang berpotensi membingungkan
- langsung berhenti di fallback itu

Tujuan:
- menghindari kesan seolah sistem punya dasar jawaban padahal sebenarnya tidak cukup

---

## Final Config Added
Day 50 menambahkan dua kontrol penting:

- `KNOWLEDGE_ANSWER_MIN_TOP_SCORE`
  - default: `0.62`
  - fungsi: minimum score tertinggi agar answer flow boleh lanjut

- `KNOWLEDGE_ANSWER_MAX_CHUNKS_PER_DOCUMENT`
  - default: `2`
  - fungsi: membatasi jumlah chunk dari satu dokumen

- `KNOWLEDGE_ANSWER_SCORE_WINDOW`
  - default: `0.05`
  - fungsi: hanya mempertahankan hasil retrieval yang masih dekat dengan score terbaik

---

## Files Changed
- `services/knowledge/internal/config/config.go`
- `services/knowledge/internal/config/config_test.go`
- `services/knowledge/internal/retrieval/postprocess.go`
- `services/knowledge/internal/retrieval/postprocess_test.go`
- `services/knowledge/internal/synthesis/citations.go`
- `services/knowledge/internal/synthesis/citations_test.go`
- `services/knowledge/internal/synthesis/prompt.go`
- `services/knowledge/cmd/answer/main.go`
- `docs/day50.md`

---

## Verification Summary

### Positive query — antibiotic rule
**Query:**
`apakah amoxicillin bisa dijual tanpa resep?`

**Expected:**
- jawaban grounded
- inline citation muncul
- sources hanya source yang benar-benar dipakai

**Observed:**
- jawaban menyatakan amoxicillin tidak boleh dijual bebas tanpa verifikasi resep
- source legend hanya memuat 2 source yang dipakai
- output bersih dan konsisten

### Positive query — controlled medicine review
**Query:**
`kapan transaksi obat keras harus direview supervisor?`

**Expected:**
- jawaban tidak terpotong
- tidak ada warning citation missing
- source legend lebih sempit dan relevan

**Observed:**
- jawaban sudah utuh
- tidak ada warning citation missing
- source legend hanya memuat source yang dipakai model

### Negative query — out of domain
**Query:**
`berapa gaji apoteker?`

**Expected:**
- fallback jujur
- tanpa source legend

**Observed:**
- sistem menjawab bahwa dasar SOP tidak cukup
- source legend tidak ditampilkan
- behavior sudah sesuai guardrail yang diinginkan

---

## Commands

### Run tests
```powershell
go test ./services/knowledge/... -count=1 -v
```

### Positive query verification
```powershell
.\scripts\run-knowledge-answer.ps1 -Query "apakah amoxicillin bisa dijual tanpa resep?"
.\scripts\run-knowledge-answer.ps1 -Query "kapan transaksi obat keras harus direview supervisor?"
```

### Negative query verification
```powershell
.\scripts\run-knowledge-answer.ps1 -Query "berapa gaji apoteker?"
```

---

## Expected Result After Day 50
- jawaban positif tetap grounded dan punya citation
- source legend lebih singkat dan lebih relevan
- query di luar SOP jatuh ke jawaban jujur
- query di luar SOP tidak menampilkan source legend
- answer flow lebih siap dijadikan endpoint HTTP

## Why this matters for portfolio
Day 50 meningkatkan sistem dari sekadar:
- “bisa retrieval”
- “bisa menjawab”

menjadi:
- lebih disiplin
- lebih hemat context
- lebih jujur saat pengetahuan tidak cukup
- lebih rapi untuk dipakai sebagai internal SOP assistant API

Ini penting untuk menunjukkan bahwa project tidak hanya mengejar happy path, tetapi juga memperhatikan:
- answer reliability
- output cleanliness
- guardrails
- production-minded behavior

---

## Final Status
**Day 50 complete.**

Setelah Day 50, fondasi RAG backend project ini sudah mencakup:
- ingestion
- retrieval
- answer synthesis
- citations
- quality control
- safe fallback