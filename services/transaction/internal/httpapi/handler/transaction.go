package handler

import (
	"net/http"
	"strconv"
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

type TransactionResponse struct {
	ID        string                          `json:"id"`
	Status    string                          `json:"status"`
	Items     []CreateTransactionItemResponse `json:"items"`
	CreatedAt time.Time                       `json:"created_at"`
}

type CreateTransactionItemResponse struct {
	MedicineID string `json:"medicine_id"`
	Qty        int    `json:"qty"`
}

type ListTransactionsResponse struct {
	Items  []TransactionResponse `json:"items"`
	Limit  int                   `json:"limit"`
	Offset int                   `json:"offset"`
	Total  int                   `json:"total"`
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
		h.handleUsecaseError(c, err, "failed to create transaction")
		return
	}

	RespondOK(c, http.StatusCreated, toTransactionResponse(result))
}

func (h *TransactionHandler) ListTransactions(c *gin.Context) {
	var req domain.ListTransactionsRequest

	if limitStr := c.Query("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "validation failed", gin.H{
				"field":  "limit",
				"reason": "must be an integer",
			})
			return
		}
		if limit <= 0 {
			RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "validation failed", gin.H{
				"field":  "limit",
				"reason": "must be > 0",
			})
			return
		}
		if limit > 100 {
			RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "validation failed", gin.H{
				"field":  "limit",
				"reason": "must be <= 100",
			})
			return
		}
		req.Limit = limit
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "validation failed", gin.H{
				"field":  "offset",
				"reason": "must be an integer",
			})
			return
		}
		if offset < 0 {
			RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "validation failed", gin.H{
				"field":  "offset",
				"reason": "must be >= 0",
			})
			return
		}
		req.Offset = offset
	}

	req.Status = c.Query("status")

	ridVal, _ := c.Get(middleware.CtxKeyRequestID)
	rid, _ := ridVal.(string)

	ctx := repository.WithRequestID(c.Request.Context(), rid)

	result, err := h.uc.ListTransactions(ctx, req)
	if err != nil {
		h.handleUsecaseError(c, err, "failed to list transactions")
		return
	}

	items := make([]TransactionResponse, 0, len(result.Items))
	for _, tx := range result.Items {
		items = append(items, toTransactionResponse(tx))
	}

	RespondOK(c, http.StatusOK, ListTransactionsResponse{
		Items:  items,
		Limit:  result.Limit,
		Offset: result.Offset,
		Total:  result.Total,
	})
}

func (h *TransactionHandler) handleUsecaseError(c *gin.Context, err error, fallbackMessage string) {
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

	RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", fallbackMessage, err.Error())
}

func toTransactionResponse(tx domain.Transaction) TransactionResponse {
	respItems := make([]CreateTransactionItemResponse, 0, len(tx.Items))
	for _, item := range tx.Items {
		respItems = append(respItems, CreateTransactionItemResponse{
			MedicineID: item.MedicineID,
			Qty:        item.Qty,
		})
	}

	return TransactionResponse{
		ID:        tx.ID,
		Status:    string(tx.Status),
		Items:     respItems,
		CreatedAt: tx.CreatedAt,
	}
}