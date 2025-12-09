-- Migration: 000003_additional_seed (DOWN)
-- Rollback additional seed data

DELETE FROM promo_usages WHERE promo_id IN (SELECT id FROM promos WHERE code = 'NEWYEAR25');
DELETE FROM audit_logs;
DELETE FROM mutations WHERE description LIKE 'Deposit via%';
DELETE FROM deposits WHERE invoice_number LIKE 'DEP%';
DELETE FROM transactions WHERE invoice_number LIKE 'GATE%';
DELETE FROM sku_pricing WHERE sku_id IN (SELECT id FROM skus WHERE code LIKE 'pubgm-%' OR code LIKE 'genshin-%');
DELETE FROM skus WHERE code LIKE 'pubgm-%' OR code LIKE 'genshin-%';
DELETE FROM promo_regions WHERE promo_id IN (SELECT id FROM promos WHERE code IN ('WELCOME50', 'WEEKEND20', 'FLAT5K', 'VIP25'));
DELETE FROM promo_products WHERE promo_id IN (SELECT id FROM promos WHERE code IN ('WELCOME50', 'WEEKEND20', 'FLAT5K', 'VIP25'));
DELETE FROM promos WHERE code IN ('WELCOME50', 'WEEKEND20', 'FLAT5K', 'VIP25');
DELETE FROM users WHERE email IN ('alex@example.com', 'maria@example.com', 'david@example.com', 'sarah@example.com', 'michael@example.com');

