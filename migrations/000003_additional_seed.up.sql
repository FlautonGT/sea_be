-- Migration: 000003_additional_seed
-- Description: Additional seed data for testing
-- Created: 2025-12-04

-- ============================================
-- ADDITIONAL PRODUCTS - PUBG Mobile
-- ============================================

-- Add product fields for PUBG Mobile and other games
INSERT INTO product_fields (product_id, name, key, field_type, label, placeholder, hint, is_required, sort_order)
SELECT id, 'User ID', 'userId', 'number', 'ID PUBG Mobile', 'Contoh: 5123456789', 'ID terletak di profil game', TRUE, 1
FROM products WHERE code = 'pubgm'
ON CONFLICT DO NOTHING;

INSERT INTO product_fields (product_id, name, key, field_type, label, placeholder, hint, is_required, sort_order)
SELECT id, 'UID', 'uid', 'number', 'UID', 'Contoh: 812345678', 'UID Genshin Impact', TRUE, 1
FROM products WHERE code = 'genshin'
ON CONFLICT DO NOTHING;

INSERT INTO product_fields (product_id, name, key, field_type, label, placeholder, hint, is_required, sort_order)
SELECT id, 'Server', 'server', 'select', 'Server', 'Pilih server', 'Pilih server yang sesuai', TRUE, 2
FROM products WHERE code = 'genshin'
ON CONFLICT DO NOTHING;

-- ============================================
-- ADDITIONAL SKUS - PUBG Mobile
-- ============================================

INSERT INTO skus (code, provider_sku_code, name, description, image, product_id, provider_id, section_id, process_time, is_active, is_featured, badge_text, badge_color, total_sold) VALUES
('pubgm-60-uc', 'pubgm60', '60 UC', '60 Unknown Cash PUBG Mobile', 'https://cdn.gate.co.id/sku/pubgm-uc.webp',
 (SELECT id FROM products WHERE code = 'pubgm'),
 (SELECT id FROM providers WHERE code = 'DIGIFLAZZ'),
 (SELECT id FROM sections WHERE code = 'topup-instant'),
 60, TRUE, FALSE, NULL, NULL, 5200),

('pubgm-325-uc', 'pubgm325', '325 UC', '325 Unknown Cash PUBG Mobile', 'https://cdn.gate.co.id/sku/pubgm-uc.webp',
 (SELECT id FROM products WHERE code = 'pubgm'),
 (SELECT id FROM providers WHERE code = 'DIGIFLAZZ'),
 (SELECT id FROM sections WHERE code = 'topup-instant'),
 60, TRUE, FALSE, NULL, NULL, 8900),

('pubgm-660-uc', 'pubgm660', '660 UC', '660 Unknown Cash PUBG Mobile', 'https://cdn.gate.co.id/sku/pubgm-uc.webp',
 (SELECT id FROM products WHERE code = 'pubgm'),
 (SELECT id FROM providers WHERE code = 'DIGIFLAZZ'),
 (SELECT id FROM sections WHERE code = 'topup-instant'),
 60, TRUE, TRUE, 'BEST SELLER', '#FF6B6B', 12500),

('pubgm-1800-uc', 'pubgm1800', '1800 UC', '1800 Unknown Cash PUBG Mobile', 'https://cdn.gate.co.id/sku/pubgm-uc.webp',
 (SELECT id FROM products WHERE code = 'pubgm'),
 (SELECT id FROM providers WHERE code = 'DIGIFLAZZ'),
 (SELECT id FROM sections WHERE code = 'topup-instant'),
 60, TRUE, FALSE, NULL, NULL, 4300)
ON CONFLICT (code) DO NOTHING;

-- PUBG Mobile SKU Pricing
INSERT INTO sku_pricing (sku_id, region_code, currency, buy_price, sell_price, original_price, is_active)
SELECT id, 'ID', 'IDR', 15000, 17000, 19000, TRUE FROM skus WHERE code = 'pubgm-60-uc'
ON CONFLICT DO NOTHING;

INSERT INTO sku_pricing (sku_id, region_code, currency, buy_price, sell_price, original_price, is_active)
SELECT id, 'ID', 'IDR', 75000, 85000, 95000, TRUE FROM skus WHERE code = 'pubgm-325-uc'
ON CONFLICT DO NOTHING;

INSERT INTO sku_pricing (sku_id, region_code, currency, buy_price, sell_price, original_price, is_active)
SELECT id, 'ID', 'IDR', 150000, 165000, 180000, TRUE FROM skus WHERE code = 'pubgm-660-uc'
ON CONFLICT DO NOTHING;

INSERT INTO sku_pricing (sku_id, region_code, currency, buy_price, sell_price, original_price, is_active)
SELECT id, 'ID', 'IDR', 400000, 440000, 480000, TRUE FROM skus WHERE code = 'pubgm-1800-uc'
ON CONFLICT DO NOTHING;

