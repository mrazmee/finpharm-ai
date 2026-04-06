package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"finpharm-ai/services/transaction/internal/domain"
)

const flaggedRiskThreshold = 0.85

type TransactionUsecase struct {
	repo      domain.TransactionRepository
	stockRepo domain.StockRepository
	auditRepo domain.AIAuditorRepository
	publisher domain.TransactionEventPublisher
}

func NewTransactionUsecase(
	repo domain.TransactionRepository,
	stockRepo domain.StockRepository,
	auditRepo domain.AIAuditorRepository,
	publisher domain.TransactionEventPublisher,
) *TransactionUsecase {
	return &TransactionUsecase{
		repo:      repo,
		stockRepo: stockRepo,
		auditRepo: auditRepo,
		publisher: publisher,
	}
}

func (u *TransactionUsecase) CreateTransaction(ctx context.Context, req domain.CreateTransactionRequest) (domain.CreateTransactionResult, error) {
	idempotencyKey := strings.TrimSpace(req.IdempotencyKey)
	if idempotencyKey == "" {
		return domain.CreateTransactionResult{}, &domain.ValidationError{
			Field:  "idempotency_key",
			Reason: "is required",
		}
	}
	if len(idempotencyKey) > 100 {
		return domain.CreateTransactionResult{}, &domain.ValidationError{
			Field:  "idempotency_key",
			Reason: "must be <= 100 characters",
		}
	}

	existing, found, err := u.repo.GetByIdempotencyKey(ctx, idempotencyKey)
	if err != nil {
		return domain.CreateTransactionResult{}, err
	}
	if found {
		return domain.CreateTransactionResult{
			Transaction: existing,
			IsReplay:    true,
		}, nil
	}

	if len(req.Items) == 0 {
		return domain.CreateTransactionResult{}, &domain.ValidationError{
			Field:  "items",
			Reason: "must contain at least 1 item",
		}
	}

	seen := make(map[string]struct{}, len(req.Items))
	items := make([]domain.TransactionItem, 0, len(req.Items))

	for idx, item := range req.Items {
		medicineID := strings.TrimSpace(item.MedicineID)
		if medicineID == "" {
			return domain.CreateTransactionResult{}, &domain.ValidationError{
				Field:  fmt.Sprintf("items[%d].medicine_id", idx),
				Reason: "is required",
			}
		}
		if item.Qty <= 0 {
			return domain.CreateTransactionResult{}, &domain.ValidationError{
				Field:  fmt.Sprintf("items[%d].qty", idx),
				Reason: "must be > 0",
			}
		}
		if _, exists := seen[medicineID]; exists {
			return domain.CreateTransactionResult{}, &domain.ValidationError{
				Field:  "items",
				Reason: fmt.Sprintf("duplicate medicine_id: %s", medicineID),
			}
		}
		seen[medicineID] = struct{}{}

		availableQty, err := u.stockRepo.GetAvailableQty(ctx, medicineID, item.Qty)
		if err != nil {
			return domain.CreateTransactionResult{}, err
		}
		if availableQty < item.Qty {
			return domain.CreateTransactionResult{}, &domain.InsufficientStockError{
				MedicineID:   medicineID,
				RequestedQty: item.Qty,
				AvailableQty: availableQty,
			}
		}

		items = append(items, domain.TransactionItem{
			MedicineID: medicineID,
			Qty:        item.Qty,
		})
	}

	createResult, err := u.repo.Create(ctx, domain.Transaction{
		ID:             generateTransactionID(time.Now()),
		IdempotencyKey: idempotencyKey,
		Status:         domain.TransactionStatusPending,
		Items:          items,
	})
	if err != nil {
		return domain.CreateTransactionResult{}, err
	}
	if createResult.IsReplay {
		return createResult, nil
	}

	tx := createResult.Transaction

	if u.auditRepo == nil {
		audit := domain.TransactionAudit{
			Decision:  domain.AuditDecisionReview,
			RiskScore: 0.50,
			Reason:    "ai auditor is not configured; manual review required",
			Provider:  "system",
			Model:     "fallback-pending-review",
			AuditedAt: time.Now().UTC(),
		}
		if err := u.repo.UpdateAudit(ctx, tx.ID, audit); err != nil {
			return domain.CreateTransactionResult{}, err
		}
		return u.markReviewStatus(ctx, tx, audit, domain.TransactionStatusPendingReview)
	}

	auditItems := make([]domain.AuditTransactionItem, 0, len(tx.Items))
	for _, item := range tx.Items {
		auditItems = append(auditItems, domain.AuditTransactionItem{
			MedicineID: item.MedicineID,
			Qty:        item.Qty,
		})
	}

	auditResult, err := u.auditRepo.AuditTransaction(ctx, domain.AuditTransactionRequest{
		TransactionID: tx.ID,
		Items:         auditItems,
	})
	if err != nil {
		audit := domain.TransactionAudit{
			Decision:  domain.AuditDecisionReview,
			RiskScore: 0.50,
			Reason:    "ai auditor unavailable; manual review required",
			Provider:  "system",
			Model:     "fallback-pending-review",
			AuditedAt: time.Now().UTC(),
		}
		if err := u.repo.UpdateAudit(ctx, tx.ID, audit); err != nil {
			return domain.CreateTransactionResult{}, err
		}
		return u.markReviewStatus(ctx, tx, audit, domain.TransactionStatusPendingReview)
	}

	audit := domain.TransactionAudit{
		Decision:  auditResult.Decision,
		RiskScore: auditResult.RiskScore,
		Reason:    auditResult.Reason,
		Provider:  auditResult.Provider,
		Model:     auditResult.Model,
		AuditedAt: time.Now().UTC(),
	}
	if err := u.repo.UpdateAudit(ctx, tx.ID, audit); err != nil {
		return domain.CreateTransactionResult{}, err
	}

	if auditResult.Decision == domain.AuditDecisionReview {
		status := domain.TransactionStatusPendingReview
		if auditResult.RiskScore >= flaggedRiskThreshold {
			status = domain.TransactionStatusFlagged
		}
		return u.markReviewStatus(ctx, tx, audit, status)
	}

	for _, item := range tx.Items {
		if err := u.stockRepo.DeductStock(ctx, item.MedicineID, item.Qty); err != nil {
			if updateErr := u.repo.UpdateStatus(ctx, tx.ID, domain.TransactionStatusFailed); updateErr != nil {
				return domain.CreateTransactionResult{}, updateErr
			}
			tx.Status = domain.TransactionStatusFailed
			tx.Audit = &audit
			return domain.CreateTransactionResult{
				Transaction: tx,
				IsReplay:    false,
			}, err
		}
	}

	if err := u.repo.UpdateStatus(ctx, tx.ID, domain.TransactionStatusApproved); err != nil {
		return domain.CreateTransactionResult{}, err
	}

	tx.Status = domain.TransactionStatusApproved
	tx.Audit = &audit

	u.publishApprovedEvent(ctx, tx)

	return domain.CreateTransactionResult{
		Transaction: tx,
		IsReplay:    false,
	}, nil
}

