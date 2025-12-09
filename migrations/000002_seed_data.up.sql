-- Migration: 000002_seed_data
-- Description: Seed initial data
-- Created: 2025-12-03

-- ============================================
-- REGIONS
-- ============================================

INSERT INTO regions (code, country, currency, currency_symbol, image, is_default, is_active, sort_order) VALUES
('ID', 'Indonesia', 'IDR', 'Rp', 'https://cdn.gate.co.id/flags/id.svg', TRUE, TRUE, 1),
('MY', 'Malaysia', 'MYR', 'RM', 'https://cdn.gate.co.id/flags/my.svg', FALSE, TRUE, 2),
('PH', 'Philippines', 'PHP', '₱', 'https://cdn.gate.co.id/flags/ph.svg', FALSE, TRUE, 3),
('SG', 'Singapore', 'SGD', 'S$', 'https://cdn.gate.co.id/flags/sg.svg', FALSE, TRUE, 4),
('TH', 'Thailand', 'THB', '฿', 'https://cdn.gate.co.id/flags/th.svg', FALSE, TRUE, 5);

-- ============================================
-- LANGUAGES
-- ============================================

INSERT INTO languages (code, name, country, image, is_default, is_active, sort_order) VALUES
('id', 'Bahasa Indonesia', 'Indonesia', 'https://cdn.gate.co.id/flags/id.svg', TRUE, TRUE, 1),
('en', 'English', 'United States', 'https://cdn.gate.co.id/flags/us.svg', FALSE, TRUE, 2);

-- ============================================
-- PERMISSIONS
-- ============================================

INSERT INTO permissions (code, name, description, category) VALUES
-- Admin Management
('admin:read', 'View Admins', 'Can view admin list and details', 'Admin'),
('admin:create', 'Create Admin', 'Can create new admin accounts', 'Admin'),
('admin:update', 'Update Admin', 'Can update admin accounts', 'Admin'),
('admin:delete', 'Delete Admin', 'Can delete admin accounts', 'Admin'),
('role:manage', 'Manage Roles', 'Can manage roles and permissions', 'Admin'),

-- Provider Management
('provider:read', 'View Providers', 'Can view provider list and details', 'Provider'),
('provider:create', 'Create Provider', 'Can create new providers', 'Provider'),
('provider:update', 'Update Provider', 'Can update providers', 'Provider'),
('provider:delete', 'Delete Provider', 'Can delete providers', 'Provider'),

-- Payment Gateway
('gateway:read', 'View Gateways', 'Can view payment gateways', 'Gateway'),
('gateway:create', 'Create Gateway', 'Can create payment gateways', 'Gateway'),
('gateway:update', 'Update Gateway', 'Can update payment gateways', 'Gateway'),
('gateway:delete', 'Delete Gateway', 'Can delete payment gateways', 'Gateway'),

-- Product Management
('product:read', 'View Products', 'Can view products', 'Product'),
('product:create', 'Create Product', 'Can create products', 'Product'),
('product:update', 'Update Product', 'Can update products', 'Product'),
('product:delete', 'Delete Product', 'Can delete products', 'Product'),

-- SKU Management
('sku:read', 'View SKUs', 'Can view SKUs', 'SKU'),
('sku:create', 'Create SKU', 'Can create SKUs', 'SKU'),
('sku:update', 'Update SKU', 'Can update SKUs', 'SKU'),
('sku:delete', 'Delete SKU', 'Can delete SKUs', 'SKU'),
('sku:sync', 'Sync SKU', 'Can sync SKUs from provider', 'SKU'),

-- Transaction Management
('transaction:read', 'View Transactions', 'Can view transactions', 'Transaction'),
('transaction:update', 'Update Transaction', 'Can update transaction status', 'Transaction'),
('transaction:refund', 'Refund Transaction', 'Can process refunds', 'Transaction'),
('transaction:manual', 'Manual Process', 'Can manually process transactions', 'Transaction'),

-- User Management
('user:read', 'View Users', 'Can view users', 'User'),
('user:update', 'Update User', 'Can update user accounts', 'User'),
('user:suspend', 'Suspend User', 'Can suspend user accounts', 'User'),
('user:balance', 'Adjust Balance', 'Can adjust user balance', 'User'),

-- Promo Management
('promo:read', 'View Promos', 'Can view promos', 'Promo'),
('promo:create', 'Create Promo', 'Can create promos', 'Promo'),
('promo:update', 'Update Promo', 'Can update promos', 'Promo'),
('promo:delete', 'Delete Promo', 'Can delete promos', 'Promo'),

-- Content Management
('content:read', 'View Content', 'Can view content', 'Content'),
('content:banner', 'Manage Banners', 'Can manage banners', 'Content'),
('content:popup', 'Manage Popups', 'Can manage popups', 'Content'),

-- Reports
('report:read', 'View Reports', 'Can view reports', 'Report'),
('report:export', 'Export Reports', 'Can export reports', 'Report'),

-- Audit
('audit:read', 'View Audit Logs', 'Can view audit logs', 'Audit'),

-- Settings
('setting:read', 'View Settings', 'Can view settings', 'Setting'),
('setting:update', 'Update Settings', 'Can update settings', 'Setting');

-- ============================================
-- ROLES
-- ============================================

INSERT INTO roles (code, name, description, level) VALUES
('SUPERADMIN', 'Super Administrator', 'Full system access, manage admins & permissions', 1),
('ADMIN', 'Administrator', 'Manage products, SKUs, promos, content', 2),
('FINANCE', 'Finance', 'View transactions, reports, manage deposits', 3),
('CS_LEAD', 'CS Lead', 'Handle escalations, manage CS team', 4),
('CS', 'Customer Service', 'View transactions, handle user issues', 5);

-- ============================================
-- ROLE PERMISSIONS
-- ============================================

-- SUPERADMIN gets all permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p WHERE r.code = 'SUPERADMIN';

-- ADMIN permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p 
WHERE r.code = 'ADMIN' AND p.code IN (
    'provider:read', 'gateway:read',
    'product:read', 'product:create', 'product:update', 'product:delete',
    'sku:read', 'sku:create', 'sku:update', 'sku:delete', 'sku:sync',
    'transaction:read', 'transaction:update', 'transaction:refund', 'transaction:manual',
    'user:read', 'user:update', 'user:suspend',
    'promo:read', 'promo:create', 'promo:update', 'promo:delete',
    'content:read', 'content:banner', 'content:popup',
    'report:read', 'report:export',
    'setting:read'
);

