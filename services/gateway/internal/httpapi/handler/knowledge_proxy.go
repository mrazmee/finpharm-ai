package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type KnowledgeProxyHandler struct {
	baseURL string
	client  *http.Client
}

type ChatSOPRequest struct {
	Question string   `json:"question"`
	TopK     *int     `json:"top_k,omitempty"`
	MinScore *float64 `json:"min_score,omitempty"`
}

func NewKnowledgeProxyHandler(knowledgeBaseURL string) *KnowledgeProxyHandler {
	return &KnowledgeProxyHandler{
		baseURL: knowledgeBaseURL,
		client:  &http.Client{Timeout: 8 * time.Second},
	}
}

func (h *KnowledgeProxyHandler) ChatSOP(c *gin.Context) {
	var req ChatSOPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body", err.Error())
		return
	}

	req.Question = strings.TrimSpace(req.Question)
	if req.Question == "" {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "validation failed", gin.H{
			"field":  "question",
			"reason": "is required",
		})
		return
	}

	topK := 5
	if req.TopK != nil {
		topK = *req.TopK
	}
	if topK <= 0 {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "validation failed", gin.H{
			"field":  "top_k",
			"reason": "must be > 0",
		})
		return
	}

	minScore := 0.45
	if req.MinScore != nil {
		minScore = *req.MinScore
	}
	if minScore < 0 || minScore > 1 {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "validation failed", gin.H{
			"field":  "min_score",
			"reason": "must be between 0 and 1",
		})
		return
	}

	bodyBytes, err := json.Marshal(gin.H{
		"question":  req.Question,
		"top_k":     topK,
		"min_score": minScore,
	})
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "GATEWAY_ERROR", "failed to encode request", nil)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 8*time.Second)
	defer cancel()

	url := h.baseURL + "/v1/chat/sop"
	upReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "GATEWAY_ERROR", "failed to create upstream request", nil)
		return
	}

	upReq.Header.Set("Content-Type", "application/json")
	setProxyForwardHeaders(c, upReq)

	resp, err := h.client.Do(upReq)
	if err != nil {
		RespondError(c, http.StatusBadGateway, "UPSTREAM_ERROR", "knowledge service unreachable", err.Error())
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