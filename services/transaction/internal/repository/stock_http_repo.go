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
	baseURL   string
	requestID string
	client    *http.Client
	callTimeout time.Duration
}

func NewStockHTTPRepo(baseURL string, requestID string) *StockHTTPRepo {
	return &StockHTTPRepo{
		baseURL:     baseURL,
		requestID:   requestID,
		client:      &http.Client{Timeout: 4 * time.Second}, // safety net
		callTimeout: 2 * time.Second,                         // SLA call inventory
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

func (r *StockHTTPRepo) GetAvailableQty(ctx context.Context, medicineID string) (int, error) {
	body, _ := json.Marshal(invReq{MedicineID: medicineID, Qty: 1})

	callCtx, cancel := context.WithTimeout(ctx, r.callTimeout)
	defer cancel()

	url := r.baseURL + "/v1/stock/check"
	req, err := http.NewRequestWithContext(callCtx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	req.Header.Set("X-Caller-Service", "transaction")

	if r.requestID != "" {
		req.Header.Set("X-Request-ID", r.requestID)
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