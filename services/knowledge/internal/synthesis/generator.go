package synthesis

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Generator struct {
	apiKey          string
	model           string
	temperature     float64
	maxOutputTokens int
	client          *http.Client
}

func NewGenerator(apiKey string, model string, temperature float64, maxOutputTokens int) *Generator {
	return &Generator{
		apiKey:          strings.TrimSpace(apiKey),
		model:           normalizeModel(model),
		temperature:     temperature,
		maxOutputTokens: maxOutputTokens,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (g *Generator) Generate(ctx context.Context, prompt string) (string, error) {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return "", fmt.Errorf("prompt is required")
	}
	if g.apiKey == "" {
		return "", fmt.Errorf("gemini api key is required")
	}
	if g.model == "" {
		return "", fmt.Errorf("answer model is required")
	}
	if g.maxOutputTokens <= 0 {
		return "", fmt.Errorf("max output tokens must be > 0")
	}

	type part struct {
		Text string `json:"text"`
	}
	type content struct {
		Parts []part `json:"parts"`
	}
	type generationConfig struct {
		Temperature     float64 `json:"temperature,omitempty"`
		MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
	}
	type requestBody struct {
		Contents         []content         `json:"contents"`
		GenerationConfig generationConfig  `json:"generationConfig,omitempty"`
	}
	type candidatePart struct {
		Text string `json:"text"`
	}
	type candidateContent struct {
		Parts []candidatePart `json:"parts"`
	}
	type candidate struct {
		Content candidateContent `json:"content"`
	}
	type responseBody struct {
		Candidates []candidate `json:"candidates"`
	}

	body := requestBody{
		Contents: []content{
			{
				Parts: []part{
					{Text: prompt},
				},
			},
		},
		GenerationConfig: generationConfig{
			Temperature:     g.temperature,
			MaxOutputTokens: g.maxOutputTokens,
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal answer request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://generativelanguage.googleapis.com/v1beta/"+g.model+":generateContent",
		bytes.NewReader(payload),
	)
	if err != nil {
		return "", fmt.Errorf("create answer request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", g.apiKey)

	resp, err := g.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("execute answer request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errBody map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&errBody)
		return "", fmt.Errorf("answer api returned status=%d body=%v", resp.StatusCode, errBody)
	}

	var decoded responseBody
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return "", fmt.Errorf("decode answer response: %w", err)
	}

	if len(decoded.Candidates) == 0 {
		return "", fmt.Errorf("answer response returned no candidates")
	}

	var textParts []string
	for _, p := range decoded.Candidates[0].Content.Parts {
		if strings.TrimSpace(p.Text) != "" {
			textParts = append(textParts, strings.TrimSpace(p.Text))
		}
	}

	answer := strings.TrimSpace(strings.Join(textParts, "\n"))
	if answer == "" {
		return "", fmt.Errorf("answer response returned empty text")
	}

	return answer, nil
}

func normalizeModel(model string) string {
	model = strings.TrimSpace(model)
	if model == "" {
		return ""
	}
	if strings.HasPrefix(model, "models/") {
		return model
	}
	return "models/" + model
}