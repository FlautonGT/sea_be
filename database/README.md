# Gate Database Schema

## Overview

Gate uses **PostgreSQL 15+** as the primary database with **Redis** for caching and session management.

## Database Structure

### Entity Relationship Diagram

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   regions   │     │  languages  │     │   contacts  │
└─────────────┘     └─────────────┘     └─────────────┘

┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│    users    │────▶│  mutations  │     │  user_      │
│             │     │             │     │  sessions   │
└─────────────┘     └─────────────┘     └─────────────┘
       │
       │  ┌─────────────┐     ┌─────────────┐
       └─▶│  deposits   │────▶│ deposit_logs│
          └─────────────┘     └─────────────┘

┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   admins    │────▶│   roles     │────▶│ permissions │
└─────────────┘     └─────────────┘     └─────────────┘

┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│ categories  │────▶│  products   │────▶│    skus     │
└─────────────┘     │             │     │             │
                    └─────────────┘     └─────────────┘
                           │                   │
                           │                   ▼
                           │            ┌─────────────┐
                           │            │ sku_pricing │
                           │            └─────────────┘
                           │
                           ▼
                    ┌─────────────┐     ┌─────────────┐
                    │  sections   │     │  providers  │
                    └─────────────┘     └─────────────┘

┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  payment_   │────▶│  payment_   │────▶│  payment_   │
│  gateways   │     │  channels   │     │  channel_   │
└─────────────┘     └─────────────┘     │  categories │
                                        └─────────────┘

┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│transactions │────▶│transaction_ │     │   promos    │
│             │     │   logs      │     │             │
└─────────────┘     └─────────────┘     └─────────────┘

┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   banners   │     │   popups    │     │ audit_logs  │
└─────────────┘     └─────────────┘     └─────────────┘
```

## Tables

### Core Tables

| Table | Description |
|-------|-------------|
| `regions` | Supported regions (ID, MY, PH, SG, TH) |
| `languages` | Supported languages |
| `settings` | Application settings |
| `contacts` | Contact information |

### User Tables

| Table | Description |
|-------|-------------|
| `users` | User accounts with multi-currency balance |
| `user_sessions` | Active user sessions |
| `password_resets` | Password reset tokens |
| `email_verifications` | Email verification tokens |
| `mutations` | Balance history (debit/credit) |

### Admin Tables

| Table | Description |
|-------|-------------|
| `admins` | Admin accounts |
| `roles` | Admin roles (SUPERADMIN, ADMIN, FINANCE, CS_LEAD, CS) |
| `permissions` | System permissions |
| `role_permissions` | Role-permission mappings |
| `admin_sessions` | Active admin sessions |
| `audit_logs` | Admin action logs |

### Product Tables

| Table | Description |
|-------|-------------|
| `categories` | Product categories |
| `products` | Products (games, e-wallets, etc) |
| `product_fields` | Input fields per product |
| `sections` | SKU groupings (Spesial Item, Topup Instan) |
| `skus` | Stock Keeping Units |
| `sku_pricing` | SKU prices per region |

### Provider Tables

| Table | Description |
|-------|-------------|
| `providers` | Product providers (Digiflazz, VIP Reseller, BangJeff) |

### Payment Tables

| Table | Description |
|-------|-------------|
| `payment_gateways` | Payment providers (LinkQu, BCA, Xendit, etc) |
| `payment_channel_categories` | Payment categories (E-Wallet, VA, etc) |
| `payment_channels` | Payment methods (QRIS, DANA, BCA VA, etc) |
| `payment_channel_gateways` | Channel-gateway assignments |

### Transaction Tables

| Table | Description |
|-------|-------------|
| `transactions` | Purchase transactions |
| `transaction_logs` | Transaction timeline/logs |
| `deposits` | Balance top-up transactions |
| `deposit_logs` | Deposit timeline/logs |
| `refunds` | Refund records |

### Promo Tables

| Table | Description |
|-------|-------------|
| `promos` | Promo codes |
| `promo_products` | Promo-product restrictions |
| `promo_payment_channels` | Promo-payment restrictions |
| `promo_regions` | Promo-region restrictions |
| `promo_usages` | Promo usage tracking |

### Content Tables

| Table | Description |
|-------|-------------|
| `banners` | Homepage banners |
| `popups` | Promotional popups |

## Key Features

### Multi-Currency Balance

Users have separate balances for each supported currency:

```sql
balance_idr BIGINT DEFAULT 0,  -- Indonesian Rupiah
balance_myr BIGINT DEFAULT 0,  -- Malaysian Ringgit
balance_php BIGINT DEFAULT 0,  -- Philippine Peso
balance_sgd BIGINT DEFAULT 0,  -- Singapore Dollar
balance_thb BIGINT DEFAULT 0,  -- Thai Baht
```

### Computed Columns

SKU pricing uses generated columns for margin and discount:

```sql
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
```

### Automatic Timestamps

All tables have `updated_at` trigger:

```sql
CREATE TRIGGER update_users_updated_at 
BEFORE UPDATE ON users 
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

## Indexes

Key indexes for performance:

- `idx_users_email` - User lookup by email
- `idx_transactions_invoice` - Invoice number search
- `idx_transactions_created_at` - Date-based queries
- `idx_skus_product` - SKUs by product
- `idx_mutations_user` - User balance history

## Migrations

Migrations are in `/migrations` folder:

| File | Description |
|------|-------------|
| `000001_init_schema.up.sql` | Initial schema |
| `000001_init_schema.down.sql` | Rollback schema |
| `000002_seed_data.up.sql` | Seed data |
| `000002_seed_data.down.sql` | Remove seed data |

### Running Migrations

```bash
# Using Make
make migrate-up
make migrate-down

# Using Docker
docker-compose run --rm migrate up
docker-compose run --rm migrate down 1
```

## Redis Usage

Redis is used for:

1. **Session Storage** - User and admin sessions
2. **Rate Limiting** - API rate limiting
3. **Caching** - Product, SKU, and pricing cache
4. **Queues** - Background job processing

### Key Patterns

```
session:user:{user_id}        - User session
session:admin:{admin_id}      - Admin session
rate_limit:{ip}:{endpoint}    - Rate limit counters
cache:products:{region}       - Product list cache
cache:skus:{product}:{region} - SKU list cache
cache:pricing:{sku}:{region}  - SKU pricing cache
```

## Security

1. **Passwords** - bcrypt with cost 12
2. **Tokens** - SHA-256 hashed before storage
3. **MFA Secrets** - Encrypted at rest
4. **API Credentials** - Stored in `.env`, NOT database

