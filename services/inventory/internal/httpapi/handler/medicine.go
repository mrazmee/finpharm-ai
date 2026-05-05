package handler

import (
	"net/http"
	"strconv"

	"finpharm-ai/services/inventory/internal/domain"

	"github.com/gin-gonic/gin"
)

type MedicineHandler struct {
	uc domain.MedicineUsecase
}

func NewMedicineHandler(uc domain.MedicineUsecase) *MedicineHandler {
	return &MedicineHandler{uc: uc}
}

type MedicineDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type ListMedicinesResponse struct {
	Items  []MedicineDTO `json:"items"`
	Limit  int           `json:"limit"`
	Offset int           `json:"offset"`
	Total  int           `json:"total"`
}

func (h *MedicineHandler) ListMedicines(c *gin.Context) {
	limit := parseIntDefault(c.Query("limit"), 10)
	offset := parseIntDefault(c.Query("offset"), 0)

	result, err := h.uc.ListMedicines(c.Request.Context(), domain.ListMedicinesQuery{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list medicines", nil)
		return
	}

	items := make([]MedicineDTO, 0, len(result.Items))
	for _, m := range result.Items {
		items = append(items, MedicineDTO{
			ID:   m.ID,
			Name: m.Name,
			Type: m.Type,
		})
	}

	RespondOK(c, http.StatusOK, ListMedicinesResponse{
		Items:  items,
		Limit:  result.Limit,
		Offset: result.Offset,
		Total:  result.Total,
	})
}

func (h *MedicineHandler) GetMedicine(c *gin.Context) {
	id := c.Param("id")

	m, err := h.uc.GetMedicine(c.Request.Context(), id)
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
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get medicine", nil)
		return
	}

	RespondOK(c, http.StatusOK, MedicineDTO{
		ID:   m.ID,
		Name: m.Name,
		Type: m.Type,
	})
}

func parseIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}