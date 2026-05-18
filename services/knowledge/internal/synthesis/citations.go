package synthesis

import (
	"regexp"
	"strings"
)

var strictCitationPattern = regexp.MustCompile(`\[S\d+\]`)
var looseCitationPattern = regexp.MustCompile(`\[S\d+\]?`)

func NormalizeAnswerCitations(answer string) string {
	answer = strings.TrimSpace(answer)
	if answer == "" {
		return answer
	}

	return looseCitationPattern.ReplaceAllStringFunc(answer, func(m string) string {
		if strings.HasSuffix(m, "]") {
			return m
		}
		return m + "]"
	})
}

func ExtractUsedSourceRefs(answer string) []string {
	answer = NormalizeAnswerCitations(answer)

	matches := strictCitationPattern.FindAllString(answer, -1)
	if len(matches) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(matches))
	out := make([]string, 0, len(matches))

	for _, m := range matches {
		if _, ok := seen[m]; ok {
			continue
		}
		seen[m] = struct{}{}
		out = append(out, m)
	}

	return out
}

func FilterSourcesByRefs(snippets []SourceSnippet, refs []string) []SourceSnippet {
	if len(snippets) == 0 || len(refs) == 0 {
		return nil
	}

	index := make(map[string]SourceSnippet, len(snippets))
	for _, s := range snippets {
		index[s.Ref] = s
	}

	out := make([]SourceSnippet, 0, len(refs))
	for _, ref := range refs {
		if s, ok := index[ref]; ok {
			out = append(out, s)
		}
	}

	return out
}