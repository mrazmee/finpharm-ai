package repository

import (
	"context"
	"database/sql"
	"errors"
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
	ID             string     `db:"id"`
	IdempotencyKey string     `db:"idempotency_key"`
	Status         string     `db:"status"`
	AuditDecision  *string    `db:"audit_decision"`
	AuditRiskScore *float64   `db:"audit_risk_score"`
	AuditReason    *string    `db:"audit_reason"`
	AuditProvider  *string    `db:"audit_provider"`
	AuditModel     *string    `db:"audit_model"`
	AuditedAt      *time.Time `db:"audited_at"`
	CreatedAt      time.Time  `db:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at"`
}

type transactionItemRow struct {
	ID            int64     `db:"id"`
	TransactionID string    `db:"transaction_id"`
	MedicineID    string    `db:"medicine_id"`
	Qty           int       `db:"qty"`
	CreatedAt     time.Time `db:"created_at"`
}

func (r *TransactionSQLXRepo) Create(ctx context.Context, tx domain.Transaction) (domain.CreateTransactionResult, error) {
	now := time.Now().UTC()
	if tx.CreatedAt.IsZero() {
		tx.CreatedAt = now
	}
	tx.UpdatedAt = tx.CreatedAt

	dbtx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return domain.CreateTransactionResult{}, err
	}
	defer func() {
		_ = dbtx.Rollback()
	}()

	_, err = dbtx.ExecContext(
		ctx,
		`INSERT INTO transactions (id, idempotency_key, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		tx.ID,
		tx.IdempotencyKey,
		string(tx.Status),
		tx.CreatedAt,
		tx.UpdatedAt,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
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

	for i := range tx.Items {
		itemCreatedAt := now
		_, err := dbtx.ExecContext(
			ctx,
			`INSERT INTO transaction_items (transaction_id, medicine_id, qty, created_at)
			 VALUES ($1, $2, $3, $4)`,
			tx.ID,
			tx.Items[i].MedicineID,
			tx.Items[i].Qty,
			itemCreatedAt,
		)
		if err != nil {
			return domain.CreateTransactionResult{}, err
		}
		tx.Items[i].TransactionID = tx.ID
		tx.Items[i].CreatedAt = itemCreatedAt
	}

	if err := dbtx.Commit(); err != nil {
		return domain.CreateTransactionResult{}, err
	}

	return domain.CreateTransactionResult{
		Transaction: tx,
		IsReplay:    false,
	}, nil
}

func (r *TransactionSQLXRepo) UpdateStatus(ctx context.Context, transactionID string, status domain.TransactionStatus) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE transactions
		 SET status = $2, updated_at = NOW()
		 WHERE id = $1`,
		transactionID,
		string(status),
	)
	return err
}

func (r *TransactionSQLXRepo) UpdateAudit(ctx context.Context, transactionID string, audit domain.TransactionAudit) error {
	auditedAt := audit.AuditedAt
	if auditedAt.IsZero() {
		auditedAt = time.Now().UTC()
	}

	_, err := r.db.ExecContext(
		ctx,
		`UPDATE transactions
		 SET audit_decision = $2,
		     audit_risk_score = $3,
		     audit_reason = $4,
		     audit_provider = $5,
		     audit_model = $6,
		     audited_at = $7,
		     updated_at = NOW()
		 WHERE id = $1`,
		transactionID,
		string(audit.Decision),
		audit.RiskScore,
		audit.Reason,
		audit.Provider,
		audit.Model,
		auditedAt,
	)
	return err
}

func (r *TransactionSQLXRepo) GetByIdempotencyKey(ctx context.Context, key string) (domain.Transaction, bool, error) {
	var row transactionRow
	err := r.db.GetContext(
		ctx,
		&row,
		`SELECT id, idempotency_key, status,
		        audit_decision, audit_risk_score, audit_reason, audit_provider, audit_model, audited_at,
		        created_at, updated_at
		   FROM transactions
		  WHERE idempotency_key = $1`,
		key,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Transaction{}, false, nil
		}
		return domain.Transaction{}, false, err
	}

	items, err := r.getItemsByTransactionID(ctx, row.ID)
	if err != nil {
		return domain.Transaction{}, false, err
	}

	return toDomainTransaction(row, items), true, nil
}

