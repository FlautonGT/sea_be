# Gate.co.id API Documentation v2.0

## Base Information

### URLs

- **Main URL:** `https://gate.id` or `https://gate.co.id`
- **API URL:** `https://gateway.gate.id` or `https://gateway.gate.co.id`

### API Paths

- **Main Path:** `/{version}` (e.g., `/v2`)
- **Admin Path:** `/admin/{version}` (e.g., `/admin/v2`)

### Timezone

- **Server Timezone:** Jakarta Time (UTC+07:00)
- **DateTime Format:** ISO 8601 format with timezone offset
- **Example:** `2025-12-31T23:59:59+07:00`

### Supported Regions

| Code | Country | Currency |
| --- | --- | --- |
| ID | Indonesia | IDR |
| MY | Malaysia | MYR |
| PH | Philippines | PHP |
| SG | Singapore | SGD |
| TH | Thailand | THB |

---

## Authentication

### Bearer Token (Optional)

Authentication is **optional** for all endpoints. Gate allows users to make purchases without logging in.

**When to include authentication:**
- ✅ **With Auth:** Transaction will be saved to user's account history
- ❌ **Without Auth:** Transaction will be processed but not linked to user account

**Header Format:**

```
Authorization: Bearer {your_access_token}
```

---

## Standard Response Structure

All API responses follow this structure with **camelCase** naming convention:

### Success Response

```json
{
    "data": {
        // Object or Array
    }
}
```

**Examples:**

Single object:
```json
{
    "data": {
        "title": "Top Up Game",
        "code": "top-up-game"
    }
}
```

Array of objects:
```json
{
    "data": [
        {
            "title": "Top Up Game",
            "code": "top-up-game"
        },
        {
            "title": "Voucher",
            "code": "voucher"
        }
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
| --- | --- |
| 200 | Success |
| 201 | Created |
| 400 | Bad Request |
| 401 | Unauthorized |
| 404 | Not Found |
| 422 | Validation Error |
| 429 | Too Many Requests |
| 500 | Internal Server Error |

---

## API Endpoints

### 1. Get Regions

Retrieve all available regions/countries with their configuration.

**Endpoint:** `GET /v2/regions`

**Response Example:**

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

**Field Descriptions:**
- `country`: Full country name
- `code`: ISO 3166-1 alpha-2 country code
- `currency`: ISO 4217 currency code
- `image`: URL to country flag image (SVG format)
- `isDefault`: Default selected region (true for one region only)

**Notes:**
- Used for region selector in the UI
- `isDefault` determines which region is selected by default
- Currency is used for price display and payment processing

---

### 2. Get Languages

Retrieve all available languages for the platform.

**Endpoint:** `GET /v2/languages`

**Response Example:**

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

**Field Descriptions:**
- `country`: Country name associated with the language
- `code`: ISO 639-1 language code
- `name`: Display name of the language
- `image`: URL to flag image representing the language
- `isDefault`: Default selected language

**Notes:**
- Used for language selector in the UI
- Content is localized based on selected language code
- API responses (error messages, labels) will be in selected language

---

### 3. Get Contacts

Retrieve all contact information and social media links.

**Endpoint:** `GET /v2/contacts`

**Response Example:**

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

**Field Descriptions:**
- `email`: Support email address
- `phone`: Customer service phone number (E.164 format)
- `whatsapp`: WhatsApp chat URL
- `instagram`: Instagram profile URL
- `facebook`: Facebook page URL
- `x`: X (formerly Twitter) profile URL
- `youtube`: YouTube channel URL
- `telegram`: Telegram group/channel URL (optional)
- `discord`: Discord server invite URL (optional)

**Notes:**
- Used in footer and contact page
- Some fields may be null/empty if not available
- WhatsApp URL includes pre-filled message support

---

### 4. Get Popups

Display promotional popups to users.

**Endpoint:** `GET /v2/popups`

**Query Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| region | string | Yes | Region code (ID, MY, PH, SG, TH) |

**Response Example:**

```json
{
    "data": {
        "title": "🔥 TOP UP DELTA FORCE MULAI HARI INI! 🔥",
        "content": "<p>🎉 <em>Diskon spesial hanya berlaku sampai 23 Desember!</em> 🎉</p>",
        "image": "https://nos.jkt-1.neo.id/gate/popups/delta-force-promo.webp",
        "link": "https://gate.co.id/id-id/delta-force",
        "isActive": true
    }
}
```

**Field Descriptions:**
- `title`: Popup title
- `content`: HTML content for popup body
- `image`: Featured image URL
- `link`: Call-to-action link (optional)
- `isActive`: Whether popup should be displayed

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

**Notes:**
- Popup is region-specific
- Can be used for announcements, promotions, or important updates
- If `isActive` is false, don't show popup

---

### 5. Get Banners

Retrieve promotional banners for the homepage.

**Endpoint:** `GET /v2/banners`

**Query Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| region | string | Yes | Region code (ID, MY, PH, SG, TH) |

**Response Example:**

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
        },
        {
            "title": "Topup PUBG Mobile Murah",
            "description": "UC PUBG Mobile dengan proses instan",
            "href": "/id-id/pubg-mobile",
            "image": "https://nos.jkt-1.neo.id/gate/banners/pubg-banner.webp",
            "order": 3
        }
    ]
}
```

**Field Descriptions:**
- `title`: Banner title
- `description`: Banner description/subtitle
- `href`: Link destination (relative or absolute URL)
- `image`: Banner image URL (recommended size: 1200x400px)
- `order`: Display order (ascending)

**Notes:**
- Banners are sorted by `order` field
- Images should be optimized (WebP format recommended)
- Use for homepage carousel/slider

---

### 6. Get Categories

Retrieve all available product categories.

**Endpoint:** `GET /v2/categories`

**Query Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| region | string | Yes | Region code (ID, MY, PH, SG, TH) |

