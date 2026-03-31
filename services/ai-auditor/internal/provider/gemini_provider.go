package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"finpharm-ai/services/ai-auditor/internal/domain"
)

const geminiGenerateURL = "https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent"

type GeminiProvider struct {
	apiKey string
	model  string
	client *http.Client
}

func NewGeminiProvider(apiKey, model string, timeout time.Duration) *GeminiProvider {
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	return &GeminiProvider{
		apiKey: strings.TrimSpace(apiKey),
		model:  strings.TrimSpace(model),
		client: &http.Client{Timeout: timeout},
	}
}

type geminiRequest struct {
	SystemInstruction *geminiSystemInstruction `json:"systemInstruction,omitempty"`
	Contents          []geminiContent          `json:"contents"`
	GenerationConfig  geminiGenerationConfig   `json:"generationConfig"`
}

type geminiSystemInstruction struct {
	Parts []geminiPart `json:"parts"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenerationConfig struct {
	Temperature      float64                `json:"temperature,omitempty"`
	ResponseMimeType string                 `json:"responseMimeType,omitempty"`
	ResponseSchema   map[string]interface{} `json:"responseSchema,omitempty"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`
	ModelVersion string `json:"modelVersion"`
	PromptFeedback struct {
		BlockReason string `json:"blockReason"`
	} `json:"promptFeedback"`
}

type geminiAuditJSON struct {
	Decision  string  `json:"decision"`
	RiskScore float64 `json:"risk_score"`
	Reason    string  `json:"reason"`
}

func (p *GeminiProvider) AuditTransaction(ctx context.Context, req domain.AuditTransactionRequest) (domain.AuditTransactionResult, error) {
	if p.apiKey == "" {
		return domain.AuditTransactionResult{}, errors.New("gemini api key is empty")
	}
	if p.model == "" {
		return domain.AuditTransactionResult{}, errors.New("gemini model is empty")
	}

	body := geminiRequest{
		SystemInstruction: &geminiSystemInstruction{
			Parts: []geminiPart{
				{Text: "You are a pharmacy transaction audit assistant. Return only valid JSON that matches the requested schema."},
			},
		},
		Contents: []geminiContent{
			{
				Role: "user",
				Parts: []geminiPart{
					{Text: buildAuditPrompt(req)},
				},
			},
		},
		GenerationConfig: geminiGenerationConfig{
			Temperature:      0.1,
			ResponseMimeType: "application/json",
			ResponseSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"decision": map[string]interface{}{
						"type": "string",
						"enum": []string{"APPROVED", "REVIEW"},
					},
					"risk_score": map[string]interface{}{
						"type": "number",
					},
					"reason": map[string]interface{}{
						"type": "string",
					},
				},
				"required": []string{"decision", "risk_score", "reason"},
			},
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return domain.AuditTransactionResult{}, fmt.Errorf("marshal gemini request: %w", err)
	}

	url := fmt.Sprintf(geminiGenerateURL, p.model)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return domain.AuditTransactionResult{}, fmt.Errorf("create gemini request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-goog-api-key", p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return domain.AuditTransactionResult{}, fmt.Errorf("gemini http error: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.AuditTransactionResult{}, fmt.Errorf("read gemini response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return domain.AuditTransactionResult{}, fmt.Errorf("gemini status=%d body=%s", resp.StatusCode, truncate(string(respBody), 300))
	}

	var decoded geminiResponse
	if err := json.Unmarshal(respBody, &decoded); err != nil {
		return domain.AuditTransactionResult{}, fmt.Errorf("decode gemini response: %w", err)
	}
	if decoded.PromptFeedback.BlockReason != "" && len(decoded.Candidates) == 0 {
		return domain.AuditTransactionResult{}, fmt.Errorf("gemini blocked prompt: %s", decoded.PromptFeedback.BlockReason)
	}
	if len(decoded.Candidates) == 0 || len(decoded.Candidates[0].Content.Parts) == 0 {
		return domain.AuditTransactionResult{}, errors.New("gemini returned no candidate text")
	}

	raw := strings.TrimSpace(decoded.Candidates[0].Content.Parts[0].Text)
	raw = trimCodeFence(raw)

	var audit geminiAuditJSON
	if err := json.Unmarshal([]byte(raw), &audit); err != nil {
		return domain.AuditTransactionResult{}, fmt.Errorf("decode gemini audit json: %w raw=%s", err, truncate(raw, 200))
	}

	decision := domain.AuditDecision(strings.ToUpper(strings.TrimSpace(audit.Decision)))
	switch decision {
	case domain.AuditDecisionApproved, domain.AuditDecisionReview:
	default:
		return domain.AuditTransactionResult{}, fmt.Errorf("unsupported gemini decision: %s", audit.Decision)
	}

	score := audit.RiskScore
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}

	reason := strings.TrimSpace(audit.Reason)
	if reason == "" {
		reason = "gemini audit result"
	}

	modelName := strings.TrimSpace(decoded.ModelVersion)
	if modelName == "" {
		modelName = p.model
	}

	return domain.AuditTransactionResult{
		Decision:  decision,
		RiskScore: score,
		Reason:    reason,
		Provider:  "gemini",
		Model:     modelName,
	}, nil
}

func buildAuditPrompt(req domain.AuditTransactionRequest) string {
	type promptPayload struct {
		TransactionID string                        `json:"transaction_id"`
		Items         []domain.AuditTransactionItem `json:"items"`
	}

	b, _ := json.Marshal(promptPayload{
		TransactionID: req.TransactionID,
		Items:         req.Items,
	})

	return "" +
		"Review this pharmacy transaction for suspicious behavior.\n" +
		"Return JSON only with keys: decision, risk_score, reason.\n" +
		"Rules:\n" +
		"- decision must be APPROVED or REVIEW\n" +
		"- risk_score must be a number between 0 and 1\n" +
		"- reason must be concise and specific\n" +
		"Transaction JSON:\n" + string(b)
}

func trimCodeFence(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}

func truncate(s string, n int) string {
	if n <= 0 || len(s) <= n {
		return s
	}
	return s[:n]
}