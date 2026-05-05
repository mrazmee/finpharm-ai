package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"finpharm-ai/services/transaction/internal/domain"
)

type AIAuditorHTTPRepo struct {
	baseURL     string
	client      *http.Client
	callTimeout time.Duration
}

func NewAIAuditorHTTPRepo(baseURL string, client *http.Client, timeout time.Duration) *AIAuditorHTTPRepo {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	if client == nil {
		client = &http.Client{Timeout: timeout + time.Second}
	}

	return &AIAuditorHTTPRepo{
		baseURL:     strings.TrimRight(baseURL, "/"),
		client:      client,
		callTimeout: timeout,
	}
}

type aiAuditRequest struct {
	TransactionID string               `json:"transaction_id"`
	Items         []aiAuditRequestItem `json:"items"`
}

type aiAuditRequestItem struct {
	MedicineID string `json:"medicine_id"`
	Qty        int    `json:"qty"`
}

type aiAuditSuccessResponse struct {
	Data struct {
		Decision  string  `json:"decision"`
		RiskScore float64 `json:"risk_score"`
		Reason    string  `json:"reason"`
		Provider  string  `json:"provider"`
		Model     string  `json:"model"`
	} `json:"data"`
	RequestID string `json:"request_id"`
}

type aiAuditErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Details struct {
			Field  string `json:"field"`
			Reason string `json:"reason"`
		} `json:"details"`
		RequestID string `json:"request_id"`
	} `json:"error"`
}

func (r *AIAuditorHTTPRepo) AuditTransaction(ctx context.Context, req domain.AuditTransactionRequest) (domain.AuditTransactionResult, error) {
	items := make([]aiAuditRequestItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, aiAuditRequestItem{
			MedicineID: item.MedicineID,
			Qty:        item.Qty,
		})
	}

	body, err := json.Marshal(aiAuditRequest{
		TransactionID: req.TransactionID,
		Items:         items,
	})
	if err != nil {
		return domain.AuditTransactionResult{}, fmt.Errorf("marshal ai audit request: %w", err)
	}

	callCtx, cancel := context.WithTimeout(ctx, r.callTimeout)
	defer cancel()

	url := r.baseURL + "/v1/audit/transaction"
	httpReq, err := http.NewRequestWithContext(callCtx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return domain.AuditTransactionResult{}, &domain.UpstreamError{
			Service: "ai-auditor",
			Reason:  err.Error(),
		}
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Caller-Service", "transaction")
	applyCommonHeadersFromContext(callCtx, httpReq)

	resp, err := r.client.Do(httpReq)
	if err != nil {
		return domain.AuditTransactionResult{}, &domain.UpstreamError{
			Service: "ai-auditor",
			Reason:  err.Error(),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		if resp.StatusCode == http.StatusBadRequest {
			var er aiAuditErrorResponse
			_ = json.NewDecoder(resp.Body).Decode(&er)
			return domain.AuditTransactionResult{}, &domain.ValidationError{
				Field:  er.Error.Details.Field,
				Reason: er.Error.Details.Reason,
			}
		}

		return domain.AuditTransactionResult{}, &domain.UpstreamError{
			Service: "ai-auditor",
			Reason:  fmt.Sprintf("status=%d", resp.StatusCode),
		}
	}

	var ok aiAuditSuccessResponse
	if err := json.NewDecoder(resp.Body).Decode(&ok); err != nil {
		return domain.AuditTransactionResult{}, &domain.UpstreamError{
			Service: "ai-auditor",
			Reason:  "invalid json response",
		}
	}

	decision := domain.AuditDecision(strings.ToUpper(strings.TrimSpace(ok.Data.Decision)))
	switch decision {
	case domain.AuditDecisionApproved, domain.AuditDecisionReview:
	default:
		return domain.AuditTransactionResult{}, &domain.UpstreamError{
			Service: "ai-auditor",
			Reason:  "invalid audit decision",
		}
	}

	return domain.AuditTransactionResult{
		Decision:  decision,
		RiskScore: ok.Data.RiskScore,
		Reason:    ok.Data.Reason,
		Provider:  ok.Data.Provider,
		Model:     ok.Data.Model,
	}, nil
}