**Response Example:**

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
            "title": "Pulsa & Paket Data",
            "code": "credit-or-data",
            "description": "Pulsa dan paket data semua operator",
            "icon": "https://nos.jkt-1.neo.id/gate/icons/phone.svg",
            "order": 4
        },
        {
            "title": "Streaming",
            "code": "streaming",
            "description": "Langganan Netflix, Spotify, Disney+ dan lainnya",
            "icon": "https://nos.jkt-1.neo.id/gate/icons/play.svg",
            "order": 5
        },
        {
            "title": "Token Listrik",
            "code": "electricity",
            "description": "Token listrik PLN",
            "icon": "https://nos.jkt-1.neo.id/gate/icons/lightning.svg",
            "order": 6
        }
    ]
}
```

**Field Descriptions:**
- `title`: Category display name
- `code`: Unique category identifier (used in filters)
- `description`: Category description
- `icon`: Category icon URL (SVG recommended)
- `order`: Display order

**Notes:**
- Categories may vary by region
- Use `code` for filtering products
- `order` determines display sequence

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

**Response Example (All Products):**

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
        },
        {
            "code": "PUBGM",
            "slug": "pubg-mobile",
            "title": "PUBG Mobile",
            "subtitle": "Tencent Games",
            "description": "Top up UC PUBG Mobile dengan harga termurah",
            "publisher": "Tencent Games",
            "thumbnail": "https://nos.jkt-1.neo.id/gate/products/pubgm-icon.webp",
            "banner": "https://nos.jkt-1.neo.id/gate/products/pubgm-banner.webp",
            "isPopular": true,
            "isAvailable": true,
            "tags": ["Battle Royale", "Shooter", "Survival"],
            "category": {
                "title": "Top Up Game",
                "code": "top-up-game"
            }
        },
        {
            "code": "GENSHIN",
            "slug": "genshin-impact",
            "title": "Genshin Impact",
            "subtitle": "HoYoverse",
            "description": "Top up Genesis Crystal Genshin Impact official",
            "publisher": "HoYoverse",
            "thumbnail": "https://nos.jkt-1.neo.id/gate/products/genshin-icon.webp",
            "banner": "https://nos.jkt-1.neo.id/gate/products/genshin-banner.webp",
            "isPopular": true,
            "isAvailable": true,
            "tags": ["RPG", "Open World", "Adventure"],
            "category": {
                "title": "Top Up Game",
                "code": "top-up-game"
            }
        },
        {
            "code": "VALORANT",
            "slug": "valorant",
            "title": "Valorant",
            "subtitle": "Riot Games",
            "description": "Top up Valorant Points (VP) untuk semua region",
            "publisher": "Riot Games",
            "thumbnail": "https://nos.jkt-1.neo.id/gate/products/valorant-icon.webp",
            "banner": "https://nos.jkt-1.neo.id/gate/products/valorant-banner.webp",
            "isPopular": false,
            "isAvailable": true,
            "tags": ["FPS", "Tactical", "Multiplayer"],
            "category": {
                "title": "Top Up Game",
                "code": "top-up-game"
            }
        },
        {
            "code": "COD",
            "slug": "call-of-duty-mobile",
            "title": "Call of Duty Mobile",
            "subtitle": "Activision",
            "description": "Top up CP Call of Duty Mobile dengan aman",
            "publisher": "Activision Publishing Inc",
            "thumbnail": "https://nos.jkt-1.neo.id/gate/products/cod-icon.webp",
            "banner": "https://nos.jkt-1.neo.id/gate/products/cod-banner.webp",
            "isPopular": false,
            "isAvailable": true,
            "tags": ["FPS", "Battle Royale", "Action"],
            "category": {
                "title": "Top Up Game",
                "code": "top-up-game"
            }
        },
        {
            "code": "DANA",
            "slug": "dana",
            "title": "Dana",
            "subtitle": "E-Wallet",
            "description": "Top up saldo Dana dengan mudah dan cepat",
            "publisher": "PT Espay Debit Indonesia Koe",
            "thumbnail": "https://nos.jkt-1.neo.id/gate/products/dana-icon.webp",
            "banner": "https://nos.jkt-1.neo.id/gate/products/dana-banner.webp",
            "isPopular": true,
            "isAvailable": true,
            "tags": ["E-Wallet", "Payment", "Digital"],
            "category": {
                "title": "E-Money",
                "code": "e-money"
            }
        },
        {
            "code": "GOPAY",
            "slug": "gopay",
            "title": "GoPay",
            "subtitle": "E-Wallet",
            "description": "Top up GoPay untuk transaksi Gojek dan merchant",
            "publisher": "PT Dompet Anak Bangsa",
            "thumbnail": "https://nos.jkt-1.neo.id/gate/products/gopay-icon.webp",
            "banner": "https://nos.jkt-1.neo.id/gate/products/gopay-banner.webp",
            "isPopular": true,
            "isAvailable": true,
            "tags": ["E-Wallet", "Gojek", "Payment"],
            "category": {
                "title": "E-Money",
                "code": "e-money"
            }
        },
        {
            "code": "OVO",
            "slug": "ovo",
            "title": "OVO",
            "subtitle": "E-Wallet",
            "description": "Top up OVO Points dengan proses instant",
            "publisher": "PT Visionet Internasional",
            "thumbnail": "https://nos.jkt-1.neo.id/gate/products/ovo-icon.webp",
            "banner": "https://nos.jkt-1.neo.id/gate/products/ovo-banner.webp",
            "isPopular": false,
            "isAvailable": true,
            "tags": ["E-Wallet", "Payment", "Cashback"],
            "category": {
                "title": "E-Money",
                "code": "e-money"
            }
        },
        {
            "code": "SHOPEEPAY",
            "slug": "shopeepay",
            "title": "ShopeePay",
            "subtitle": "E-Wallet",
            "description": "Top up ShopeePay untuk belanja di Shopee",
            "publisher": "PT Airpay International",
            "thumbnail": "https://nos.jkt-1.neo.id/gate/products/shopeepay-icon.webp",
            "banner": "https://nos.jkt-1.neo.id/gate/products/shopeepay-banner.webp",
            "isPopular": false,
            "isAvailable": true,
            "tags": ["E-Wallet", "Shopee", "Shopping"],
            "category": {
                "title": "E-Money",
                "code": "e-money"
            }
        },
        {
            "code": "TELKOMSEL",
            "slug": "telkomsel",
            "title": "Telkomsel",
            "subtitle": "Pulsa & Paket Data",
            "description": "Pulsa dan paket data Telkomsel dengan harga terbaik",
            "publisher": "PT Telkomsel",
            "thumbnail": "https://nos.jkt-1.neo.id/gate/products/telkomsel-icon.webp",
            "banner": "https://nos.jkt-1.neo.id/gate/products/telkomsel-banner.webp",
            "isPopular": true,
            "isAvailable": true,
            "tags": ["Pulsa", "Data", "Operator"],
            "category": {
                "title": "Pulsa & Paket Data",
                "code": "credit-or-data"
            }
        },
        {
            "code": "XL",
            "slug": "xl-axiata",
            "title": "XL Axiata",
            "subtitle": "Pulsa & Paket Data",
            "description": "Pulsa dan paket data XL dengan proses cepat",
            "publisher": "PT XL Axiata",
            "thumbnail": "https://nos.jkt-1.neo.id/gate/products/xl-icon.webp",
            "banner": "https://nos.jkt-1.neo.id/gate/products/xl-banner.webp",
            "isPopular": false,
            "isAvailable": true,
            "tags": ["Pulsa", "Data", "Operator"],
            "category": {
                "title": "Pulsa & Paket Data",
                "code": "credit-or-data"
            }
        },
        {
            "code": "INDOSAT",
            "slug": "indosat-ooredoo",
            "title": "Indosat Ooredoo",
            "subtitle": "Pulsa & Paket Data",
            "description": "Pulsa dan paket data Indosat dengan harga murah",
            "publisher": "PT Indosat Ooredoo Hutchison",
            "thumbnail": "https://nos.jkt-1.neo.id/gate/products/indosat-icon.webp",
            "banner": "https://nos.jkt-1.neo.id/gate/products/indosat-banner.webp",
            "isPopular": false,
            "isAvailable": true,
            "tags": ["Pulsa", "Data", "Operator"],
            "category": {
                "title": "Pulsa & Paket Data",
                "code": "credit-or-data"
            }
        },
        {
            "code": "NETFLIX",
            "slug": "netflix",
            "title": "Netflix",
            "subtitle": "Streaming",
            "description": "Voucher Netflix untuk nonton film dan series",
            "publisher": "Netflix Inc",
            "thumbnail": "https://nos.jkt-1.neo.id/gate/products/netflix-icon.webp",
            "banner": "https://nos.jkt-1.neo.id/gate/products/netflix-banner.webp",
            "isPopular": true,
            "isAvailable": true,
            "tags": ["Streaming", "Entertainment", "Movies"],
            "category": {
                "title": "Streaming",
                "code": "streaming"
            }
        },
        {
            "code": "SPOTIFY",
            "slug": "spotify",
            "title": "Spotify",
            "subtitle": "Music Streaming",
            "description": "Spotify Premium untuk musik tanpa iklan",
            "publisher": "Spotify AB",
            "thumbnail": "https://nos.jkt-1.neo.id/gate/products/spotify-icon.webp",
            "banner": "https://nos.jkt-1.neo.id/gate/products/spotify-banner.webp",
            "isPopular": true,
            "isAvailable": true,
            "tags": ["Music", "Streaming", "Premium"],
            "category": {
                "title": "Streaming",
                "code": "streaming"
            }
        },
        {
            "code": "PLN",
            "slug": "token-listrik-pln",
            "title": "Token Listrik PLN",
            "subtitle": "PLN",
            "description": "Beli token listrik PLN dengan mudah",
            "publisher": "PT PLN (Persero)",
            "thumbnail": "https://nos.jkt-1.neo.id/gate/products/pln-icon.webp",
            "banner": "https://nos.jkt-1.neo.id/gate/products/pln-banner.webp",
            "isPopular": true,
            "isAvailable": true,
            "tags": ["Electricity", "Utility", "PLN"],
            "category": {
                "title": "Token Listrik",
                "code": "electricity"
            }
        }
    ]
}
```

