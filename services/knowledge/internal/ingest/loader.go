package ingest

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type SourceDocument struct {
	SourceKey  string
	SourcePath string
	Title      string
	Category   string
	DocType    string
	Version    string
	Owner      string
	Body       string
	Checksum   string
	Metadata   map[string]string
}

func LoadMarkdownDocuments(root string) ([]SourceDocument, error) {
	var docs []SourceDocument

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if strings.ToLower(filepath.Ext(path)) != ".md" {
			return nil
		}

		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		meta, body := parseFrontMatter(string(raw))
		body = strings.TrimSpace(body)
		if body == "" {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		title := firstNonEmpty(meta["title"], inferTitleFromFilename(path))
		category := firstNonEmpty(meta["category"], "general")
		docType := firstNonEmpty(meta["doc_type"], "sop")
		version := firstNonEmpty(meta["version"], "1.0")
		owner := firstNonEmpty(meta["owner"], "operations")

		sum := sha256.Sum256(raw)

		docs = append(docs, SourceDocument{
			SourceKey:  filepath.ToSlash(rel),
			SourcePath: filepath.ToSlash(path),
			Title:      title,
			Category:   category,
			DocType:    docType,
			Version:    version,
			Owner:      owner,
			Body:       body,
			Checksum:   hex.EncodeToString(sum[:]),
			Metadata:   meta,
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(docs, func(i, j int) bool {
		return docs[i].SourceKey < docs[j].SourceKey
	})

	return docs, nil
}

func parseFrontMatter(raw string) (map[string]string, string) {
	meta := map[string]string{}

	normalized := strings.ReplaceAll(raw, "\r\n", "\n")
	if !strings.HasPrefix(normalized, "---\n") {
		return meta, normalized
	}

	rest := strings.TrimPrefix(normalized, "---\n")
	endIdx := strings.Index(rest, "\n---\n")
	if endIdx == -1 {
		return meta, normalized
	}

	frontMatter := rest[:endIdx]
	body := rest[endIdx+5:]

	for _, line := range strings.Split(frontMatter, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(strings.ToLower(parts[0]))
		val := strings.TrimSpace(parts[1])
		if key != "" && val != "" {
			meta[key] = val
		}
	}

	return meta, body
}

func inferTitleFromFilename(path string) string {
	base := filepath.Base(path)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	base = strings.ReplaceAll(base, "-", " ")
	base = strings.ReplaceAll(base, "_", " ")
	return base
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v != "" {
			return v
		}
	}
	return ""
}