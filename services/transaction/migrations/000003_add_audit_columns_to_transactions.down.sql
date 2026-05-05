ALTER TABLE transactions
DROP COLUMN IF EXISTS audited_at,
DROP COLUMN IF EXISTS audit_model,
DROP COLUMN IF EXISTS audit_provider,
DROP COLUMN IF EXISTS audit_reason,
DROP COLUMN IF EXISTS audit_risk_score,
DROP COLUMN IF EXISTS audit_decision;