func (r *TransactionSQLXRepo) List(ctx context.Context, req domain.ListTransactionsRequest) (domain.ListTransactionsResult, error) {
	var total int
	var rows []transactionRow
	var err error

	if req.Status == "" {
		err = r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM transactions`)
		if err != nil {
			return domain.ListTransactionsResult{}, err
		}

		err = r.db.SelectContext(
			ctx,
			&rows,
			`SELECT id, idempotency_key, status,
			        audit_decision, audit_risk_score, audit_reason, audit_provider, audit_model, audited_at,
			        created_at, updated_at
			   FROM transactions
			  ORDER BY created_at DESC
			  LIMIT $1 OFFSET $2`,
			req.Limit,
			req.Offset,
		)
		if err != nil {
			return domain.ListTransactionsResult{}, err
		}
	} else {
		err = r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM transactions WHERE status = $1`, req.Status)
		if err != nil {
			return domain.ListTransactionsResult{}, err
		}

		err = r.db.SelectContext(
			ctx,
			&rows,
			`SELECT id, idempotency_key, status,
			        audit_decision, audit_risk_score, audit_reason, audit_provider, audit_model, audited_at,
			        created_at, updated_at
			   FROM transactions
			  WHERE status = $1
			  ORDER BY created_at DESC
			  LIMIT $2 OFFSET $3`,
			req.Status,
			req.Limit,
			req.Offset,
		)
		if err != nil {
			return domain.ListTransactionsResult{}, err
		}
	}

	items := make([]domain.Transaction, 0, len(rows))
	for _, row := range rows {
		txItems, err := r.getItemsByTransactionID(ctx, row.ID)
		if err != nil {
			return domain.ListTransactionsResult{}, err
		}
		items = append(items, toDomainTransaction(row, txItems))
	}

	return domain.ListTransactionsResult{
		Items:  items,
		Limit:  req.Limit,
		Offset: req.Offset,
		Total:  total,
	}, nil
}

func (r *TransactionSQLXRepo) getItemsByTransactionID(ctx context.Context, transactionID string) ([]transactionItemRow, error) {
	var rows []transactionItemRow
	err := r.db.SelectContext(
		ctx,
		&rows,
		`SELECT id, transaction_id, medicine_id, qty, created_at
		   FROM transaction_items
		  WHERE transaction_id = $1
		  ORDER BY id ASC`,
		transactionID,
	)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func toDomainTransaction(row transactionRow, items []transactionItemRow) domain.Transaction {
	txItems := make([]domain.TransactionItem, 0, len(items))
	for _, item := range items {
		txItems = append(txItems, domain.TransactionItem{
			ID:            item.ID,
			TransactionID: item.TransactionID,
			MedicineID:    item.MedicineID,
			Qty:           item.Qty,
			CreatedAt:     item.CreatedAt,
		})
	}

	var audit *domain.TransactionAudit
	if row.AuditDecision != nil || row.AuditRiskScore != nil || row.AuditReason != nil || row.AuditProvider != nil || row.AuditModel != nil || row.AuditedAt != nil {
		audit = &domain.TransactionAudit{}
		if row.AuditDecision != nil {
			audit.Decision = domain.AuditDecision(*row.AuditDecision)
		}
		if row.AuditRiskScore != nil {
			audit.RiskScore = *row.AuditRiskScore
		}
		if row.AuditReason != nil {
			audit.Reason = *row.AuditReason
		}
		if row.AuditProvider != nil {
			audit.Provider = *row.AuditProvider
		}
		if row.AuditModel != nil {
			audit.Model = *row.AuditModel
		}
		if row.AuditedAt != nil {
			audit.AuditedAt = *row.AuditedAt
		}
	}

	return domain.Transaction{
		ID:             row.ID,
		IdempotencyKey: row.IdempotencyKey,
		Status:         domain.TransactionStatus(row.Status),
		Items:          txItems,
		Audit:          audit,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}
}