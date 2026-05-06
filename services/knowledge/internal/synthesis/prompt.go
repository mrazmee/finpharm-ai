package synthesis

import (
	"fmt"
	"strings"
)

type SourceSnippet struct {
	Ref       string
	Title     string
	Category  string
	SourceKey string
	Heading   string
	Content   string
	Score     float64
}

func BuildGroundedAnswerPrompt(question string, snippets []SourceSnippet) string {
	var b strings.Builder

	b.WriteString("Anda adalah asisten SOP farmasi internal.\n")
	b.WriteString("Tugas Anda adalah menjawab pertanyaan user hanya berdasarkan sumber yang diberikan.\n")
	b.WriteString("Aturan:\n")
	b.WriteString("1. Jangan menggunakan pengetahuan di luar sumber.\n")
	b.WriteString("2. Jika sumber tidak cukup, jawab persis: Saya belum menemukan dasar SOP yang cukup untuk menjawab pertanyaan ini.\n")
	b.WriteString("3. Gunakan bahasa yang sama dengan bahasa pertanyaan user.\n")
	b.WriteString("4. Berikan jawaban yang ringkas, jelas, dan operasional.\n")
	b.WriteString("5. Setiap poin penting harus memakai citation inline seperti [S1] atau [S2].\n")
	b.WriteString("6. Jangan membuat aturan yang tidak tertulis di sumber.\n")
	b.WriteString("7. Jangan menuliskan bagian SOURCES. Hanya tulis jawaban final.\n\n")

	b.WriteString("PERTANYAAN USER:\n")
	b.WriteString(strings.TrimSpace(question))
	b.WriteString("\n\n")

	b.WriteString("SUMBER TERSEDIA:\n")
	for _, s := range snippets {
		b.WriteString(fmt.Sprintf("%s\n", s.Ref))
		b.WriteString(fmt.Sprintf("title: %s\n", strings.TrimSpace(s.Title)))
		b.WriteString(fmt.Sprintf("category: %s\n", strings.TrimSpace(s.Category)))
		b.WriteString(fmt.Sprintf("source_key: %s\n", strings.TrimSpace(s.SourceKey)))
		b.WriteString(fmt.Sprintf("heading: %s\n", strings.TrimSpace(s.Heading)))
		b.WriteString("content:\n")
		b.WriteString(strings.TrimSpace(s.Content))
		b.WriteString("\n\n")
	}

	b.WriteString("Tulis jawaban final sekarang.\n")

	return b.String()
}