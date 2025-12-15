-- Add backup provider SKU codes for retry mechanism
-- When primary SKU fails with certain RC codes, system will retry with backup SKUs

ALTER TABLE public.skus
ADD COLUMN IF NOT EXISTS provider_sku_code_backup1 character varying(100),
ADD COLUMN IF NOT EXISTS provider_sku_code_backup2 character varying(100);

-- Add comment for documentation
COMMENT ON COLUMN public.skus.provider_sku_code_backup1 IS 'First backup SKU code for retry on failed transactions';
COMMENT ON COLUMN public.skus.provider_sku_code_backup2 IS 'Second backup SKU code for retry on failed transactions';

