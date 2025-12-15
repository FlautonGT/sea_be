-- Remove payment_logs and provider_logs columns
ALTER TABLE public.transactions
DROP COLUMN IF EXISTS payment_logs,
DROP COLUMN IF EXISTS provider_logs;

