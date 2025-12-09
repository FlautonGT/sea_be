-- Add gateway_code column to payment_channels
-- This stores the gateway-specific code (e.g., "002" for BRI, "014" for BCA)
ALTER TABLE payment_channels 
ADD COLUMN IF NOT EXISTS gateway_code VARCHAR(50) DEFAULT '';

-- Add comments
COMMENT ON COLUMN payment_channels.gateway_code IS 'Gateway-specific code for this payment channel (e.g., bank code "002" for BRI)';

-- Update existing channels with default gateway codes
UPDATE payment_channels SET gateway_code = 'QRIS' WHERE code = 'QRIS';
UPDATE payment_channels SET gateway_code = 'DANA' WHERE code = 'DANA';
UPDATE payment_channels SET gateway_code = 'GOPAY' WHERE code = 'GOPAY';
UPDATE payment_channels SET gateway_code = 'SHOPEEPAY' WHERE code = 'SHOPEEPAY';
UPDATE payment_channels SET gateway_code = '014' WHERE code = 'BCA_VA';
UPDATE payment_channels SET gateway_code = '002' WHERE code = 'BRI_VA';
UPDATE payment_channels SET gateway_code = '008' WHERE code = 'MANDIRI_VA';
UPDATE payment_channels SET gateway_code = '013' WHERE code = 'PERMATA_VA';
UPDATE payment_channels SET gateway_code = 'ALFAMART' WHERE code = 'ALFAMART';
UPDATE payment_channels SET gateway_code = 'INDOMARET' WHERE code = 'INDOMARET';

