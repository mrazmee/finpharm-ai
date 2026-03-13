package handler

import (
	"net/http"
	"time"

	"finpharm-ai/services/transaction/internal/domain"
	"finpharm-ai/services/transaction/internal/httpapi/middleware"
	"finpharm-ai/services/transaction/internal/repository"

	"github.com/gin-gonic/gin"
)

type TransactionHandler struct {
	uc domain.TransactionUsecase
}

func NewTransactionHandler(uc domain.TransactionUsecase) *TransactionHandler {
	return &TransactionHandler{uc: uc}
}

type CreateTransactionRequest struct {
	Items []CreateTransactionItemRequest `json:"items" binding:"required"`
}

type CreateTransactionItemRequest struct {
	MedicineID string `json:"medicine_id"`
	Qty        int    `json:"qty"`
}

type CreateTransactionResponse struct {
	ID        string                          `json:"id"`
	Status    string                          `json:"status"`
	Items     []CreateTransactionItemResponse `json:"items"`
	CreatedAt time.Time                       `json:"created_at"`
}

type CreateTransactionItemResponse struct {
	MedicineID string `json:"medicine_id"`
	Qty        int    `json:"qty"`
}

func (h *TransactionHandler) CreateTransaction(c *gin.Context) {
	var req CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body", err.Error())
		return
	}

	ridVal, _ := c.Get(middleware.CtxKeyRequestID)
	rid, _ := ridVal.(string)

	ctx := repository.WithRequestID(c.Request.Context(), rid)

	items := make([]domain.TransactionItemInput, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, domain.TransactionItemInput{
			MedicineID: item.MedicineID,
			Qty:        item.Qty,
		})
	}

	result, err := h.uc.CreateTransaction(ctx, domain.CreateTransactionRequest{Items: items})
	if err != nil {
		if ve, ok := domain.IsValidation(err); ok {
			RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "validation failed", gin.H{
				"field":  ve.Field,
				"reason": ve.Reason,
			})
			return
		}
		if nf, ok := domain.IsNotFound(err); ok {
			RespondError(c, http.StatusNotFound, "MEDICINE_NOT_FOUND", "medicine not found", gin.H{
				"resource": nf.Resource,
				"key":      nf.Key,
			})
			return
		}
		if ie, ok := domain.IsInsufficientStock(err); ok {
			RespondError(c, http.StatusConflict, "INSUFFICIENT_STOCK", "stock is not enough", gin.H{
				"medicine_id":   ie.MedicineID,
				"requested_qty": ie.RequestedQty,
				"available_qty": ie.AvailableQty,
			})
			return
		}
		if ue, ok := domain.IsUpstream(err); ok {
			RespondError(c, http.StatusBadGateway, "UPSTREAM_ERROR", "inventory service error", gin.H{
				"service": ue.Service,
				"reason":  ue.Reason,
			})
			return
		}

		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create transaction", err.Error())
		return
	}

	respItems := make([]CreateTransactionItemResponse, 0, len(result.Items))
	for _, item := range result.Items {
		respItems = append(respItems, CreateTransactionItemResponse{
			MedicineID: item.MedicineID,
			Qty:        item.Qty,
		})
	}

	RespondOK(c, http.StatusCreated, CreateTransactionResponse{
		ID:        result.ID,
		Status:    string(result.Status),
		Items:     respItems,
		CreatedAt: result.CreatedAt,
	})
}
