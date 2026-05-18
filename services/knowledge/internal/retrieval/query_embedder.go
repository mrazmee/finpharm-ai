package retrieval

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type QueryEmbedder struct {
	apiKey    string
	model     string
	outputDim int
	client    *http.Client
}

func NewQueryEmbedder(apiKey string, model string, outputDim int) *QueryEmbedder {
	return &QueryEmbedder{
		apiKey:    strings.TrimSpace(apiKey),
		model:     normalizeModel(model),
		outputDim: outputDim,
		client: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

func (e *QueryEmbedder) EmbedQuery(ctx context.Context, query string) ([]float32, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}
	if e.apiKey == "" {
		return nil, fmt.Errorf("gemini api key is required")
	}
	if e.model == "" {
		return nil, fmt.Errorf("embedding model is required")
	}
	if e.outputDim <= 0 {
		return nil, fmt.Errorf("output dimension must be > 0")
	}

	type part struct {
		Text string `json:"text"`
	}
	type content struct {
		Parts []part `json:"parts"`
	}
	type requestBody struct {
		Content              content `json:"content"`
		TaskType             string  `json:"taskType"`
		OutputDimensionality int     `json:"outputDimensionality"`
	}
	type embedding struct {
		Values []float32 `json:"values"`
	}
	type responseBody struct {
		Embedding embedding `json:"embedding"`
	}

	body := requestBody{
		Content: content{
			Parts: []part{
				{Text: query},
			},
		},
		TaskType:             "RETRIEVAL_QUERY",
		OutputDimensionality: e.outputDim,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal query embedding request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://generativelanguage.googleapis.com/v1beta/"+e.model+":embedContent",
		bytes.NewReader(payload),
	)
	if err != nil {
		return nil, fmt.Errorf("create query embedding request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", e.apiKey)

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute query embedding request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errBody map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&errBody)
		return nil, fmt.Errorf("query embedding api returned status=%d body=%v", resp.StatusCode, errBody)
	}

	var decoded responseBody
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, fmt.Errorf("decode query embedding response: %w", err)
	}

	if len(decoded.Embedding.Values) == 0 {
		return nil, fmt.Errorf("query embedding response returned empty vector")
	}

	return decoded.Embedding.Values, nil
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