**Response Example (Single Product by productCode):**

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

**Field Descriptions:**
- `code`: Unique product identifier (used in API calls)
- `slug`: URL-friendly identifier (used in web URLs)
- `title`: Product display name
- `subtitle`: Product subtitle (usually publisher or category)
- `description`: Product description
- `publisher`: Official publisher name
- `thumbnail`: Product icon/logo (recommended size: 200x200px)
- `banner`: Product banner image (recommended size: 1200x400px) **NEW**
- `isPopular`: Flag to display product in "Popular" section
- `isAvailable`: Product availability status
- `tags`: Product tags/labels for filtering
- `category`: Product category information
- `features`: Product features (only in single product response)
- `howToOrder`: Step-by-step ordering guide (only in single product response)

**Notes:**
- `banner` is displayed above the `thumbnail` in product detail page
- When `productCode` is provided, returns single product with additional details
- Products with `isAvailable: false` should be shown as "Coming Soon" or hidden

---

### 8. Get Product Fields

Retrieve input fields required for a specific product.

**Endpoint:** `GET /v2/fields`

**Query Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| region | string | Yes | Region code (ID, MY, PH, SG, TH) |
| productCode | string | Yes | Product code (e.g., DANA, MLBB) |

**Response Example (Mobile Legends):**

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

**Response Example (DANA E-Wallet):**

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

**Response Example (Telkomsel Pulsa):**

```json
{
    "data": [
        {
            "name": "Nomor Telepon",
            "key": "phoneNumber",
            "type": "number",
            "label": "Masukkan Nomor Telkomsel",
            "required": true,
            "minLength": 10,
            "maxLength": 13,
            "placeholder": "0812xxxxxxxx",
            "pattern": "^(0811|0812|0813|0821|0822|0823|0851|0852|0853)[0-9]{7,9}$",
            "hint": "Nomor Telkomsel yang akan diisi pulsa"
        }
    ]
}
```

**Field Descriptions:**
- `name`: Field display name (shown above input) **NEW**
- `key`: Field identifier (used as request body key)
- `type`: Input type (number, text, email, select)
- `label`: Field label (placeholder-style label)
- `required`: Whether field must be filled
- `minLength`: Minimum character length
- `maxLength`: Maximum character length
- `placeholder`: Example input text
- `pattern`: Regex pattern for validation (optional)
- `hint`: Help text shown below input (optional)

**Field Types:**
- `number`: Numeric input only
- `text`: Text input
- `email`: Email format validation
- `select`: Dropdown selection

**Notes:**
- `name` is displayed above the input field (like "Id Pengguna" in screenshot)
- `label` is shown as placeholder or floating label
- `hint` provides additional guidance to users

---

### 9. Get Sections

Retrieve product sections for organizing SKUs (e.g., Promo, Regular).

**Endpoint:** `GET /v2/sections`

**Query Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| region | string | Yes | Region code (ID, MY, PH, SG, TH) |
| productCode | string | Yes | Product code (e.g., DANA, MLBB) |

**Response Example:**

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

**Field Descriptions:**
- `title`: Section display name
- `code`: Unique section identifier
- `icon`: Optional emoji or icon
- `order`: Display order

**Notes:**
- Sections are used as tabs in SKU selection (see screenshot)
- Each section contains specific SKU groups
- "Semua Item" typically shows all SKUs without filtering

---

### 10. Get SKUs (Stock Keeping Units)

Retrieve available product variants/denominations with pricing.

**Endpoint:** `GET /v2/skus`

**Query Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| region | string | Yes | Region code (ID, MY, PH, SG, TH) |
| productCode | string | Yes | Product code (e.g., DANA, MLBB) |
| sectionCode | string | No | Filter by section code |

