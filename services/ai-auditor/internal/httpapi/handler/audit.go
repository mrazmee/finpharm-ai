package handler

import (
	"net/http"

	"finpharm-ai/services/ai-auditor/internal/domain"

	"github.com/gin-gonic/gin"
)

type AuditHandler struct {
	uc domain.AuditUsecase
}

func NewAuditHandler(uc domain.AuditUsecase) *AuditHandler {
	return &AuditHandler{uc: uc}
}

type AuditTransactionRequest struct {
	TransactionID string                        `json:"transaction_id"`
	Items         []AuditTransactionItemRequest `json:"items"`
}

type AuditTransactionItemRequest struct {
	MedicineID string `json:"medicine_id"`
	Qty        int    `json:"qty"`
}

type AuditTransactionResponse struct {
	Decision  string  `json:"decision"`
	RiskScore float64 `json:"risk_score"`
	Reason    string  `json:"reason"`
	Provider  string  `json:"provider"`
	Model     string  `json:"model"`
}

func (h *AuditHandler) AuditTransaction(c *gin.Context) {
	var req AuditTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body", err.Error())
		return
	}

	items := make([]domain.AuditTransactionItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, domain.AuditTransactionItem{
			MedicineID: item.MedicineID,
			Qty:        item.Qty,
		})
	}

	result, err := h.uc.AuditTransaction(c.Request.Context(), domain.AuditTransactionRequest{
		TransactionID: req.TransactionID,
		Items:         items,
	})
	if err != nil {
		if ve, ok := domain.IsValidation(err); ok {
			RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "validation failed", gin.H{
				"field":  ve.Field,
				"reason": ve.Reason,
			})
			return
		}
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to audit transaction", nil)
		return
	}

	RespondOK(c, http.StatusOK, AuditTransactionResponse{
		Decision:  string(result.Decision),
		RiskScore: result.RiskScore,
		Reason:    result.Reason,
		Provider:  result.Provider,
		Model:     result.Model,
	})
}