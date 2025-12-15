-- Remove retry_count column
ALTER TABLE public.transactions
DROP COLUMN IF EXISTS retry_count;

