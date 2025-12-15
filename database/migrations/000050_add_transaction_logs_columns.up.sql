-- Add payment_logs and provider_logs columns to transactions table
-- These store timestamped logs from payment gateway and provider callbacks

ALTER TABLE public.transactions
ADD COLUMN IF NOT EXISTS payment_logs text DEFAULT '',
ADD COLUMN IF NOT EXISTS provider_logs text DEFAULT '';

COMMENT ON COLUMN public.transactions.payment_logs IS 'Timestamped logs from payment gateway (creation, callbacks, etc)';
COMMENT ON COLUMN public.transactions.provider_logs IS 'Timestamped logs from provider (Digiflazz, etc) callbacks and responses';

