-- Revert payment_logs and provider_logs back to text
ALTER TABLE public.transactions
ALTER COLUMN payment_logs TYPE text USING '',
ALTER COLUMN payment_logs SET DEFAULT '';

ALTER TABLE public.transactions
ALTER COLUMN provider_logs TYPE text USING '',
ALTER COLUMN provider_logs SET DEFAULT '';

