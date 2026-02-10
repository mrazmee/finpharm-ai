package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"finpharm-ai/services/gateway/internal/httpapi/middleware"

	"github.com/gin-gonic/gin"
)

type StockProxyHandler struct {
	baseURL string
	client  *http.Client
}

func NewStockProxyHandler(transactionBaseURL string) *StockProxyHandler {
	return &StockProxyHandler{
		baseURL: transactionBaseURL,
		client:  &http.Client{Timeout: 3 * time.Second},
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

	ridVal, _ := c.Get(middleware.CtxKeyRequestID)
	rid, _ := ridVal.(string)
	upReq.Header.Set(middleware.HeaderRequestID, rid)

	upReq.Header.Set("X-From-Gateway", "finpharm-gateway")

	resp, err := h.client.Do(upReq)
	if err != nil {
		RespondError(c, http.StatusBadGateway, "UPSTREAM_ERROR", "transaction service unreachable", err.Error())
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

	c.Data(resp.StatusCode, ct, respBody)
}
