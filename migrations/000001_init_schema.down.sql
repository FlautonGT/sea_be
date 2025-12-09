-- Migration: 000001_init_schema (ROLLBACK)
-- Description: Drop all tables and types
-- Created: 2025-12-03

-- Drop triggers first
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_admins_updated_at ON admins;
DROP TRIGGER IF EXISTS update_products_updated_at ON products;
DROP TRIGGER IF EXISTS update_skus_updated_at ON skus;
DROP TRIGGER IF EXISTS update_sku_pricing_updated_at ON sku_pricing;
DROP TRIGGER IF EXISTS update_transactions_updated_at ON transactions;
DROP TRIGGER IF EXISTS update_deposits_updated_at ON deposits;
DROP TRIGGER IF EXISTS update_providers_updated_at ON providers;
DROP TRIGGER IF EXISTS update_payment_gateways_updated_at ON payment_gateways;
DROP TRIGGER IF EXISTS update_payment_channels_updated_at ON payment_channels;
DROP TRIGGER IF EXISTS update_promos_updated_at ON promos;
DROP TRIGGER IF EXISTS update_banners_updated_at ON banners;
DROP TRIGGER IF EXISTS update_popups_updated_at ON popups;
DROP TRIGGER IF EXISTS update_categories_updated_at ON categories;
DROP TRIGGER IF EXISTS update_sections_updated_at ON sections;
DROP TRIGGER IF EXISTS update_regions_updated_at ON regions;
DROP TRIGGER IF EXISTS update_languages_updated_at ON languages;
DROP TRIGGER IF EXISTS update_roles_updated_at ON roles;
DROP TRIGGER IF EXISTS update_settings_updated_at ON settings;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS contacts CASCADE;
DROP TABLE IF EXISTS settings CASCADE;
DROP TABLE IF EXISTS popups CASCADE;
DROP TABLE IF EXISTS banner_regions CASCADE;
DROP TABLE IF EXISTS banners CASCADE;
DROP TABLE IF EXISTS refunds CASCADE;
DROP TABLE IF EXISTS mutations CASCADE;
DROP TABLE IF EXISTS deposit_logs CASCADE;
DROP TABLE IF EXISTS deposits CASCADE;
DROP TABLE IF EXISTS transaction_logs CASCADE;
DROP TABLE IF EXISTS transactions CASCADE;
DROP TABLE IF EXISTS promo_usages CASCADE;
DROP TABLE IF EXISTS promo_regions CASCADE;
DROP TABLE IF EXISTS promo_payment_channels CASCADE;
DROP TABLE IF EXISTS promo_products CASCADE;
DROP TABLE IF EXISTS promos CASCADE;
DROP TABLE IF EXISTS payment_channel_gateways CASCADE;
DROP TABLE IF EXISTS payment_channel_regions CASCADE;
DROP TABLE IF EXISTS payment_channels CASCADE;
DROP TABLE IF EXISTS payment_channel_categories CASCADE;
DROP TABLE IF EXISTS payment_gateways CASCADE;
DROP TABLE IF EXISTS sku_pricing CASCADE;
DROP TABLE IF EXISTS skus CASCADE;
DROP TABLE IF EXISTS providers CASCADE;
DROP TABLE IF EXISTS product_sections CASCADE;
DROP TABLE IF EXISTS sections CASCADE;
DROP TABLE IF EXISTS product_fields CASCADE;
DROP TABLE IF EXISTS product_regions CASCADE;
DROP TABLE IF EXISTS products CASCADE;
DROP TABLE IF EXISTS category_regions CASCADE;
DROP TABLE IF EXISTS categories CASCADE;
DROP TABLE IF EXISTS admin_sessions CASCADE;
DROP TABLE IF EXISTS admins CASCADE;
DROP TABLE IF EXISTS role_permissions CASCADE;
DROP TABLE IF EXISTS roles CASCADE;
DROP TABLE IF EXISTS permissions CASCADE;
DROP TABLE IF EXISTS email_verifications CASCADE;
DROP TABLE IF EXISTS password_resets CASCADE;
DROP TABLE IF EXISTS user_sessions CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS languages CASCADE;
DROP TABLE IF EXISTS regions CASCADE;

-- Drop enum types
DROP TYPE IF EXISTS membership_level CASCADE;
DROP TYPE IF EXISTS audit_action CASCADE;
DROP TYPE IF EXISTS field_type CASCADE;
DROP TYPE IF EXISTS fee_type CASCADE;
DROP TYPE IF EXISTS payment_type CASCADE;
DROP TYPE IF EXISTS health_status CASCADE;
DROP TYPE IF EXISTS mutation_type CASCADE;
DROP TYPE IF EXISTS deposit_status CASCADE;
DROP TYPE IF EXISTS payment_status CASCADE;
DROP TYPE IF EXISTS transaction_status CASCADE;
DROP TYPE IF EXISTS currency_code CASCADE;
DROP TYPE IF EXISTS region_code CASCADE;
DROP TYPE IF EXISTS admin_role CASCADE;
DROP TYPE IF EXISTS mfa_status CASCADE;
DROP TYPE IF EXISTS user_status CASCADE;

-- Drop extensions
DROP EXTENSION IF EXISTS "pgcrypto";
DROP EXTENSION IF EXISTS "uuid-ossp";

