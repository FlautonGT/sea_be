-- ============================================
-- Gate.co.id Database Schema
-- PostgreSQL 15+
-- ============================================

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================
-- ENUM TYPES
-- ============================================

-- User & Admin Status
CREATE TYPE user_status AS ENUM ('ACTIVE', 'INACTIVE', 'SUSPENDED');
CREATE TYPE mfa_status AS ENUM ('ACTIVE', 'INACTIVE');

-- Admin Roles
CREATE TYPE admin_role AS ENUM ('SUPERADMIN', 'ADMIN', 'FINANCE', 'CS_LEAD', 'CS');

-- Region Codes
CREATE TYPE region_code AS ENUM ('ID', 'MY', 'PH', 'SG', 'TH');

-- Currency Codes
CREATE TYPE currency_code AS ENUM ('IDR', 'MYR', 'PHP', 'SGD', 'THB');

-- Transaction Status
CREATE TYPE transaction_status AS ENUM ('PENDING', 'PAID', 'PROCESSING', 'SUCCESS', 'FAILED', 'EXPIRED', 'REFUNDED');

-- Payment Status
CREATE TYPE payment_status AS ENUM ('UNPAID', 'PAID', 'EXPIRED', 'REFUNDED');

-- Deposit Status
CREATE TYPE deposit_status AS ENUM ('PENDING', 'SUCCESS', 'FAILED', 'EXPIRED', 'REFUNDED');

-- Mutation Type
CREATE TYPE mutation_type AS ENUM ('CREDIT', 'DEBIT');

-- Provider Health Status
CREATE TYPE health_status AS ENUM ('HEALTHY', 'DEGRADED', 'UNHEALTHY');

-- Payment Type
CREATE TYPE payment_type AS ENUM ('purchase', 'deposit');

-- Fee Type
CREATE TYPE fee_type AS ENUM ('FIXED', 'PERCENTAGE', 'MIXED');

-- Field Type
CREATE TYPE field_type AS ENUM ('text', 'number', 'email', 'select', 'phone');

-- Audit Action
CREATE TYPE audit_action AS ENUM ('CREATE', 'UPDATE', 'DELETE', 'LOGIN', 'LOGOUT');

-- Membership Level
CREATE TYPE membership_level AS ENUM ('CLASSIC', 'PRESTIGE', 'ROYAL');

-- ============================================
-- CORE TABLES
-- ============================================

-- Regions
CREATE TABLE regions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code region_code UNIQUE NOT NULL,
    country VARCHAR(100) NOT NULL,
    currency currency_code NOT NULL,
    currency_symbol VARCHAR(10) NOT NULL,
    image VARCHAR(500),
    is_default BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    sort_order INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Languages
CREATE TABLE languages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(10) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    country VARCHAR(100) NOT NULL,
    image VARCHAR(500),
    is_default BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    sort_order INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================
-- USER TABLES
-- ============================================

-- Users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100),
    email VARCHAR(255) UNIQUE NOT NULL,
    email_verified_at TIMESTAMPTZ,
    phone_number VARCHAR(20) UNIQUE,
    password_hash VARCHAR(255),
    profile_picture VARCHAR(500),
    status user_status DEFAULT 'INACTIVE',
    primary_region region_code DEFAULT 'ID',
    current_region region_code DEFAULT 'ID',
    
    -- Multi-currency balances (stored as smallest unit: cents/sen)
    balance_idr BIGINT DEFAULT 0,
    balance_myr BIGINT DEFAULT 0,
    balance_php BIGINT DEFAULT 0,
    balance_sgd BIGINT DEFAULT 0,
    balance_thb BIGINT DEFAULT 0,
    
    -- Membership
    membership_level membership_level DEFAULT 'CLASSIC',
    total_spent_idr BIGINT DEFAULT 0,
    
    -- MFA
    mfa_status mfa_status DEFAULT 'INACTIVE',
    mfa_secret VARCHAR(255),
    mfa_backup_codes TEXT[],
    
    -- Google OAuth
    google_id VARCHAR(255) UNIQUE,
    
    -- Timestamps
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    -- Indexes will be created separately
    CONSTRAINT valid_email CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$')
);

