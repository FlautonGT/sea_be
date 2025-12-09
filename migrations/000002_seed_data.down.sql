-- Migration: 000002_seed_data (ROLLBACK)
-- Description: Remove seed data
-- Created: 2025-12-03

DELETE FROM settings;
DELETE FROM contacts;
DELETE FROM popups;
DELETE FROM category_regions;
DELETE FROM categories;
DELETE FROM sections;
DELETE FROM providers;
DELETE FROM payment_gateways;
DELETE FROM payment_channel_categories;
DELETE FROM admins;
DELETE FROM role_permissions;
DELETE FROM roles;
DELETE FROM permissions;
DELETE FROM languages;
DELETE FROM regions;

