-- Drop payment_data table
DROP TRIGGER IF EXISTS update_payment_data_updated_at ON payment_data;
DROP TABLE IF EXISTS public.payment_data;

