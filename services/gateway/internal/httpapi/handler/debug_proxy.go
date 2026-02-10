package handler

import (
	"context"
	"io"
	"net/http"
	"time"

	"finpharm-ai/services/gateway/internal/httpapi/middleware"

	"github.com/gin-gonic/gin"
)

type DebugProxyHandler struct {
	baseURL string
	client  *http.Client
}

func NewDebugProxyHandler(transactionBaseURL string) *DebugProxyHandler {
	return &DebugProxyHandler{
		baseURL: transactionBaseURL,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// GET /v1/debug/sleep?ms=3000 -> forward to transaction
func (h *DebugProxyHandler) Sleep(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	url := h.baseURL + "/v1/debug/sleep?ms=" + c.Query("ms")
	upReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "GATEWAY_ERROR", "failed to create upstream request", nil)
		return
	}

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
