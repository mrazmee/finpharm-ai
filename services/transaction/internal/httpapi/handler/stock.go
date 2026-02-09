package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type StockHandler struct {
	// Day 2 masih in-memory. Nanti diganti repository/DB.
	stock map[string]int
}

func NewStockHandler() *StockHandler {
	return &StockHandler{
		stock: map[string]int{
			"AMOX500": 120,
			"PARA500": 80,
			"OBATKERAS-X": 5,
		},
	}
}

type CheckStockRequest struct {
	MedicineID string `json:"medicine_id" binding:"required"`
	Qty        int    `json:"qty" binding:"required,gt=0"`
}

type CheckStockResponse struct {
	MedicineID    string `json:"medicine_id"`
	RequestedQty  int    `json:"requested_qty"`
	AvailableQty  int    `json:"available_qty"`
	IsAvailable   bool   `json:"is_available"`
}

func (h *StockHandler) CheckStock(c *gin.Context) {
	var req CheckStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body", err.Error())
		return
	}
	

	available := h.stock[req.MedicineID] // kalau tidak ada, default 0
	resp := CheckStockResponse{
		MedicineID:   req.MedicineID,
		RequestedQty: req.Qty,
		AvailableQty: available,
		IsAvailable:  available >= req.Qty,
	}

	c.JSON(http.StatusOK, resp)
}
