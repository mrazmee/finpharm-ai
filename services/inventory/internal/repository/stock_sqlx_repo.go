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