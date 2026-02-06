package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"time"

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
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

type CheckStockRequest struct {
	MedicineID string `json:"medicine_id" binding:"required"`
	Qty        int    `json:"qty" binding:"required,gt=0"`
}

func (h *StockProxyHandler) CheckStock(c *gin.Context) {
	// 1) Validasi input di gateway juga (best practice: fail fast)
	var req CheckStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "invalid request body",
				"details": err.Error(),
			},
		})
		return
	}

	// 2) Serialize request untuk dikirim ke transaction service
	bodyBytes, _ := json.Marshal(req)

	// 3) Buat request ke transaction service
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	url := h.baseURL + "/v1/stock/check"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "GATEWAY_ERROR", "message": "failed to create upstream request"},
		})
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Forward request-id untuk tracing (kalau ada)
	if rid := c.GetHeader("X-Request-ID"); rid != "" {
		httpReq.Header.Set("X-Request-ID", rid)
	}

	// Forward from-gateway untuk perbedaan asal request
	httpReq.Header.Set("X-From-Gateway", "finpharm-gateway")

	// 4) Call upstream
	resp, err := h.client.Do(httpReq)
	if err != nil {
		// Bedakan timeout vs error lain
		if errors.Is(err, context.DeadlineExceeded) {
			c.JSON(http.StatusGatewayTimeout, gin.H{
				"error": gin.H{"code": "UPSTREAM_TIMEOUT", "message": "transaction service timeout"},
			})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{
			"error": gin.H{"code": "UPSTREAM_ERROR", "message": "transaction service unreachable", "details": err.Error()},
		})
		return
	}
	defer resp.Body.Close()

	// 5) Relay response apa adanya (simple gateway pattern)
	respBody, _ := io.ReadAll(resp.Body)
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}
