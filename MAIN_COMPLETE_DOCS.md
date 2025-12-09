# Gate.co.id API Documentation v2.0

> **Last Updated:** December 3, 2025 | **API Version:** v2.0

---

## Table of Contents

1. [Base Information](#base-information)
2. [Authentication](#authentication)
3. [Response Structure](#response-structure)
4. [Public Endpoints](#public-endpoints)
5. [Product Endpoints](#product-endpoints)
6. [Transaction Endpoints](#transaction-endpoints)
7. [Auth Endpoints](#auth-endpoints)
8. [User Dashboard Endpoints](#user-dashboard-endpoints)
9. [Error Codes](#error-codes)
10. [Best Practices](#best-practices)

---

## Base Information

### URLs

| Type | URL |
|------|-----|
| Main URL | `https://gate.id` or `https://gate.co.id` |
| API URL | `https://gateway.gate.id` or `https://gateway.gate.co.id` |

### API Paths

| Path | Description |
|------|-------------|
| `/{version}` | Main API path (e.g., `/v2`) |
| `/admin/{version}` | Admin API path (e.g., `/admin/v2`) |

### Timezone & DateTime

- **Server Timezone:** Jakarta Time (UTC+07:00)
- **DateTime Format:** ISO 8601 with timezone offset
- **Example:** `2025-12-31T23:59:59+07:00`

### Supported Regions

| Code | Country | Currency | Symbol |
|------|---------|----------|--------|
| ID | Indonesia | IDR | Rp |
| MY | Malaysia | MYR | RM |
| PH | Philippines | PHP | ₱ |
| SG | Singapore | SGD | S$ |
| TH | Thailand | THB | ฿ |

---

## Authentication

### Bearer Token (Optional)

Authentication is **optional** for most endpoints. Gate allows users to make purchases without logging in.

**When to include authentication:**

| Scenario | Auth Required | Benefit |
|----------|---------------|---------|
| Guest Purchase | ❌ No | Transaction processed but not linked to account |
| Logged-in Purchase | ✅ Yes | Transaction saved to user's account history |
| Dashboard Access | ✅ Yes | Required for all dashboard endpoints |

**Header Format:**

```
Authorization: Bearer {your_access_token}
```

---

## Response Structure

All API responses follow camelCase naming convention.

### Success Response

```json
{
    "data": {
        // Object or Array
    }
}
```

**Single Object Example:**

```json
{
    "data": {
        "title": "Top Up Game",
        "code": "top-up-game"
    }
}
```

**Array Example:**

```json
{
    "data": [
        { "title": "Top Up Game", "code": "top-up-game" },
        { "title": "Voucher", "code": "voucher" }
    ]
}
```

### Error Response

```json
{
    "error": {
        "code": "ERROR_CODE",
        "message": "Human readable error message",
        "details": "Additional error details (optional)",
        "fields": {
            "fieldName": "Field specific error message"
        }
    }
}
```

### HTTP Status Codes

| Code | Description |
|------|-------------|
| 200 | Success |
| 201 | Created |
| 400 | Bad Request |
| 401 | Unauthorized |
| 404 | Not Found |
| 422 | Validation Error |
| 429 | Too Many Requests |
| 500 | Internal Server Error |

---

## Public Endpoints

### 1. Get Regions

Retrieve all available regions/countries with their configuration.

**Endpoint:** `GET /v2/regions`

**Response:**

```json
{
    "data": [
        {
            "country": "Indonesia",
            "code": "ID",
            "currency": "IDR",
            "image": "https://nos.jkt-1.neo.id/gate/flags/id.svg",
            "isDefault": true
        },
        {
            "country": "Malaysia",
            "code": "MY",
            "currency": "MYR",
            "image": "https://nos.jkt-1.neo.id/gate/flags/my.svg",
            "isDefault": false
        },
        {
            "country": "Philippines",
            "code": "PH",
            "currency": "PHP",
            "image": "https://nos.jkt-1.neo.id/gate/flags/ph.svg",
            "isDefault": false
        },
        {
            "country": "Singapore",
            "code": "SG",
            "currency": "SGD",
            "image": "https://nos.jkt-1.neo.id/gate/flags/sg.svg",
            "isDefault": false
        },
        {
            "country": "Thailand",
            "code": "TH",
            "currency": "THB",
            "image": "https://nos.jkt-1.neo.id/gate/flags/th.svg",
            "isDefault": false
        }
    ]
}
```

| Field | Description |
|-------|-------------|
| `country` | Full country name |
| `code` | ISO 3166-1 alpha-2 country code |
| `currency` | ISO 4217 currency code |
| `image` | URL to country flag image (SVG) |
| `isDefault` | Default selected region |

---

### 2. Get Languages

Retrieve all available languages for the platform.

**Endpoint:** `GET /v2/languages`

**Response:**

```json
{
    "data": [
        {
            "country": "Indonesia",
            "code": "id",
            "name": "Bahasa Indonesia",
            "image": "https://nos.jkt-1.neo.id/gate/flags/id.svg",
            "isDefault": true
        },
        {
            "country": "United States",
            "code": "en",
            "name": "English",
            "image": "https://nos.jkt-1.neo.id/gate/flags/us.svg",
            "isDefault": false
        }
    ]
}
```

| Field | Description |
|-------|-------------|
| `country` | Country name associated with the language |
| `code` | ISO 639-1 language code |
| `name` | Display name of the language |
| `image` | URL to flag image |
| `isDefault` | Default selected language |

---

### 3. Get Contacts

Retrieve all contact information and social media links.

**Endpoint:** `GET /v2/contacts`

**Response:**

```json
{
    "data": {
        "email": "support@gate.co.id",
        "phone": "+6281234567890",
        "whatsapp": "https://wa.me/6281234567890",
        "instagram": "https://instagram.com/gate.official",
        "facebook": "https://facebook.com/gate.official",
        "x": "https://x.com/gate_official",
        "youtube": "https://youtube.com/@gateofficial",
        "telegram": "https://t.me/gate_official",
        "discord": "https://discord.gg/gate"
    }
}
```

---

### 4. Get Popups

Display promotional popups to users.

**Endpoint:** `GET /v2/popups`

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| region | string | Yes | Region code (ID, MY, PH, SG, TH) |

**Response:**

```json
{
    "data": {
        "title": "🔥 TOP UP TERMURAH HANYA DI GATE! 🔥",
        "content": "<p>🎉 <em>Diskon spesial hanya berlaku sampai 23 Desember!</em> 🎉</p>",
        "image": "https://nos.jkt-1.neo.id/gate/popups/delta-force-promo.webp",
        "href": "/id-id/delta-force",
        "isActive": true
    }
}
```

**Error Response:**

```json
{
    "error": {
        "code": "INVALID_REGION",
        "message": "Region tidak valid",
        "details": "Region code must be one of: ID, MY, PH, SG, TH"
    }
}
```

---

### 5. Get Banners

Retrieve promotional banners for the homepage.

**Endpoint:** `GET /v2/banners`

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| region | string | Yes | Region code (ID, MY, PH, SG, TH) |

**Response:**

```json
{
    "data": [
        {
            "title": "Topup Game Cepat & Murah",
            "description": "Dapatkan diamond dan UC dengan harga terbaik",
            "href": "/id-id/mobile-legends",
            "image": "https://nos.jkt-1.neo.id/gate/banners/mobile-legends-banner.webp",
            "order": 1
        },
        {
            "title": "Promo Spesial Free Fire",
            "description": "Diskon hingga 20% untuk semua diamond Free Fire",
            "href": "/id-id/free-fire",
            "image": "https://nos.jkt-1.neo.id/gate/banners/free-fire-promo.webp",
            "order": 2
        }
    ]
}
```

| Field | Description |
|-------|-------------|
| `title` | Banner title |
| `description` | Banner description/subtitle |
| `href` | Link destination (relative or absolute URL) |
| `image` | Banner image URL (recommended: 1200x400px) |
| `order` | Display order (ascending) |

---

## Product Endpoints

### 6. Get Categories

Retrieve all available product categories.

**Endpoint:** `GET /v2/categories`

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| region | string | Yes | Region code (ID, MY, PH, SG, TH) |

**Response:**

```json
{
    "data": [
        {
            "title": "Top Up Game",
            "code": "top-up-game",
            "description": "Top up diamond, UC, dan in-game currency lainnya",
            "icon": "https://nos.jkt-1.neo.id/gate/icons/game-controller.svg",
            "order": 1
        },
        {
            "title": "Voucher",
            "code": "voucher",
            "description": "Voucher game dan digital content",
            "icon": "https://nos.jkt-1.neo.id/gate/icons/ticket.svg",
            "order": 2
        },
        {
            "title": "E-Money",
            "code": "e-money",
            "description": "Top up saldo e-wallet dan e-money",
            "icon": "https://nos.jkt-1.neo.id/gate/icons/wallet.svg",
            "order": 3
        },
        {
            "title": "Streaming",
            "code": "streaming",
            "description": "Langganan Netflix, Spotify, Disney+ dan lainnya",
            "icon": "https://nos.jkt-1.neo.id/gate/icons/play.svg",
            "order": 4
        }
    ]
}
```

---

### 7. Get Products

Retrieve products, optionally filtered by category.

**Endpoint:** `GET /v2/products`

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| region | string | Yes | Region code (ID, MY, PH, SG, TH) |
| categoryCode | string | No | Filter by category code |
| productCode | string | No | Get specific product by code |

**Response (List):**

```json
{
    "data": [
        {
            "code": "MLBB",
            "slug": "mobile-legends",
            "title": "Mobile Legends: Bang Bang",
            "subtitle": "Moonton",
            "description": "Top up diamond Mobile Legends dengan proses cepat dan aman",
            "publisher": "Moonton",
            "thumbnail": "https://nos.jkt-1.neo.id/gate/products/mlbb-icon.webp",
            "banner": "https://nos.jkt-1.neo.id/gate/products/mlbb-banner.webp",
            "isPopular": true,
            "isAvailable": true,
            "tags": ["MOBA", "Multiplayer", "Strategy"],
            "category": {
                "title": "Top Up Game",
                "code": "top-up-game"
            }
        },
        {
            "code": "FF",
            "slug": "free-fire",
            "title": "Free Fire",
            "subtitle": "Garena",
            "description": "Top up diamond Free Fire instant tanpa ribet",
            "publisher": "Garena International",
            "thumbnail": "https://nos.jkt-1.neo.id/gate/products/ff-icon.webp",
            "banner": "https://nos.jkt-1.neo.id/gate/products/ff-banner.webp",
            "isPopular": true,
            "isAvailable": true,
            "tags": ["Battle Royale", "Shooter", "Multiplayer"],
            "category": {
                "title": "Top Up Game",
                "code": "top-up-game"
            }
        }
    ]
}
```

**Response (Single Product with productCode):**

```json
{
    "data": {
        "code": "MLBB",
        "slug": "mobile-legends",
        "title": "Mobile Legends: Bang Bang",
        "subtitle": "Moonton",
        "description": "Top up diamond Mobile Legends dengan proses cepat dan aman. Dapatkan diamond untuk membeli hero, skin, dan item premium lainnya.",
        "publisher": "Moonton",
        "thumbnail": "https://nos.jkt-1.neo.id/gate/products/mlbb-icon.webp",
        "banner": "https://nos.jkt-1.neo.id/gate/products/mlbb-banner.webp",
        "isPopular": true,
        "isAvailable": true,
        "tags": ["MOBA", "Multiplayer", "Strategy"],
        "category": {
            "title": "Top Up Game",
            "code": "top-up-game"
        },
        "features": [
            "⚡ Proses Instan",
            "🔒 Aman & Terpercaya",
            "💰 Harga Termurah",
            "🎁 Bonus Diamond"
        ],
        "howToOrder": [
            "Masukkan User ID dan Zone ID",
            "Pilih nominal diamond yang diinginkan",
            "Pilih metode pembayaran",
            "Selesaikan pembayaran",
            "Diamond akan masuk otomatis"
        ]
    }
}
```

| Field | Description |
|-------|-------------|
| `code` | Unique product identifier (used in API calls) |
| `slug` | URL-friendly identifier (used in web URLs) |
| `title` | Product display name |
| `subtitle` | Product subtitle (usually publisher) |
| `description` | Product description |
| `publisher` | Official publisher name |
| `thumbnail` | Product icon/logo (200x200px) |
| `banner` | Product banner image (1200x400px) |
| `isPopular` | Flag for "Popular" section |
| `isAvailable` | Product availability status |
| `tags` | Product tags for filtering |
| `category` | Product category information |
| `features` | Product features (single product only) |
| `howToOrder` | Step-by-step guide (single product only) |

---

### 8. Get Popular Products

Retrieve products where `isPopular = true`.

**Endpoint:** `GET /v2/populars`

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| region | string | Yes | Region code (ID, MY, PH, SG, TH) |

**Response:** Same structure as Get Products (list format).

---

### 9. Get Product Fields

Retrieve input fields required for a specific product.

**Endpoint:** `GET /v2/fields`

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| region | string | Yes | Region code (ID, MY, PH, SG, TH) |
| productCode | string | Yes | Product code (e.g., DANA, MLBB) |

**Response (Mobile Legends):**

```json
{
    "data": [
        {
            "name": "User ID",
            "key": "userId",
            "type": "number",
            "label": "Masukkan User ID",
            "required": true,
            "minLength": 1,
            "maxLength": 12,
            "placeholder": "123456789",
            "hint": "Cek di profil game, menu bagian kanan atas"
        },
        {
            "name": "Zone ID",
            "key": "zoneId",
            "type": "number",
            "label": "Masukkan Zone ID",
            "required": true,
            "minLength": 1,
            "maxLength": 8,
            "placeholder": "1234",
            "hint": "Zone ID berada di samping User ID"
        }
    ]
}
```

**Response (DANA E-Wallet):**

```json
{
    "data": [
        {
            "name": "Nomor Telepon",
            "key": "phoneNumber",
            "type": "number",
            "label": "Masukkan Nomor Dana",
            "required": true,
            "minLength": 10,
            "maxLength": 13,
            "placeholder": "08xxxxxxxxxx",
            "pattern": "^08[0-9]{8,11}$",
            "hint": "Nomor telepon yang terdaftar di DANA"
        }
    ]
}
```

| Field | Description |
|-------|-------------|
| `name` | Field display name (shown above input) |
| `key` | Field identifier (used as request body key) |
| `type` | Input type: `number`, `text`, `email`, `select` |
| `label` | Field label (placeholder-style) |
| `required` | Whether field must be filled |
| `minLength` | Minimum character length |
| `maxLength` | Maximum character length |
| `placeholder` | Example input text |
| `pattern` | Regex pattern for validation |
| `hint` | Help text shown below input |

---

### 10. Get Sections

Retrieve product sections for organizing SKUs.

**Endpoint:** `GET /v2/sections`

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| region | string | Yes | Region code (ID, MY, PH, SG, TH) |
| productCode | string | Yes | Product code (e.g., DANA, MLBB) |

**Response:**

```json
{
    "data": [
        {
            "title": "Spesial Item",
            "code": "special-item",
            "icon": "⭐",
            "order": 1
        },
        {
            "title": "Topup Instan",
            "code": "topup-instant",
            "icon": "⚡",
            "order": 2
        },
        {
            "title": "Semua Item",
            "code": "all-items",
            "icon": "",
            "order": 3
        }
    ]
}
```

---

### 11. Get SKUs

Retrieve available product variants/denominations with pricing.

**Endpoint:** `GET /v2/skus`

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| region | string | Yes | Region code (ID, MY, PH, SG, TH) |
| productCode | string | Yes | Product code (e.g., DANA, MLBB) |
| sectionCode | string | No | Filter by section code |

**Response:**

```json
{
    "data": [
        {
            "code": "MLBB_86",
            "name": "86 Diamonds",
            "description": "86 (78+8) Diamonds",
            "currency": "IDR",
            "price": 24750,
            "originalPrice": 25000,
            "discount": 1.0,
            "image": "https://nos.jkt-1.neo.id/gate/products/mlbb-diamond.webp",
            "info": "Bonus +8 Diamonds",
            "processTime": 0,
            "isAvailable": true,
            "isFeatured": false,
            "section": {
                "title": "Topup Instan",
                "code": "topup-instant"
            }
        },
        {
            "code": "MLBB_172",
            "name": "172 Diamonds",
            "description": "172 (156+16) Diamonds",
            "currency": "IDR",
            "price": 49450,
            "originalPrice": 50000,
            "discount": 1.1,
            "image": "https://nos.jkt-1.neo.id/gate/products/mlbb-diamond.webp",
            "info": "Bonus +16 Diamonds",
            "processTime": 0,
            "isAvailable": true,
            "isFeatured": true,
            "section": {
                "title": "Spesial Item",
                "code": "special-item"
            },
            "badge": {
                "text": "Diskon 20%",
                "color": "#FF6B6B"
            }
        }
    ]
}
```

| Field | Description |
|-------|-------------|
| `code` | Unique SKU identifier |
| `name` | Display name of the variant |
| `description` | Detailed description |
| `currency` | ISO 4217 currency code |
| `price` | Final price after discount |
| `originalPrice` | Original price before discount |
| `discount` | Discount percentage |
| `image` | SKU image URL |
| `info` | Additional information (e.g., bonus) |
| `processTime` | Processing time in minutes (0 = instant) |
| `isAvailable` | Stock availability status |
| `isFeatured` | Whether to highlight this SKU |
| `section` | Section/category information |
| `badge` | Optional badge for promotions |

---

### 12. Get SKU Promos

Retrieve all SKUs that are currently on promotion (across all products).

**Endpoint:** `GET /v2/sku/promos`

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| region | string | Yes | Region code (ID, MY, PH, SG, TH) |

**Response:**

```json
{
    "data": [
        {
            "code": "MLBB_86",
            "name": "86 Diamonds",
            "description": "86 (78+8) Diamonds",
            "currency": "IDR",
            "price": 24750,
            "originalPrice": 25000,
            "discount": 1.0,
            "image": "https://nos.jkt-1.neo.id/gate/products/mlbb-diamond.webp",
            "info": "Bonus +8 Diamonds",
            "processTime": 0,
            "isAvailable": true,
            "isFeatured": false,
            "section": {
                "title": "Topup Instan",
                "code": "topup-instant"
            }
        },
        {
            "code": "PUBGM_1800",
            "name": "1800 UC",
            "description": "1800 Unknown Cash",
            "currency": "IDR",
            "price": 400000,
            "originalPrice": 425000,
            "discount": 5.88,
            "image": "https://nos.jkt-1.neo.id/gate/products/pubgm-uc.webp",
            "info": "",
            "processTime": 0,
            "isAvailable": true,
            "isFeatured": false,
            "section": {
                "title": "Topup Instan",
                "code": "topup-instant"
            }
        }
    ]
}
```

> **Note:** This endpoint returns SKUs from all products that have active promotions (discount > 0). Used for displaying promo items on the homepage.

---

## Transaction Endpoints

### 13. Validate Account

Validate user account before transaction.

**Endpoint:** `POST /v2/account/inquiries`

**Headers:**

```
Content-Type: application/json
Authorization: Bearer {token} (Optional)
```

**Request Body (Mobile Legends):**

```json
{
    "productCode": "MLBB",
    "userId": "656696292",
    "zoneId": "8610"
}
```

**Request Body (DANA):**

```json
{
    "productCode": "DANA",
    "phoneNumber": "081234567890"
}
```

**Response (Success):**

```json
{
    "data": {
        "product": {
            "name": "Mobile Legends",
            "code": "MLBB"
        },
        "account": {
            "region": "ID",
            "nickname": "り い こ ✧"
        }
    }
}
```

**Response (Error - Account Not Found):**

```json
{
    "error": {
        "code": "ACCOUNT_NOT_FOUND",
        "message": "Account not found",
        "details": "The provided User ID and Zone ID combination does not exist"
    }
}
```

**Response (Error - Inconsistent Provider):**

```json
{
    "error": {
        "code": "INCONSISTENT_PROVIDER",
        "message": "Sepertinya nomor telepon tersebut adalah nomor By.U, mohon masukkan nomor telepon yang valid atau lanjutkan bertransaksi pada halaman By.U",
        "details": "Phone number prefix does not match selected provider",
        "suggestion": {
            "productCode": "BYU",
            "productName": "By.U",
            "productSlug": "byu"
        }
    }
}
```

---

### 14. Get Payment Channel Categories

Retrieve payment channel categories.

**Endpoint:** `GET /v2/payment-channel/categories`

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| region | string | Yes | Region code (ID, MY, PH, SG, TH) |
| paymentType | string | Yes | Payment type: `purchase`, `deposit` |

**Response:**

```json
{
    "data": [
        {
            "title": "E-Wallet",
            "code": "E_WALLET",
            "icon": "https://nos.jkt-1.neo.id/gate/icons/wallet.svg",
            "order": 1
        },
        {
            "title": "Virtual Account",
            "code": "VIRTUAL_ACCOUNT",
            "icon": "https://nos.jkt-1.neo.id/gate/icons/bank.svg",
            "order": 2
        },
        {
            "title": "Convenience Store",
            "code": "RETAIL",
            "icon": "https://nos.jkt-1.neo.id/gate/icons/store.svg",
            "order": 3
        },
        {
            "title": "Credit or Debit Card",
            "code": "CARD",
            "icon": "https://nos.jkt-1.neo.id/gate/icons/card.svg",
            "order": 4
        }
    ]
}
```

---

### 15. Get Payment Channels

Retrieve all available payment channels.

**Endpoint:** `GET /v2/payment-channels`

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| region | string | Yes | Region code (ID, MY, PH, SG, TH) |
| categoryCode | string | No | Filter by category code |
| paymentType | string | Yes | Payment type: `purchase`, `deposit` |

**Response:**

```json
{
    "data": [
        {
            "code": "QRIS",
            "name": "QRIS",
            "description": "Bayar menggunakan QRIS dari semua aplikasi e-wallet dan mobile banking",
            "image": "https://nos.jkt-1.neo.id/gate/payment/qris.webp",
            "currency": "IDR",
            "feeAmount": 0,
            "feePercentage": 0.7,
            "minAmount": 1000,
            "maxAmount": 10000000,
            "featured": true,
            "instruction": "<p>Gunakan <strong>E-wallet</strong> atau <strong>aplikasi mobile banking</strong> yang tersedia untuk scan QRIS</p>",
            "category": {
                "title": "E-Wallet",
                "code": "E_WALLET"
            }
        },
        {
            "code": "DANA",
            "name": "DANA",
            "description": "Bayar menggunakan DANA",
            "image": "https://nos.jkt-1.neo.id/gate/payment/dana.webp",
            "currency": "IDR",
            "feeAmount": 0,
            "feePercentage": 1.5,
            "minAmount": 1000,
            "maxAmount": 5000000,
            "featured": true,
            "instruction": "<ol><li>Setelah klik bayar, kamu akan diarahkan ke aplikasi DANA</li><li>Pastikan saldo DANA kamu mencukupi</li><li>Konfirmasi pembayaran dengan PIN DANA</li><li>Transaksi selesai</li></ol>",
            "category": {
                "title": "E-Wallet",
                "code": "E_WALLET"
            }
        },
        {
            "code": "BCA_VA",
            "name": "BCA Virtual Account",
            "description": "Bayar menggunakan BCA Virtual Account",
            "image": "https://nos.jkt-1.neo.id/gate/payment/bca.webp",
            "currency": "IDR",
            "feeAmount": 4000,
            "feePercentage": 0,
            "minAmount": 10000,
            "maxAmount": 50000000,
            "featured": false,
            "instruction": "<ol><li>Pilih <strong>m-Transfer</strong> &gt; <strong>BCA Virtual Account</strong></li><li>Masukkan nomor Virtual Account</li><li>Periksa informasi yang tertera di layar</li><li>Masukkan <strong>PIN m-BCA</strong></li><li>Transaksi selesai</li></ol>",
            "category": {
                "title": "Virtual Account",
                "code": "VIRTUAL_ACCOUNT"
            }
        },
        {
            "code": "BALANCE",
            "name": "Saldo Gate",
            "description": "Bayar menggunakan saldo akun Gate",
            "image": "https://nos.jkt-1.neo.id/gate/payment/balance.webp",
            "currency": "IDR",
            "feeAmount": 0,
            "feePercentage": 0,
            "minAmount": 100,
            "maxAmount": 100000000,
            "featured": true,
            "instruction": "<p>Pastikan saldo akun Gate kamu mencukupi untuk melakukan transaksi</p>",
            "category": {
                "title": "E-Wallet",
                "code": "E_WALLET"
            }
        }
    ]
}
```

| Field | Description |
|-------|-------------|
| `code` | Unique payment channel code |
| `name` | Display name |
| `description` | Payment description |
| `image` | Payment channel logo |
| `currency` | Supported currency |
| `feeAmount` | Fixed fee amount |
| `feePercentage` | Percentage fee |
| `minAmount` | Minimum transaction amount |
| `maxAmount` | Maximum transaction amount |
| `featured` | Show at top of list |
| `instruction` | Payment instructions (HTML) |
| `category` | Payment category |

> **Fee Calculation:** `totalFee = feeAmount + (price * feePercentage / 100)`

---

### 16. Get Promo Codes

Retrieve available promo codes.

**Endpoint:** `GET /v2/promos`

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| region | string | Yes | Region code |
| productCode | string | Yes | Product code |

**Response:**

```json
{
    "data": [
        {
            "code": "WELCOME10",
            "title": "Diskon 10% Pengguna Baru",
            "description": "Khusus untuk pengguna baru, dapatkan diskon 10%",
            "products": [],
            "paymentChannels": [],
            "daysAvailable": [],
            "maxDailyUsage": 100,
            "maxUsage": 10000,
            "maxUsagePerId": 1,
            "maxUsagePerDevice": 1,
            "maxUsagePerIp": 1,
            "expiredAt": "2025-12-31T23:59:59+07:00",
            "minAmount": 50000,
            "maxPromoAmount": 10000,
            "promoFlat": 0,
            "promoPercentage": 10,
            "isAvailable": true,
            "note": "Berlaku untuk semua produk",
            "totalUsage": 5421,
            "totalDailyUsage": 87
        },
        {
            "code": "MLBB50",
            "title": "Cashback 50% Mobile Legends",
            "description": "Bayar pakai DANA dapat cashback 50%",
            "products": [
                { "code": "MLBB", "name": "Mobile Legends" }
            ],
            "paymentChannels": [
                { "code": "DANA", "name": "DANA" }
            ],
            "daysAvailable": ["MON", "WED", "FRI"],
            "maxDailyUsage": 200,
            "maxUsage": 5000,
            "maxUsagePerId": 3,
            "maxUsagePerDevice": 3,
            "maxUsagePerIp": 3,
            "expiredAt": "2025-12-31T23:59:59+07:00",
            "minAmount": 100000,
            "maxPromoAmount": 50000,
            "promoFlat": 0,
            "promoPercentage": 50,
            "isAvailable": true,
            "note": "Khusus hari Senin, Rabu, Jumat",
            "totalUsage": 3245,
            "totalDailyUsage": 54
        }
    ]
}
```

---

### 17. Validate Promo Code

Validate promo code before order.

**Endpoint:** `POST /v2/promos/validate`

**Request Body:**

```json
{
    "promoCode": "WELCOME10",
    "productCode": "MLBB",
    "skuCode": "MLBB_172",
    "paymentCode": "QRIS",
    "region": "ID",
    "userId": "656696292",
    "zoneId": "8610"
}
```

**Response (Valid):**

```json
{
    "data": {
        "promoCode": "WELCOME10",
        "discountAmount": 4945,
        "originalAmount": 49450,
        "finalAmount": 44505,
        "promoDetails": {
            "title": "Diskon 10% Pengguna Baru",
            "promoPercentage": 10,
            "maxPromoAmount": 10000
        }
    }
}
```

**Response (Invalid):**

```json
{
    "error": {
        "code": "PROMO_EXPIRED",
        "message": "Kode promo telah kadaluarsa",
        "details": "Promo ini berakhir pada 30 November 2025"
    }
}
```

---

### 18. Order Inquiry

Pre-validate order before creation.

**Endpoint:** `POST /v2/orders/inquiries`

**Headers:**

```
Content-Type: application/json
Authorization: Bearer {token} (Optional)
```

| Auth Status | Behavior |
|-------------|----------|
| ✅ With Token | Transaction will be linked to user account |
| ❌ Without Token | Guest checkout, transaction not linked to account |

**Request Body:**

```json
{
    "productCode": "MLBB",
    "skuCode": "MLBB_172",
    "userId": "656696292",
    "zoneId": "8610",
    "quantity": 1,
    "paymentCode": "QRIS",
    "promoCode": "WELCOME10",
    "email": "user@example.com",
    "phoneNumber": "+6281234567890"
}
```

**Response:**

```json
{
    "data": {
        "validationToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "expiresAt": "2025-12-03T15:30:00+07:00",
        "order": {
            "product": {
                "code": "MLBB",
                "name": "Mobile Legends: Bang Bang"
            },
            "sku": {
                "code": "MLBB_172",
                "name": "172 Diamonds",
                "quantity": 1
            },
            "account": {
                "nickname": "り い こ ✧",
                "userId": "656696292",
                "zoneId": "8610"
            },
            "payment": {
                "code": "QRIS",
                "name": "QRIS",
                "currency": "IDR"
            },
            "pricing": {
                "subtotal": 49450,
                "discount": 4945,
                "paymentFee": 346,
                "total": 44851
            },
            "promo": {
                "code": "WELCOME10",
                "discountAmount": 4945
            },
            "contact": {
                "email": "user@example.com",
                "phoneNumber": "+6281234567890"
            }
        }
    }
}
```

---

### 19. Create Order

Create order with validation token.

**Endpoint:** `POST /v2/orders`

**Headers:**

```
Content-Type: application/json
Authorization: Bearer {token} (Optional)
```

| Auth Status | Behavior |
|-------------|----------|
| ✅ With Token | Transaction linked to user account, visible in dashboard |
| ❌ Without Token | Transaction processed but not linked to any account |

**Request Body:**

```json
{
    "validationToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response:**

```json
{
    "data": {
        "step": "SUCCESS",
        "order": {
            "invoiceNumber": "GATE1A11BB97DF88D56530993",
            "status": "PENDING",
            "productCode": "MLBB",
            "productName": "Mobile Legends: Bang Bang",
            "skuCode": "MLBB_172",
            "skuName": "172 Diamonds",
            "quantity": 1,
            "account": {
                "nickname": "り い こ ✧",
                "inputs": "656696292 - 8610"
            },
            "pricing": {
                "subtotal": 49450,
                "discount": 4945,
                "paymentFee": 346,
                "total": 44851,
                "currency": "IDR"
            },
            "payment": {
                "code": "QRIS",
                "name": "QRIS",
                "instruction": "<p>Gunakan E-wallet atau aplikasi mobile banking untuk scan QRIS</p>",
                "qrCode": "00020101021226660016ID.CO.QRIS.WWW...",
                "qrCodeImage": "https://api.gate.co.id/v2/qr/generate?data=...",
                "expiredAt": "2025-12-03T16:25:00+07:00"
            },
            "contact": {
                "email": "user@example.com",
                "phoneNumber": "+6281234567890"
            },
            "createdAt": "2025-12-03T15:25:00+07:00",
            "expiredAt": "2025-12-03T16:25:00+07:00"
        }
    }
}
```

---

### 20. Get Invoice Details

Get order details by invoice number.

**Endpoint:** `GET /v2/invoices`

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| invoiceNumber | string | Yes | Invoice number |

**Response:**

```json
{
    "data": {
        "invoiceNumber": "GATE1A11BB97DF88D56530993",
        "status": "SUCCESS",
        "productCode": "MLBB",
        "productName": "Mobile Legends: Bang Bang",
        "skuCode": "MLBB_172",
        "skuName": "172 Diamonds",
        "quantity": 1,
        "account": {
            "nickname": "り い こ ✧",
            "inputs": "656696292 - 8610"
        },
        "pricing": {
            "subtotal": 49450,
            "discount": 4945,
            "paymentFee": 346,
            "total": 44851,
            "currency": "IDR"
        },
        "payment": {
            "code": "QRIS",
            "name": "QRIS",
            "paidAt": "2025-12-03T15:30:45+07:00"
        },
        "promo": {
            "code": "WELCOME10",
            "discountAmount": 4945
        },
        "contact": {
            "email": "user@example.com",
            "phoneNumber": "081234567890"
        },
        "timeline": [
            {
                "status": "PENDING",
                "message": "Order created, waiting for payment",
                "timestamp": "2025-12-03T15:25:00+07:00"
            },
            {
                "status": "PAID",
                "message": "Payment received successfully",
                "timestamp": "2025-12-03T15:30:45+07:00"
            },
            {
                "status": "PROCESSING",
                "message": "Processing your order",
                "timestamp": "2025-12-03T15:30:50+07:00"
            },
            {
                "status": "SUCCESS",
                "message": "Diamonds has been added to your account",
                "timestamp": "2025-12-03T15:31:15+07:00"
            }
        ],
        "createdAt": "2025-12-03T15:25:00+07:00",
        "expiredAt": "2025-12-03T16:25:00+07:00",
        "completedAt": "2025-12-03T15:31:15+07:00"
    }
}
```

---

## Auth Endpoints

### 21. Register

Register new user account.

**Endpoint:** `POST /v2/auth/register`

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| region | string | No | Region code (ID, MY, PH, SG, TH). Default: ID |

**Request Body:**

```json
{
    "firstName": "John",
    "lastName": "Doe",
    "email": "john.doe@example.com",
    "phoneNumber": "+628123456789",
    "password": "SecureP@ssw0rd",
    "confirmPassword": "SecureP@ssw0rd"
}
```

**Response:**

```json
{
    "data": {
        "step": "EMAIL_VERIFICATION",
        "user": {
            "id": "usr_1a2b3c4d5e6f",
            "firstName": "John",
            "lastName": "Doe",
            "email": "john.doe@example.com",
            "phoneNumber": "+628123456789",
            "profilePicture": null,
            "status": "INACTIVE",
            "primaryRegion": "ID",
            "currentRegion": "ID",
            "currency": "IDR",
            "balance": {
                "IDR": 0,
                "MYR": 0,
                "PHP": 0,
                "SGD": 0,
                "THB": 0
            },
            "membership": {
                "level": "CLASSIC",
                "name": "Classic"
            },
            "mfaStatus": "INACTIVE",
            "createdAt": "2025-12-03T10:00:00+07:00"
        }
    }
}
```

| Field | Description |
|-------|-------------|
| `primaryRegion` | Set at registration, **permanent** |
| `currentRegion` | Initially same as primaryRegion, can change |
| `currency` | Based on currentRegion |
| `balance` | Multi-currency balance (all currencies) |

---

### 22. Register with Google

Register using Google OAuth.

**Endpoint:** `POST /v2/auth/register/google`

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| region | string | No | Region code. Default: ID |

**Request Body:**

```json
{
    "idToken": "eyJhbGciOiJSUzI1NiIsImtpZCI6IjZmODI4..."
}
```

**Response:**

```json
{
    "data": {
        "step": "SUCCESS",
        "token": {
            "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
            "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
            "expiresIn": 3600,
            "tokenType": "Bearer"
        },
        "user": {
            "id": "usr_1a2b3c4d5e6f",
            "firstName": "John",
            "lastName": "Doe",
            "email": "john.doe@gmail.com",
            "phoneNumber": null,
            "profilePicture": "https://lh3.googleusercontent.com/...",
            "status": "ACTIVE",
            "primaryRegion": "ID",
            "currentRegion": "ID",
            "currency": "IDR",
            "balance": {
                "IDR": 0,
                "MYR": 0,
                "PHP": 0,
                "SGD": 0,
                "THB": 0
            },
            "membership": {
                "level": "CLASSIC",
                "name": "Classic"
            },
            "mfaStatus": "INACTIVE",
            "googleId": "117562748392847562",
            "createdAt": "2025-12-03T10:00:00+07:00"
        }
    }
}
```

---

### 23. Verify Email

Verify email address after registration.

**Endpoint:** `POST /v2/auth/verify-email`

**Request Body:**

```json
{
    "email": "john.doe@example.com",
    "token": "abc123xyz"
}
```

**Response:**

```json
{
    "data": {
        "step": "SUCCESS",
        "token": {
            "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
            "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
            "expiresIn": 3600,
            "tokenType": "Bearer"
        },
        "user": {
            "id": "usr_1a2b3c4d5e6f",
            "status": "ACTIVE",
            "emailVerifiedAt": "2025-12-03T10:15:00+07:00"
        }
    }
}
```

---

### 24. Resend Verification Email

Resend email verification link.

**Endpoint:** `POST /v2/auth/resend-verification`

**Request Body:**

```json
{
    "email": "john.doe@example.com"
}
```

**Response:**

```json
{
    "data": {
        "message": "Verification email has been sent"
    }
}
```

---

### 25. Login

Login with email and password.

**Endpoint:** `POST /v2/auth/login`

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| region | string | No | Region to set as current. Default: user's primaryRegion |

**Request Body:**

```json
{
    "email": "john.doe@example.com",
    "password": "SecureP@ssw0rd"
}
```

**Response (Success - No MFA):**

```json
{
    "data": {
        "step": "SUCCESS",
        "token": {
            "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
            "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
            "expiresIn": 3600,
            "tokenType": "Bearer"
        },
        "user": {
            "id": "usr_1a2b3c4d5e6f",
            "firstName": "John",
            "lastName": "Doe",
            "email": "john.doe@example.com",
            "phoneNumber": "+628123456789",
            "profilePicture": "https://nos.jkt-1.neo.id/gate/profiles/user123.jpg",
            "status": "ACTIVE",
            "primaryRegion": "ID",
            "currentRegion": "ID",
            "currency": "IDR",
            "balance": {
                "IDR": 150000,
                "MYR": 500,
                "PHP": 0,
                "SGD": 0,
                "THB": 0
            },
            "membership": {
                "level": "PRESTIGE",
                "name": "Prestige"
            },
            "mfaStatus": "INACTIVE",
            "createdAt": "2025-11-01T10:00:00+07:00",
            "lastLoginAt": "2025-12-03T10:30:00+07:00"
        }
    }
}
```

**Response (MFA Required):**

```json
{
    "data": {
        "step": "MFA_VERIFICATION",
        "mfaToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "expiresAt": "2025-12-03T10:35:00+07:00"
    }
}
```

---

### 26. Login with Google

Login using Google OAuth.

**Endpoint:** `POST /v2/auth/login/google`

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| region | string | No | Region to set as current |

**Request Body:**

```json
{
    "idToken": "eyJhbGciOiJSUzI1NiIsImtpZCI6IjZmODI4..."
}
```

**Response:** Same as regular login success response.

---

### 27. Verify MFA

Verify MFA code during login.

**Endpoint:** `POST /v2/auth/verify-mfa`

**Request Body:**

```json
{
    "mfaToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "code": "123456"
}
```

**Response:**

```json
{
    "data": {
        "step": "SUCCESS",
        "token": {
            "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
            "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
            "expiresIn": 3600,
            "tokenType": "Bearer"
        },
        "user": { ... }
    }
}
```

---

### 28. Forgot Password

Request password reset link.

**Endpoint:** `POST /v2/auth/forgot-password`

**Request Body:**

```json
{
    "email": "john.doe@example.com"
}
```

**Response:**

```json
{
    "data": {
        "message": "Password reset link has been sent to your email"
    }
}
```

---

### 29. Reset Password

Reset password using token from email.

**Endpoint:** `POST /v2/auth/reset-password`

**Request Body:**

```json
{
    "token": "reset_token_from_email",
    "newPassword": "NewSecureP@ssw0rd",
    "confirmPassword": "NewSecureP@ssw0rd"
}
```

**Response:**

```json
{
    "data": {
        "message": "Password has been reset successfully"
    }
}
```

---

### 30. Enable MFA

Enable two-factor authentication.

**Endpoint:** `POST /v2/auth/mfa/enable`

**Headers:**

```
Authorization: Bearer {access_token}
```

**Response:**

```json
{
    "data": {
        "step": "SETUP",
        "qrCode": "otpauth://totp/Gate:john.doe@example.com?secret=JBSWY3DPEHPK3PXP&issuer=Gate",
        "qrCodeImage": "https://api.gate.co.id/v2/qr/generate?data=...",
        "secret": "JBSWY3DPEHPK3PXP",
        "backupCodes": [
            "ABC12345",
            "DEF67890",
            "GHI11111",
            "JKL22222",
            "MNO33333"
        ]
    }
}
```

---

### 31. Verify MFA Setup

Verify MFA setup with code from authenticator app.

**Endpoint:** `POST /v2/auth/mfa/verify-setup`

**Headers:**

```
Authorization: Bearer {access_token}
```

**Request Body:**

```json
{
    "code": "123456"
}
```

**Response:**

```json
{
    "data": {
        "step": "SUCCESS",
        "message": "MFA has been enabled successfully",
        "mfaStatus": "ACTIVE"
    }
}
```

---

### 32. Disable MFA

Disable two-factor authentication.

**Endpoint:** `POST /v2/auth/mfa/disable`

**Headers:**

```
Authorization: Bearer {access_token}
```

**Request Body:**

```json
{
    "code": "123456",
    "password": "CurrentP@ssw0rd"
}
```

**Response:**

```json
{
    "data": {
        "message": "MFA has been disabled",
        "mfaStatus": "INACTIVE"
    }
}
```

---

### 33. Refresh Token

Get new access token using refresh token.

**Endpoint:** `POST /v2/auth/refresh-token`

**Request Body:**

```json
{
    "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response:**

```json
{
    "data": {
        "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "expiresIn": 3600,
        "tokenType": "Bearer"
    }
}
```

---

### 34. Logout

Invalidate current session.

**Endpoint:** `POST /v2/auth/logout`

**Headers:**

```
Authorization: Bearer {access_token}
```

**Response:**

```json
{
    "data": {
        "message": "Logged out successfully"
    }
}
```

---

### 35. Get User Profile

Get current user profile. Can also switch region.

**Endpoint:** `GET /v2/user/profile`

**Headers:**

```
Authorization: Bearer {access_token}
```

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| region | string | No | Region to switch to |

**Response:**

```json
{
    "data": {
        "id": "usr_1a2b3c4d5e6f",
        "firstName": "John",
        "lastName": "Doe",
        "email": "john.doe@example.com",
        "phoneNumber": "+628123456789",
        "profilePicture": "https://nos.jkt-1.neo.id/gate/profiles/user123.jpg",
        "status": "ACTIVE",
        "primaryRegion": "ID",
        "currentRegion": "ID",
        "currency": "IDR",
        "balance": {
            "IDR": 150000,
            "MYR": 500,
            "PHP": 0,
            "SGD": 100,
            "THB": 0
        },
        "membership": {
            "level": "PRESTIGE",
            "name": "Prestige",
            "benefits": [
                "Diskon eksklusif hingga 5%",
                "Priority customer support",
                "Bonus poin 3%",
                "Akses promo premium"
            ],
            "progress": {
                "current": 5420000,
                "target": 10000000,
                "percentage": 54.2,
                "nextLevel": "ROYAL",
                "currency": "IDR"
            }
        },
        "mfaStatus": "ACTIVE",
        "emailVerifiedAt": "2025-11-01T10:15:00+07:00",
        "createdAt": "2025-11-01T10:00:00+07:00",
        "lastLoginAt": "2025-12-03T10:30:00+07:00",
        "updatedAt": "2025-12-03T10:30:00+07:00"
    }
}
```

> **Note:** Calling with `?region=XX` updates user's `currentRegion` and `currency`.

---

### 36. Update User Profile

Update user profile information.

**Endpoint:** `PUT /v2/user/profile`

**Headers:**

```
Authorization: Bearer {access_token}
Content-Type: multipart/form-data
```

**Request Body (Form Data):**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| firstName | string | No | First name |
| lastName | string | No | Last name |
| phoneNumber | string | No | Phone number (must be unique) |
| profilePicture | file | No | Profile image (max 5MB, JPG/PNG/WebP) |

**Response:**

```json
{
    "data": {
        "id": "usr_1a2b3c4d5e6f",
        "firstName": "John",
        "lastName": "Doe",
        "email": "john.doe@example.com",
        "phoneNumber": "+628123456789",
        "profilePicture": "https://nos.jkt-1.neo.id/gate/profiles/user123-updated.jpg",
        "status": "ACTIVE",
        "updatedAt": "2025-12-03T10:45:00+07:00"
    }
}
```

> **Note:** Email cannot be changed for security reasons.

---

### 37. Change Password

Change user password.

**Endpoint:** `POST /v2/user/change-password`

**Headers:**

```
Authorization: Bearer {access_token}
```

**Request Body:**

```json
{
    "currentPassword": "CurrentP@ssw0rd",
    "newPassword": "NewSecureP@ssw0rd",
    "confirmPassword": "NewSecureP@ssw0rd"
}
```

**Response:**

```json
{
    "data": {
        "message": "Password changed successfully"
    }
}
```

**Error Response:**

```json
{
    "error": {
        "code": "INVALID_PASSWORD",
        "message": "Password saat ini salah",
        "details": "Please enter your correct current password"
    }
}
```

> **Note:** All refresh tokens are invalidated after password change.

---

## User Dashboard Endpoints

All dashboard endpoints require authentication.

### 38. Get Transactions

Retrieve transaction history with filtering and pagination.

**Endpoint:** `GET /v2/transactions`

**Headers:**

```
Authorization: Bearer {access_token}
```

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| region | string | No | Region code. Default: user's currentRegion |
| limit | integer | No | Items per page. Default: 10, Max: 100 |
| page | integer | No | Page number. Default: 1 |
| search | string | No | Search by invoice number |
| status | string | No | Filter: ALL, SUCCESS, PROCESSING, PENDING, FAILED |
| paymentStatus | string | No | Filter: PAID, UNPAID, EXPIRED |
| startDate | string | No | Start date (YYYY-MM-DD) |
| endDate | string | No | End date (YYYY-MM-DD) |

**Response:**

```json
{
    "data": {
        "overview": {
            "totalTransaction": 125,
            "totalPurchase": 1250000,
            "success": 100,
            "processing": 10,
            "pending": 15,
            "failed": 0
        },
        "transactions": [
            {
                "invoiceNumber": "GATE1A11BB97DF88D56530993",
                "status": "SUCCESS",
                "productCode": "MLBB",
                "productName": "Mobile Legends: Bang Bang",
                "skuCode": "MLBB_172",
                "skuName": "172 (156+16) Diamonds",
                "account": {
                    "nickname": "り い こ ✧",
                    "inputs": "656696292 - 8610"
                },
                "pricing": {
                    "subtotal": 49450,
                    "discount": 4945,
                    "paymentFee": 346,
                    "total": 44851,
                    "currency": "IDR"
                },
                "payment": {
                    "code": "QRIS",
                    "name": "QRIS"
                },
                "createdAt": "2025-09-25T10:30:00+07:00"
            }
        ],
        "pagination": {
            "limit": 10,
            "page": 1,
            "totalRows": 125,
            "totalPages": 13
        }
    }
}
```

**Transaction Status:**

| Status | Description | Color |
|--------|-------------|-------|
| SUCCESS | Completed successfully | Green |
| PROCESSING | Being processed | Yellow |
| PENDING | Waiting for payment | Orange |
| FAILED | Transaction failed | Red |
| EXPIRED | Payment deadline exceeded | Gray |

---

### 39. Get Mutations

Retrieve balance mutation history (debits and credits).

**Endpoint:** `GET /v2/mutations`

**Headers:**

```
Authorization: Bearer {access_token}
```

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| region | string | No | Region code. Default: user's currentRegion |
| limit | integer | No | Items per page. Default: 10, Max: 100 |
| page | integer | No | Page number. Default: 1 |
| search | string | No | Search by invoice number |
| type | string | No | Filter: ALL, DEBIT, CREDIT |
| startDate | string | No | Start date (YYYY-MM-DD) |
| endDate | string | No | End date (YYYY-MM-DD) |

**Response:**

```json
{
    "data": {
        "overview": {
            "totalDebit": 500000,
            "totalCredit": 650000,
            "netBalance": 150000,
            "transactionCount": 45
        },
        "mutations": [
            {
                "invoiceNumber": "GATE1A11BB97DF88D56530993",
                "description": "Pembelian Mobile Legends - 172 Diamonds",
                "amount": 44851,
                "type": "DEBIT",
                "balanceBefore": 200000,
                "balanceAfter": 155149,
                "currency": "IDR",
                "createdAt": "2025-09-25T10:30:00+07:00"
            },
            {
                "invoiceNumber": "DEP5E55FF11IJ22H90974337",
                "description": "Isi Ulang Saldo via QRIS",
                "amount": 200000,
                "type": "CREDIT",
                "balanceBefore": 0,
                "balanceAfter": 200000,
                "currency": "IDR",
                "createdAt": "2025-09-25T10:00:00+07:00"
            }
        ],
        "pagination": {
            "limit": 10,
            "page": 1,
            "totalRows": 45,
            "totalPages": 5
        }
    }
}
```

**Mutation Types:**

| Type | Description | Icon |
|------|-------------|------|
| DEBIT | Money out (purchases) | ↓ Red |
| CREDIT | Money in (top-ups, refunds) | ↑ Green |

---

### 40. Get Reports

Retrieve daily transaction reports (aggregated by date).

**Endpoint:** `GET /v2/reports`

**Headers:**

```
Authorization: Bearer {access_token}
```

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| region | string | No | Region code. Default: user's currentRegion |
| limit | integer | No | Items per page. Default: 10, Max: 100 |
| page | integer | No | Page number. Default: 1 |
| startDate | string | No | Start date (YYYY-MM-DD) |
| endDate | string | No | End date (YYYY-MM-DD) |

**Response:**

```json
{
    "data": {
        "overview": {
            "totalDays": 30,
            "totalTransactions": 125,
            "totalAmount": 1250000,
            "averagePerDay": 41666.67,
            "highestDay": {
                "date": "2025-09-25",
                "amount": 150000
            },
            "lowestDay": {
                "date": "2025-09-05",
                "amount": 15000
            }
        },
        "reports": [
            {
                "date": "2025-09-25",
                "totalTransactions": 100,
                "totalAmount": 150000,
                "currency": "IDR"
            },
            {
                "date": "2025-09-24",
                "totalTransactions": 85,
                "totalAmount": 120000,
                "currency": "IDR"
            }
        ],
        "pagination": {
            "limit": 10,
            "page": 1,
            "totalRows": 30,
            "totalPages": 3
        }
    }
}
```

---

### 41. Get Deposits

Retrieve deposit/top-up transaction history.

**Endpoint:** `GET /v2/deposits`

**Headers:**

```
Authorization: Bearer {access_token}
```

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| region | string | No | Region code. Default: user's currentRegion |
| limit | integer | No | Items per page. Default: 10, Max: 100 |
| page | integer | No | Page number. Default: 1 |
| search | string | No | Search by invoice number |
| status | string | No | Filter: ALL, SUCCESS, PENDING, EXPIRED, FAILED |
| startDate | string | No | Start date (YYYY-MM-DD) |
| endDate | string | No | End date (YYYY-MM-DD) |

**Response:**

```json
{
    "data": {
        "overview": {
            "totalDeposits": 15,
            "totalAmount": 1500000,
            "successCount": 12,
            "pendingCount": 2,
            "failedCount": 1
        },
        "deposits": [
            {
                "invoiceNumber": "DEP5E55FF11IJ22H90974337",
                "status": "SUCCESS",
                "amount": 200000,
                "payment": {
                    "code": "QRIS",
                    "name": "QRIS"
                },
                "currency": "IDR",
                "createdAt": "2025-09-25T10:00:00+07:00",
                "paidAt": "2025-09-25T10:01:30+07:00"
            },
            {
                "invoiceNumber": "DEP6F66GG22JK33I01085448",
                "status": "PENDING",
                "amount": 100000,
                "payment": {
                    "code": "BCA_VA",
                    "name": "BCA Virtual Account"
                },
                "currency": "IDR",
                "createdAt": "2025-09-25T09:30:00+07:00",
                "expiredAt": "2025-09-26T09:30:00+07:00"
            }
        ],
        "pagination": {
            "limit": 10,
            "page": 1,
            "totalRows": 15,
            "totalPages": 2
        }
    }
}
```

---

### 42. Create Deposit Inquiry

Pre-validate deposit request before creating actual deposit.

**Endpoint:** `POST /v2/deposits/inquiries`

**Headers:**

```
Authorization: Bearer {access_token}
Content-Type: application/json
```

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| region | string | No | Region code. Default: user's currentRegion |

**Request Body:**

```json
{
    "amount": 100000,
    "paymentCode": "QRIS"
}
```

**Response:**

```json
{
    "data": {
        "validationToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "expiresAt": "2025-09-25T10:35:00+07:00",
        "deposit": {
            "amount": 100000,
            "pricing": {
                "subtotal": 100000,
                "paymentFee": 700,
                "total": 100700,
                "currency": "IDR"
            },
            "payment": {
                "code": "QRIS",
                "name": "QRIS",
                "currency": "IDR",
                "minAmount": 1000,
                "maxAmount": 10000000,
                "feeAmount": 0,
                "feePercentage": 0.7
            }
        }
    }
}
```

**Error Response:**

```json
{
    "error": {
        "code": "AMOUNT_TOO_LOW",
        "message": "Jumlah deposit terlalu kecil",
        "details": "Minimum deposit untuk QRIS adalah Rp 1.000"
    }
}
```

---

### 43. Create Deposit

Create deposit transaction using validation token.

**Endpoint:** `POST /v2/deposits`

**Headers:**

```
Authorization: Bearer {access_token}
Content-Type: application/json
```

**Request Body:**

```json
{
    "validationToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response (QRIS):**

```json
{
    "data": {
        "step": "SUCCESS",
        "deposit": {
            "invoiceNumber": "DEP5E55FF11IJ22H90974337",
            "status": "PENDING",
            "amount": 100000,
            "pricing": {
                "subtotal": 100000,
                "paymentFee": 700,
                "total": 100700,
                "currency": "IDR"
            },
            "payment": {
                "code": "QRIS",
                "name": "QRIS",
                "instruction": "<p>Gunakan E-wallet atau aplikasi mobile banking untuk scan QRIS</p>",
                "qrCode": "00020101021226660016ID.CO.QRIS.WWW...",
                "qrCodeImage": "https://api.gate.co.id/v2/qr/generate?data=...",
                "expiredAt": "2025-09-25T11:00:00+07:00"
            },
            "createdAt": "2025-09-25T10:00:00+07:00",
            "expiredAt": "2025-09-25T11:00:00+07:00"
        }
    }
}
```

**Response (Virtual Account):**

```json
{
    "data": {
        "step": "SUCCESS",
        "deposit": {
            "invoiceNumber": "DEP6F66GG22JK33I01085448",
            "status": "PENDING",
            "amount": 100000,
            "pricing": {
                "subtotal": 100000,
                "paymentFee": 4000,
                "total": 104000,
                "currency": "IDR"
            },
            "payment": {
                "code": "BCA_VA",
                "name": "BCA Virtual Account",
                "instruction": "<ol><li>Pilih m-Transfer > BCA Virtual Account</li>...</ol>",
                "accountNumber": "80777123456789012",
                "bankName": "BCA",
                "accountName": "GATE INDONESIA",
                "expiredAt": "2025-09-26T10:00:00+07:00"
            },
            "createdAt": "2025-09-25T10:00:00+07:00",
            "expiredAt": "2025-09-26T10:00:00+07:00"
        }
    }
}
```

**Response (E-Wallet Redirect):**

```json
{
    "data": {
        "step": "SUCCESS",
        "deposit": {
            "invoiceNumber": "DEP7G77HH33KL44J12196559",
            "status": "PENDING",
            "amount": 100000,
            "pricing": {
                "subtotal": 100000,
                "paymentFee": 1500,
                "total": 101500,
                "currency": "IDR"
            },
            "payment": {
                "code": "DANA",
                "name": "DANA",
                "instruction": "<ol><li>Kamu akan diarahkan ke aplikasi DANA</li>...</ol>",
                "redirectUrl": "https://app.dana.id/pay/...",
                "deeplink": "dana://pay/...",
                "expiredAt": "2025-09-25T10:30:00+07:00"
            },
            "createdAt": "2025-09-25T10:00:00+07:00",
            "expiredAt": "2025-09-25T10:30:00+07:00"
        }
    }
}
```

---

## Error Codes

### Common Error Codes

| Code | Description |
|------|-------------|
| `VALIDATION_ERROR` | Request validation failed |
| `NOT_FOUND` | Resource not found |
| `UNAUTHORIZED` | Invalid or missing authentication |
| `RATE_LIMIT_EXCEEDED` | Too many requests |
| `ACCOUNT_NOT_FOUND` | Game/service account not found |
| `SKU_UNAVAILABLE` | Product variant out of stock |
| `PROMO_EXPIRED` | Promo code has expired |
| `PROMO_NOT_FOUND` | Promo code not found |
| `PROMO_LIMIT_REACHED` | Promo usage limit reached |
| `INVALID_PAYMENT_METHOD` | Payment method not allowed |
| `TOKEN_EXPIRED` | Validation token has expired |
| `INVALID_TOKEN` | Invalid token |
| `INVALID_CREDENTIALS` | Wrong email or password |
| `INVALID_PASSWORD` | Current password is incorrect |
| `EMAIL_NOT_VERIFIED` | Email not verified |
| `ACCOUNT_SUSPENDED` | Account has been suspended |
| `MFA_REQUIRED` | MFA verification required |
| `INVALID_MFA_CODE` | Invalid MFA code |
| `AMOUNT_TOO_LOW` | Amount below minimum |
| `AMOUNT_TOO_HIGH` | Amount above maximum |
| `INSUFFICIENT_BALANCE` | Not enough balance |

---

## Best Practices

### Response Handling

```javascript
const response = await fetch('/v2/products?region=ID');
const json = await response.json();

if (json.data) {
    // Success
    console.log(json.data);
} else if (json.error) {
    // Error
    console.error(json.error.message);
    if (json.error.fields) {
        // Validation errors
        Object.entries(json.error.fields).forEach(([field, message]) => {
            console.error(`${field}: ${message}`);
        });
    }
}
```

### TypeScript Types

```typescript
interface ApiResponse<T> {
    data?: T;
    error?: {
        code: string;
        message: string;
        details?: string;
        fields?: Record<string, string>;
    };
}

interface User {
    id: string;
    firstName: string;
    lastName: string;
    email: string;
    phoneNumber: string | null;
    profilePicture: string | null;
    status: 'ACTIVE' | 'INACTIVE' | 'SUSPENDED';
    primaryRegion: 'ID' | 'MY' | 'PH' | 'SG' | 'TH';
    currentRegion: 'ID' | 'MY' | 'PH' | 'SG' | 'TH';
    currency: 'IDR' | 'MYR' | 'PHP' | 'SGD' | 'THB';
    balance: {
        IDR: number;
        MYR: number;
        PHP: number;
        SGD: number;
        THB: number;
    };
    membership: {
        level: 'CLASSIC' | 'PRESTIGE' | 'ROYAL';
        name: string;
    };
    mfaStatus: 'ACTIVE' | 'INACTIVE';
    createdAt: string;
    lastLoginAt: string;
}
```

### Currency Formatting

```javascript
const formatCurrency = (amount, currency) => {
    const formats = {
        IDR: { locale: 'id-ID', symbol: 'Rp' },
        MYR: { locale: 'ms-MY', symbol: 'RM' },
        PHP: { locale: 'en-PH', symbol: '₱' },
        SGD: { locale: 'en-SG', symbol: 'S$' },
        THB: { locale: 'th-TH', symbol: '฿' }
    };
    
    const format = formats[currency];
    return new Intl.NumberFormat(format.locale, {
        style: 'currency',
        currency: currency
    }).format(amount);
};

// Examples:
// formatCurrency(150000, 'IDR') → "Rp150.000"
// formatCurrency(500, 'MYR') → "RM500.00"
// formatCurrency(100, 'SGD') → "S$100.00"
```

---

## Membership Levels

| Level | Name | Min Transaction | Benefits |
|-------|------|-----------------|----------|
| CLASSIC | Classic | Rp 0 | Standard transactions, 24/7 support, 1% bonus points |
| PRESTIGE | Prestige | Rp 5,000,000 | Up to 5% discount, priority support, 3% bonus points, premium promos |
| ROYAL | Royal | Rp 10,000,000 | Up to 10% discount, dedicated manager, 5% bonus points, VIP promos, priority transactions |

> **Note:** Membership level automatically upgrades based on total transaction volume (last 90 days).

---

## Authentication Flow Diagrams

### Regular Registration

```
User fills form → POST /v2/auth/register → step: EMAIL_VERIFICATION
                                                    ↓
                     User logged in ← POST /v2/auth/verify-email ← User checks email
```

### Google Registration

```
User clicks "Sign up with Google" → Google OAuth → POST /v2/auth/register/google → step: SUCCESS + token → User logged in
```

### Login with MFA

```
User enters credentials → POST /v2/auth/login → step: MFA_VERIFICATION + mfaToken
                                                            ↓
                        User logged in ← POST /v2/auth/verify-mfa ← User enters MFA code
```

### Forgot Password

```
User clicks "Forgot Password" → POST /v2/auth/forgot-password → User checks email
                                                                      ↓
                     User must login again ← POST /v2/auth/reset-password ← User enters new password
```

---

## Support

| Channel | Contact |
|---------|---------|
| Email | support@gate.co.id |
| WhatsApp | Check `/v2/contacts` endpoint |
| Documentation | https://docs.gate.co.id |
| Status Page | https://status.gate.co.id |

---

## Changelog

### Version 2.0 (December 2025)

**New Endpoints:**
- `GET /v2/regions` - Get supported regions
- `GET /v2/languages` - Get supported languages
- `GET /v2/contacts` - Get contact information
- `GET /v2/sku/promos` - Get all promotional SKUs
- Complete authentication endpoints (register, login, MFA, password reset)
- Complete dashboard endpoints (transactions, mutations, reports, deposits)

**Breaking Changes:**
- Changed base path from `/v1` to `/v2`
- Changed `inquirys` to `inquiries` in endpoint paths
- Added `banner` field to products response
- Added `name` field to fields response
- Simplified response structure (removed success, code, message wrapper)

**Improvements:**
- Multi-region support with currency switching
- Multi-currency balance per user
- Comprehensive payment channels for SEA region
- Enhanced SKU data with badges and sections
- Better field descriptions with hints

---

**Last Updated:** December 3, 2025 | **API Version:** v2.0 | **Document Version:** 2.0
