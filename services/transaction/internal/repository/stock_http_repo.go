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

type invCheckSuccess struct {
	Data struct {
		MedicineID   string `json:"medicine_id"`
		RequestedQty int    `json:"requested_qty"`
		AvailableQty int    `json:"available_qty"`
		IsAvailable  bool   `json:"is_available"`
	} `json:"data"`
	RequestID string `json:"request_id"`
}

type invDeductSuccess struct {
	Data struct {
		MedicineID   string `json:"medicine_id"`
		DeductedQty  int    `json:"deducted_qty"`
		RemainingQty int    `json:"remaining_qty"`
	} `json:"data"`
	RequestID string `json:"request_id"`
}

type invErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Details struct {
			Field        string `json:"field"`
			Reason       string `json:"reason"`
			MedicineID   string `json:"medicine_id"`
			RequestedQty int    `json:"requested_qty"`
			AvailableQty int    `json:"available_qty"`
		} `json:"details"`
	} `json:"error"`
}

func (r *StockHTTPRepo) GetAvailableQty(ctx context.Context, medicineID string, requestedQty int) (int, error) {
	return r.getAvailableQty(ctx, medicineID, requestedQty)
}

func (r *StockHTTPRepo) getAvailableQty(ctx context.Context, medicineID string, requestedQty int) (int, error) {
	now := time.Now()
	if !r.breaker.Allow(now) {
		return 0, &domain.UpstreamError{Service: "inventory", Reason: "circuit breaker open"}
	}

	var lastErr error
	for attempt := 0; attempt <= r.retries; attempt++ {
		qty, err := r.callInventoryCheck(ctx, medicineID, requestedQty)
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
		if _, ok := domain.IsInsufficientStock(err); ok {
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

func (r *StockHTTPRepo) DeductStock(ctx context.Context, medicineID string, qty int) error {
	now := time.Now()
	if !r.breaker.Allow(now) {
		return &domain.UpstreamError{Service: "inventory", Reason: "circuit breaker open"}
	}

	err := r.callInventoryDeduct(ctx, medicineID, qty)
	if err != nil {
		if _, ok := domain.IsNotFound(err); ok {
			r.breaker.OnSuccess(time.Now())
			return err
		}
		if _, ok := domain.IsValidation(err); ok {
			r.breaker.OnSuccess(time.Now())
			return err
		}
		if _, ok := domain.IsInsufficientStock(err); ok {
			r.breaker.OnSuccess(time.Now())
			return err
		}
		r.breaker.OnFailure(time.Now())
		return err
	}

	r.breaker.OnSuccess(time.Now())
	return nil
}

func (r *StockHTTPRepo) callInventoryCheck(ctx context.Context, medicineID string, requestedQty int) (int, error) {
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
	applyCommonHeadersFromContext(callCtx, req)

	resp, err := r.client.Do(req)
	if err != nil {
		return 0, &domain.UpstreamError{Service: "inventory", Reason: err.Error()}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return 0, decodeInventoryError(resp, medicineID, requestedQty)
	}

	var ok invCheckSuccess
	if err := json.NewDecoder(resp.Body).Decode(&ok); err != nil {
		return 0, &domain.UpstreamError{Service: "inventory", Reason: "invalid json response"}
	}

	return ok.Data.AvailableQty, nil
}

func (r *StockHTTPRepo) callInventoryDeduct(ctx context.Context, medicineID string, qty int) error {
	body, err := json.Marshal(invReq{MedicineID: medicineID, Qty: qty})
	if err != nil {
		return fmt.Errorf("marshal inventory request: %w", err)
	}

	callCtx, cancel := context.WithTimeout(ctx, r.callTimeout)
	defer cancel()

	url := r.baseURL + "/v1/stock/deduct"
	req, err := http.NewRequestWithContext(callCtx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Caller-Service", "transaction")
	applyCommonHeadersFromContext(callCtx, req)

	resp, err := r.client.Do(req)
	if err != nil {
		return &domain.UpstreamError{Service: "inventory", Reason: err.Error()}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return decodeInventoryError(resp, medicineID, qty)
	}

	var ok invDeductSuccess
	if err := json.NewDecoder(resp.Body).Decode(&ok); err != nil {
		return &domain.UpstreamError{Service: "inventory", Reason: "invalid json response"}
	}

	return nil
}

func decodeInventoryError(resp *http.Response, medicineID string, qty int) error {
	switch resp.StatusCode {
	case http.StatusNotFound:
		return &domain.NotFoundError{
			Resource: "medicine",
			Key:      medicineID,
		}
	case http.StatusBadRequest:
		var er invErrorResponse
		_ = json.NewDecoder(resp.Body).Decode(&er)
		return &domain.ValidationError{
			Field:  er.Error.Details.Field,
			Reason: er.Error.Details.Reason,
		}
	case http.StatusConflict:
		var er invErrorResponse
		_ = json.NewDecoder(resp.Body).Decode(&er)
		return &domain.InsufficientStockError{
			MedicineID:   medicineID,
			RequestedQty: qty,
			AvailableQty: er.Error.Details.AvailableQty,
		}
	default:
		return &domain.UpstreamError{Service: "inventory", Reason: fmt.Sprintf("status=%d", resp.StatusCode)}
	}
}