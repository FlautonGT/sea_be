-- Make category_id nullable in payment_channels table
-- This allows payment channels like QRIS and BALANCE to exist without a category

ALTER TABLE payment_channels 
  ALTER COLUMN category_id DROP NOT NULL;

