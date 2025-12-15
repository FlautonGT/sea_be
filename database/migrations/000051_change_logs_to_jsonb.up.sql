-- Change payment_logs and provider_logs from text to JSONB array
-- This stores full JSON responses/callbacks from payment gateway and provider

-- First drop defaults
ALTER TABLE public.transactions
ALTER COLUMN payment_logs DROP DEFAULT;

ALTER TABLE public.transactions
ALTER COLUMN provider_logs DROP DEFAULT;

-- Change type with proper conversion
ALTER TABLE public.transactions
ALTER COLUMN payment_logs TYPE jsonb USING '[]'::jsonb;

ALTER TABLE public.transactions
ALTER COLUMN provider_logs TYPE jsonb USING '[]'::jsonb;

-- Set new defaults
ALTER TABLE public.transactions
ALTER COLUMN payment_logs SET DEFAULT '[]'::jsonb;

ALTER TABLE public.transactions
ALTER COLUMN provider_logs SET DEFAULT '[]'::jsonb;

COMMENT ON COLUMN public.transactions.payment_logs IS 'Array of full JSON responses/callbacks from payment gateway';
COMMENT ON COLUMN public.transactions.provider_logs IS 'Array of full JSON responses/callbacks from provider (Digiflazz, etc)';
