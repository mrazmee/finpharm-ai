package handler

import (
	"context"
	"io"
	"net/http"
	"time"

	"finpharm-ai/services/gateway/internal/httpapi/middleware"

	"github.com/gin-gonic/gin"
)

type InventoryProxyHandler struct {
	baseURL string
	client  *http.Client
}

func NewInventoryProxyHandler(inventoryBaseURL string) *InventoryProxyHandler {
	return &InventoryProxyHandler{
		baseURL: inventoryBaseURL,
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

// GET /v1/medicines?limit=&offset=
func (h *InventoryProxyHandler) ListMedicines(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	url := h.baseURL + "/v1/medicines"
	if q := c.Request.URL.RawQuery; q != "" {
		url += "?" + q
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "GATEWAY_ERROR", "failed to create upstream request", nil)
		return
	}

	ridVal, _ := c.Get(middleware.CtxKeyRequestID)
	rid, _ := ridVal.(string)
	req.Header.Set(middleware.HeaderRequestID, rid)

	req.Header.Set("X-Caller-Service", "gateway")

	resp, err := h.client.Do(req)
	if err != nil {
		RespondError(c, http.StatusBadGateway, "UPSTREAM_ERROR", "inventory service unreachable", err.Error())
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		ct = "application/json"
	}
	c.Data(resp.StatusCode, ct, body)
}

// GET /v1/medicines/:id
func (h *InventoryProxyHandler) GetMedicine(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	id := c.Param("id")
	url := h.baseURL + "/v1/medicines/" + id

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "GATEWAY_ERROR", "failed to create upstream request", nil)
		return
	}

	ridVal, _ := c.Get(middleware.CtxKeyRequestID)
	rid, _ := ridVal.(string)
	req.Header.Set(middleware.HeaderRequestID, rid)

	req.Header.Set("X-Caller-Service", "gateway")

	resp, err := h.client.Do(req)
	if err != nil {
		RespondError(c, http.StatusBadGateway, "UPSTREAM_ERROR", "inventory service unreachable", err.Error())
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		ct = "application/json"
	}
	c.Data(resp.StatusCode, ct, body)
}