-- FINANCE permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p 
WHERE r.code = 'FINANCE' AND p.code IN (
    'gateway:read',
    'product:read', 'sku:read',
    'transaction:read', 'transaction:update', 'transaction:refund',
    'user:read', 'user:balance',
    'promo:read',
    'report:read', 'report:export'
);

-- CS_LEAD permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p 
WHERE r.code = 'CS_LEAD' AND p.code IN (
    'product:read', 'sku:read',
    'transaction:read', 'transaction:update',
    'user:read', 'user:update', 'user:suspend',
    'promo:read',
    'report:read'
);

-- CS permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p 
WHERE r.code = 'CS' AND p.code IN (
    'product:read', 'sku:read',
    'transaction:read',
    'user:read',
    'promo:read'
);

-- ============================================
-- DEFAULT SUPERADMIN
-- Password: SuperAdmin@123 (change this!)
-- ============================================

INSERT INTO admins (name, email, password_hash, role_id, status)
SELECT 
    'Super Admin',
    'superadmin@gate.co.id',
    crypt('SuperAdmin@123', gen_salt('bf', 12)),
    r.id,
    'ACTIVE'
FROM roles r WHERE r.code = 'SUPERADMIN';

-- ============================================
-- PAYMENT CHANNEL CATEGORIES
-- ============================================

INSERT INTO payment_channel_categories (code, title, icon, is_active, sort_order) VALUES
('E_WALLET', 'E-Wallet', 'https://cdn.gate.co.id/icons/wallet.svg', TRUE, 1),
('VIRTUAL_ACCOUNT', 'Virtual Account', 'https://cdn.gate.co.id/icons/bank.svg', TRUE, 2),
('RETAIL', 'Convenience Store', 'https://cdn.gate.co.id/icons/store.svg', TRUE, 3),
('CARD', 'Credit or Debit Card', 'https://cdn.gate.co.id/icons/card.svg', TRUE, 4);

-- ============================================
-- PAYMENT GATEWAYS
-- ============================================

INSERT INTO payment_gateways (code, name, base_url, callback_url, is_active, supported_methods, supported_types, env_credential_keys) VALUES
('LINKQU', 'LinkQu', 'https://api.linkqu.id', 'https://gateway.gate.id/callbacks/linkqu', TRUE, 
    ARRAY['QRIS'], ARRAY['purchase', 'deposit']::payment_type[],
    '{"clientId": "LINKQU_CLIENT_ID", "clientSecret": "LINKQU_CLIENT_SECRET", "username": "LINKQU_USERNAME", "pin": "LINKQU_PIN"}'
),
('BCA_DIRECT', 'BCA Direct API', 'https://sandbox.bca.co.id', 'https://gateway.gate.id/callbacks/bca', TRUE,
    ARRAY['BCA_VA'], ARRAY['purchase', 'deposit']::payment_type[],
    '{"clientId": "BCA_CLIENT_ID", "clientSecret": "BCA_CLIENT_SECRET", "apiKey": "BCA_API_KEY", "apiSecret": "BCA_API_SECRET"}'
),
('BRI_DIRECT', 'BRI Direct API', 'https://sandbox.bri.co.id', 'https://gateway.gate.id/callbacks/bri', TRUE,
    ARRAY['BRI_VA'], ARRAY['purchase', 'deposit']::payment_type[],
    '{"clientId": "BRI_CLIENT_ID", "clientSecret": "BRI_CLIENT_SECRET"}'
),
('XENDIT', 'Xendit', 'https://api.xendit.co', 'https://gateway.gate.id/callbacks/xendit', TRUE,
    ARRAY['PERMATA_VA', 'MANDIRI_VA', 'CARD'], ARRAY['purchase', 'deposit']::payment_type[],
    '{"secretKey": "XENDIT_SECRET_KEY", "callbackToken": "XENDIT_CALLBACK_TOKEN"}'
),
('MIDTRANS', 'Midtrans', 'https://api.midtrans.com', 'https://gateway.gate.id/callbacks/midtrans', TRUE,
    ARRAY['GOPAY', 'SHOPEEPAY'], ARRAY['purchase']::payment_type[],
    '{"serverKey": "MIDTRANS_SERVER_KEY", "clientKey": "MIDTRANS_CLIENT_KEY"}'
),
('DANA_DIRECT', 'DANA Direct', 'https://api.dana.id', 'https://gateway.gate.id/callbacks/dana', TRUE,
    ARRAY['DANA'], ARRAY['purchase']::payment_type[],
    '{"clientId": "DANA_CLIENT_ID", "clientSecret": "DANA_CLIENT_SECRET"}'
);

-- ============================================
-- PROVIDERS
-- ============================================

INSERT INTO providers (code, name, base_url, webhook_url, is_active, priority, supported_types, env_credential_keys) VALUES
('DIGIFLAZZ', 'Digiflazz', 'https://api.digiflazz.com/v1', 'https://gateway.gate.id/webhooks/digiflazz', TRUE, 1,
    ARRAY['PULSA', 'DATA', 'GAME', 'EWALLET', 'PLN'],
    '{"username": "DIGIFLAZZ_USERNAME", "apiKey": "DIGIFLAZZ_API_KEY", "webhookSecret": "DIGIFLAZZ_WEBHOOK_SECRET"}'
),
('VIPRESELLER', 'VIP Reseller', 'https://vip-reseller.co.id/api', 'https://gateway.gate.id/webhooks/vipreseller', TRUE, 2,
    ARRAY['GAME', 'VOUCHER'],
    '{"apiId": "VIPRESELLER_API_ID", "apiKey": "VIPRESELLER_API_KEY"}'
),
('BANGJEFF', 'BangJeff', 'https://api.bangjeff.com', 'https://gateway.gate.id/webhooks/bangjeff', TRUE, 3,
    ARRAY['GAME', 'STREAMING'],
    '{"memberId": "BANGJEFF_MEMBER_ID", "secretKey": "BANGJEFF_SECRET_KEY"}'
);

-- ============================================
-- CATEGORIES
-- ============================================

INSERT INTO categories (code, title, description, icon, is_active, sort_order) VALUES
('top-up-game', 'Top Up Game', 'Top up diamond, UC, dan in-game currency lainnya', 'https://cdn.gate.co.id/icons/game-controller.svg', TRUE, 1),
('voucher', 'Voucher', 'Voucher game dan digital content', 'https://cdn.gate.co.id/icons/ticket.svg', TRUE, 2),
('e-money', 'E-Money', 'Top up saldo e-wallet dan e-money', 'https://cdn.gate.co.id/icons/wallet.svg', TRUE, 3),
('credit-or-data', 'Pulsa & Paket Data', 'Pulsa dan paket data semua operator', 'https://cdn.gate.co.id/icons/phone.svg', TRUE, 4),
('streaming', 'Streaming', 'Langganan Netflix, Spotify, Disney+ dan lainnya', 'https://cdn.gate.co.id/icons/play.svg', TRUE, 5),
('electricity', 'Token Listrik', 'Token listrik PLN prabayar', 'https://cdn.gate.co.id/icons/lightning.svg', TRUE, 6);

