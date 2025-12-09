-- Ensure min_length and max_length columns exist in product_fields table
-- This migration is idempotent - it will only add columns if they don't exist

DO $$
BEGIN
    -- Add min_length column if it doesn't exist
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'product_fields' AND column_name = 'min_length'
    ) THEN
        ALTER TABLE product_fields ADD COLUMN min_length INT;
    END IF;

    -- Add max_length column if it doesn't exist
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'product_fields' AND column_name = 'max_length'
    ) THEN
        ALTER TABLE product_fields ADD COLUMN max_length INT;
    END IF;
END $$;