**Response Example (Mobile Legends):**

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
        },
        {
            "code": "MLBB_257",
            "name": "257 Diamonds",
            "description": "257 (234+23) Diamonds",
            "currency": "IDR",
            "price": 73700,
            "originalPrice": 75000,
            "discount": 1.73,
            "image": "https://nos.jkt-1.neo.id/gate/products/mlbb-diamond.webp",
            "info": "Bonus +23 Diamonds",
            "processTime": 0,
            "isAvailable": true,
            "isFeatured": false,
            "section": {
                "title": "Topup Instan",
                "code": "topup-instant"
            }
        },
        {
            "code": "MLBB_344",
            "name": "344 Diamonds",
            "description": "344 (312+32) Diamonds",
            "currency": "IDR",
            "price": 98450,
            "originalPrice": 100000,
            "discount": 1.55,
            "image": "https://nos.jkt-1.neo.id/gate/products/mlbb-diamond.webp",
            "info": "Bonus +32 Diamonds",
            "processTime": 0,
            "isAvailable": true,
            "isFeatured": false,
            "section": {
                "title": "Topup Instan",
                "code": "topup-instant"
            }
        },
        {
            "code": "MLBB_429",
            "name": "429 Diamonds",
            "description": "429 (390+39) Diamonds",
            "currency": "IDR",
            "price": 122700,
            "originalPrice": 125000,
            "discount": 1.84,
            "image": "https://nos.jkt-1.neo.id/gate/products/mlbb-diamond.webp",
            "info": "Bonus +39 Diamonds",
            "processTime": 0,
            "isAvailable": true,
            "isFeatured": false,
            "section": {
                "title": "Topup Instan",
                "code": "topup-instant"
            }
        },
        {
            "code": "MLBB_514",
            "name": "514 Diamonds",
            "description": "514 (468+46) Diamonds",
            "currency": "IDR",
            "price": 147200,
            "originalPrice": 150000,
            "discount": 1.87,
            "image": "https://nos.jkt-1.neo.id/gate/products/mlbb-diamond.webp",
            "info": "Bonus +46 Diamonds",
            "processTime": 0,
            "isAvailable": true,
            "isFeatured": false,
            "section": {
                "title": "Topup Instan",
                "code": "topup-instant"
            }
        },
        {
            "code": "MLBB_706",
            "name": "706 Diamonds",
            "description": "706 (625+81) Diamonds",
            "currency": "IDR",
            "price": 196400,
            "originalPrice": 200000,
            "discount": 1.8,
            "image": "https://nos.jkt-1.neo.id/gate/products/mlbb-diamond.webp",
            "info": "Bonus +81 Diamonds",
            "processTime": 0,
            "isAvailable": true,
            "isFeatured": false,
            "section": {
                "title": "Topup Instan",
                "code": "topup-instant"
            }
        },
        {
            "code": "MLBB_878",
            "name": "878 Diamonds",
            "description": "878 (781+97) Diamonds",
            "currency": "IDR",
            "price": 245100,
            "originalPrice": 250000,
            "discount": 1.96,
            "image": "https://nos.jkt-1.neo.id/gate/products/mlbb-diamond.webp",
            "info": "Bonus +97 Diamonds",
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

**Response Example (PUBG Mobile):**

```json
{
    "data": [
        {
            "code": "PUBGM_60",
            "name": "60 UC",
            "description": "60 Unknown Cash",
            "currency": "IDR",
            "price": 16000,
            "originalPrice": 16000,
            "discount": 0,
            "image": "https://nos.jkt-1.neo.id/gate/products/pubgm-uc.webp",
            "info": "",
            "processTime": 0,
            "isAvailable": true,
            "isFeatured": false,
            "section": {
                "title": "Topup Instan",
                "code": "topup-instant"
            }
        },
        {
            "code": "PUBGM_325",
            "name": "325 UC",
            "description": "325 Unknown Cash",
            "currency": "IDR",
            "price": 80000,
            "originalPrice": 85000,
            "discount": 5.88,
            "image": "https://nos.jkt-1.neo.id/gate/products/pubgm-uc.webp",
            "info": "",
            "processTime": 0,
            "isAvailable": true,
            "isFeatured": true,
            "section": {
                "title": "Spesial Item",
                "code": "special-item"
            },
            "badge": {
                "text": "Best Seller",
                "color": "#4CAF50"
            }
        },
        {
            "code": "PUBGM_660",
            "name": "660 UC",
            "description": "660 Unknown Cash",
            "currency": "IDR",
            "price": 160000,
            "originalPrice": 170000,
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
        },
        {
            "code": "PUBGM_3850",
            "name": "3850 UC",
            "description": "3850 Unknown Cash",
            "currency": "IDR",
            "price": 800000,
            "originalPrice": 850000,
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
        },
        {
            "code": "PUBGM_8100",
            "name": "8100 UC",
            "description": "8100 Unknown Cash",
            "currency": "IDR",
            "price": 1600000,
            "originalPrice": 1700000,
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

**Field Descriptions:**
- `code`: Unique SKU identifier
- `name`: Display name of the variant
- `description`: Detailed description
- `currency`: ISO 4217 currency code
- `price`: Final price after discount
- `originalPrice`: Original price before discount
- `discount`: Discount percentage
- `image`: SKU image URL
- `info`: Additional information (e.g., bonus details)
- `processTime`: Processing time in minutes (0 = instant)
- `isAvailable`: Stock availability status
- `isFeatured`: Whether to highlight this SKU
- `section`: Section/category information
- `badge`: Optional badge for special promotions (optional)

**Notes:**
- SKUs with `discount > 0` should show original price as strikethrough
- `isFeatured` SKUs can be highlighted with special badges
- `processTime: 0` means instant delivery
- Match the layout shown in screenshot with card grid

---

### 11. Validate Account/User ID

Validate user account before transaction.

**Endpoint:** `POST /v2/account/inquirys`

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

**Response Example (Success):**

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

**Response Example (Account Not Found):**

```json
{
    "error": {
        "code": "ACCOUNT_NOT_FOUND",
        "message": "Account not found",
        "details": "The provided User ID and Zone ID combination does not exist"
    }
}
```

**Response Example (Inconsistent Provider):**

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

### 12. Get Payment Channel Categories

Retrieve payment channel categories.

**Endpoint:** `GET /v2/payment-channel/categories`

**Query Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| region | string | Yes | Region code (ID, MY, PH, SG, TH) |

**Response Example:**

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

### 13. Get Payment Channels

Retrieve all available payment channels.

**Endpoint:** `GET /v2/payment-channels`

**Query Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| region | string | Yes | Region code (ID, MY, PH, SG, TH) |
| categoryCode | string | No | Filter by category code |

**Response Example:**

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
            "code": "OVO",
            "name": "OVO",
            "description": "Bayar menggunakan OVO",
            "image": "https://nos.jkt-1.neo.id/gate/payment/ovo.webp",
            "currency": "IDR",
            "feeAmount": 0,
            "feePercentage": 1.5,
            "minAmount": 10000,
            "maxAmount": 5000000,
            "featured": false,
            "instruction": "<ol><li>Kamu akan diarahkan ke aplikasi OVO</li><li>Pastikan saldo OVO mencukupi</li><li>Konfirmasi pembayaran dengan PIN OVO</li><li>Transaksi selesai</li></ol>",
            "category": {
                "title": "E-Wallet",
                "code": "E_WALLET"
            }
        },
        {
            "code": "GOPAY",
            "name": "GoPay",
            "description": "Bayar menggunakan GoPay",
            "image": "https://nos.jkt-1.neo.id/gate/payment/gopay.webp",
            "currency": "IDR",
            "feeAmount": 0,
            "feePercentage": 2.0,
            "minAmount": 1000,
            "maxAmount": 5000000,
            "featured": false,
            "instruction": "<ol><li>Kamu akan diarahkan ke aplikasi Gojek</li><li>Pastikan saldo GoPay mencukupi</li><li>Konfirmasi pembayaran dengan PIN GoPay</li><li>Transaksi selesai</li></ol>",
            "category": {
                "title": "E-Wallet",
                "code": "E_WALLET"
            }
        },
        {
            "code": "SHOPEEPAY",
            "name": "ShopeePay",
            "description": "Bayar menggunakan ShopeePay",
            "image": "https://nos.jkt-1.neo.id/gate/payment/shopeepay.webp",
            "currency": "IDR",
            "feeAmount": 0,
            "feePercentage": 2.0,
            "minAmount": 1000,
            "maxAmount": 5000000,
            "featured": false,
            "instruction": "<ol><li>Kamu akan diarahkan ke aplikasi Shopee</li><li>Pastikan saldo ShopeePay mencukupi</li><li>Konfirmasi pembayaran dengan PIN ShopeePay</li><li>Transaksi selesai</li></ol>",
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
            "code": "BNI_VA",
            "name": "BNI Virtual Account",
            "description": "Bayar menggunakan BNI Virtual Account",
            "image": "https://nos.jkt-1.neo.id/gate/payment/bni.webp",
            "currency": "IDR",
            "feeAmount": 4000,
            "feePercentage": 0,
            "minAmount": 10000,
            "maxAmount": 50000000,
            "featured": false,
            "instruction": "<ol><li>Pilih <strong>Transfer</strong> &gt; <strong>Virtual Account Billing</strong></li><li>Masukkan nomor Virtual Account</li><li>Periksa informasi yang tertera</li><li>Masukkan <strong>PIN BNI Mobile</strong></li><li>Transaksi selesai</li></ol>",
            "category": {
                "title": "Virtual Account",
                "code": "VIRTUAL_ACCOUNT"
            }
        },
        {
            "code": "BRI_VA",
            "name": "BRI Virtual Account",
            "description": "Bayar menggunakan BRI Virtual Account",
            "image": "https://nos.jkt-1.neo.id/gate/payment/bri.webp",
            "currency": "IDR",
            "feeAmount": 4000,
            "feePercentage": 0,
            "minAmount": 10000,
            "maxAmount": 50000000,
            "featured": false,
            "instruction": "<ol><li>Pilih <strong>Pembayaran</strong> &gt; <strong>BRIVA</strong></li><li>Masukkan nomor BRIVA</li><li>Periksa informasi yang tertera</li><li>Masukkan <strong>PIN BRI Mobile</strong></li><li>Transaksi selesai</li></ol>",
            "category": {
                "title": "Virtual Account",
                "code": "VIRTUAL_ACCOUNT"
            }
        },
        {
            "code": "MANDIRI_VA",
            "name": "Mandiri Virtual Account",
            "description": "Bayar menggunakan Mandiri Virtual Account",
            "image": "https://nos.jkt-1.neo.id/gate/payment/mandiri.webp",
            "currency": "IDR",
            "feeAmount": 4000,
            "feePercentage": 0,
            "minAmount": 10000,
            "maxAmount": 50000000,
            "featured": false,
            "instruction": "<ol><li>Pilih <strong>Bayar</strong> &gt; <strong>Multipayment</strong></li><li>Masukkan kode perusahaan <strong>70012</strong></li><li>Masukkan nomor Virtual Account</li><li>Periksa informasi yang tertera</li><li>Masukkan <strong>PIN Mandiri</strong></li><li>Transaksi selesai</li></ol>",
            "category": {
                "title": "Virtual Account",
                "code": "VIRTUAL_ACCOUNT"
            }
        },
        {
            "code": "PERMATA_VA",
            "name": "Permata Virtual Account",
            "description": "Bayar menggunakan Permata Virtual Account",
            "image": "https://nos.jkt-1.neo.id/gate/payment/permata.webp",
            "currency": "IDR",
            "feeAmount": 4000,
            "feePercentage": 0,
            "minAmount": 10000,
            "maxAmount": 50000000,
            "featured": false,
            "instruction": "<ol><li>Pilih <strong>Pembayaran Tagihan</strong></li><li>Pilih <strong>Virtual Account</strong></li><li>Masukkan nomor Virtual Account</li><li>Periksa informasi yang tertera</li><li>Masukkan <strong>PIN PermataNet</strong></li><li>Transaksi selesai</li></ol>",
            "category": {
                "title": "Virtual Account",
                "code": "VIRTUAL_ACCOUNT"
            }
        },
        {
            "code": "INDOMARET",
            "name": "Indomaret",
            "description": "Bayar di gerai Indomaret terdekat",
            "image": "https://nos.jkt-1.neo.id/gate/payment/indomaret.webp",
            "currency": "IDR",
            "feeAmount": 2500,
            "feePercentage": 0,
            "minAmount": 10000,
            "maxAmount": 5000000,
            "featured": false,
            "instruction": "<ol><li>Setelah checkout, kamu akan mendapatkan kode pembayaran</li><li>Datang ke gerai Indomaret terdekat</li><li>Berikan kode pembayaran ke kasir</li><li>Lakukan pembayaran</li><li>Simpan struk sebagai bukti</li></ol>",
            "category": {
                "title": "Convenience Store",
                "code": "RETAIL"
            }
        },
        {
            "code": "ALFAMART",
            "name": "Alfamart",
            "description": "Bayar di gerai Alfamart terdekat",
            "image": "https://nos.jkt-1.neo.id/gate/payment/alfamart.webp",
            "currency": "IDR",
            "feeAmount": 2500,
            "feePercentage": 0,
            "minAmount": 10000,
            "maxAmount": 5000000,
            "featured": false,
            "instruction": "<ol><li>Setelah checkout, kamu akan mendapatkan kode pembayaran</li><li>Datang ke gerai Alfamart terdekat</li><li>Berikan kode pembayaran ke kasir</li><li>Lakukan pembayaran</li><li>Simpan struk sebagai bukti</li></ol>",
            "category": {
                "title": "Convenience Store",
                "code": "RETAIL"
            }
        },
        {
            "code": "CARD",
            "name": "Credit/Debit Card",
            "description": "Bayar menggunakan kartu Debit atau Credit (Visa/Mastercard)",
            "image": "https://nos.jkt-1.neo.id/gate/payment/card.webp",
            "currency": "IDR",
            "feeAmount": 0,
            "feePercentage": 2.9,
            "minAmount": 10000,
            "maxAmount": 100000000,
            "featured": false,
            "instruction": "<ol><li>Masukkan nomor kartu, expired date, dan CVV</li><li>Masukkan kode OTP yang dikirim ke nomor HP terdaftar</li><li>Transaksi selesai</li></ol>",
            "category": {
                "title": "Credit or Debit Card",
                "code": "CARD"
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

**Notes:**
- Payment channels match the screenshot layout
- Featured channels appear at the top
- Calculate total fee: `feeAmount + (price * feePercentage / 100)`

---

### 14. Get Promo Codes

Retrieve available promo codes.

**Endpoint:** `GET /v2/promos`

**Query Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| region | string | Yes | Region code |
| productCode | string | Yes | Product code |

**Response Example:**

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
                {
                    "code": "MLBB",
                    "name": "Mobile Legends"
                }
            ],
            "paymentChannels": [
                {
                    "code": "DANA",
                    "name": "DANA"
                }
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

### 15. Validate Promo Code

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

**Response Example (Valid):**

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

**Response Example (Invalid):**

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

### 16. Order Inquiry

Pre-validate order before creation.

**Endpoint:** `POST /v2/orders/inquirys`

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
    "phoneNumber": "081234567890"
}
```

**Response Example:**

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
                "phoneNumber": "081234567890"
            }
        }
    }
}
```

---

### 17. Create Order

Create order with validation token.

**Endpoint:** `POST /v2/orders`

**Request Body:**

```json
{
    "validationToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response Example:**

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
                "phoneNumber": "081234567890"
            },
            "createdAt": "2025-12-03T15:25:00+07:00",
            "expiredAt": "2025-12-03T16:25:00+07:00"
        }
    }
}
```

---

### 18. Get Invoice Details

Get order details by invoice number.

**Endpoint:** `GET /v2/invoices`

**Query Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| invoiceNumber | string | Yes | Invoice number |

**Response Example:**

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

## Error Handling

### Common Error Codes

| Code | Description |
| --- | --- |
| VALIDATION_ERROR | Request validation failed |
| NOT_FOUND | Resource not found |
| UNAUTHORIZED | Invalid or missing authentication |
| RATE_LIMIT_EXCEEDED | Too many requests |
| ACCOUNT_NOT_FOUND | Game/service account not found |
| SKU_UNAVAILABLE | Product variant out of stock |
| PROMO_EXPIRED | Promo code has expired |
| INVALID_PAYMENT_METHOD | Payment method not allowed |
| TOKEN_EXPIRED | Validation token has expired |

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

function isError(response: ApiResponse<any>): response is { error: NonNullable<ApiResponse<any>['error']> } {
    return 'error' in response && response.error !== undefined;
}
```

