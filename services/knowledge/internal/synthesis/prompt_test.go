package synthesis

import (
	"strings"
	"testing"
)

func TestBuildGroundedAnswerPrompt_ContainsQuestionAndSources(t *testing.T) {
	prompt := BuildGroundedAnswerPrompt(
		"apakah amoxicillin bisa dijual tanpa resep?",
		[]SourceSnippet{
			{
				Ref:       "[S1]",
				Title:     "SOP Penjualan Antibiotik",
				Category:  "antibiotic-dispensation",
				SourceKey: "antibiotic-amoxicillin.md",
				Heading:   "Aturan Dasar",
				Content:   "Antibiotik tidak boleh dijual bebas tanpa verifikasi resep.",
				Score:     0.72,
			},
		},
	)

	if !strings.Contains(prompt, "PERTANYAAN USER") {
		t.Fatal("expected prompt to contain question section")
	}
	if !strings.Contains(prompt, "[S1]") {
		t.Fatal("expected prompt to contain source reference")
	}
	if !strings.Contains(prompt, "Antibiotik tidak boleh dijual bebas") {
		t.Fatal("expected prompt to contain source content")
	}
}