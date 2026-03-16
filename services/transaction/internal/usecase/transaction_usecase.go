package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"finpharm-ai/services/transaction/internal/domain"
)

type TransactionUsecase struct {
	repo      domain.TransactionRepository
	stockRepo domain.StockRepository
}

func NewTransactionUsecase(repo domain.TransactionRepository, stockRepo domain.StockRepository) *TransactionUsecase {
	return &TransactionUsecase{
		repo:      repo,
		stockRepo: stockRepo,
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

	tx := domain.Transaction{
		ID:             generateTransactionID(time.Now()),
		IdempotencyKey: idempotencyKey,
		Status:         domain.TransactionStatusPending,
		Items:          items,
	}

	return u.repo.Create(ctx, tx)
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