package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"finpharm-ai/services/transaction/internal/domain"
)

type StockHTTPRepo struct {
	baseURL   string
	requestID string
	client    *http.Client
}

func NewStockHTTPRepo(baseURL string, requestID string) *StockHTTPRepo {
	return &StockHTTPRepo{
		baseURL:   baseURL,
		requestID: requestID,
		client:    &http.Client{Timeout: 4 * time.Second}, // safety net
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

type invError struct {
	Error struct {
		Code      string      `json:"code"`
		Message   string      `json:"message"`
		Details   interface{} `json:"details"`
		RequestID string      `json:"request_id"`
	} `json:"error"`
}

func (r *StockHTTPRepo) GetAvailableQty(ctx context.Context, medicineID string) (int, error) {
	body, _ := json.Marshal(invReq{MedicineID: medicineID, Qty: 1})

	url := r.baseURL + "/v1/stock/check"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-From-Service", "transaction")

	// ✅ propagate request-id
	if r.requestID != "" {
		req.Header.Set("X-Request-ID", r.requestID)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("inventory unreachable: %w", err)
	}
	defer resp.Body.Close()

	// Handle non-2xx by reading inventory error body if possible
	if resp.StatusCode >= 400 {
		var ie invError
		_ = json.NewDecoder(resp.Body).Decode(&ie)

		switch resp.StatusCode {
		case http.StatusNotFound:
			return 0, &domain.NotFoundError{Resource: "medicine", Key: medicineID}
		case http.StatusBadRequest:
			// map to validation error (keep it generic)
			return 0, &domain.ValidationError{Field: "request", Reason: "invalid"}
		default:
			// 5xx or others
			if ie.Error.Message != "" {
				return 0, errors.New(ie.Error.Message)
			}
			return 0, fmt.Errorf("inventory error status=%d", resp.StatusCode)
		}
	}

	var ok invSuccess
	if err := json.NewDecoder(resp.Body).Decode(&ok); err != nil {
		return 0, err
	}

	return ok.Data.AvailableQty, nil
}