-- ============================================
-- ADDITIONAL SKUS - Genshin Impact
-- ============================================

INSERT INTO skus (code, provider_sku_code, name, description, image, product_id, provider_id, section_id, process_time, is_active, is_featured, badge_text, badge_color, total_sold) VALUES
('genshin-60-gc', 'genshin60', '60 Genesis Crystals', '60 Genesis Crystals Genshin Impact', 'https://cdn.gate.co.id/sku/genshin-gc.webp',
 (SELECT id FROM products WHERE code = 'genshin'),
 (SELECT id FROM providers WHERE code = 'DIGIFLAZZ'),
 (SELECT id FROM sections WHERE code = 'topup-instant'),
 60, TRUE, FALSE, NULL, NULL, 3400),

('genshin-330-gc', 'genshin330', '330 Genesis Crystals', '330 Genesis Crystals Genshin Impact', 'https://cdn.gate.co.id/sku/genshin-gc.webp',
 (SELECT id FROM products WHERE code = 'genshin'),
 (SELECT id FROM providers WHERE code = 'DIGIFLAZZ'),
 (SELECT id FROM sections WHERE code = 'topup-instant'),
 60, TRUE, FALSE, NULL, NULL, 5600),

('genshin-1090-gc', 'genshin1090', '1090 Genesis Crystals', '1090 Genesis Crystals Genshin Impact', 'https://cdn.gate.co.id/sku/genshin-gc.webp',
 (SELECT id FROM products WHERE code = 'genshin'),
 (SELECT id FROM providers WHERE code = 'DIGIFLAZZ'),
 (SELECT id FROM sections WHERE code = 'topup-instant'),
 60, TRUE, TRUE, 'POPULAR', '#4ECDC4', 8900),

('genshin-blessing', 'genshin-welkin', 'Blessing of the Welkin Moon', 'Blessing of the Welkin Moon (30 hari)', 'https://cdn.gate.co.id/sku/genshin-welkin.webp',
 (SELECT id FROM products WHERE code = 'genshin'),
 (SELECT id FROM providers WHERE code = 'DIGIFLAZZ'),
 (SELECT id FROM sections WHERE code = 'monthly-pass'),
 300, TRUE, FALSE, 'HEMAT', '#45B7D1', 6700)
ON CONFLICT (code) DO NOTHING;

-- Genshin SKU Pricing
INSERT INTO sku_pricing (sku_id, region_code, currency, buy_price, sell_price, original_price, is_active)
SELECT id, 'ID', 'IDR', 15000, 17000, 19000, TRUE FROM skus WHERE code = 'genshin-60-gc'
ON CONFLICT DO NOTHING;

INSERT INTO sku_pricing (sku_id, region_code, currency, buy_price, sell_price, original_price, is_active)
SELECT id, 'ID', 'IDR', 80000, 89000, 99000, TRUE FROM skus WHERE code = 'genshin-330-gc'
ON CONFLICT DO NOTHING;

INSERT INTO sku_pricing (sku_id, region_code, currency, buy_price, sell_price, original_price, is_active)
SELECT id, 'ID', 'IDR', 250000, 275000, 299000, TRUE FROM skus WHERE code = 'genshin-1090-gc'
ON CONFLICT DO NOTHING;

INSERT INTO sku_pricing (sku_id, region_code, currency, buy_price, sell_price, original_price, is_active)
SELECT id, 'ID', 'IDR', 75000, 82000, 89000, TRUE FROM skus WHERE code = 'genshin-blessing'
ON CONFLICT DO NOTHING;

-- ============================================
-- ADDITIONAL USERS FOR TESTING
-- ============================================

INSERT INTO users (first_name, last_name, email, email_verified_at, phone_number, password_hash, status, primary_region, current_region, balance_idr, membership_level) VALUES
('Alex', 'Johnson', 'alex@example.com', NOW(), '081234567893', crypt('User@123', gen_salt('bf', 12)), 'ACTIVE', 'ID', 'ID', 250000, 'CLASSIC'),
('Maria', 'Garcia', 'maria@example.com', NOW(), '081234567894', crypt('User@123', gen_salt('bf', 12)), 'ACTIVE', 'ID', 'ID', 750000, 'CLASSIC'),
('David', 'Chen', 'david@example.com', NOW(), '081234567895', crypt('User@123', gen_salt('bf', 12)), 'ACTIVE', 'ID', 'ID', 2500000, 'ROYAL'),
('Sarah', 'Kim', 'sarah@example.com', NOW(), '081234567896', crypt('User@123', gen_salt('bf', 12)), 'SUSPENDED', 'ID', 'ID', 100000, 'CLASSIC'),
('Michael', 'Brown', 'michael@example.com', NOW(), '081234567897', crypt('User@123', gen_salt('bf', 12)), 'ACTIVE', 'MY', 'MY', 0, 'CLASSIC')
ON CONFLICT DO NOTHING;