-- User Sessions
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash VARCHAR(255) NOT NULL,
    ip_address INET,
    user_agent TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Password Reset Tokens
CREATE TABLE password_resets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Email Verification Tokens
CREATE TABLE email_verifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================
-- ADMIN TABLES
-- ============================================

-- Admin Permissions
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    category VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Admin Roles Configuration
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code admin_role UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    level INT NOT NULL, -- 1 = highest (SUPERADMIN), 5 = lowest (CS)
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Role Permissions (Many-to-Many)
CREATE TABLE role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (role_id, permission_id)
);

-- Admins
CREATE TABLE admins (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(200) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone_number VARCHAR(20),
    password_hash VARCHAR(255) NOT NULL,
    role_id UUID NOT NULL REFERENCES roles(id),
    status user_status DEFAULT 'ACTIVE',
    
    -- MFA
    mfa_enabled BOOLEAN DEFAULT FALSE,
    mfa_secret VARCHAR(255),
    mfa_backup_codes TEXT[],
    
    -- Timestamps
    last_login_at TIMESTAMPTZ,
    created_by UUID REFERENCES admins(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Admin Sessions
CREATE TABLE admin_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    admin_id UUID NOT NULL REFERENCES admins(id) ON DELETE CASCADE,
    refresh_token_hash VARCHAR(255) NOT NULL,
    ip_address INET,
    user_agent TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================
-- PRODUCT TABLES
-- ============================================

-- Categories
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(50) UNIQUE NOT NULL,
    title VARCHAR(100) NOT NULL,
    description TEXT,
    icon VARCHAR(500),
    is_active BOOLEAN DEFAULT TRUE,
    sort_order INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Category Regions (Many-to-Many)
CREATE TABLE category_regions (
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    region_code region_code NOT NULL,
    PRIMARY KEY (category_id, region_code)
);

-- Products
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(50) UNIQUE NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    title VARCHAR(200) NOT NULL,
    subtitle VARCHAR(200),
    description TEXT,
    publisher VARCHAR(200),
    thumbnail VARCHAR(500),
    banner VARCHAR(500),
    category_id UUID NOT NULL REFERENCES categories(id),
    is_active BOOLEAN DEFAULT TRUE,
    is_popular BOOLEAN DEFAULT FALSE,
    features JSONB DEFAULT '[]',
    how_to_order JSONB DEFAULT '[]',
    tags TEXT[],
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Product Regions (Many-to-Many)
CREATE TABLE product_regions (
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    region_code region_code NOT NULL,
    PRIMARY KEY (product_id, region_code)
);

-- Product Fields (Input fields for each product)
CREATE TABLE product_fields (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    key VARCHAR(50) NOT NULL,
    field_type field_type NOT NULL,
    label VARCHAR(200) NOT NULL,
    placeholder VARCHAR(200),
    hint TEXT,
    pattern VARCHAR(255),
    is_required BOOLEAN DEFAULT TRUE,
    min_length INT,
    max_length INT,
    options JSONB, -- For select type
    sort_order INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (product_id, key)
);

-- Sections (SKU groupings)
CREATE TABLE sections (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(50) UNIQUE NOT NULL,
    title VARCHAR(100) NOT NULL,
    icon VARCHAR(50),
    is_active BOOLEAN DEFAULT TRUE,
    sort_order INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Product Sections (Many-to-Many)
CREATE TABLE product_sections (
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    section_id UUID NOT NULL REFERENCES sections(id) ON DELETE CASCADE,
    PRIMARY KEY (product_id, section_id)
);

-- ============================================
-- PROVIDER TABLES
-- ============================================

-- Providers (Digiflazz, VIP Reseller, BangJeff, etc)
CREATE TABLE providers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    base_url VARCHAR(500) NOT NULL,
    webhook_url VARCHAR(500),
    is_active BOOLEAN DEFAULT TRUE,
    priority INT DEFAULT 0, -- Lower = higher priority
    supported_types TEXT[], -- ['PULSA', 'DATA', 'GAME', 'EWALLET', 'PLN']
    health_status health_status DEFAULT 'HEALTHY',
    last_health_check TIMESTAMPTZ,
    
    -- API Configuration (non-sensitive)
    api_config JSONB DEFAULT '{
        "timeout": 30000,
        "retryAttempts": 3,
        "retryDelay": 1000
    }',
    
    -- Status mapping
    status_mapping JSONB DEFAULT '{
        "success": ["Sukses", "success", "1", "00"],
        "pending": ["Pending", "pending", "0", "01"],
        "failed": ["Gagal", "failed", "-1", "02"]
    }',
    
    -- Env credential keys (reference to .env variables)
    env_credential_keys JSONB DEFAULT '{}',
    
    -- Stats (updated periodically)
    total_skus INT DEFAULT 0,
    active_skus INT DEFAULT 0,
    success_rate DECIMAL(5,2) DEFAULT 0,
    avg_response_time INT DEFAULT 0, -- milliseconds
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================
-- SKU TABLES
-- ============================================

-- SKUs (Stock Keeping Units)
CREATE TABLE skus (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(100) UNIQUE NOT NULL,
    provider_sku_code VARCHAR(100) NOT NULL, -- Code from provider
    name VARCHAR(200) NOT NULL,
    description TEXT,
    image VARCHAR(500),
    info TEXT, -- Bonus info, etc
    
    product_id UUID NOT NULL REFERENCES products(id),
    provider_id UUID NOT NULL REFERENCES providers(id),
    section_id UUID REFERENCES sections(id),
    
    process_time INT DEFAULT 0, -- minutes, 0 = instant
    is_active BOOLEAN DEFAULT TRUE,
    is_featured BOOLEAN DEFAULT FALSE,
    
    -- Badge
    badge_text VARCHAR(50),
    badge_color VARCHAR(20),
    
    -- Stock status
    stock_status VARCHAR(20) DEFAULT 'AVAILABLE',
    
    -- Stats
    total_sold INT DEFAULT 0,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE (provider_id, provider_sku_code)
);

-- SKU Pricing (Per Region)
CREATE TABLE sku_pricing (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    sku_id UUID NOT NULL REFERENCES skus(id) ON DELETE CASCADE,
    region_code region_code NOT NULL,
    currency currency_code NOT NULL,
    
    buy_price BIGINT NOT NULL, -- Price from provider (smallest unit)
    sell_price BIGINT NOT NULL, -- Our selling price
    original_price BIGINT NOT NULL, -- Price before discount (for strikethrough)
    
    -- Calculated fields
    margin_percentage DECIMAL(5,2) GENERATED ALWAYS AS (
        CASE WHEN buy_price > 0 
        THEN ((sell_price - buy_price)::DECIMAL / buy_price * 100)
        ELSE 0 END
    ) STORED,
    discount_percentage DECIMAL(5,2) GENERATED ALWAYS AS (
        CASE WHEN original_price > 0 
        THEN ((original_price - sell_price)::DECIMAL / original_price * 100)
        ELSE 0 END
    ) STORED,
    
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE (sku_id, region_code)
);

-- ============================================
-- PAYMENT TABLES
-- ============================================

-- Payment Gateways (LinkQu, BCA, BRI, Xendit, etc)
CREATE TABLE payment_gateways (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    base_url VARCHAR(500) NOT NULL,
    callback_url VARCHAR(500),
    is_active BOOLEAN DEFAULT TRUE,
    supported_methods TEXT[], -- ['QRIS', 'VA', 'EWALLET']
    supported_types payment_type[], -- ['purchase', 'deposit']
    health_status health_status DEFAULT 'HEALTHY',
    last_health_check TIMESTAMPTZ,
    
    -- API Configuration
    api_config JSONB DEFAULT '{
        "timeout": 30000,
        "retryAttempts": 3
    }',
    
    -- Status mapping
    status_mapping JSONB DEFAULT '{}',
    
    -- Env credential keys
    env_credential_keys JSONB DEFAULT '{}',
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Payment Channel Categories
CREATE TABLE payment_channel_categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(50) UNIQUE NOT NULL,
    title VARCHAR(100) NOT NULL,
    icon VARCHAR(500),
    is_active BOOLEAN DEFAULT TRUE,
    sort_order INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Payment Channels (User-facing payment methods)
CREATE TABLE payment_channels (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    image VARCHAR(500),
    category_id UUID NOT NULL REFERENCES payment_channel_categories(id),
    
    -- Fee configuration
    fee_type fee_type DEFAULT 'PERCENTAGE',
    fee_amount BIGINT DEFAULT 0, -- Fixed fee amount
    fee_percentage DECIMAL(5,2) DEFAULT 0, -- Percentage fee
    min_fee BIGINT DEFAULT 0,
    max_fee BIGINT DEFAULT 0,
    
    -- Limits
    min_amount BIGINT DEFAULT 0,
    max_amount BIGINT DEFAULT 0,
    
    -- Supported types
    supported_types payment_type[] DEFAULT ARRAY['purchase', 'deposit']::payment_type[],
    
    -- Instructions (HTML)
    instruction TEXT,
    
    is_active BOOLEAN DEFAULT TRUE,
    is_featured BOOLEAN DEFAULT FALSE,
    sort_order INT DEFAULT 0,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Payment Channel Regions
CREATE TABLE payment_channel_regions (
    channel_id UUID NOT NULL REFERENCES payment_channels(id) ON DELETE CASCADE,
    region_code region_code NOT NULL,
    PRIMARY KEY (channel_id, region_code)
);

-- Payment Channel Gateway Assignments
CREATE TABLE payment_channel_gateways (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    channel_id UUID NOT NULL REFERENCES payment_channels(id) ON DELETE CASCADE,
    gateway_id UUID NOT NULL REFERENCES payment_gateways(id),
    payment_type payment_type NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (channel_id, payment_type)
);

-- ============================================
-- PROMO TABLES
-- ============================================

-- Promos
CREATE TABLE promos (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(50) UNIQUE NOT NULL,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    note TEXT,
    
    -- Limits
    max_usage INT, -- Total max usage
    max_daily_usage INT, -- Max per day
    max_usage_per_id INT DEFAULT 1, -- Max per user
    max_usage_per_device INT DEFAULT 1,
    max_usage_per_ip INT DEFAULT 1,
    
    -- Amount constraints
    min_amount BIGINT DEFAULT 0,
    max_promo_amount BIGINT, -- Max discount amount
    
    -- Discount
    promo_flat BIGINT DEFAULT 0, -- Flat discount
    promo_percentage DECIMAL(5,2) DEFAULT 0, -- Percentage discount
    
    -- Validity
    days_available TEXT[], -- ['MON', 'TUE', 'WED', 'THU', 'FRI', 'SAT', 'SUN']
    start_at TIMESTAMPTZ,
    expired_at TIMESTAMPTZ,
    
    -- Status
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Stats
    total_usage INT DEFAULT 0,
    total_discount_given BIGINT DEFAULT 0,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Promo Products (Many-to-Many, empty = all products)
CREATE TABLE promo_products (
    promo_id UUID NOT NULL REFERENCES promos(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    PRIMARY KEY (promo_id, product_id)
);

-- Promo Payment Channels (Many-to-Many, empty = all channels)
CREATE TABLE promo_payment_channels (
    promo_id UUID NOT NULL REFERENCES promos(id) ON DELETE CASCADE,
    channel_id UUID NOT NULL REFERENCES payment_channels(id) ON DELETE CASCADE,
    PRIMARY KEY (promo_id, channel_id)
);

-- Promo Regions
CREATE TABLE promo_regions (
    promo_id UUID NOT NULL REFERENCES promos(id) ON DELETE CASCADE,
    region_code region_code NOT NULL,
    PRIMARY KEY (promo_id, region_code)
);

-- Promo Usage Tracking
CREATE TABLE promo_usages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    promo_id UUID NOT NULL REFERENCES promos(id),
    user_id UUID REFERENCES users(id),
    transaction_id UUID, -- Will reference transactions table
    device_id VARCHAR(255),
    ip_address INET,
    discount_amount BIGINT NOT NULL,
    used_at TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================
-- TRANSACTION TABLES
-- ============================================

-- Transactions
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    invoice_number VARCHAR(50) UNIQUE NOT NULL,
    
    -- User (nullable for guest checkout)
    user_id UUID REFERENCES users(id),
    
    -- Product & SKU
    product_id UUID NOT NULL REFERENCES products(id),
    sku_id UUID NOT NULL REFERENCES skus(id),
    quantity INT DEFAULT 1,
    
    -- Account info (game ID, phone number, etc)
    account_inputs JSONB NOT NULL, -- {"userId": "123", "zoneId": "456"}
    account_nickname VARCHAR(200),
    
    -- Provider
    provider_id UUID NOT NULL REFERENCES providers(id),
    provider_ref_id VARCHAR(255), -- Reference ID from provider
    provider_serial_number VARCHAR(255), -- SN from provider
    provider_response JSONB,
    
    -- Payment
    payment_channel_id UUID NOT NULL REFERENCES payment_channels(id),
    payment_gateway_id UUID REFERENCES payment_gateways(id),
    payment_gateway_ref_id VARCHAR(255),
    
    -- Promo
    promo_id UUID REFERENCES promos(id),
    promo_code VARCHAR(50),
    
    -- Pricing (smallest unit)
    buy_price BIGINT NOT NULL,
    sell_price BIGINT NOT NULL,
    discount_amount BIGINT DEFAULT 0,
    payment_fee BIGINT DEFAULT 0,
    total_amount BIGINT NOT NULL,
    profit BIGINT GENERATED ALWAYS AS (sell_price - buy_price - discount_amount) STORED,
    currency currency_code NOT NULL,
    
    -- Status
    status transaction_status DEFAULT 'PENDING',
    payment_status payment_status DEFAULT 'UNPAID',
    
    -- Contact
    contact_email VARCHAR(255),
    contact_phone VARCHAR(20),
    
    -- Region & Meta
    region region_code NOT NULL,
    ip_address INET,
    user_agent TEXT,
    device_id VARCHAR(255),
    
    -- Timestamps
    paid_at TIMESTAMPTZ,
    processed_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    expired_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Transaction Timeline/Logs
CREATE TABLE transaction_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL,
    message TEXT,
    data JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================
-- DEPOSIT TABLES
-- ============================================

-- Deposits
CREATE TABLE deposits (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    invoice_number VARCHAR(50) UNIQUE NOT NULL,
    
    user_id UUID NOT NULL REFERENCES users(id),
    
    -- Amount
    amount BIGINT NOT NULL,
    payment_fee BIGINT DEFAULT 0,
    total_amount BIGINT NOT NULL,
    currency currency_code NOT NULL,
    
    -- Payment
    payment_channel_id UUID NOT NULL REFERENCES payment_channels(id),
    payment_gateway_id UUID REFERENCES payment_gateways(id),
    payment_gateway_ref_id VARCHAR(255),
    payment_data JSONB, -- QR code, VA number, etc
    
    -- Status
    status deposit_status DEFAULT 'PENDING',
    
    -- Balance change
    balance_before BIGINT,
    balance_after BIGINT,
    
    -- Region & Meta
    region region_code NOT NULL,
    ip_address INET,
    user_agent TEXT,
    
    -- Admin actions
    confirmed_by UUID REFERENCES admins(id),
    confirmed_at TIMESTAMPTZ,
    cancelled_by UUID REFERENCES admins(id),
    cancelled_at TIMESTAMPTZ,
    cancel_reason TEXT,
    
    -- Timestamps
    paid_at TIMESTAMPTZ,
    expired_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Deposit Logs
CREATE TABLE deposit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    deposit_id UUID NOT NULL REFERENCES deposits(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL,
    message TEXT,
    data JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================
-- MUTATION TABLES (Balance History)
-- ============================================

-- Mutations
CREATE TABLE mutations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    
    -- Reference
    invoice_number VARCHAR(50),
    reference_type VARCHAR(50) NOT NULL, -- 'TRANSACTION', 'DEPOSIT', 'REFUND', 'ADJUSTMENT', 'BONUS'
    reference_id UUID,
    
    description TEXT NOT NULL,
    
    -- Amount
    mutation_type mutation_type NOT NULL,
    amount BIGINT NOT NULL,
    balance_before BIGINT NOT NULL,
    balance_after BIGINT NOT NULL,
    currency currency_code NOT NULL,
    
    -- Admin (for manual adjustments)
    admin_id UUID REFERENCES admins(id),
    admin_note TEXT,
    
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================
-- REFUND TABLES
-- ============================================

-- Refunds
CREATE TABLE refunds (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Reference (either transaction or deposit)
    transaction_id UUID REFERENCES transactions(id),
    deposit_id UUID REFERENCES deposits(id),
    
    invoice_number VARCHAR(50) NOT NULL,
    
    -- Amount
    amount BIGINT NOT NULL,
    currency currency_code NOT NULL,
    
    -- Refund destination
    refund_to VARCHAR(50) NOT NULL, -- 'BALANCE', 'ORIGINAL_METHOD'
    
    -- Status
    status VARCHAR(50) DEFAULT 'PROCESSING',
    
    reason TEXT,
    
    -- Admin
    processed_by UUID NOT NULL REFERENCES admins(id),
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    
    CONSTRAINT refund_has_reference CHECK (
        (transaction_id IS NOT NULL AND deposit_id IS NULL) OR
        (transaction_id IS NULL AND deposit_id IS NOT NULL)
    )
);

-- ============================================
-- CONTENT TABLES
-- ============================================

-- Banners
CREATE TABLE banners (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(200) NOT NULL,
    description TEXT,
    href VARCHAR(500),
    image VARCHAR(500) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    sort_order INT DEFAULT 0,
    start_at TIMESTAMPTZ,
    expired_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Banner Regions
CREATE TABLE banner_regions (
    banner_id UUID NOT NULL REFERENCES banners(id) ON DELETE CASCADE,
    region_code region_code NOT NULL,
    PRIMARY KEY (banner_id, region_code)
);

-- Popups (One per region)
CREATE TABLE popups (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    region_code region_code UNIQUE NOT NULL,
    title VARCHAR(200),
    content TEXT,
    image VARCHAR(500),
    href VARCHAR(500),
    is_active BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================
-- SETTINGS & AUDIT TABLES
-- ============================================

-- Settings
CREATE TABLE settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    category VARCHAR(50) NOT NULL,
    key VARCHAR(100) NOT NULL,
    value JSONB NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (category, key)
);

-- Contacts (Special settings)
CREATE TABLE contacts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255),
    phone VARCHAR(50),
    whatsapp VARCHAR(500),
    instagram VARCHAR(500),
    facebook VARCHAR(500),
    x VARCHAR(500),
    youtube VARCHAR(500),
    telegram VARCHAR(500),
    discord VARCHAR(500),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Audit Logs
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    admin_id UUID REFERENCES admins(id),
    admin_name VARCHAR(200),
    admin_email VARCHAR(255),
    
    action audit_action NOT NULL,
    resource VARCHAR(100) NOT NULL, -- 'USER', 'PRODUCT', 'SKU', etc
    resource_id UUID,
    description TEXT,
    
    changes JSONB, -- {"before": {...}, "after": {...}}
    
    ip_address INET,
    user_agent TEXT,
    
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================
-- INDEXES
-- ============================================

-- Users
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_phone ON users(phone_number);
CREATE INDEX idx_users_google_id ON users(google_id);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_primary_region ON users(primary_region);
CREATE INDEX idx_users_membership ON users(membership_level);
CREATE INDEX idx_users_created_at ON users(created_at DESC);

-- User Sessions
CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_expires_at ON user_sessions(expires_at);

-- Admins
CREATE INDEX idx_admins_email ON admins(email);
CREATE INDEX idx_admins_role ON admins(role_id);
CREATE INDEX idx_admins_status ON admins(status);

-- Products
CREATE INDEX idx_products_code ON products(code);
CREATE INDEX idx_products_slug ON products(slug);
CREATE INDEX idx_products_category ON products(category_id);
CREATE INDEX idx_products_is_active ON products(is_active);
CREATE INDEX idx_products_is_popular ON products(is_popular);

-- SKUs
CREATE INDEX idx_skus_code ON skus(code);
CREATE INDEX idx_skus_product ON skus(product_id);
CREATE INDEX idx_skus_provider ON skus(provider_id);
CREATE INDEX idx_skus_section ON skus(section_id);
CREATE INDEX idx_skus_is_active ON skus(is_active);

-- SKU Pricing
CREATE INDEX idx_sku_pricing_sku ON sku_pricing(sku_id);
CREATE INDEX idx_sku_pricing_region ON sku_pricing(region_code);

-- Transactions
CREATE INDEX idx_transactions_invoice ON transactions(invoice_number);
CREATE INDEX idx_transactions_user ON transactions(user_id);
CREATE INDEX idx_transactions_product ON transactions(product_id);
CREATE INDEX idx_transactions_sku ON transactions(sku_id);
CREATE INDEX idx_transactions_provider ON transactions(provider_id);
CREATE INDEX idx_transactions_status ON transactions(status);
CREATE INDEX idx_transactions_payment_status ON transactions(payment_status);
CREATE INDEX idx_transactions_region ON transactions(region);
CREATE INDEX idx_transactions_created_at ON transactions(created_at DESC);
CREATE INDEX idx_transactions_date ON transactions(DATE(created_at));

-- Deposits
CREATE INDEX idx_deposits_invoice ON deposits(invoice_number);
CREATE INDEX idx_deposits_user ON deposits(user_id);
CREATE INDEX idx_deposits_status ON deposits(status);
CREATE INDEX idx_deposits_created_at ON deposits(created_at DESC);

-- Mutations
CREATE INDEX idx_mutations_user ON mutations(user_id);
CREATE INDEX idx_mutations_type ON mutations(mutation_type);
CREATE INDEX idx_mutations_created_at ON mutations(created_at DESC);

-- Promos
CREATE INDEX idx_promos_code ON promos(code);
CREATE INDEX idx_promos_is_active ON promos(is_active);
CREATE INDEX idx_promos_expired_at ON promos(expired_at);

-- Promo Usages
CREATE INDEX idx_promo_usages_promo ON promo_usages(promo_id);
CREATE INDEX idx_promo_usages_user ON promo_usages(user_id);
CREATE INDEX idx_promo_usages_date ON promo_usages(DATE(used_at));

-- Audit Logs
CREATE INDEX idx_audit_logs_admin ON audit_logs(admin_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);

-- ============================================
-- TRIGGERS
-- ============================================

-- Updated At Trigger Function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply updated_at trigger to tables
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_admins_updated_at BEFORE UPDATE ON admins
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_products_updated_at BEFORE UPDATE ON products
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_skus_updated_at BEFORE UPDATE ON skus
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_sku_pricing_updated_at BEFORE UPDATE ON sku_pricing
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_transactions_updated_at BEFORE UPDATE ON transactions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_deposits_updated_at BEFORE UPDATE ON deposits
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_providers_updated_at BEFORE UPDATE ON providers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_payment_gateways_updated_at BEFORE UPDATE ON payment_gateways
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_payment_channels_updated_at BEFORE UPDATE ON payment_channels
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_promos_updated_at BEFORE UPDATE ON promos
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_banners_updated_at BEFORE UPDATE ON banners
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_popups_updated_at BEFORE UPDATE ON popups
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_categories_updated_at BEFORE UPDATE ON categories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_sections_updated_at BEFORE UPDATE ON sections
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_regions_updated_at BEFORE UPDATE ON regions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_languages_updated_at BEFORE UPDATE ON languages
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_roles_updated_at BEFORE UPDATE ON roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_settings_updated_at BEFORE UPDATE ON settings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

