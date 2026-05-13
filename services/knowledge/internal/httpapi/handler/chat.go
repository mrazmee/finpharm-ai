package handler

import (
	"net/http"
	"strings"

	"finpharm-ai/services/knowledge/internal/chat"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	service *chat.Service
}

func NewChatHandler(service *chat.Service) *ChatHandler {
	return &ChatHandler{service: service}
}

type ChatSOPRequest struct {
	Question string   `json:"question"`
	TopK     *int     `json:"top_k,omitempty"`
	MinScore *float64 `json:"min_score,omitempty"`
}

func (h *ChatHandler) HandleSOPChat(c *gin.Context) {
	var req ChatSOPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body", err.Error())
		return
	}

	question := strings.TrimSpace(req.Question)
	if question == "" {
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

	result, err := h.service.Answer(c.Request.Context(), chat.Request{
		Question: question,
		TopK:     topK,
		MinScore: minScore,
	})
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "KNOWLEDGE_CHAT_ERROR", "failed to generate SOP answer", nil)
		return
	}

	RespondOK(c, http.StatusOK, result)
}