-- All categories available in all regions
INSERT INTO category_regions (category_id, region_code)
SELECT c.id, r.code::region_code FROM categories c, regions r;

-- ============================================
-- SECTIONS
-- ============================================

INSERT INTO sections (code, title, icon, is_active, sort_order) VALUES
('special-item', 'Spesial Item', '⭐', TRUE, 1),
('topup-instant', 'Topup Instan', '⚡', TRUE, 2),
('weekly-pass', 'Weekly Pass', '📅', TRUE, 3),
('monthly-pass', 'Monthly Pass', '📆', TRUE, 4),
('all-items', 'Semua Item', '', TRUE, 99);

-- ============================================
-- POPUPS (One per region, initially inactive)
-- ============================================

INSERT INTO popups (region_code, title, content, is_active) VALUES
('ID', NULL, NULL, FALSE),
('MY', NULL, NULL, FALSE),
('PH', NULL, NULL, FALSE),
('SG', NULL, NULL, FALSE),
('TH', NULL, NULL, FALSE);

-- ============================================
-- CONTACTS
-- ============================================

INSERT INTO contacts (email, phone, whatsapp, instagram, facebook, x, youtube, telegram, discord) VALUES
('support@gate.co.id', '+6281234567890', 'https://wa.me/6281234567890', 
 'https://instagram.com/gate.official', 'https://facebook.com/gate.official',
 'https://x.com/gate_official', 'https://youtube.com/@gateofficial',
 'https://t.me/gate_official', 'https://discord.gg/gate');

-- ============================================
-- DEFAULT SETTINGS
-- ============================================

INSERT INTO settings (category, key, value, description) VALUES
('general', 'siteName', '"Gate.co.id"', 'Site name'),
('general', 'siteDescription', '"Top Up Game & Voucher Digital Terpercaya"', 'Site description'),
('general', 'maintenanceMode', 'false', 'Enable maintenance mode'),
('general', 'maintenanceMessage', 'null', 'Maintenance message'),

('transaction', 'orderExpiry', '3600', 'Order expiry time in seconds'),
('transaction', 'autoRefundOnFail', 'true', 'Auto refund on failed transaction'),
('transaction', 'maxRetryAttempts', '3', 'Max retry attempts for provider'),

('notification', 'emailEnabled', 'true', 'Enable email notifications'),
('notification', 'whatsappEnabled', 'true', 'Enable WhatsApp notifications'),
('notification', 'telegramEnabled', 'false', 'Enable Telegram notifications'),

('security', 'maxLoginAttempts', '5', 'Max login attempts before lockout'),
('security', 'lockoutDuration', '900', 'Lockout duration in seconds'),
('security', 'sessionTimeout', '3600', 'Session timeout in seconds'),
('security', 'mfaRequired', 'true', 'Require MFA for admin');

-- ============================================
-- SAMPLE PRODUCTS
-- ============================================

INSERT INTO products (code, slug, title, subtitle, description, publisher, thumbnail, banner, category_id, is_active, is_popular, features, how_to_order, tags) VALUES
-- Mobile Legends
('mlbb', 'mobile-legends', 'Mobile Legends: Bang Bang', 'Top Up Diamond ML Murah & Cepat', 
 'Mobile Legends: Bang Bang adalah game MOBA mobile yang paling populer di Asia Tenggara. Top up diamond untuk membeli skin hero, battle pass, dan item eksklusif lainnya.',
 'Moonton',
 'https://cdn.gate.co.id/products/mlbb/thumbnail.webp',
 'https://cdn.gate.co.id/products/mlbb/banner.webp',
 (SELECT id FROM categories WHERE code = 'top-up-game'),
 TRUE, TRUE,
 '[{"icon": "⚡", "text": "Proses Instan 1-5 Menit"}, {"icon": "🔒", "text": "100% Aman & Legal"}, {"icon": "💎", "text": "Harga Termurah"}]',
 '[{"step": 1, "text": "Masukkan User ID dan Server ID"}, {"step": 2, "text": "Pilih nominal diamond"}, {"step": 3, "text": "Pilih metode pembayaran"}, {"step": 4, "text": "Diamond akan masuk otomatis"}]',
 ARRAY['moba', 'mobile legends', 'ml', 'diamond', 'moonton']),

-- Free Fire
('ff', 'free-fire', 'Free Fire', 'Top Up Diamond FF Murah & Cepat',
 'Garena Free Fire adalah game battle royale yang sangat populer. Top up diamond untuk membeli skin, karakter, dan item eksklusif lainnya.',
 'Garena',
 'https://cdn.gate.co.id/products/ff/thumbnail.webp',
 'https://cdn.gate.co.id/products/ff/banner.webp',
 (SELECT id FROM categories WHERE code = 'top-up-game'),
 TRUE, TRUE,
 '[{"icon": "⚡", "text": "Proses Instan"}, {"icon": "🔒", "text": "Aman & Terpercaya"}]',
 '[{"step": 1, "text": "Masukkan ID Free Fire"}, {"step": 2, "text": "Pilih nominal diamond"}, {"step": 3, "text": "Bayar dan diamond masuk otomatis"}]',
 ARRAY['battle royale', 'free fire', 'ff', 'diamond', 'garena']),

-- PUBG Mobile
('pubgm', 'pubg-mobile', 'PUBG Mobile', 'Top Up UC PUBG Mobile Murah',
 'PUBG Mobile adalah game battle royale terbaik. Top up UC untuk membeli Royale Pass, skin senjata, dan outfit keren.',
 'Krafton',
 'https://cdn.gate.co.id/products/pubgm/thumbnail.webp',
 'https://cdn.gate.co.id/products/pubgm/banner.webp',
 (SELECT id FROM categories WHERE code = 'top-up-game'),
 TRUE, TRUE,
 '[{"icon": "⚡", "text": "Proses 1-5 Menit"}, {"icon": "🎮", "text": "Support Semua Server"}]',
 '[{"step": 1, "text": "Masukkan ID PUBG Mobile"}, {"step": 2, "text": "Pilih nominal UC"}, {"step": 3, "text": "Bayar dan UC masuk otomatis"}]',
 ARRAY['battle royale', 'pubg', 'pubgm', 'uc', 'krafton']),

