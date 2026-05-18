package retrieval

func FilterByTopScoreWindow(results []SearchResult, window float64) []SearchResult {
	if len(results) == 0 {
		return nil
	}
	if window < 0 {
		window = 0
	}

	threshold := results[0].Score - window
	out := make([]SearchResult, 0, len(results))

	for _, item := range results {
		if item.Score >= threshold {
			out = append(out, item)
		}
	}

	return out
}

func DiversifyResults(results []SearchResult, limit int, maxPerDocument int) []SearchResult {
	if len(results) == 0 || limit <= 0 {
		return nil
	}
	if maxPerDocument <= 0 {
		maxPerDocument = limit
	}

	counts := make(map[int64]int)
	out := make([]SearchResult, 0, limit)

	for _, item := range results {
		if counts[item.DocumentID] >= maxPerDocument {
			continue
		}

		out = append(out, item)
		counts[item.DocumentID]++

		if len(out) == limit {
			break
		}
	}

	return out
}