-- ============================================
-- SAMPLE TRANSACTIONS FOR TESTING
-- ============================================

-- Create sample transactions
INSERT INTO transactions (invoice_number, status, payment_status, user_id, product_id, sku_id, payment_channel_id, provider_id, account_inputs, account_nickname, buy_price, sell_price, discount_amount, payment_fee, total_amount, currency, region, ip_address)
SELECT
    'GATE' || UPPER(SUBSTR(MD5(RANDOM()::TEXT), 1, 20)),
    CASE WHEN RANDOM() > 0.1 THEN 'SUCCESS'::transaction_status ELSE 'PENDING'::transaction_status END,
    'PAID'::payment_status,
    u.id,
    p.id,
    s.id,
    pc.id,
    s.provider_id,
    '{"userId": "656696292", "serverId": "8610"}'::JSONB,
    'TestUser' || FLOOR(RANDOM() * 1000)::TEXT,
    sp.buy_price,
    sp.sell_price,
    0,
    CASE WHEN pc.fee_type = 'PERCENTAGE' THEN ROUND(sp.sell_price * pc.fee_percentage / 100) ELSE pc.fee_amount END,
    sp.sell_price + CASE WHEN pc.fee_type = 'PERCENTAGE' THEN ROUND(sp.sell_price * pc.fee_percentage / 100) ELSE pc.fee_amount END,
    'IDR',
    'ID',
    ('103.123.' || FLOOR(RANDOM() * 255)::TEXT || '.' || FLOOR(RANDOM() * 255)::TEXT)::INET
FROM users u
CROSS JOIN products p
CROSS JOIN skus s
CROSS JOIN payment_channels pc
CROSS JOIN sku_pricing sp
WHERE s.product_id = p.id
  AND sp.sku_id = s.id
  AND sp.region_code = 'ID'
  AND pc.code IN ('QRIS', 'DANA', 'BCA_VA')
  AND u.email IN ('john@example.com', 'jane@example.com', 'alex@example.com')
LIMIT 50
ON CONFLICT DO NOTHING;

-- ============================================
-- SAMPLE DEPOSITS FOR TESTING
-- ============================================

INSERT INTO deposits (invoice_number, user_id, amount, payment_fee, total_amount, currency, status, payment_channel_id, region, ip_address)
SELECT
    'DEP' || UPPER(SUBSTR(MD5(RANDOM()::TEXT), 1, 20)),
    u.id,
    (ARRAY[50000, 100000, 200000, 500000, 1000000])[FLOOR(RANDOM() * 5 + 1)],
    0,
    (ARRAY[50000, 100000, 200000, 500000, 1000000])[FLOOR(RANDOM() * 5 + 1)],
    'IDR',
    CASE WHEN RANDOM() > 0.2 THEN 'SUCCESS'::deposit_status ELSE 'PENDING'::deposit_status END,
    pc.id,
    'ID'::region_code,
    ('103.123.' || FLOOR(RANDOM() * 255)::TEXT || '.' || FLOOR(RANDOM() * 255)::TEXT)::INET
FROM users u
CROSS JOIN payment_channels pc
WHERE pc.code IN ('QRIS', 'BCA_VA', 'BRI_VA')
  AND 'deposit' = ANY(pc.supported_types)
  AND u.email IN ('john@example.com', 'jane@example.com', 'david@example.com')
LIMIT 20
ON CONFLICT DO NOTHING;

-- ============================================
-- SAMPLE BALANCE MUTATIONS
-- ============================================

INSERT INTO mutations (user_id, mutation_type, amount, currency, balance_before, balance_after, reference_type, description)
SELECT
    u.id,
    'CREDIT'::mutation_type,
    100000,
    'IDR',
    u.balance_idr - 100000,
    u.balance_idr,
    'DEPOSIT',
    'Deposit via QRIS'
FROM users u
WHERE u.email = 'john@example.com'
ON CONFLICT DO NOTHING;

-- ============================================
-- ADDITIONAL PROMOS
-- ============================================