-- Genshin Impact
('genshin', 'genshin-impact', 'Genshin Impact', 'Top Up Genesis Crystal Murah',
 'Genshin Impact adalah game open-world action RPG dari miHoYo. Top up Genesis Crystal untuk gacha karakter dan senjata.',
 'miHoYo',
 'https://cdn.gate.co.id/products/genshin/thumbnail.webp',
 'https://cdn.gate.co.id/products/genshin/banner.webp',
 (SELECT id FROM categories WHERE code = 'top-up-game'),
 TRUE, TRUE,
 '[{"icon": "⚡", "text": "Proses Instan"}, {"icon": "🌍", "text": "Support Server Asia"}]',
 '[{"step": 1, "text": "Masukkan UID Genshin Impact"}, {"step": 2, "text": "Pilih server"}, {"step": 3, "text": "Pilih nominal Genesis Crystal"}]',
 ARRAY['rpg', 'genshin', 'genesis crystal', 'mihoyo']),

-- Valorant
('valorant', 'valorant', 'Valorant', 'Top Up Valorant Points Murah',
 'Valorant adalah game tactical shooter dari Riot Games. Top up VP untuk membeli skin senjata dan battle pass.',
 'Riot Games',
 'https://cdn.gate.co.id/products/valorant/thumbnail.webp',
 'https://cdn.gate.co.id/products/valorant/banner.webp',
 (SELECT id FROM categories WHERE code = 'top-up-game'),
 TRUE, FALSE,
 '[{"icon": "⚡", "text": "Proses Via Gift"}, {"icon": "🎮", "text": "Butuh Login Riot"}]',
 '[{"step": 1, "text": "Masukkan Riot ID"}, {"step": 2, "text": "Pilih nominal VP"}, {"step": 3, "text": "Login untuk klaim gift"}]',
 ARRAY['fps', 'valorant', 'vp', 'riot games']),

-- Honkai Star Rail
('hsr', 'honkai-star-rail', 'Honkai Star Rail', 'Top Up Oneiric Shard Murah',
 'Honkai Star Rail adalah game turn-based RPG dari HoYoverse. Top up untuk gacha karakter dan light cone.',
 'HoYoverse',
 'https://cdn.gate.co.id/products/hsr/thumbnail.webp',
 'https://cdn.gate.co.id/products/hsr/banner.webp',
 (SELECT id FROM categories WHERE code = 'top-up-game'),
 TRUE, FALSE,
 '[{"icon": "⚡", "text": "Proses Instan"}, {"icon": "🌍", "text": "Support Server Asia"}]',
 '[{"step": 1, "text": "Masukkan UID"}, {"step": 2, "text": "Pilih server"}, {"step": 3, "text": "Pilih nominal"}]',
 ARRAY['rpg', 'honkai', 'star rail', 'hoyoverse']),

-- DANA Top Up
('dana', 'dana', 'DANA', 'Top Up Saldo DANA',
 'DANA adalah dompet digital terpercaya di Indonesia. Top up saldo DANA dengan mudah dan cepat.',
 'DANA Indonesia',
 'https://cdn.gate.co.id/products/dana/thumbnail.webp',
 'https://cdn.gate.co.id/products/dana/banner.webp',
 (SELECT id FROM categories WHERE code = 'e-money'),
 TRUE, FALSE,
 '[{"icon": "⚡", "text": "Proses Instan"}, {"icon": "📱", "text": "Masuk Otomatis"}]',
 '[{"step": 1, "text": "Masukkan Nomor DANA"}, {"step": 2, "text": "Pilih nominal"}, {"step": 3, "text": "Bayar dan saldo masuk"}]',
 ARRAY['e-wallet', 'dana', 'saldo']),

-- GoPay Top Up
('gopay', 'gopay', 'GoPay', 'Top Up Saldo GoPay',
 'GoPay adalah e-wallet dari Gojek. Top up saldo GoPay dengan mudah untuk berbagai transaksi.',
 'Gojek',
 'https://cdn.gate.co.id/products/gopay/thumbnail.webp',
 'https://cdn.gate.co.id/products/gopay/banner.webp',
 (SELECT id FROM categories WHERE code = 'e-money'),
 TRUE, FALSE,
 '[{"icon": "⚡", "text": "Proses Instan"}, {"icon": "📱", "text": "Masuk Otomatis"}]',
 '[{"step": 1, "text": "Masukkan Nomor GoPay"}, {"step": 2, "text": "Pilih nominal"}, {"step": 3, "text": "Bayar dan saldo masuk"}]',
 ARRAY['e-wallet', 'gopay', 'gojek', 'saldo']);

-- All products available in Indonesia
INSERT INTO product_regions (product_id, region_code)
SELECT p.id, 'ID'::region_code FROM products p;

-- Add product fields for Mobile Legends
INSERT INTO product_fields (product_id, name, key, field_type, label, placeholder, hint, pattern, is_required, sort_order) VALUES
((SELECT id FROM products WHERE code = 'mlbb'), 'User ID', 'userId', 'number', 'User ID', 'Contoh: 123456789', 'Buka profile game untuk melihat User ID', '^[0-9]{6,12}$', TRUE, 1),
((SELECT id FROM products WHERE code = 'mlbb'), 'Server ID', 'serverId', 'number', 'Server ID', 'Contoh: 1234', 'Terletak di sebelah User ID dalam kurung', '^[0-9]{4,5}$', TRUE, 2);

-- Add product fields for Free Fire
INSERT INTO product_fields (product_id, name, key, field_type, label, placeholder, hint, is_required, sort_order) VALUES
((SELECT id FROM products WHERE code = 'ff'), 'User ID', 'userId', 'number', 'ID Free Fire', 'Contoh: 123456789', 'Buka profile untuk melihat ID', TRUE, 1);

-- Add product fields for PUBG Mobile
INSERT INTO product_fields (product_id, name, key, field_type, label, placeholder, hint, is_required, sort_order) VALUES
((SELECT id FROM products WHERE code = 'pubgm'), 'User ID', 'userId', 'number', 'ID PUBG Mobile', 'Contoh: 5123456789', 'Buka profile untuk melihat ID', TRUE, 1);

-- Add product fields for Genshin Impact
INSERT INTO product_fields (product_id, name, key, field_type, label, placeholder, hint, is_required, sort_order) VALUES
((SELECT id FROM products WHERE code = 'genshin'), 'UID', 'uid', 'number', 'UID Genshin', 'Contoh: 812345678', 'Buka menu Paimon untuk melihat UID', TRUE, 1),
((SELECT id FROM products WHERE code = 'genshin'), 'Server', 'server', 'select', 'Server', '', 'Pilih server yang sesuai', TRUE, 2);

