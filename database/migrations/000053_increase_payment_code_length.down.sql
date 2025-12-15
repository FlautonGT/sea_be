-- Revert payment_code column to VARCHAR(500)
-- Note: This may fail if there are existing values longer than 500 chars
ALTER TABLE public.payment_data
ALTER COLUMN payment_code TYPE VARCHAR(500);

