package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"finpharm-ai/services/transaction/internal/domain"
)

type StockHTTPRepo struct {
	baseURL string
	client  *http.Client
}

func NewStockHTTPRepo(baseURL string) *StockHTTPRepo {
	return &StockHTTPRepo{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 4 * time.Second}, // safety net
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

	// propagate request-id if present in context? (we don't store it in ctx yet)
	// For now, just mark caller service:
	req.Header.Set("X-From-Service", "transaction")

	resp, err := r.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// parse optional error body; but mapping is enough
		return 0, &domain.NotFoundError{Resource: "medicine", Key: medicineID}
	}

	if resp.StatusCode >= 400 {
		var ie invError
		_ = json.NewDecoder(resp.Body).Decode(&ie)
		return 0, errors.New("inventory error")
	}

	var ok invSuccess
	if err := json.NewDecoder(resp.Body).Decode(&ok); err != nil {
		return 0, err
	}

	return ok.Data.AvailableQty, nil
}