-- Add product fields for E-Wallet
INSERT INTO product_fields (product_id, name, key, field_type, label, placeholder, hint, pattern, is_required, sort_order) VALUES
((SELECT id FROM products WHERE code = 'dana'), 'Nomor HP', 'phone', 'phone', 'Nomor DANA', '08xx xxxx xxxx', 'Nomor HP yang terdaftar di DANA', '^08[0-9]{9,12}$', TRUE, 1),
((SELECT id FROM products WHERE code = 'gopay'), 'Nomor HP', 'phone', 'phone', 'Nomor GoPay', '08xx xxxx xxxx', 'Nomor HP yang terdaftar di GoPay', '^08[0-9]{9,12}$', TRUE, 1);

-- ============================================
-- SAMPLE SKUs - Mobile Legends
-- ============================================

INSERT INTO skus (code, provider_sku_code, name, description, image, product_id, provider_id, section_id, process_time, is_active, is_featured, badge_text, badge_color, total_sold) VALUES
-- Mobile Legends Diamond
('mlbb-86-dm', 'mlbb86', '86 Diamonds', '86 Diamonds (78+8 Bonus)', 'https://cdn.gate.co.id/sku/mlbb-diamond.webp',
 (SELECT id FROM products WHERE code = 'mlbb'),
 (SELECT id FROM providers WHERE code = 'DIGIFLAZZ'),
 (SELECT id FROM sections WHERE code = 'topup-instant'),
 60, TRUE, FALSE, NULL, NULL, 15420),

('mlbb-172-dm', 'mlbb172', '172 Diamonds', '172 Diamonds (156+16 Bonus)', 'https://cdn.gate.co.id/sku/mlbb-diamond.webp',
 (SELECT id FROM products WHERE code = 'mlbb'),
 (SELECT id FROM providers WHERE code = 'DIGIFLAZZ'),
 (SELECT id FROM sections WHERE code = 'topup-instant'),
 60, TRUE, FALSE, NULL, NULL, 12350),

('mlbb-257-dm', 'mlbb257', '257 Diamonds', '257 Diamonds (234+23 Bonus)', 'https://cdn.gate.co.id/sku/mlbb-diamond.webp',
 (SELECT id FROM products WHERE code = 'mlbb'),
 (SELECT id FROM providers WHERE code = 'DIGIFLAZZ'),
 (SELECT id FROM sections WHERE code = 'topup-instant'),
 60, TRUE, FALSE, NULL, NULL, 9870),

('mlbb-344-dm', 'mlbb344', '344 Diamonds', '344 Diamonds (312+32 Bonus)', 'https://cdn.gate.co.id/sku/mlbb-diamond.webp',
 (SELECT id FROM products WHERE code = 'mlbb'),
 (SELECT id FROM providers WHERE code = 'DIGIFLAZZ'),
 (SELECT id FROM sections WHERE code = 'topup-instant'),
 60, TRUE, TRUE, 'BEST SELLER', '#FF6B6B', 18920),

('mlbb-429-dm', 'mlbb429', '429 Diamonds', '429 Diamonds (390+39 Bonus)', 'https://cdn.gate.co.id/sku/mlbb-diamond.webp',
 (SELECT id FROM products WHERE code = 'mlbb'),
 (SELECT id FROM providers WHERE code = 'DIGIFLAZZ'),
 (SELECT id FROM sections WHERE code = 'topup-instant'),
 60, TRUE, FALSE, NULL, NULL, 7650),

('mlbb-514-dm', 'mlbb514', '514 Diamonds', '514 Diamonds (468+46 Bonus)', 'https://cdn.gate.co.id/sku/mlbb-diamond.webp',
 (SELECT id FROM products WHERE code = 'mlbb'),
 (SELECT id FROM providers WHERE code = 'DIGIFLAZZ'),
 (SELECT id FROM sections WHERE code = 'topup-instant'),
 60, TRUE, FALSE, NULL, NULL, 6540),

('mlbb-706-dm', 'mlbb706', '706 Diamonds', '706 Diamonds (642+64 Bonus)', 'https://cdn.gate.co.id/sku/mlbb-diamond.webp',
 (SELECT id FROM products WHERE code = 'mlbb'),
 (SELECT id FROM providers WHERE code = 'DIGIFLAZZ'),
 (SELECT id FROM sections WHERE code = 'topup-instant'),
 60, TRUE, FALSE, NULL, NULL, 5430),

('mlbb-878-dm', 'mlbb878', '878 Diamonds', '878 Diamonds (798+80 Bonus)', 'https://cdn.gate.co.id/sku/mlbb-diamond.webp',
 (SELECT id FROM products WHERE code = 'mlbb'),
 (SELECT id FROM providers WHERE code = 'DIGIFLAZZ'),
 (SELECT id FROM sections WHERE code = 'topup-instant'),
 60, TRUE, FALSE, NULL, NULL, 4320),

('mlbb-1050-dm', 'mlbb1050', '1050 Diamonds', '1050 Diamonds (955+95 Bonus)', 'https://cdn.gate.co.id/sku/mlbb-diamond.webp',
 (SELECT id FROM products WHERE code = 'mlbb'),
 (SELECT id FROM providers WHERE code = 'DIGIFLAZZ'),
 (SELECT id FROM sections WHERE code = 'topup-instant'),
 60, TRUE, TRUE, 'POPULAR', '#4ECDC4', 8760),

('mlbb-2195-dm', 'mlbb2195', '2195 Diamonds', '2195 Diamonds (1996+199 Bonus)', 'https://cdn.gate.co.id/sku/mlbb-diamond.webp',
 (SELECT id FROM products WHERE code = 'mlbb'),
 (SELECT id FROM providers WHERE code = 'DIGIFLAZZ'),
 (SELECT id FROM sections WHERE code = 'topup-instant'),
 60, TRUE, FALSE, NULL, NULL, 3210),

-- Mobile Legends Weekly Pass
('mlbb-weekly-pass', 'mlbb-wp', 'Weekly Diamond Pass', 'Weekly Diamond Pass (Total 225 Diamonds)', 'https://cdn.gate.co.id/sku/mlbb-weekly.webp',
 (SELECT id FROM products WHERE code = 'mlbb'),
 (SELECT id FROM providers WHERE code = 'DIGIFLAZZ'),
 (SELECT id FROM sections WHERE code = 'weekly-pass'),
 300, TRUE, FALSE, 'HEMAT', '#45B7D1', 2340),

