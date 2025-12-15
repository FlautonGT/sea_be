-- Increase payment_code column length to TEXT for long URLs (e.g., DANA redirect URLs with signatures)
ALTER TABLE public.payment_data
ALTER COLUMN payment_code TYPE TEXT;

COMMENT ON COLUMN public.payment_data.payment_code IS 'VA number, QRIS string, redirect URL (can be long with signatures), or retail code';

