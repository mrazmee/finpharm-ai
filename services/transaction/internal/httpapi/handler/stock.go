package handler

import (
	"net/http"

	"finpharm-ai/services/transaction/internal/domain"

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
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to check stock", err.Error())
		return
	}

	c.JSON(http.StatusOK, CheckStockResponse{
		MedicineID:   result.MedicineID,
		RequestedQty: result.RequestedQty,
		AvailableQty: result.AvailableQty,
		IsAvailable:  result.IsAvailable,
	})
}
