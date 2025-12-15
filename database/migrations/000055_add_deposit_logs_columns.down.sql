-- Remove payment_logs column from deposits table

DROP INDEX IF EXISTS idx_deposits_payment_logs;

ALTER TABLE deposits
DROP COLUMN IF EXISTS payment_logs;

