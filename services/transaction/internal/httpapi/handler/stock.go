package handler

import (
	"net/http"

	"finpharm-ai/services/transaction/internal/domain"
	"finpharm-ai/services/transaction/internal/httpapi/middleware"
	"finpharm-ai/services/transaction/internal/repository"

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

	ridVal, _ := c.Get(middleware.CtxKeyRequestID)
	rid, _ := ridVal.(string)

	ctx := repository.WithRequestID(c.Request.Context(), rid)

	result, err := h.uc.CheckStock(ctx, domain.StockCheckRequest{
		MedicineID: req.MedicineID,
		Qty:        req.Qty,
	})
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
		if ue, ok := domain.IsUpstream(err); ok {
			RespondError(c, http.StatusBadGateway, "UPSTREAM_ERROR", "inventory service error", gin.H{
				"service": ue.Service,
				"reason":  ue.Reason,
			})
			return
		}

		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to check stock", err.Error())
		return
	}

	RespondOK(c, http.StatusOK, CheckStockResponse{
		MedicineID:   result.MedicineID,
		RequestedQty: result.RequestedQty,
		AvailableQty: result.AvailableQty,
		IsAvailable:  result.IsAvailable,
	})
}