---

## Changelog

### Version 2.0 (Current - December 2025)

**New Endpoints:**
- `GET /v2/regions` - Get supported regions
- `GET /v2/languages` - Get supported languages
- `GET /v2/contacts` - Get contact information

**Breaking Changes:**
- Changed base path from `/v1` to `/v2`
- Added `banner` field to products response
- Added `name` field to fields response (shown above label)
- Added `productCode` query parameter to `/v2/products`
- Simplified response structure (removed success, code, message wrapper)

**Improvements:**
- Complete production-ready data for all endpoints
- Full category list with descriptions and icons
- Comprehensive payment channels for Indonesia
- Enhanced SKU data with badges and sections
- Better field descriptions with hints

---

## Authentication & User Management

### 19. Register User

Register a new user account.

**Endpoint:** `POST /v2/auth/register`

**Request Body:**

```json
{
    "firstName": "John",
    "lastName": "Doe",
    "email": "john.doe@example.com",
    "phoneNumber": "628123456789",
    "password": "SecureP@ssw0rd",
    "confirmPassword": "SecureP@ssw0rd"
}
```

**Request Fields:**
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| firstName | string | Yes | User's first name (2-50 characters) |
| lastName | string | No | User's last name (2-50 characters) |
| email | string | Yes | Valid email address |
| phoneNumber | string | Yes | Phone number in E.164 format (e.g., +628123456789) |
| password | string | Yes | Password (min 8 chars, must include uppercase, lowercase, number, special char) |
| confirmPassword | string | Yes | Must match password |

