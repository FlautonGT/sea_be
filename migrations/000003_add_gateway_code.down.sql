-- Remove gateway_code column from payment_channels
ALTER TABLE payment_channels DROP COLUMN IF EXISTS gateway_code;