-- Mobile Legends Twilight Pass
('mlbb-twilight-pass', 'mlbb-tp', 'Twilight Pass', 'Twilight Pass Season Ini', 'https://cdn.gate.co.id/sku/mlbb-twilight.webp',
 (SELECT id FROM products WHERE code = 'mlbb'),
 (SELECT id FROM providers WHERE code = 'DIGIFLAZZ'),
 (SELECT id FROM sections WHERE code = 'monthly-pass'),
 300, TRUE, TRUE, 'EXCLUSIVE', '#9B59B6', 1890);

-- ============================================
-- SAMPLE SKUs - Free Fire
-- ============================================

INSERT INTO skus (code, provider_sku_code, name, description, image, product_id, provider_id, section_id, process_time, is_active, is_featured, badge_text, badge_color, total_sold) VALUES
('ff-50-dm', 'ff50', '50 Diamonds', '50 Diamonds Free Fire', 'https://cdn.gate.co.id/sku/ff-diamond.webp',
 (SELECT id FROM products WHERE code = 'ff'),
 (SELECT id FROM providers WHERE code = 'DIGIFLAZZ'),
 (SELECT id FROM sections WHERE code = 'topup-instant'),
 60, TRUE, FALSE, NULL, NULL, 8970),

('ff-100-dm', 'ff100', '100 Diamonds', '100 Diamonds Free Fire', 'https://cdn.gate.co.id/sku/ff-diamond.webp',
 (SELECT id FROM products WHERE code = 'ff'),
 (SELECT id FROM providers WHERE code = 'DIGIFLAZZ'),
 (SELECT id FROM sections WHERE code = 'topup-instant'),
 60, TRUE, FALSE, NULL, NULL, 11230),

('ff-210-dm', 'ff210', '210 Diamonds', '210 Diamonds Free Fire', 'https://cdn.gate.co.id/sku/ff-diamond.webp',
 (SELECT id FROM products WHERE code = 'ff'),
 (SELECT id FROM providers WHERE code = 'DIGIFLAZZ'),
 (SELECT id FROM sections WHERE code = 'topup-instant'),
 60, TRUE, TRUE, 'BEST SELLER', '#FF6B6B', 15670),

('ff-520-dm', 'ff520', '520 Diamonds', '520 Diamonds Free Fire', 'https://cdn.gate.co.id/sku/ff-diamond.webp',
 (SELECT id FROM products WHERE code = 'ff'),
 (SELECT id FROM providers WHERE code = 'DIGIFLAZZ'),
 (SELECT id FROM sections WHERE code = 'topup-instant'),
 60, TRUE, FALSE, NULL, NULL, 7890),

('ff-1060-dm', 'ff1060', '1060 Diamonds', '1060 Diamonds Free Fire', 'https://cdn.gate.co.id/sku/ff-diamond.webp',
 (SELECT id FROM products WHERE code = 'ff'),
 (SELECT id FROM providers WHERE code = 'DIGIFLAZZ'),
 (SELECT id FROM sections WHERE code = 'topup-instant'),
 60, TRUE, FALSE, NULL, NULL, 4560),

('ff-2180-dm', 'ff2180', '2180 Diamonds', '2180 Diamonds Free Fire', 'https://cdn.gate.co.id/sku/ff-diamond.webp',
 (SELECT id FROM products WHERE code = 'ff'),
 (SELECT id FROM providers WHERE code = 'DIGIFLAZZ'),
 (SELECT id FROM sections WHERE code = 'topup-instant'),
 60, TRUE, FALSE, NULL, NULL, 2340),

-- Free Fire Membership
('ff-weekly-member', 'ff-wm', 'Weekly Membership', 'Weekly Membership Free Fire', 'https://cdn.gate.co.id/sku/ff-member.webp',
 (SELECT id FROM products WHERE code = 'ff'),
 (SELECT id FROM providers WHERE code = 'DIGIFLAZZ'),
 (SELECT id FROM sections WHERE code = 'weekly-pass'),
 300, TRUE, FALSE, 'HEMAT', '#45B7D1', 3450),

('ff-monthly-member', 'ff-mm', 'Monthly Membership', 'Monthly Membership Free Fire', 'https://cdn.gate.co.id/sku/ff-member.webp',
 (SELECT id FROM products WHERE code = 'ff'),
 (SELECT id FROM providers WHERE code = 'DIGIFLAZZ'),
 (SELECT id FROM sections WHERE code = 'monthly-pass'),
 300, TRUE, FALSE, NULL, NULL, 1890);

-- ============================================
-- SKU PRICING - Mobile Legends (IDR)
-- ============================================

INSERT INTO sku_pricing (sku_id, region_code, currency, buy_price, sell_price, original_price, is_active) VALUES
((SELECT id FROM skus WHERE code = 'mlbb-86-dm'), 'ID', 'IDR', 18500, 20000, 22000, TRUE),
((SELECT id FROM skus WHERE code = 'mlbb-172-dm'), 'ID', 'IDR', 37000, 40000, 44000, TRUE),
((SELECT id FROM skus WHERE code = 'mlbb-257-dm'), 'ID', 'IDR', 55500, 60000, 66000, TRUE),
((SELECT id FROM skus WHERE code = 'mlbb-344-dm'), 'ID', 'IDR', 74000, 80000, 88000, TRUE),
((SELECT id FROM skus WHERE code = 'mlbb-429-dm'), 'ID', 'IDR', 92500, 100000, 110000, TRUE),
((SELECT id FROM skus WHERE code = 'mlbb-514-dm'), 'ID', 'IDR', 111000, 120000, 132000, TRUE),
((SELECT id FROM skus WHERE code = 'mlbb-706-dm'), 'ID', 'IDR', 148000, 160000, 176000, TRUE),
((SELECT id FROM skus WHERE code = 'mlbb-878-dm'), 'ID', 'IDR', 185000, 200000, 220000, TRUE),
((SELECT id FROM skus WHERE code = 'mlbb-1050-dm'), 'ID', 'IDR', 222000, 240000, 264000, TRUE),
((SELECT id FROM skus WHERE code = 'mlbb-2195-dm'), 'ID', 'IDR', 463000, 500000, 550000, TRUE),
((SELECT id FROM skus WHERE code = 'mlbb-weekly-pass'), 'ID', 'IDR', 25000, 28000, 30000, TRUE),
((SELECT id FROM skus WHERE code = 'mlbb-twilight-pass'), 'ID', 'IDR', 115000, 125000, 140000, TRUE);

-- ============================================
-- SKU PRICING - Free Fire (IDR)
-- ============================================

