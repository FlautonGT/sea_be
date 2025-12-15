-- Add retry_count column to track retry attempts for failed transactions
ALTER TABLE public.transactions
ADD COLUMN IF NOT EXISTS retry_count integer DEFAULT 0;

COMMENT ON COLUMN public.transactions.retry_count IS 'Number of retry attempts for failed transactions';

