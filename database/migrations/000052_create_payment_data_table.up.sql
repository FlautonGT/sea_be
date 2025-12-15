-- Create payment_data table to store payment gateway responses
CREATE TABLE IF NOT EXISTS public.payment_data (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    payment_channel_id UUID NOT NULL REFERENCES payment_channels(id),
    gateway_id UUID REFERENCES payment_gateways(id),
    
    -- Invoice reference
    invoice_number VARCHAR(50) NOT NULL,
    
    -- Payment details
    payment_code TEXT, -- VA number, QRIS string, redirect URL (can be long with signatures), retail code
    payment_type VARCHAR(50), -- VIRTUAL_ACCOUNT, QRIS, E_WALLET, RETAIL, BALANCE
    gateway_ref_id VARCHAR(255), -- Reference ID from gateway
    
    -- Virtual Account specific
    bank_code VARCHAR(20),
    account_name VARCHAR(200),
    
    -- Amount details
    amount BIGINT NOT NULL,
    fee BIGINT DEFAULT 0,
    total_amount BIGINT NOT NULL,
    currency VARCHAR(10) DEFAULT 'IDR',
    
    -- Status
    status VARCHAR(20) DEFAULT 'PENDING', -- PENDING, PAID, EXPIRED, FAILED, REFUNDED
    
    -- Timestamps
    expired_at TIMESTAMPTZ,
    paid_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    -- Raw gateway response for audit
    raw_request JSONB,
    raw_response JSONB
);

-- Indexes
CREATE INDEX idx_payment_data_transaction ON payment_data(transaction_id);
CREATE INDEX idx_payment_data_invoice ON payment_data(invoice_number);
CREATE INDEX idx_payment_data_gateway_ref ON payment_data(gateway_ref_id);
CREATE INDEX idx_payment_data_status ON payment_data(status);

-- Trigger for updated_at
CREATE TRIGGER update_payment_data_updated_at BEFORE UPDATE ON payment_data
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Comments
COMMENT ON TABLE public.payment_data IS 'Stores payment gateway response data for transactions';
COMMENT ON COLUMN public.payment_data.payment_code IS 'VA number, QRIS string, redirect URL, or retail code';
COMMENT ON COLUMN public.payment_data.payment_type IS 'Type of payment: VIRTUAL_ACCOUNT, QRIS, E_WALLET, RETAIL, BALANCE';
COMMENT ON COLUMN public.payment_data.raw_request IS 'Raw request sent to payment gateway';
COMMENT ON COLUMN public.payment_data.raw_response IS 'Raw response from payment gateway';

