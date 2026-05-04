package ingest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type EmbeddingInput struct {
	Title string
	Text  string
}

type GeminiBatchEmbedder struct {
	apiKey    string
	model     string
	outputDim int
	client    *http.Client
}

func NewGeminiBatchEmbedder(apiKey string, model string, outputDim int) *GeminiBatchEmbedder {
	return &GeminiBatchEmbedder{
		apiKey:    strings.TrimSpace(apiKey),
		model:     strings.TrimSpace(model),
		outputDim: outputDim,
		client:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (e *GeminiBatchEmbedder) Embed(ctx context.Context, inputs []EmbeddingInput) ([][]float32, error) {
	if len(inputs) == 0 {
		return nil, nil
	}
	if e.apiKey == "" {
		return nil, fmt.Errorf("gemini api key is required")
	}
	if e.model == "" {
		return nil, fmt.Errorf("embedding model is required")
	}

	type part struct {
		Text string `json:"text"`
	}
	type content struct {
		Parts []part `json:"parts"`
	}
	type embedRequest struct {
		Model                string  `json:"model"`
		TaskType             string  `json:"taskType,omitempty"`
		Title                string  `json:"title,omitempty"`
		OutputDimensionality int     `json:"outputDimensionality,omitempty"`
		Content              content `json:"content"`
	}
	type batchRequest struct {
		Requests []embedRequest `json:"requests"`
	}
	type contentEmbedding struct {
		Values []float32 `json:"values"`
	}
	type batchResponse struct {
		Embeddings []contentEmbedding `json:"embeddings"`
	}

	reqBody := batchRequest{Requests: make([]embedRequest, 0, len(inputs))}
	for _, input := range inputs {
		reqBody.Requests = append(reqBody.Requests, embedRequest{
			Model:                e.model,
			TaskType:             "RETRIEVAL_DOCUMENT",
			Title:                strings.TrimSpace(input.Title),
			OutputDimensionality: e.outputDim,
			Content:              content{Parts: []part{{Text: strings.TrimSpace(input.Text)}}},
		})
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal embedding request: %w", err)
	}

	url := "https://generativelanguage.googleapis.com/v1beta/" + strings.TrimPrefix(e.model, "models/") + ":batchEmbedContents"
	if strings.HasPrefix(e.model, "models/") {
		url = "https://generativelanguage.googleapis.com/v1beta/" + e.model + ":batchEmbedContents"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create embedding request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", e.apiKey)

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute embedding request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errBody map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&errBody)
		return nil, fmt.Errorf("embedding api returned status=%d body=%v", resp.StatusCode, errBody)
	}

	var decoded batchResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, fmt.Errorf("decode embedding response: %w", err)
	}

	if len(decoded.Embeddings) != len(inputs) {
		return nil, fmt.Errorf("embedding response mismatch: expected %d embeddings, got %d", len(inputs), len(decoded.Embeddings))
	}

	result := make([][]float32, 0, len(decoded.Embeddings))
	for _, emb := range decoded.Embeddings {
		result = append(result, emb.Values)
	}
	return result, nil
}