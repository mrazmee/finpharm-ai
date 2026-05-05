package repository

import (
	"context"

	"finpharm-ai/services/inventory/internal/domain"

	"github.com/jmoiron/sqlx"
)

type MedicineSQLXRepo struct {
	db *sqlx.DB
}

func NewMedicineSQLXRepo(db *sqlx.DB) *MedicineSQLXRepo {
	return &MedicineSQLXRepo{db: db}
}

type medicineRow struct {
	ID   string `db:"id"`
	Name string `db:"name"`
	Type string `db:"type"`
}

func (r *MedicineSQLXRepo) List(ctx context.Context, q domain.ListMedicinesQuery) (domain.ListMedicinesResult, error) {
	limit := q.Limit
	offset := q.Offset

	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	var total int
	if err := r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM medicines`); err != nil {
		return domain.ListMedicinesResult{}, err
	}

	rows := make([]medicineRow, 0)
	query := `
		SELECT id, name, type
		FROM medicines
		ORDER BY id
		LIMIT $1 OFFSET $2
	`
	if err := r.db.SelectContext(ctx, &rows, query, limit, offset); err != nil {
		return domain.ListMedicinesResult{}, err
	}

	items := make([]domain.Medicine, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.Medicine{
			ID:   row.ID,
			Name: row.Name,
			Type: row.Type,
		})
	}

	return domain.ListMedicinesResult{
		Items:  items,
		Limit:  limit,
		Offset: offset,
		Total:  total,
	}, nil
}

func (r *MedicineSQLXRepo) GetByID(ctx context.Context, id string) (domain.Medicine, error) {
	var row medicineRow

	query := `
		SELECT id, name, type
		FROM medicines
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &row, query, id)
	if err != nil {
		if isNoRows(err) {
			return domain.Medicine{}, &domain.NotFoundError{
				Resource: "medicine",
				Key:      id,
			}
		}
		return domain.Medicine{}, err
	}

	return domain.Medicine{
		ID:   row.ID,
		Name: row.Name,
		Type: row.Type,
	}, nil
}
