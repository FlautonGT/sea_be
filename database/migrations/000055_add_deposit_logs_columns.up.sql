-- Add payment_logs column to deposits table
-- Similar to transactions table for admin visibility
-- Note: deposits don't need provider_logs as there's no provider processing

ALTER TABLE deposits
ADD COLUMN IF NOT EXISTS payment_logs JSONB DEFAULT '[]'::jsonb;

-- Add index for better query performance
CREATE INDEX IF NOT EXISTS idx_deposits_payment_logs ON deposits USING GIN (payment_logs);

