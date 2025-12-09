-- Remove min_length and max_length columns from product_fields table
-- Note: This will remove the columns if they exist

ALTER TABLE product_fields DROP COLUMN IF EXISTS min_length;
ALTER TABLE product_fields DROP COLUMN IF EXISTS max_length;