**Response Example (Email Verification Required):**

```json
{
    "data": {
        "step": "EMAIL_VERIFICATION",
        "message": "Please verify your email to complete registration",
        "email": "john.doe@example.com",
        "user": {
            "id": "usr_1a2b3c4d5e6f",
            "firstName": "John",
            "lastName": "Doe",
            "email": "john.doe@example.com",
            "phoneNumber": "628123456789",
            "profilePicture": null,
            "status": "INACTIVE",
            "membership": {
                "level": "CLASSIC",
                "name": "Classic",
                "benefits": [
                    "Transaksi standar",
                    "Customer support 24/7",
                    "Bonus poin 1%"
                ]
            },
            "mfaStatus": "INACTIVE",
            "createdAt": "2025-12-03T10:00:00+07:00"
        }
    }
}
```

**Response Example (Success - Google Register):**

```json
{
    "data": {
        "step": "SUCCESS",
        "message": "Registration successful",
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
            "phoneNumber": "628123456789",
            "profilePicture": "https://lh3.googleusercontent.com/...",
            "status": "ACTIVE",
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
                "name": "Classic",
                "benefits": [
                    "Transaksi standar",
                    "Customer support 24/7",
                    "Bonus poin 1%"
                ]
            },
            "mfaStatus": "INACTIVE",
            "googleId": "117562748392847562",
            "createdAt": "2025-12-03T10:00:00+07:00"
        }
    }
}
```

**Error Response:**

```json
{
    "error": {
        "code": "VALIDATION_ERROR",
        "message": "Data tidak valid",
        "fields": {
            "email": "Email sudah terdaftar",
            "password": "Password harus minimal 8 karakter dengan kombinasi huruf besar, kecil, angka, dan simbol",
            "phoneNumber": "Format nomor telepon tidak valid (gunakan format: 628xxxxxxxxx)"
        }
    }
}
```

**Notes:**
- Regular registration requires email verification
- Google registration auto-verifies email (status: ACTIVE)
- Phone number must be in E.164 format (e.g., +628123456789, +60123456789)
- Password validation: min 8 chars, uppercase, lowercase, number, special char
- `step: EMAIL_VERIFICATION` means user must verify email before login
- `step: SUCCESS` means user can login immediately (Google register)

**Membership Levels:**
- **CLASSIC**: Default level for new users
- **PRESTIGE**: Premium level (requires qualification)
- **ROYAL**: VIP level (by invitation or high transaction volume)

---

### 20. Register with Google

Register using Google OAuth.

**Endpoint:** `POST /v2/auth/register/google`

**Request Body:**

```json
{
    "idToken": "eyJhbGciOiJSUzI1NiIsImtpZCI6IjZmODI4..."
}
```

**Request Fields:**
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| idToken | string | Yes | Google ID token from OAuth |

**Response Example:**

```json
{
    "data": {
        "step": "SUCCESS",
        "message": "Registration successful",
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
            "profilePicture": "https://lh3.googleusercontent.com/a/default-user=s96-c",
            "status": "ACTIVE",
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
                "name": "Classic",
                "benefits": [
                    "Transaksi standar",
                    "Customer support 24/7",
                    "Bonus poin 1%"
                ]
            },
            "mfaStatus": "INACTIVE",
            "googleId": "117562748392847562",
            "createdAt": "2025-12-03T10:00:00+07:00"
        }
    }
}
```

**Notes:**
- No email verification required (Google already verified)
- No MFA verification required for Google register
- No phone number required (can be added later in profile settings)
- User status is immediately ACTIVE
- Profile picture automatically imported from Google
- First name and last name extracted from Google profile

---

### 21. Verify Email

Verify email address after registration.

**Endpoint:** `POST /v2/auth/verify-email`

**Request Body:**

```json
{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Request Fields:**
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| token | string | Yes | Verification token from email link |

**Response Example:**

```json
{
    "data": {
        "step": "SUCCESS",
        "message": "Email verified successfully. You can now login.",
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
            "profilePicture": null,
            "status": "ACTIVE",
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
                "name": "Classic",
                "benefits": [
                    "Transaksi standar",
                    "Customer support 24/7",
                    "Bonus poin 1%"
                ]
            },
            "mfaStatus": "INACTIVE",
            "createdAt": "2025-12-03T10:00:00+07:00",
            "emailVerifiedAt": "2025-12-03T10:15:00+07:00"
        }
    }
}
```

**Error Response:**

```json
{
    "error": {
        "code": "INVALID_TOKEN",
        "message": "Token verifikasi tidak valid atau sudah kadaluarsa",
        "details": "Please request a new verification email"
    }
}
```

**Notes:**
- Verification token expires in 24 hours
- Token is embedded in the email link: `https://gate.co.id/verify-email/{token}`
- User clicks the link and frontend extracts token from URL, then calls this endpoint
- After verification, user status changes from INACTIVE to ACTIVE
- Returns access token for immediate login
- User can request new verification email if token expired

---

### 22. Resend Verification Email

Request new verification email.

**Endpoint:** `POST /v2/auth/resend-verification`

**Request Body:**

```json
{
    "email": "john.doe@example.com"
}
```

**Response Example:**

```json
{
    "data": {
        "message": "Verification email sent successfully",
        "email": "john.doe@example.com",
        "expiresIn": 86400
    }
}
```

**Error Response:**

```json
{
    "error": {
        "code": "EMAIL_ALREADY_VERIFIED",
        "message": "Email sudah terverifikasi",
        "details": "You can login with your account"
    }
}
```

---

### 23. Login

Login with email and password.

**Endpoint:** `POST /v2/auth/login`

**Request Body:**

```json
{
    "email": "john.doe@example.com",
    "password": "SecureP@ssw0rd"
}
```

**Request Fields:**
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| email | string | Yes | User's email address |
| password | string | Yes | User's password |

**Response Example (Success - No MFA):**

```json
{
    "data": {
        "step": "SUCCESS",
        "message": "Login successful",
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
            "currency": "IDR",
            "balance": {
                "IDR": 150000,
                "MYR": 0,
                "PHP": 0,
                "SGD": 0,
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
                ]
            },
            "mfaStatus": "INACTIVE",
            "createdAt": "2025-11-01T10:00:00+07:00",
            "lastLoginAt": "2025-12-03T10:30:00+07:00"
        }
    }
}
```

**Response Example (MFA Required):**

```json
{
    "data": {
        "step": "MFA_VERIFICATION",
        "message": "Please enter your MFA code",
        "mfaToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "expiresIn": 300
    }
}
```

**Error Response (Invalid Credentials):**

```json
{
    "error": {
        "code": "INVALID_CREDENTIALS",
        "message": "Email atau password salah",
        "details": "Please check your credentials and try again"
    }
}
```

**Error Response (Email Not Verified):**