INSERT INTO sku_pricing (sku_id, region_code, currency, buy_price, sell_price, original_price, is_active) VALUES
((SELECT id FROM skus WHERE code = 'ff-50-dm'), 'ID', 'IDR', 7000, 8000, 10000, TRUE),
((SELECT id FROM skus WHERE code = 'ff-100-dm'), 'ID', 'IDR', 14000, 16000, 20000, TRUE),
((SELECT id FROM skus WHERE code = 'ff-210-dm'), 'ID', 'IDR', 28000, 32000, 40000, TRUE),
((SELECT id FROM skus WHERE code = 'ff-520-dm'), 'ID', 'IDR', 70000, 79000, 95000, TRUE),
((SELECT id FROM skus WHERE code = 'ff-1060-dm'), 'ID', 'IDR', 140000, 155000, 190000, TRUE),
((SELECT id FROM skus WHERE code = 'ff-2180-dm'), 'ID', 'IDR', 280000, 310000, 380000, TRUE),
((SELECT id FROM skus WHERE code = 'ff-weekly-member'), 'ID', 'IDR', 18000, 22000, 25000, TRUE),
((SELECT id FROM skus WHERE code = 'ff-monthly-member'), 'ID', 'IDR', 85000, 95000, 110000, TRUE);

-- ============================================
-- PAYMENT CHANNELS
-- ============================================

INSERT INTO payment_channels (code, name, description, image, category_id, fee_type, fee_amount, fee_percentage, min_amount, max_amount, supported_types, instruction, is_active, is_featured, sort_order) VALUES
-- E-Wallet
('QRIS', 'QRIS', 'Bayar dengan scan QRIS', 'https://cdn.gate.co.id/payment/qris.webp',
 (SELECT id FROM payment_channel_categories WHERE code = 'E_WALLET'),
 'PERCENTAGE', 0, 0.70, 1000, 10000000, ARRAY['purchase', 'deposit']::payment_type[],
 'Scan QR code menggunakan aplikasi e-wallet favoritmu', TRUE, TRUE, 1),

('DANA', 'DANA', 'Bayar dengan DANA', 'https://cdn.gate.co.id/payment/dana.webp',
 (SELECT id FROM payment_channel_categories WHERE code = 'E_WALLET'),
 'FIXED', 1500, 0, 10000, 10000000, ARRAY['purchase', 'deposit']::payment_type[],
 'Kamu akan diarahkan ke aplikasi DANA untuk menyelesaikan pembayaran', TRUE, FALSE, 2),

('GOPAY', 'GoPay', 'Bayar dengan GoPay', 'https://cdn.gate.co.id/payment/gopay.webp',
 (SELECT id FROM payment_channel_categories WHERE code = 'E_WALLET'),
 'PERCENTAGE', 0, 2.00, 10000, 10000000, ARRAY['purchase']::payment_type[],
 'Kamu akan diarahkan ke aplikasi Gojek untuk menyelesaikan pembayaran', TRUE, FALSE, 3),

('SHOPEEPAY', 'ShopeePay', 'Bayar dengan ShopeePay', 'https://cdn.gate.co.id/payment/shopeepay.webp',
 (SELECT id FROM payment_channel_categories WHERE code = 'E_WALLET'),
 'PERCENTAGE', 0, 1.50, 10000, 10000000, ARRAY['purchase']::payment_type[],
 'Kamu akan diarahkan ke aplikasi Shopee untuk menyelesaikan pembayaran', TRUE, FALSE, 4),

-- Virtual Account
('BCA_VA', 'BCA Virtual Account', 'Transfer via BCA', 'https://cdn.gate.co.id/payment/bca.webp',
 (SELECT id FROM payment_channel_categories WHERE code = 'VIRTUAL_ACCOUNT'),
 'FIXED', 4000, 0, 10000, 100000000, ARRAY['purchase', 'deposit']::payment_type[],
 'Transfer ke nomor Virtual Account yang akan diberikan', TRUE, TRUE, 1),

('BRI_VA', 'BRI Virtual Account', 'Transfer via BRI', 'https://cdn.gate.co.id/payment/bri.webp',
 (SELECT id FROM payment_channel_categories WHERE code = 'VIRTUAL_ACCOUNT'),
 'FIXED', 4000, 0, 10000, 100000000, ARRAY['purchase', 'deposit']::payment_type[],
 'Transfer ke nomor Virtual Account yang akan diberikan', TRUE, FALSE, 2),

('MANDIRI_VA', 'Mandiri Virtual Account', 'Transfer via Mandiri', 'https://cdn.gate.co.id/payment/mandiri.webp',
 (SELECT id FROM payment_channel_categories WHERE code = 'VIRTUAL_ACCOUNT'),
 'FIXED', 4000, 0, 10000, 100000000, ARRAY['purchase', 'deposit']::payment_type[],
 'Transfer ke nomor Virtual Account yang akan diberikan', TRUE, FALSE, 3),

('PERMATA_VA', 'Permata Virtual Account', 'Transfer via Permata', 'https://cdn.gate.co.id/payment/permata.webp',
 (SELECT id FROM payment_channel_categories WHERE code = 'VIRTUAL_ACCOUNT'),
 'FIXED', 4000, 0, 10000, 100000000, ARRAY['purchase', 'deposit']::payment_type[],
 'Transfer ke nomor Virtual Account yang akan diberikan', TRUE, FALSE, 4),

-- Saldo Gate (internal balance)
('BALANCE', 'Saldo Gate', 'Bayar dengan saldo Gate', 'https://cdn.gate.co.id/payment/balance.webp',
 (SELECT id FROM payment_channel_categories WHERE code = 'E_WALLET'),
 'FIXED', 0, 0, 1000, 100000000, ARRAY['purchase']::payment_type[],
 'Pembayaran langsung menggunakan saldo Gate kamu', TRUE, TRUE, 0);

-- Payment channels available in Indonesia
INSERT INTO payment_channel_regions (channel_id, region_code)
SELECT pc.id, 'ID'::region_code FROM payment_channels pc;

