package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"finpharm-ai/services/gateway/internal/httpapi/middleware"

	"github.com/gin-gonic/gin"
)

type StockProxyHandler struct {
	baseURL string
	client  *http.Client
}

func NewStockProxyHandler() *StockProxyHandler {
	baseURL := os.Getenv("TRANSACTION_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8081"
	}

	return &StockProxyHandler{
		baseURL: baseURL,
		client: &http.Client{Timeout: 3 * time.Second},
	}
}

type CheckStockRequest struct {
	MedicineID string `json:"medicine_id" binding:"required"`
	Qty        int    `json:"qty" binding:"required,gt=0"`
}

func (h *StockProxyHandler) CheckStock(c *gin.Context) {
	var req CheckStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body", err.Error())
		return
	}

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "GATEWAY_ERROR", "failed to encode request", nil)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	url := h.baseURL + "/v1/stock/check"
	upReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "GATEWAY_ERROR", "failed to create upstream request", nil)
		return
	}

	upReq.Header.Set("Content-Type", "application/json")

	// ✅ propagate request-id (sekarang selalu ada dari middleware)
	ridVal, _ := c.Get(middleware.CtxKeyRequestID)
	rid, _ := ridVal.(string)
	upReq.Header.Set(middleware.HeaderRequestID, rid)

	// ✅ PR kamu: tandai asal request dari gateway
	upReq.Header.Set("X-From-Gateway", "finpharm-gateway")

	resp, err := h.client.Do(upReq)
	if err != nil {
		RespondError(c, http.StatusBadGateway, "UPSTREAM_ERROR", "transactions service unreachable", err.Error())
		return
	}
	defer resp.Body.Close()

	respBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		RespondError(c, http.StatusBadGateway, "UPSTREAM_ERROR", "failed to read upstream response", nil)
		return
	}

	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		ct = "application/json"
	}

	// pass-through response upstream (kalau error format upstream sudah standar, client dapat format itu)
	c.Data(resp.StatusCode, ct, respBody)
}
