package repository

import (
	"context"

	"finpharm-ai/services/inventory/internal/domain"

	"github.com/jmoiron/sqlx"
)

type StockSQLXRepo struct {
	db *sqlx.DB
}

func NewStockSQLXRepo(db *sqlx.DB) *StockSQLXRepo {
	return &StockSQLXRepo{db: db}
}

type stockRow struct {
	MedicineID   string `db:"medicine_id"`
	AvailableQty int    `db:"available_qty"`
}

func (r *StockSQLXRepo) GetAvailableQty(ctx context.Context, medicineID string) (int, error) {
	var row stockRow

	query := `
		SELECT medicine_id, available_qty
		FROM stocks
		WHERE medicine_id = $1
	`

	err := r.db.GetContext(ctx, &row, query, medicineID)
	if err != nil {
		if isNoRows(err) {
			return 0, &domain.NotFoundError{
				Resource: "medicine",
				Key:      medicineID,
			}
		}
		return 0, err
	}

	return row.AvailableQty, nil
}

func (r *StockSQLXRepo) DeductStock(ctx context.Context, medicineID string, qty int) (int, error) {
	const deductQuery = `
		UPDATE stocks
		SET available_qty = available_qty - $2
		WHERE medicine_id = $1
		  AND available_qty >= $2
		RETURNING available_qty
	`

	var remaining int
	err := r.db.GetContext(ctx, &remaining, deductQuery, medicineID, qty)
	if err == nil {
		return remaining, nil
	}

	if !isNoRows(err) {
		return 0, err
	}

	available, getErr := r.GetAvailableQty(ctx, medicineID)
	if getErr != nil {
		return 0, getErr
	}

	return 0, &domain.InsufficientStockError{
		MedicineID:   medicineID,
		RequestedQty: qty,
		AvailableQty: available,
	}
}