DROP INDEX IF EXISTS idx_transactions_idempotency_key;

ALTER TABLE transactions
DROP COLUMN IF EXISTS idempotency_key;