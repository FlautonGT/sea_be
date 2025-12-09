-- Add inquiry_slug column to products table
ALTER TABLE products ADD COLUMN IF NOT EXISTS inquiry_slug VARCHAR(100);

-- Add index for faster lookups
CREATE INDEX IF NOT EXISTS idx_products_inquiry_slug ON products(inquiry_slug) WHERE inquiry_slug IS NOT NULL;

