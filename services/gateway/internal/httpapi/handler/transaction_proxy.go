package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"finpharm-ai/services/gateway/internal/httpapi/middleware"

	"github.com/gin-gonic/gin"
)

type TransactionProxyHandler struct {
	baseURL string
	client  *http.Client
}

func NewTransactionProxyHandler(transactionBaseURL string) *TransactionProxyHandler {
	return &TransactionProxyHandler{
		baseURL: transactionBaseURL,
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

type CreateTransactionRequest struct {
	Items []CreateTransactionItemRequest `json:"items" binding:"required"`
}

type CreateTransactionItemRequest struct {
	MedicineID string `json:"medicine_id"`
	Qty        int    `json:"qty"`
}

func (h *TransactionProxyHandler) CreateTransaction(c *gin.Context) {
	var req CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body", err.Error())
		return
	}

	if len(req.Items) == 0 {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "validation failed", gin.H{
			"field":  "items",
			"reason": "must contain at least 1 item",
		})
		return
	}

	for i, item := range req.Items {
		if strings.TrimSpace(item.MedicineID) == "" {
			RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "validation failed", gin.H{
				"field":  "items[" + strconv.Itoa(i) + "].medicine_id",
				"reason": "is required",
			})
			return
		}
		if item.Qty <= 0 {
			RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "validation failed", gin.H{
				"field":  "items[" + strconv.Itoa(i) + "].qty",
				"reason": "must be > 0",
			})
			return
		}
	}

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "GATEWAY_ERROR", "failed to encode request", nil)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	url := h.baseURL + "/v1/transactions"
	upReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "GATEWAY_ERROR", "failed to create upstream request", nil)
		return
	}

	upReq.Header.Set("Content-Type", "application/json")

	ridVal, _ := c.Get(middleware.CtxKeyRequestID)
	rid, _ := ridVal.(string)
	upReq.Header.Set(middleware.HeaderRequestID, rid)
	upReq.Header.Set("X-Caller-Service", "gateway")

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