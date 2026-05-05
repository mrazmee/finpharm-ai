ALTER TABLE transactions
ADD COLUMN IF NOT EXISTS idempotency_key VARCHAR(100);

UPDATE transactions
SET idempotency_key = 'legacy-' || id
WHERE idempotency_key IS NULL OR idempotency_key = '';

ALTER TABLE transactions
ALTER COLUMN idempotency_key SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_transactions_idempotency_key
ON transactions (idempotency_key);