```json
{
    "error": {
        "code": "EMAIL_NOT_VERIFIED",
        "message": "Email belum diverifikasi",
        "details": "Please verify your email before logging in",
        "action": {
            "type": "RESEND_VERIFICATION",
            "endpoint": "/v2/auth/resend-verification"
        }
    }
}
```

**Error Response (Account Suspended):**

```json
{
    "error": {
        "code": "ACCOUNT_SUSPENDED",
        "message": "Akun Anda telah dinonaktifkan",
        "details": "Please contact support for more information",
        "contact": {
            "email": "support@gate.co.id",
            "whatsapp": "https://wa.me/6281234567890"
        }
    }
}
```

**Notes:**
- If MFA is enabled, `step: MFA_VERIFICATION` is returned
- User must verify MFA code before getting access token
- `mfaToken` is valid for 5 minutes
- Account gets locked after 5 failed login attempts (30 minutes cooldown)

---

### 24. Login with Google

Login using Google OAuth.

**Endpoint:** `POST /v2/auth/login/google`

**Request Body:**

```json
{
    "idToken": "eyJhbGciOiJSUzI1NiIsImtpZCI6IjZmODI4..."
}
```

**Request Fields:**
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| idToken | string | Yes | Google ID token from OAuth |

**Response Example:**

```json
{
    "data": {
        "step": "SUCCESS",
        "message": "Login successful",
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
            "phoneNumber": "+628123456789",
            "profilePicture": "https://lh3.googleusercontent.com/a/default-user=s96-c",
            "status": "ACTIVE",
            "currency": "IDR",
            "balance": {
                "IDR": 150000,
                "MYR": 0,
                "PHP": 0,
                "SGD": 0,
                "THB": 0
            },
            "membership": {
                "level": "CLASSIC",
                "name": "Classic",
                "benefits": [
                    "Transaksi standar",
                    "Customer support 24/7",
                    "Bonus poin 1%"
                ]
            },
            "mfaStatus": "INACTIVE",
            "googleId": "117562748392847562",
            "createdAt": "2025-11-01T10:00:00+07:00",
            "lastLoginAt": "2025-12-03T10:30:00+07:00"
        }
    }
}
```

**Error Response:**

```json
{
    "error": {
        "code": "GOOGLE_AUTH_FAILED",
        "message": "Autentikasi Google gagal",
        "details": "Invalid or expired Google ID token"
    }
}
```

**Notes:**
- No MFA verification required for Google login
- Automatically updates profile picture from Google
- If Google account not registered, returns error suggesting registration

---

### 25. Verify MFA Code

Verify MFA code during login.

**Endpoint:** `POST /v2/auth/verify-mfa`

**Headers:**

```
Content-Type: application/json
```

**Request Body:**

```json
{
    "mfaToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "code": "123456"
}
```

**Request Fields:**
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| mfaToken | string | Yes | MFA token from login response |
| code | string | Yes | 6-digit MFA code from authenticator app |

**Response Example:**

```json
{
    "data": {
        "step": "SUCCESS",
        "message": "MFA verification successful",
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
            "phoneNumber": "628123456789",
            "profilePicture": "https://nos.jkt-1.neo.id/gate/profiles/user123.jpg",
            "status": "ACTIVE",
            "currency": "IDR",
            "balance": {
                "IDR": 150000,
                "MYR": 0,
                "PHP": 0,
                "SGD": 0,
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
                ]
            },
            "mfaStatus": "ACTIVE",
            "createdAt": "2025-11-01T10:00:00+07:00",
            "lastLoginAt": "2025-12-03T10:30:00+07:00"
        }
    }
}
```

**Error Response:**

```json
{
    "error": {
        "code": "INVALID_MFA_CODE",
        "message": "Kode MFA tidak valid",
        "details": "Please check your authenticator app and try again",
        "remainingAttempts": 2
    }
}
```

**Notes:**
- MFA code is valid for 30 seconds (TOTP standard)
- Maximum 3 attempts allowed
- After 3 failed attempts, user must login again
- `mfaToken` expires in 5 minutes

---

### 26. Forgot Password

Request password reset link.

**Endpoint:** `POST /v2/auth/forgot-password`

**Request Body:**

```json
{
    "email": "john.doe@example.com"
}
```

**Response Example:**

```json
{
    "data": {
        "message": "Password reset link sent to your email",
        "email": "john.doe@example.com",
        "expiresIn": 3600
    }
}
```

**Error Response:**

```json
{
    "error": {
        "code": "EMAIL_NOT_FOUND",
        "message": "Email tidak ditemukan",
        "details": "Please check your email or register a new account"
    }
}
```

**Email Content:**

Email will contain a link: `https://gate.co.id/reset-password/{token}`

**Notes:**
- Reset token expires in 1 hour
- Can only request once every 5 minutes (rate limited)
- Token is single-use only
- For security, always returns success even if email doesn't exist (but no email sent)

---

### 27. Reset Password

Reset password using token from email.

**Endpoint:** `POST /v2/auth/reset-password`

**Request Body:**

```json
{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "password": "NewSecureP@ssw0rd",
    "confirmPassword": "NewSecureP@ssw0rd"
}
```

**Request Fields:**
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| token | string | Yes | Reset token from email |
| password | string | Yes | New password (min 8 chars) |
| confirmPassword | string | Yes | Must match password |

**Response Example:**

```json
{
    "data": {
        "message": "Password reset successful",
        "email": "john.doe@example.com"
    }
}
```

**Error Response:**

```json
{
    "error": {
        "code": "INVALID_TOKEN",
        "message": "Token tidak valid atau sudah kadaluarsa",
        "details": "Please request a new password reset link"
    }
}
```

**Notes:**
- Token can only be used once
- After successful reset, all refresh tokens are invalidated
- User must login again with new password
- Password validation same as registration

---

### 28. Enable MFA

Enable Multi-Factor Authentication.

**Endpoint:** `POST /v2/auth/mfa/enable`

**Headers:**

```
Authorization: Bearer {access_token}
```

**Response Example:**

```json
{
    "data": {
        "qrCode": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA...",
        "secret": "JBSWY3DPEHPK3PXP",
        "backupCodes": [
            "12345678",
            "23456789",
            "34567890",
            "45678901",
            "56789012",
            "67890123",
            "78901234",
            "89012345"
        ],
        "message": "Scan QR code with your authenticator app and verify with a code"
    }
}
```

**Notes:**
- **Requires authentication** - User must be logged in
- Returns QR code and secret key for authenticator app setup
- Generates 8 backup codes for account recovery
- User must verify MFA code to complete activation
- Supports Google Authenticator, Authy, Microsoft Authenticator, etc.

---

### 29. Verify MFA Setup

Verify MFA setup after enabling.

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

**Response Example:**

```json
{
    "data": {
        "message": "MFA enabled successfully",
        "mfaStatus": "ACTIVE",
        "enabledAt": "2025-12-03T10:30:00+07:00"
    }
}
```

**Error Response:**

```json
{
    "error": {
        "code": "INVALID_MFA_CODE",
        "message": "Kode MFA tidak valid",
        "details": "Please check your authenticator app"
    }
}
```

**Notes:**
- **Requires authentication** - User must be logged in
- Code must be verified within setup session
- After verification, MFA is immediately active
- User will need MFA code for next login
- Backup codes are already generated in enable step

---

### 30. Disable MFA

Disable Multi-Factor Authentication.

**Endpoint:** `POST /v2/auth/mfa/disable`

**Headers:**

```
Authorization: Bearer {access_token}
```

**Request Body:**

