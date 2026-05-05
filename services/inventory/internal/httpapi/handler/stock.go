package handler

import (
	"net/http"

	"finpharm-ai/services/inventory/internal/domain"

	"github.com/gin-gonic/gin"
)

type StockHandler struct {
	uc domain.StockUsecase
}

func NewStockHandler(uc domain.StockUsecase) *StockHandler {
	return &StockHandler{uc: uc}
}

type CheckStockRequest struct {
	MedicineID string `json:"medicine_id" binding:"required"`
	Qty        int    `json:"qty" binding:"required,gt=0"`
}

type CheckStockResponse struct {
	MedicineID   string `json:"medicine_id"`
	RequestedQty int    `json:"requested_qty"`
	AvailableQty int    `json:"available_qty"`
	IsAvailable  bool   `json:"is_available"`
}

type DeductStockRequest struct {
	MedicineID string `json:"medicine_id" binding:"required"`
	Qty        int    `json:"qty" binding:"required,gt=0"`
}

type DeductStockResponse struct {
	MedicineID   string `json:"medicine_id"`
	DeductedQty  int    `json:"deducted_qty"`
	RemainingQty int    `json:"remaining_qty"`
}

func (h *StockHandler) CheckStock(c *gin.Context) {
	var req CheckStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body", err.Error())
		return
	}

	result, err := h.uc.CheckStock(c.Request.Context(), domain.StockCheckRequest{
		MedicineID: req.MedicineID,
		Qty:        req.Qty,
	})
	if err != nil {
		h.handleStockError(c, err, "failed to check stock")
		return
	}

	RespondOK(c, http.StatusOK, CheckStockResponse{
		MedicineID:   result.MedicineID,
		RequestedQty: result.RequestedQty,
		AvailableQty: result.AvailableQty,
		IsAvailable:  result.IsAvailable,
	})
}

func (h *StockHandler) DeductStock(c *gin.Context) {
	var req DeductStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body", err.Error())
		return
	}

	result, err := h.uc.DeductStock(c.Request.Context(), domain.StockDeductRequest{
		MedicineID: req.MedicineID,
		Qty:        req.Qty,
	})
	if err != nil {
		h.handleStockError(c, err, "failed to deduct stock")
		return
	}

	RespondOK(c, http.StatusOK, DeductStockResponse{
		MedicineID:   result.MedicineID,
		DeductedQty:  result.DeductedQty,
		RemainingQty: result.RemainingQty,
	})
}

func (h *StockHandler) handleStockError(c *gin.Context, err error, fallbackMessage string) {
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

	RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", fallbackMessage, nil)
}