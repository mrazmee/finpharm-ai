package synthesis

import "testing"

func TestNormalizeAnswerCitations_ClosesDanglingRef(t *testing.T) {
	answer := "Transaksi harus direview supervisor [S2"
	got := NormalizeAnswerCitations(answer)

	if got != "Transaksi harus direview supervisor [S2]" {
		t.Fatalf("unexpected normalized answer: %q", got)
	}
}

func TestExtractUsedSourceRefs_UniqueAndOrdered(t *testing.T) {
	answer := "Antibiotik butuh resep [S2]. Jika tidak ada resep, tahan transaksi [S1]. Lihat juga [S2]."

	got := ExtractUsedSourceRefs(answer)

	if len(got) != 2 {
		t.Fatalf("expected 2 refs, got %d", len(got))
	}
	if got[0] != "[S2]" || got[1] != "[S1]" {
		t.Fatalf("unexpected refs order: %#v", got)
	}
}

func TestFilterSourcesByRefs_FiltersUsedOnly(t *testing.T) {
	snippets := []SourceSnippet{
		{Ref: "[S1]", Title: "Doc 1"},
		{Ref: "[S2]", Title: "Doc 2"},
		{Ref: "[S3]", Title: "Doc 3"},
	}

	got := FilterSourcesByRefs(snippets, []string{"[S2]", "[S3]"})

	if len(got) != 2 {
		t.Fatalf("expected 2 sources, got %d", len(got))
	}
	if got[0].Ref != "[S2]" || got[1].Ref != "[S3]" {
		t.Fatalf("unexpected filtered refs: %#v", got)
	}
}