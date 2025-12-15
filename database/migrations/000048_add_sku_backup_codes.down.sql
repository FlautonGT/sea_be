-- Remove backup provider SKU codes
ALTER TABLE public.skus
DROP COLUMN IF EXISTS provider_sku_code_backup1,
DROP COLUMN IF EXISTS provider_sku_code_backup2;

