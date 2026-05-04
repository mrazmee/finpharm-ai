package ingest

import "testing"

func TestChunkDocument_PreservesHeadingsAndChunks(t *testing.T) {
	doc := SourceDocument{
		Title: "SOP Test",
		Body: `
# Tujuan
Paragraf pertama yang cukup panjang untuk dipotong menjadi beberapa chunk jika diperlukan.

# Prosedur
Langkah satu.

Langkah dua yang juga cukup panjang supaya chunking tetap jalan dengan heading yang benar.
`,
	}

	chunks := ChunkDocument(doc, 80, 20)
	if len(chunks) == 0 {
		t.Fatal("expected chunks, got none")
	}
	if chunks[0].Heading == "" || chunks[0].ContentLength == 0 || chunks[0].TokenEstimate == 0 {
		t.Fatal("expected heading, content length, and token estimate to be populated")
	}
}

func TestChunkDocument_HardSplitLongParagraph(t *testing.T) {
	doc := SourceDocument{
		Title: "Long Doc",
		Body:  "satu dua tiga empat lima enam tujuh delapan sembilan sepuluh sebelas dua belas tiga belas empat belas lima belas",
	}

	chunks := ChunkDocument(doc, 30, 0)
	if len(chunks) < 2 {
		t.Fatalf("expected at least 2 chunks, got %d", len(chunks))
	}
}