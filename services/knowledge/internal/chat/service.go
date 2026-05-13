package chat

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"finpharm-ai/services/knowledge/internal/config"
	"finpharm-ai/services/knowledge/internal/retrieval"
	"finpharm-ai/services/knowledge/internal/synthesis"
)

const insufficientContextMessage = "Saya belum menemukan dasar SOP yang cukup untuk menjawab pertanyaan ini."

type Request struct {
	Question string
	TopK     int
	MinScore float64
}

type Source struct {
	Ref       string  `json:"ref"`
	Title     string  `json:"title"`
	Category  string  `json:"category"`
	SourceKey string  `json:"source_key"`
	Heading   string  `json:"heading"`
	Score     float64 `json:"score"`
}

type Confidence struct {
	TopScore        float64 `json:"top_score"`
	MinTopScore     float64 `json:"min_top_score"`
	RetrievedCount  int     `json:"retrieved_count"`
	UsedSourceCount int     `json:"used_source_count"`
}

type Response struct {
	Question   string     `json:"question"`
	Answer     string     `json:"answer"`
	Fallback   bool       `json:"fallback"`
	Citations  []string   `json:"citations"`
	Sources    []Source   `json:"sources"`
	Confidence Confidence `json:"confidence"`
}

type Service struct {
	cfg           config.Config
	queryEmbedder *retrieval.QueryEmbedder
	searchStore   *retrieval.Store
	generator     *synthesis.Generator
}

func NewService(db *sql.DB, cfg config.Config) *Service {
	return &Service{
		cfg:           cfg,
		queryEmbedder: retrieval.NewQueryEmbedder(cfg.GeminiAPIKey, cfg.EmbeddingModel, cfg.EmbeddingOutputDimension),
		searchStore:   retrieval.NewStore(db),
		generator: synthesis.NewGenerator(
			cfg.GeminiAPIKey,
			cfg.AnswerModel,
			cfg.AnswerTemperature,
			cfg.AnswerMaxOutputTokens,
		),
	}
}

func (s *Service) Answer(ctx context.Context, req Request) (Response, error) {
	question := strings.TrimSpace(req.Question)
	if question == "" {
		return Response{}, fmt.Errorf("question is required")
	}

	if req.TopK <= 0 {
		req.TopK = 5
	}
	if req.MinScore == 0 {
		req.MinScore = 0.45
	}
	if req.MinScore < 0 || req.MinScore > 1 {
		req.MinScore = 0.45
	}

	queryEmbedding, err := s.queryEmbedder.EmbedQuery(ctx, question)
	if err != nil {
		return Response{}, err
	}

	candidateLimit := req.TopK * 3
	if candidateLimit < req.TopK {
		candidateLimit = req.TopK
	}

	rawResults, err := s.searchStore.Search(ctx, queryEmbedding, candidateLimit, req.MinScore)
	if err != nil {
		return Response{}, err
	}

	topScore := 0.0
	if len(rawResults) > 0 {
		topScore = rawResults[0].Score
	}

	if len(rawResults) == 0 || topScore < s.cfg.AnswerMinTopScore {
		return Response{
			Question:  question,
			Answer:    insufficientContextMessage,
			Fallback:  true,
			Citations: []string{},
			Sources:   []Source{},
			Confidence: Confidence{
				TopScore:        topScore,
				MinTopScore:     s.cfg.AnswerMinTopScore,
				RetrievedCount:  0,
				UsedSourceCount: 0,
			},
		}, nil
	}

	windowedResults := retrieval.FilterByTopScoreWindow(rawResults, s.cfg.AnswerScoreWindow)
	results := retrieval.DiversifyResults(windowedResults, req.TopK, s.cfg.AnswerMaxChunksPerDocument)

	if len(results) == 0 || results[0].Score < s.cfg.AnswerMinTopScore {
		return Response{
			Question:  question,
			Answer:    insufficientContextMessage,
			Fallback:  true,
			Citations: []string{},
			Sources:   []Source{},
			Confidence: Confidence{
				TopScore:        topScore,
				MinTopScore:     s.cfg.AnswerMinTopScore,
				RetrievedCount:  len(results),
				UsedSourceCount: 0,
			},
		}, nil
	}

	snippets := make([]synthesis.SourceSnippet, 0, len(results))
	for i, item := range results {
		ref := fmt.Sprintf("[S%d]", i+1)
		snippets = append(snippets, synthesis.SourceSnippet{
			Ref:       ref,
			Title:     item.Title,
			Category:  item.Category,
			SourceKey: item.SourceKey,
			Heading:   item.Heading,
			Content:   item.Content,
			Score:     item.Score,
		})
	}

	prompt := synthesis.BuildGroundedAnswerPrompt(question, snippets)

	answer, err := s.generator.Generate(ctx, prompt)
	if err != nil {
		return Response{}, err
	}

	answer = synthesis.NormalizeAnswerCitations(strings.TrimSpace(answer))

	if answer == insufficientContextMessage {
		return Response{
			Question:  question,
			Answer:    answer,
			Fallback:  true,
			Citations: []string{},
			Sources:   []Source{},
			Confidence: Confidence{
				TopScore:        topScore,
				MinTopScore:     s.cfg.AnswerMinTopScore,
				RetrievedCount:  len(results),
				UsedSourceCount: 0,
			},
		}, nil
	}

	usedRefs := synthesis.ExtractUsedSourceRefs(answer)
	usedSources := synthesis.FilterSourcesByRefs(snippets, usedRefs)

	if len(usedSources) == 0 {
		if len(snippets) > 2 {
			usedSources = snippets[:2]
		} else {
			usedSources = snippets
		}
	}

	citations := make([]string, 0, len(usedSources))
	sources := make([]Source, 0, len(usedSources))

	for _, src := range usedSources {
		citations = append(citations, src.Ref)
		sources = append(sources, Source{
			Ref:       src.Ref,
			Title:     src.Title,
			Category:  src.Category,
			SourceKey: src.SourceKey,
			Heading:   src.Heading,
			Score:     src.Score,
		})
	}

	return Response{
		Question:  question,
		Answer:    answer,
		Fallback:  false,
		Citations: citations,
		Sources:   sources,
		Confidence: Confidence{
			TopScore:        topScore,
			MinTopScore:     s.cfg.AnswerMinTopScore,
			RetrievedCount:  len(results),
			UsedSourceCount: len(sources),
		},
	}, nil
}