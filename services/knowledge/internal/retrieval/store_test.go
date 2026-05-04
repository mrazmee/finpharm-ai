package retrieval

import (
	"context"
	"testing"
)

func TestSearch_RejectsEmptyEmbedding(t *testing.T) {
	store := NewStore(nil)

	_, err := store.Search(context.Background(), nil, 5, 0.0)
	if err == nil {
		t.Fatal("expected error for empty embedding, got nil")
	}
}

func TestSearch_RejectsNonPositiveLimit(t *testing.T) {
	store := NewStore(nil)

	_, err := store.Search(context.Background(), []float32{0.1, 0.2}, 0, 0.0)
	if err == nil {
		t.Fatal("expected error for non-positive limit, got nil")
	}
}

func TestVectorLiteral_NotEmpty(t *testing.T) {
	got := vectorLiteral([]float32{0.1, 0.2, 0.3})
	if got == "" || got == "[]" {
		t.Fatalf("expected vector literal, got %q", got)
	}
}