-- Link payment channels to gateways
INSERT INTO payment_channel_gateways (channel_id, gateway_id, payment_type, is_active) VALUES
((SELECT id FROM payment_channels WHERE code = 'QRIS'), (SELECT id FROM payment_gateways WHERE code = 'LINKQU'), 'purchase', TRUE),
((SELECT id FROM payment_channels WHERE code = 'QRIS'), (SELECT id FROM payment_gateways WHERE code = 'LINKQU'), 'deposit', TRUE),
((SELECT id FROM payment_channels WHERE code = 'DANA'), (SELECT id FROM payment_gateways WHERE code = 'DANA_DIRECT'), 'purchase', TRUE),
((SELECT id FROM payment_channels WHERE code = 'GOPAY'), (SELECT id FROM payment_gateways WHERE code = 'MIDTRANS'), 'purchase', TRUE),
((SELECT id FROM payment_channels WHERE code = 'SHOPEEPAY'), (SELECT id FROM payment_gateways WHERE code = 'MIDTRANS'), 'purchase', TRUE),
((SELECT id FROM payment_channels WHERE code = 'BCA_VA'), (SELECT id FROM payment_gateways WHERE code = 'BCA_DIRECT'), 'purchase', TRUE),
((SELECT id FROM payment_channels WHERE code = 'BCA_VA'), (SELECT id FROM payment_gateways WHERE code = 'BCA_DIRECT'), 'deposit', TRUE),
((SELECT id FROM payment_channels WHERE code = 'BRI_VA'), (SELECT id FROM payment_gateways WHERE code = 'BRI_DIRECT'), 'purchase', TRUE),
((SELECT id FROM payment_channels WHERE code = 'BRI_VA'), (SELECT id FROM payment_gateways WHERE code = 'BRI_DIRECT'), 'deposit', TRUE),
((SELECT id FROM payment_channels WHERE code = 'MANDIRI_VA'), (SELECT id FROM payment_gateways WHERE code = 'XENDIT'), 'purchase', TRUE),
((SELECT id FROM payment_channels WHERE code = 'MANDIRI_VA'), (SELECT id FROM payment_gateways WHERE code = 'XENDIT'), 'deposit', TRUE),
((SELECT id FROM payment_channels WHERE code = 'PERMATA_VA'), (SELECT id FROM payment_gateways WHERE code = 'XENDIT'), 'purchase', TRUE),
((SELECT id FROM payment_channels WHERE code = 'PERMATA_VA'), (SELECT id FROM payment_gateways WHERE code = 'XENDIT'), 'deposit', TRUE);

-- ============================================
-- SAMPLE BANNERS
-- ============================================

INSERT INTO banners (title, description, href, image, is_active, sort_order, start_at, expired_at) VALUES
('Promo Tahun Baru 2025', 'Diskon hingga 20% untuk semua top up game!', '/promo/new-year-2025', 'https://cdn.gate.co.id/banners/new-year-2025.webp', TRUE, 1, NOW(), NOW() + INTERVAL '30 days'),
('Mobile Legends x Transformers', 'Event kolaborasi terbaru sudah hadir!', '/products/mobile-legends', 'https://cdn.gate.co.id/banners/mlbb-transformers.webp', TRUE, 2, NOW(), NOW() + INTERVAL '14 days'),
('Top Up Pertama Bonus 10%', 'Khusus pengguna baru, top up pertama dapat bonus!', '/promo/first-topup', 'https://cdn.gate.co.id/banners/first-topup.webp', TRUE, 3, NOW(), NOW() + INTERVAL '90 days');

-- All banners for Indonesia
INSERT INTO banner_regions (banner_id, region_code)
SELECT b.id, 'ID'::region_code FROM banners b;

-- ============================================
-- SAMPLE PROMO
-- ============================================

INSERT INTO promos (code, title, description, note, max_usage, max_daily_usage, max_usage_per_id, min_amount, max_promo_amount, promo_flat, promo_percentage, start_at, expired_at, is_active) VALUES
('NEWYEAR25', 'Promo Tahun Baru 2025', 'Diskon 10% untuk semua transaksi', 'Berlaku untuk semua produk', 1000, 100, 1, 50000, 50000, 0, 10.00, NOW(), NOW() + INTERVAL '30 days', TRUE),
('FIRSTBUY', 'Promo First Buy', 'Bonus 5% untuk pembelian pertama', 'Khusus pengguna baru', NULL, NULL, 1, 20000, 25000, 0, 5.00, NOW(), NOW() + INTERVAL '90 days', TRUE),
('MLBB10K', 'Diskon MLBB 10K', 'Potongan Rp10.000 untuk Mobile Legends', 'Khusus produk Mobile Legends', 500, 50, 2, 80000, 10000, 10000, 0, NOW(), NOW() + INTERVAL '14 days', TRUE);

-- Link promo to products
INSERT INTO promo_products (promo_id, product_id)
SELECT p.id, pr.id FROM promos p, products pr WHERE p.code = 'MLBB10K' AND pr.code = 'mlbb';

-- Link promo to regions
INSERT INTO promo_regions (promo_id, region_code)
SELECT p.id, 'ID'::region_code FROM promos p;

-- ============================================
-- SAMPLE USER
-- Password: User@123
-- ============================================

INSERT INTO users (first_name, last_name, email, email_verified_at, phone_number, password_hash, status, primary_region, current_region, balance_idr, membership_level) VALUES
('John', 'Doe', 'john@example.com', NOW(), '081234567890', crypt('User@123', gen_salt('bf', 12)), 'ACTIVE', 'ID', 'ID', 500000, 'CLASSIC'),
('Jane', 'Smith', 'jane@example.com', NOW(), '081234567891', crypt('User@123', gen_salt('bf', 12)), 'ACTIVE', 'ID', 'ID', 1500000, 'PRESTIGE'),
('Admin', 'Test', 'admin@example.com', NOW(), '081234567892', crypt('User@123', gen_salt('bf', 12)), 'ACTIVE', 'ID', 'ID', 0, 'CLASSIC');

-- ============================================
-- ADDITIONAL ADMIN ACCOUNTS
-- ============================================

-- Admin (regular) - Password: Admin@123
INSERT INTO admins (name, email, password_hash, role_id, status)
SELECT 
    'Admin Gate',
    'admin@gate.co.id',
    crypt('Admin@123', gen_salt('bf', 12)),
    r.id,
    'ACTIVE'
FROM roles r WHERE r.code = 'ADMIN';

-- Finance - Password: Finance@123
INSERT INTO admins (name, email, password_hash, role_id, status)
SELECT 
    'Finance Gate',
    'finance@gate.co.id',
    crypt('Finance@123', gen_salt('bf', 12)),
    r.id,
    'ACTIVE'
FROM roles r WHERE r.code = 'FINANCE';

-- CS Lead - Password: CSLead@123
INSERT INTO admins (name, email, password_hash, role_id, status)
SELECT 
    'CS Lead Gate',
    'cslead@gate.co.id',
    crypt('CSLead@123', gen_salt('bf', 12)),
    r.id,
    'ACTIVE'
FROM roles r WHERE r.code = 'CS_LEAD';

-- CS - Password: CS@12345
INSERT INTO admins (name, email, password_hash, role_id, status)
SELECT 
    'CS Gate',
    'cs@gate.co.id',
    crypt('CS@12345', gen_salt('bf', 12)),
    r.id,
    'ACTIVE'
FROM roles r WHERE r.code = 'CS';