INSERT INTO promos (code, title, description, note, max_usage, max_daily_usage, max_usage_per_id, min_amount, max_promo_amount, promo_flat, promo_percentage, start_at, expired_at, is_active) VALUES
('WELCOME50', 'Welcome Promo', 'Diskon 50% untuk pengguna baru', 'Khusus pengguna baru, maks diskon Rp25.000', NULL, NULL, 1, 10000, 25000, 0, 50.00, NOW(), NOW() + INTERVAL '365 days', TRUE),
('WEEKEND20', 'Weekend Special', 'Diskon 20% untuk transaksi di akhir pekan', 'Berlaku Sabtu-Minggu', 2000, 200, 2, 50000, 30000, 0, 20.00, NOW(), NOW() + INTERVAL '30 days', TRUE),
('FLAT5K', 'Flat 5K Off', 'Potongan langsung Rp5.000', 'Minimal transaksi Rp25.000', 5000, 500, 3, 25000, 5000, 5000, 0, NOW(), NOW() + INTERVAL '60 days', TRUE),
('VIP25', 'VIP Member 25%', 'Diskon khusus member VIP', 'Khusus member Prestige/Royal', 1000, 100, 5, 100000, 100000, 0, 25.00, NOW(), NOW() + INTERVAL '90 days', TRUE)
ON CONFLICT DO NOTHING;

-- Link promos to all products
INSERT INTO promo_products (promo_id, product_id)
SELECT p.id, pr.id FROM promos p, products pr 
WHERE p.code IN ('WEEKEND20', 'FLAT5K')
ON CONFLICT DO NOTHING;

-- Link promos to all regions
INSERT INTO promo_regions (promo_id, region_code)
SELECT p.id, 'ID'::region_code FROM promos p WHERE p.code NOT IN ('NEWYEAR25', 'FIRSTBUY', 'MLBB10K')
ON CONFLICT DO NOTHING;

INSERT INTO promo_regions (promo_id, region_code)
SELECT p.id, 'MY'::region_code FROM promos p WHERE p.code IN ('WEEKEND20', 'WELCOME50')
ON CONFLICT DO NOTHING;

-- ============================================
-- UPDATE PROVIDER STATS
-- ============================================

UPDATE providers SET 
    total_skus = (SELECT COUNT(*) FROM skus WHERE provider_id = providers.id),
    active_skus = (SELECT COUNT(*) FROM skus WHERE provider_id = providers.id AND is_active = TRUE),
    success_rate = 98.5,
    avg_response_time = 1200,
    health_status = 'HEALTHY',
    last_health_check = NOW()
WHERE code = 'DIGIFLAZZ';

UPDATE providers SET 
    total_skus = (SELECT COUNT(*) FROM skus WHERE provider_id = providers.id),
    active_skus = (SELECT COUNT(*) FROM skus WHERE provider_id = providers.id AND is_active = TRUE),
    success_rate = 97.8,
    avg_response_time = 1500,
    health_status = 'HEALTHY',
    last_health_check = NOW()
WHERE code = 'VIPRESELLER';

UPDATE providers SET 
    total_skus = (SELECT COUNT(*) FROM skus WHERE provider_id = providers.id),
    active_skus = (SELECT COUNT(*) FROM skus WHERE provider_id = providers.id AND is_active = TRUE),
    success_rate = 95.2,
    avg_response_time = 2100,
    health_status = 'DEGRADED',
    last_health_check = NOW()
WHERE code = 'BANGJEFF';

-- ============================================
-- AUDIT LOGS SAMPLE
-- ============================================

INSERT INTO audit_logs (admin_id, action, resource_type, resource_id, description, changes, ip_address, user_agent)
SELECT 
    a.id,
    'CREATE',
    'PRODUCT',
    p.id::TEXT,
    'Created product ' || p.title,
    ('{"after": {"code": "' || p.code || '", "title": "' || p.title || '"}}')::JSONB,
    '103.123.45.67',
    'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36'
FROM admins a, products p
WHERE a.email = 'superadmin@gate.co.id'
LIMIT 5
ON CONFLICT DO NOTHING;

INSERT INTO audit_logs (admin_id, action, resource_type, resource_id, description, changes, ip_address, user_agent)
SELECT 
    a.id,
    'UPDATE',
    'SKU',
    s.id::TEXT,
    'Updated SKU ' || s.name || ' price',
    ('{"before": {"sellPrice": 20000}, "after": {"sellPrice": ' || sp.sell_price || '}}')::JSONB,
    '103.123.45.67',
    'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36'
FROM admins a, skus s, sku_pricing sp
WHERE a.email = 'admin@gate.co.id'
AND sp.sku_id = s.id
LIMIT 5
ON CONFLICT DO NOTHING;

-- ============================================
-- PROMO USAGE SAMPLE
-- ============================================

INSERT INTO promo_usages (promo_id, user_id, transaction_id, discount_amount, used_at)
SELECT 
    p.id,
    u.id,
    t.id,
    5000,
    t.created_at
FROM promos p
CROSS JOIN users u
CROSS JOIN transactions t
WHERE p.code = 'NEWYEAR25'
AND t.user_id = u.id
LIMIT 10
ON CONFLICT DO NOTHING;

