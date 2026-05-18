package ingest

import "strings"

type Chunk struct {
	Index         int
	Heading       string
	Content       string
	ContentLength int
	TokenEstimate int
}

func ChunkDocument(doc SourceDocument, maxChars int, overlapChars int) []Chunk {
	body := strings.ReplaceAll(doc.Body, "\r\n", "\n")

	type section struct {
		heading string
		text    string
	}

	var sections []section
	currentHeading := doc.Title
	var currentLines []string

	flushSection := func() {
		text := strings.TrimSpace(strings.Join(currentLines, "\n"))
		if text == "" {
			return
		}
		sections = append(sections, section{
			heading: currentHeading,
			text:    text,
		})
		currentLines = nil
	}

	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			flushSection()

			currentHeading = strings.TrimSpace(strings.TrimLeft(trimmed, "#"))
			if currentHeading == "" {
				currentHeading = doc.Title
			}
			continue
		}

		currentLines = append(currentLines, line)
	}

	flushSection()

	var chunks []Chunk
	nextIndex := 0

	for _, sec := range sections {
		paragraphs := splitParagraphs(sec.text)
		if len(paragraphs) == 0 {
			continue
		}

		var buffer string

		for _, p := range paragraphs {
			candidate := p
			if buffer != "" {
				candidate = buffer + "\n\n" + p
			}

			if len(candidate) <= maxChars {
				buffer = candidate
				continue
			}

			if strings.TrimSpace(buffer) != "" {
				chunks = append(chunks, makeChunk(nextIndex, sec.heading, buffer))
				nextIndex++

				overlapSeed := tail(buffer, overlapChars)
				if overlapSeed != "" {
					buffer = strings.TrimSpace(overlapSeed + "\n\n" + p)
				} else {
					buffer = p
				}
			} else {
				for _, hard := range hardSplitParagraph(p, maxChars) {
					chunks = append(chunks, makeChunk(nextIndex, sec.heading, hard))
					nextIndex++
				}
				buffer = ""
			}
		}

		if strings.TrimSpace(buffer) != "" {
			chunks = append(chunks, makeChunk(nextIndex, sec.heading, buffer))
			nextIndex++
		}
	}

	return chunks
}

func splitParagraphs(text string) []string {
	raw := strings.Split(text, "\n\n")
	out := make([]string, 0, len(raw))

	for _, p := range raw {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}

	return out
}

func hardSplitParagraph(text string, maxChars int) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	if len(text) <= maxChars {
		return []string{text}
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{text}
	}

	var parts []string
	var current string

	for _, word := range words {
		candidate := word
		if current != "" {
			candidate = current + " " + word
		}

		if len(candidate) <= maxChars {
			current = candidate
			continue
		}

		if current != "" {
			parts = append(parts, current)
		}
		current = word
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}

func tail(text string, overlapChars int) string {
	text = strings.TrimSpace(text)
	if overlapChars <= 0 || text == "" {
		return ""
	}
	if len(text) <= overlapChars {
		return text
	}
	return strings.TrimSpace(text[len(text)-overlapChars:])
}

func makeChunk(index int, heading string, content string) Chunk {
	content = strings.TrimSpace(content)

	return Chunk{
		Index:         index,
		Heading:       strings.TrimSpace(heading),
		Content:       content,
		ContentLength: len(content),
		TokenEstimate: estimateTokens(content),
	}
}

func estimateTokens(text string) int {
	text = strings.TrimSpace(text)
	if text == "" {
		return 0
	}

	estimated := len(text) / 4
	if estimated <= 0 {
		return 1
	}

	return estimated
}