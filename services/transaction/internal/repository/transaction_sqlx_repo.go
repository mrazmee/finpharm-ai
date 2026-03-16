package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"finpharm-ai/services/transaction/internal/domain"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type TransactionSQLXRepo struct {
	db *sqlx.DB
}

func NewTransactionSQLXRepo(db *sqlx.DB) *TransactionSQLXRepo {
	return &TransactionSQLXRepo{db: db}
}

type transactionRow struct {
	ID             string    `db:"id"`
	IdempotencyKey string    `db:"idempotency_key"`
	Status         string    `db:"status"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

type transactionItemRow struct {
	ID            int64     `db:"id"`
	TransactionID string    `db:"transaction_id"`
	MedicineID    string    `db:"medicine_id"`
	Qty           int       `db:"qty"`
	CreatedAt     time.Time `db:"created_at"`
}

func (r *TransactionSQLXRepo) Create(ctx context.Context, tx domain.Transaction) (domain.CreateTransactionResult, error) {
	dbtx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return domain.CreateTransactionResult{}, err
	}

	committed := false
	defer func() {
		if !committed {
			_ = dbtx.Rollback()
		}
	}()

	const insertTransaction = `
		INSERT INTO transactions (id, idempotency_key, status)
		VALUES ($1, $2, $3)
		RETURNING id, idempotency_key, status, created_at, updated_at
	`

	var header transactionRow
	if err := dbtx.GetContext(ctx, &header, insertTransaction, tx.ID, tx.IdempotencyKey, string(tx.Status)); err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" && pqErr.Constraint == "idx_transactions_idempotency_key" {
			existing, found, getErr := r.GetByIdempotencyKey(ctx, tx.IdempotencyKey)
			if getErr != nil {
				return domain.CreateTransactionResult{}, getErr
			}
			if found {
				return domain.CreateTransactionResult{
					Transaction: existing,
					IsReplay:    true,
				}, nil
			}
		}
		return domain.CreateTransactionResult{}, err
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
			return domain.CreateTransactionResult{}, err
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
		return domain.CreateTransactionResult{}, err
	}
	committed = true

	return domain.CreateTransactionResult{
		Transaction: domain.Transaction{
			ID:             header.ID,
			IdempotencyKey: header.IdempotencyKey,
			Status:         domain.TransactionStatus(header.Status),
			Items:          items,
			CreatedAt:      header.CreatedAt,
			UpdatedAt:      header.UpdatedAt,
		},
		IsReplay: false,
	}, nil
}

func (r *TransactionSQLXRepo) UpdateStatus(ctx context.Context, transactionID string, status domain.TransactionStatus) error {
	const query = `
		UPDATE transactions
		SET status = $2, updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, transactionID, string(status))
	return err
}

func (r *TransactionSQLXRepo) GetByIdempotencyKey(ctx context.Context, key string) (domain.Transaction, bool, error) {
	const query = `
		SELECT id, idempotency_key, status, created_at, updated_at
		FROM transactions
		WHERE idempotency_key = $1
	`

	var header transactionRow
	if err := r.db.GetContext(ctx, &header, query, key); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Transaction{}, false, nil
		}
		return domain.Transaction{}, false, err
	}

	itemsByTxID, err := r.loadItemsByTransactionIDs(ctx, []string{header.ID})
	if err != nil {
		return domain.Transaction{}, false, err
	}

	return domain.Transaction{
		ID:             header.ID,
		IdempotencyKey: header.IdempotencyKey,
		Status:         domain.TransactionStatus(header.Status),
		Items:          itemsByTxID[header.ID],
		CreatedAt:      header.CreatedAt,
		UpdatedAt:      header.UpdatedAt,
	}, true, nil
}

func (r *TransactionSQLXRepo) List(ctx context.Context, req domain.ListTransactionsRequest) (domain.ListTransactionsResult, error) {
	status := strings.TrimSpace(req.Status)

	const countQuery = `
		SELECT COUNT(*)
		FROM transactions
		WHERE ($1 = '' OR status = $1)
	`

	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, status); err != nil {
		return domain.ListTransactionsResult{}, err
	}

	const listTransactionsQuery = `
		SELECT id, idempotency_key, status, created_at, updated_at
		FROM transactions
		WHERE ($1 = '' OR status = $1)
		ORDER BY created_at DESC, id DESC
		LIMIT $2 OFFSET $3
	`

	var headers []transactionRow
	if err := r.db.SelectContext(ctx, &headers, listTransactionsQuery, status, req.Limit, req.Offset); err != nil {
		return domain.ListTransactionsResult{}, err
	}

	transactions := make([]domain.Transaction, 0, len(headers))
	if len(headers) == 0 {
		return domain.ListTransactionsResult{
			Items:  transactions,
			Limit:  req.Limit,
			Offset: req.Offset,
			Total:  total,
		}, nil
	}

	ids := make([]string, 0, len(headers))
	for _, row := range headers {
		transactions = append(transactions, domain.Transaction{
			ID:             row.ID,
			IdempotencyKey: row.IdempotencyKey,
			Status:         domain.TransactionStatus(row.Status),
			Items:          []domain.TransactionItem{},
			CreatedAt:      row.CreatedAt,
			UpdatedAt:      row.UpdatedAt,
		})
		ids = append(ids, row.ID)
	}

	itemsByTxID, err := r.loadItemsByTransactionIDs(ctx, ids)
	if err != nil {
		return domain.ListTransactionsResult{}, err
	}

	for i := range transactions {
		transactions[i].Items = itemsByTxID[transactions[i].ID]
	}

	return domain.ListTransactionsResult{
		Items:  transactions,
		Limit:  req.Limit,
		Offset: req.Offset,
		Total:  total,
	}, nil
}

func (r *TransactionSQLXRepo) loadItemsByTransactionIDs(ctx context.Context, ids []string) (map[string][]domain.TransactionItem, error) {
	itemsByTxID := make(map[string][]domain.TransactionItem, len(ids))
	if len(ids) == 0 {
		return itemsByTxID, nil
	}

	query, args, err := sqlx.In(`
		SELECT id, transaction_id, medicine_id, qty, created_at
		FROM transaction_items
		WHERE transaction_id IN (?)
		ORDER BY id ASC
	`, ids)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)

	var rows []transactionItemRow
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, err
	}

	for _, row := range rows {
		itemsByTxID[row.TransactionID] = append(itemsByTxID[row.TransactionID], domain.TransactionItem{
			ID:            row.ID,
			TransactionID: row.TransactionID,
			MedicineID:    row.MedicineID,
			Qty:           row.Qty,
			CreatedAt:     row.CreatedAt,
		})
	}

	return itemsByTxID, nil
}