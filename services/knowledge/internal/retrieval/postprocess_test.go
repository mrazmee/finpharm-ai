package retrieval

import "testing"

func TestFilterByTopScoreWindow_KeepsNearTop(t *testing.T) {
	results := []SearchResult{
		{DocumentID: 1, Score: 0.80},
		{DocumentID: 1, Score: 0.77},
		{DocumentID: 2, Score: 0.74},
		{DocumentID: 3, Score: 0.69},
	}

	got := FilterByTopScoreWindow(results, 0.05)

	if len(got) != 2 {
		t.Fatalf("expected 2 results, got %d", len(got))
	}
}

func TestDiversifyResults_LimitsPerDocument(t *testing.T) {
	results := []SearchResult{
		{DocumentID: 1, ChunkIndex: 0},
		{DocumentID: 1, ChunkIndex: 1},
		{DocumentID: 1, ChunkIndex: 2},
		{DocumentID: 2, ChunkIndex: 0},
		{DocumentID: 2, ChunkIndex: 1},
	}

	got := DiversifyResults(results, 4, 2)

	if len(got) != 4 {
		t.Fatalf("expected 4 results, got %d", len(got))
	}

	doc1 := 0
	doc2 := 0
	for _, item := range got {
		if item.DocumentID == 1 {
			doc1++
		}
		if item.DocumentID == 2 {
			doc2++
		}
	}

	if doc1 > 2 {
		t.Fatalf("expected max 2 chunks for doc1, got %d", doc1)
	}
	if doc2 > 2 {
		t.Fatalf("expected max 2 chunks for doc2, got %d", doc2)
	}
}

func TestDiversifyResults_ReturnsNilWhenEmpty(t *testing.T) {
	got := DiversifyResults(nil, 5, 2)
	if got != nil {
		t.Fatal("expected nil result")
	}
}