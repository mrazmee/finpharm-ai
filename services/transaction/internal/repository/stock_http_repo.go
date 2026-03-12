package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"finpharm-ai/services/transaction/internal/domain"
)

type StockHTTPRepo struct {
	baseURL string
	client  *http.Client

	callTimeout time.Duration

	breaker *CircuitBreaker
	retries int
}

func NewStockHTTPRepo(baseURL string, client *http.Client, breaker *CircuitBreaker) *StockHTTPRepo {
	if client == nil {
		client = &http.Client{Timeout: 4 * time.Second}
	}
	if breaker == nil {
		breaker = NewCircuitBreaker(3, 5*time.Second)
	}
	return &StockHTTPRepo{
		baseURL:     baseURL,
		client:      client,
		callTimeout: 2 * time.Second,
		breaker:     breaker,
		retries:     1,
	}
}

type invReq struct {
	MedicineID string `json:"medicine_id"`
	Qty        int    `json:"qty"`
}

type invSuccess struct {
	Data struct {
		MedicineID   string `json:"medicine_id"`
		RequestedQty int    `json:"requested_qty"`
		AvailableQty int    `json:"available_qty"`
		IsAvailable  bool   `json:"is_available"`
	} `json:"data"`
	RequestID string `json:"request_id"`
}

func (r *StockHTTPRepo) GetAvailableQty(ctx context.Context, medicineID string, requestedQty int) (int, error) {
	rid := RequestIDFromContext(ctx)
	return r.getAvailableQty(ctx, medicineID, requestedQty, rid)
}

func (r *StockHTTPRepo) getAvailableQty(ctx context.Context, medicineID string, requestedQty int, requestID string) (int, error) {
	now := time.Now()
	if !r.breaker.Allow(now) {
		return 0, &domain.UpstreamError{Service: "inventory", Reason: "circuit breaker open"}
	}

	var lastErr error
	for attempt := 0; attempt <= r.retries; attempt++ {
		qty, err := r.callInventory(ctx, medicineID, requestedQty, requestID)
		if err == nil {
			r.breaker.OnSuccess(time.Now())
			return qty, nil
		}

		if _, ok := domain.IsNotFound(err); ok {
			r.breaker.OnSuccess(time.Now())
			return 0, err
		}
		if _, ok := domain.IsValidation(err); ok {
			r.breaker.OnSuccess(time.Now())
			return 0, err
		}

		lastErr = err
		if _, ok := domain.IsUpstream(err); !ok {
			break
		}

		time.Sleep(time.Duration(50*(attempt+1)) * time.Millisecond)
	}

	r.breaker.OnFailure(time.Now())
	return 0, lastErr
}

func (r *StockHTTPRepo) callInventory(ctx context.Context, medicineID string, requestedQty int, requestID string) (int, error) {
	body, err := json.Marshal(invReq{MedicineID: medicineID, Qty: requestedQty})
	if err != nil {
		return 0, fmt.Errorf("marshal inventory request: %w", err)
	}

	callCtx, cancel := context.WithTimeout(ctx, r.callTimeout)
	defer cancel()

	url := r.baseURL + "/v1/stock/check"
	req, err := http.NewRequestWithContext(callCtx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Caller-Service", "transaction")
	if requestID != "" {
		req.Header.Set("X-Request-ID", requestID)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return 0, &domain.UpstreamError{Service: "inventory", Reason: err.Error()}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		switch resp.StatusCode {
		case http.StatusNotFound:
			return 0, &domain.NotFoundError{Resource: "medicine", Key: medicineID}
		case http.StatusBadRequest:
			return 0, &domain.ValidationError{Field: "request", Reason: "invalid"}
		default:
			return 0, &domain.UpstreamError{Service: "inventory", Reason: fmt.Sprintf("status=%d", resp.StatusCode)}
		}
	}

	var ok invSuccess
	if err := json.NewDecoder(resp.Body).Decode(&ok); err != nil {
		return 0, &domain.UpstreamError{Service: "inventory", Reason: "invalid json response"}
	}

	return ok.Data.AvailableQty, nil
}