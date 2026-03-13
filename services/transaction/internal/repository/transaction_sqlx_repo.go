package repository

import (
	"context"
	"time"

	"finpharm-ai/services/transaction/internal/domain"

	"github.com/jmoiron/sqlx"
)

type TransactionSQLXRepo struct {
	db *sqlx.DB
}

func NewTransactionSQLXRepo(db *sqlx.DB) *TransactionSQLXRepo {
	return &TransactionSQLXRepo{db: db}
}

type transactionRow struct {
	ID        string    `db:"id"`
	Status    string    `db:"status"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type transactionItemRow struct {
	ID            int64     `db:"id"`
	TransactionID string    `db:"transaction_id"`
	MedicineID    string    `db:"medicine_id"`
	Qty           int       `db:"qty"`
	CreatedAt     time.Time `db:"created_at"`
}

func (r *TransactionSQLXRepo) Create(ctx context.Context, tx domain.Transaction) (domain.Transaction, error) {
	dbtx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return domain.Transaction{}, err
	}

	committed := false
	defer func() {
		if !committed {
			_ = dbtx.Rollback()
		}
	}()

	const insertTransaction = `
		INSERT INTO transactions (id, status)
		VALUES ($1, $2)
		RETURNING id, status, created_at, updated_at
	`

	var header transactionRow
	if err := dbtx.GetContext(ctx, &header, insertTransaction, tx.ID, string(tx.Status)); err != nil {
		return domain.Transaction{}, err
	}

	const insertItem = `
		INSERT INTO transaction_items (transaction_id, medicine_id, qty)
		VALUES ($1, $2, $3)
		RETURNING id, transaction_id, medicine_id, qty, created_at
	`

	items := make([]domain.TransactionItem, 0, len(tx.Items))
	for _, item := range tx.Items {
		var row transactionItemRow
		if err := dbtx.GetContext(ctx, &row, insertItem, header.ID, item.MedicineID, item.Qty); err != nil {
			return domain.Transaction{}, err
		}

		items = append(items, domain.TransactionItem{
			ID:            row.ID,
			TransactionID: row.TransactionID,
			MedicineID:    row.MedicineID,
			Qty:           row.Qty,
			CreatedAt:     row.CreatedAt,
		})
	}

	if err := dbtx.Commit(); err != nil {
		return domain.Transaction{}, err
	}
	committed = true

	return domain.Transaction{
		ID:        header.ID,
		Status:    domain.TransactionStatus(header.Status),
		Items:     items,
		CreatedAt: header.CreatedAt,
		UpdatedAt: header.UpdatedAt,
	}, nil
}