func (u *TransactionUsecase) publishApprovedEvent(ctx context.Context, tx domain.Transaction) {
	if u.publisher == nil {
		return
	}

	event := domain.TransactionApprovedEvent{
		EventName:      "transaction.approved",
		TransactionID:  tx.ID,
		IdempotencyKey: tx.IdempotencyKey,
		Status:         string(tx.Status),
		Items:          make([]domain.TransactionApprovedEventItem, 0, len(tx.Items)),
		CreatedAt:      tx.CreatedAt,
		PublishedAt:    time.Now().UTC(),
	}

	for _, item := range tx.Items {
		event.Items = append(event.Items, domain.TransactionApprovedEventItem{
			MedicineID: item.MedicineID,
			Qty:        item.Qty,
		})
	}

	if tx.Audit != nil {
		event.Audit = &domain.TransactionApprovedEventAudit{
			Decision:  string(tx.Audit.Decision),
			RiskScore: tx.Audit.RiskScore,
			Reason:    tx.Audit.Reason,
			Provider:  tx.Audit.Provider,
			Model:     tx.Audit.Model,
			AuditedAt: tx.Audit.AuditedAt,
		}
	}

	if err := u.publisher.PublishTransactionApproved(ctx, event); err != nil {
		slog.Warn("publish_transaction_approved_failed",
			"transaction_id", tx.ID,
			"event_name", event.EventName,
			"error", err,
		)
	}
}

func (u *TransactionUsecase) markReviewStatus(ctx context.Context, tx domain.Transaction, audit domain.TransactionAudit, status domain.TransactionStatus) (domain.CreateTransactionResult, error) {
	if err := u.repo.UpdateStatus(ctx, tx.ID, status); err != nil {
		return domain.CreateTransactionResult{}, err
	}
	tx.Status = status
	tx.Audit = &audit
	return domain.CreateTransactionResult{
		Transaction: tx,
		IsReplay:    false,
	}, nil
}

func (u *TransactionUsecase) ListTransactions(ctx context.Context, req domain.ListTransactionsRequest) (domain.ListTransactionsResult, error) {
	limit := req.Limit
	if limit == 0 {
		limit = 10
	}
	if limit < 0 {
		return domain.ListTransactionsResult{}, &domain.ValidationError{
			Field:  "limit",
			Reason: "must be >= 0",
		}
	}
	if limit > 100 {
		return domain.ListTransactionsResult{}, &domain.ValidationError{
			Field:  "limit",
			Reason: "must be <= 100",
		}
	}

	offset := req.Offset
	if offset < 0 {
		return domain.ListTransactionsResult{}, &domain.ValidationError{
			Field:  "offset",
			Reason: "must be >= 0",
		}
	}

	status := strings.ToUpper(strings.TrimSpace(req.Status))

	return u.repo.List(ctx, domain.ListTransactionsRequest{
		Limit:  limit,
		Offset: offset,
		Status: status,
	})
}

func generateTransactionID(now time.Time) string {
	stamp := now.UTC().Format("20060102150405")
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("TXN-%s-%d", stamp, now.UTC().UnixNano())
	}
	return fmt.Sprintf("TXN-%s-%s", stamp, strings.ToUpper(hex.EncodeToString(b[:])))
}