```json
{
    "password": "CurrentP@ssw0rd",
    "code": "123456"
}
```

**Request Fields:**
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| password | string | Yes | Current account password |
| code | string | Yes | Current MFA code from authenticator |

**Response Example:**

```json
{
    "data": {
        "message": "MFA disabled successfully",
        "mfaStatus": "INACTIVE",
        "disabledAt": "2025-12-03T10:45:00+07:00"
    }
}
```

**Error Response:**

```json
{
    "error": {
        "code": "INVALID_CREDENTIALS",
        "message": "Password atau kode MFA salah",
        "details": "Please verify your password and MFA code"
    }
}
```

**Notes:**
- **Requires authentication** - User must be logged in
- Requires both password and valid MFA code for security
- All backup codes are invalidated after disabling
- Secret key is removed from system
- User can enable MFA again anytime (new secret will be generated)

---

### 31. Refresh Access Token

Get new access token using refresh token.

**Endpoint:** `POST /v2/auth/refresh`

**Request Body:**

```json
{
    "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response Example:**

```json
{
    "data": {
        "token": {
            "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
            "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
            "expiresIn": 3600,
            "tokenType": "Bearer"
        }
    }
}
```

**Error Response:**

```json
{
    "error": {
        "code": "INVALID_REFRESH_TOKEN",
        "message": "Refresh token tidak valid",
        "details": "Please login again"
    }
}
```

**Notes:**
- Access token expires in 1 hour
- Refresh token expires in 30 days
- New refresh token is issued with each refresh
- Old refresh token is invalidated after use

---

### 32. Logout

Logout and invalidate tokens.

**Endpoint:** `POST /v2/auth/logout`

**Headers:**

```
Authorization: Bearer {access_token}
```

**Request Body:**

```json
{
    "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response Example:**

```json
{
    "data": {
        "message": "Logout successful"
    }
}
```

**Notes:**
- Invalidates both access and refresh tokens
- User must login again to get new tokens
- All active sessions on other devices remain active

---

### 33. Get User Profile

Get current user profile information.

**Endpoint:** `GET /v2/user/profile`

**Headers:**

```
Authorization: Bearer {access_token}
```

**Response Example:**

```json
{
    "data": {
        "id": "usr_1a2b3c4d5e6f",
        "firstName": "John",
        "lastName": "Doe",
        "email": "john.doe@example.com",
        "phoneNumber": "628123456789",
        "profilePicture": "https://nos.jkt-1.neo.id/gate/profiles/user123.jpg",
        "status": "ACTIVE",
        "currency": "IDR",
        "balance": {
            "IDR": 150000,
            "MYR": 0,
            "PHP": 0,
            "SGD": 0,
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
                "nextLevel": "ROYAL"
            }
        },
        "mfaStatus": "ACTIVE",
        "googleId": null,
        "emailVerifiedAt": "2025-11-01T10:15:00+07:00",
        "createdAt": "2025-11-01T10:00:00+07:00",
        "lastLoginAt": "2025-12-03T10:30:00+07:00"
    }
}
```

**Notes:**
- Requires authentication
- Shows complete user information
- Includes membership progress to next level
- Balance shown in all supported currencies

---

### 34. Update User Profile

Update user profile information.

**Endpoint:** `PUT /v2/user/profile`

**Headers:**

```
Authorization: Bearer {access_token}
Content-Type: multipart/form-data
```

**Request Body (Form Data):**

```
firstName: John
lastName: Doe
phoneNumber: 628123456789
profilePicture: [file]
```

**Response Example:**

```json
{
    "data": {
        "id": "usr_1a2b3c4d5e6f",
        "firstName": "John",
        "lastName": "Doe",
        "email": "john.doe@example.com",
        "phoneNumber": "628123456789",
        "profilePicture": "https://nos.jkt-1.neo.id/gate/profiles/user123-updated.jpg",
        "status": "ACTIVE",
        "updatedAt": "2025-12-03T10:45:00+07:00"
    }
}
```

**Notes:**
- Email cannot be changed (security reason)
- Profile picture max size: 5MB
- Accepted formats: JPG, PNG, WebP
- Phone number must be unique

---

### 35. Change Password

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

**Response Example:**

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

**Notes:**
- Requires current password for verification
- New password must meet validation requirements
- All refresh tokens are invalidated after password change
- User must login again with new password

---

## Membership Levels

### Level Details

| Level | Name | Min Transaction | Benefits |
|-------|------|----------------|----------|
| CLASSIC | Classic | Rp 0 | - Transaksi standar<br>- Customer support 24/7<br>- Bonus poin 1% |
| PRESTIGE | Prestige | Rp 5,000,000 | - Diskon eksklusif hingga 5%<br>- Priority customer support<br>- Bonus poin 3%<br>- Akses promo premium |
| ROYAL | Royal | Rp 10,000,000 | - Diskon eksklusif hingga 10%<br>- Dedicated account manager<br>- Bonus poin 5%<br>- Akses promo VIP<br>- Transaksi prioritas |

**Notes:**
- Membership level automatically upgraded based on total transaction volume (last 90 days)
- Downgrade if transaction volume drops below threshold for 90 days
- Special benefits vary by level
- Royal members get early access to new products and exclusive promotions

---

## Authentication Flow Diagrams

### Regular Registration Flow

```
User fills form
    ↓
POST /v2/auth/register
    ↓
step: EMAIL_VERIFICATION
    ↓
User checks email
    ↓
POST /v2/auth/verify-email
    ↓
step: SUCCESS + token
    ↓
User logged in
```

### Google Registration Flow

```
User clicks "Sign up with Google"
    ↓
Google OAuth popup
    ↓
POST /v2/auth/register/google
    ↓
step: SUCCESS + token
    ↓
User logged in (no email verification needed)
```

### Regular Login Flow (No MFA)

```
User enters credentials
    ↓
POST /v2/auth/login
    ↓
step: SUCCESS + token
    ↓
User logged in
```

### Regular Login Flow (With MFA)

```
User enters credentials
    ↓
POST /v2/auth/login
    ↓
step: MFA_VERIFICATION + mfaToken
    ↓
User enters MFA code
    ↓
POST /v2/auth/verify-mfa
    ↓
step: SUCCESS + token
    ↓
User logged in
```

### Google Login Flow

```
User clicks "Sign in with Google"
    ↓
Google OAuth popup
    ↓
POST /v2/auth/login/google
    ↓
step: SUCCESS + token
    ↓
User logged in (no MFA needed)
```

### Forgot Password Flow

```
User clicks "Lupa Password"
    ↓
POST /v2/auth/forgot-password
    ↓
User checks email
    ↓
Clicks link: gate.co.id/reset-password/{token}
    ↓
User enters new password
    ↓
POST /v2/auth/reset-password
    ↓
Password changed
    ↓
User must login again
```

### Enable MFA Flow

```
User goes to Settings > Security
    ↓
Clicks "Enable MFA"
    ↓
POST /v2/auth/mfa/enable
    ↓
Shows QR code + backup codes
    ↓
User scans QR with authenticator app
    ↓
User enters code from app
    ↓
POST /v2/auth/mfa/verify-setup
    ↓
MFA enabled
    ↓
Next login requires MFA code
```

---

## Support

- **Email:** support@gate.co.id
- **WhatsApp:** Check `/v2/contacts` endpoint
- **Documentation:** https://docs.gate.co.id
- **Status Page:** https://status.gate.co.id

---

**Last Updated:** December 3, 2025
**API Version:** v2.0
**Document Version:** 2.0
