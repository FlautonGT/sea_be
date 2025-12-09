# Gate.co.id Admin API Documentation v2.0

> **Last Updated:** December 3, 2025 | **API Version:** v2.0

---

## Table of Contents

1. [Base Information](#base-information)
2. [Authentication](#authentication)
3. [Roles & Permissions](#roles--permissions)
4. [Environment Variables](#environment-variables)
5. [Admin Management](#admin-management)
6. [Provider Management](#provider-management)
7. [Payment Gateway Management](#payment-gateway-management)
8. [Product Management](#product-management)
9. [SKU Management](#sku-management)
10. [Transaction Management](#transaction-management)
11. [User Management](#user-management)
12. [Promo Management](#promo-management)
13. [Content Management](#content-management)
14. [Reports & Analytics](#reports--analytics)
15. [Audit Logs](#audit-logs)
16. [Settings](#settings)
17. [Region Management](#region-management)
18. [Language Management](#language-management)
19. [Category Management](#category-management)
20. [Section Management](#section-management)
21. [Payment Channel Management](#payment-channel-management)
22. [Deposit Management](#deposit-management)
23. [Invoice Management](#invoice-management)

---

## Base Information

### URLs

| Type | URL |
|------|-----|
| Admin API URL | `https://gateway.gate.id/admin/v2` |
| Alternative | `https://gateway.gate.co.id/admin/v2` |

### API Path Pattern

```
/admin/v2/{resource}
```

### Timezone & DateTime

- **Server Timezone:** Jakarta Time (UTC+07:00)
- **DateTime Format:** ISO 8601 with timezone offset
- **Example:** `2025-12-31T23:59:59+07:00`

---

## Authentication

### Admin Login

**Endpoint:** `POST /admin/v2/auth/login`

**Request Body:**

```json
{
    "email": "admin@gate.co.id",
    "password": "SecureAdminP@ss"
}
```

**Response (Success):**

```json
{
    "data": {
        "token": {
            "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
            "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
            "expiresIn": 3600,
            "tokenType": "Bearer"
        },
        "admin": {
            "id": "adm_1a2b3c4d5e6f",
            "name": "John Admin",
            "email": "admin@gate.co.id",
            "role": {
                "code": "SUPERADMIN",
                "name": "Super Administrator"
            },
            "status": "ACTIVE",
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

### Verify Admin MFA

**Endpoint:** `POST /admin/v2/auth/verify-mfa`

**Request Body:**

```json
{
    "mfaToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "code": "123456"
}
```

### Refresh Token

**Endpoint:** `POST /admin/v2/auth/refresh-token`

**Request Body:**

```json
{
    "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Logout

**Endpoint:** `POST /admin/v2/auth/logout`

**Headers:**

```
Authorization: Bearer {admin_access_token}
```

---

## Roles & Permissions

### Role Hierarchy

```
SUPERADMIN (Level 1 - Highest)
    ↓
ADMIN (Level 2)
    ↓
FINANCE (Level 3)
    ↓
CS_LEAD (Level 4)
    ↓
CS (Level 5 - Lowest)
```

### Role Definitions

| Role | Code | Level | Description |
|------|------|-------|-------------|
| Super Administrator | `SUPERADMIN` | 1 | Full system access, manage admins & permissions |
| Administrator | `ADMIN` | 2 | Manage products, SKUs, promos, content |
| Finance | `FINANCE` | 3 | View transactions, reports, manage deposits |
| CS Lead | `CS_LEAD` | 4 | Handle escalations, manage CS team |
| Customer Service | `CS` | 5 | View transactions, handle user issues |

### Permission Matrix

| Permission | Code | SUPERADMIN | ADMIN | FINANCE | CS_LEAD | CS |
|------------|------|------------|-------|---------|---------|-----|
| **Admin Management** |
| View Admins | `admin:read` | ✅ | ❌ | ❌ | ❌ | ❌ |
| Create Admin | `admin:create` | ✅ | ❌ | ❌ | ❌ | ❌ |
| Update Admin | `admin:update` | ✅ | ❌ | ❌ | ❌ | ❌ |
| Delete Admin | `admin:delete` | ✅ | ❌ | ❌ | ❌ | ❌ |
| Manage Roles | `role:manage` | ✅ | ❌ | ❌ | ❌ | ❌ |
| **Provider Management** |
| View Providers | `provider:read` | ✅ | ✅ | ❌ | ❌ | ❌ |
| Create Provider | `provider:create` | ✅ | ❌ | ❌ | ❌ | ❌ |
| Update Provider | `provider:update` | ✅ | ❌ | ❌ | ❌ | ❌ |
| Delete Provider | `provider:delete` | ✅ | ❌ | ❌ | ❌ | ❌ |
| **Payment Gateway** |
| View Gateways | `gateway:read` | ✅ | ✅ | ✅ | ❌ | ❌ |
| Create Gateway | `gateway:create` | ✅ | ❌ | ❌ | ❌ | ❌ |
| Update Gateway | `gateway:update` | ✅ | ❌ | ❌ | ❌ | ❌ |
| Delete Gateway | `gateway:delete` | ✅ | ❌ | ❌ | ❌ | ❌ |
| **Product Management** |
| View Products | `product:read` | ✅ | ✅ | ✅ | ✅ | ✅ |
| Create Product | `product:create` | ✅ | ✅ | ❌ | ❌ | ❌ |
| Update Product | `product:update` | ✅ | ✅ | ❌ | ❌ | ❌ |
| Delete Product | `product:delete` | ✅ | ✅ | ❌ | ❌ | ❌ |
| **SKU Management** |
| View SKUs | `sku:read` | ✅ | ✅ | ✅ | ✅ | ✅ |
| Create SKU | `sku:create` | ✅ | ✅ | ❌ | ❌ | ❌ |
| Update SKU | `sku:update` | ✅ | ✅ | ❌ | ❌ | ❌ |
| Delete SKU | `sku:delete` | ✅ | ✅ | ❌ | ❌ | ❌ |
| Sync SKU | `sku:sync` | ✅ | ✅ | ❌ | ❌ | ❌ |
| **Transaction Management** |
| View Transactions | `transaction:read` | ✅ | ✅ | ✅ | ✅ | ✅ |
| Update Transaction | `transaction:update` | ✅ | ✅ | ✅ | ✅ | ❌ |
| Refund Transaction | `transaction:refund` | ✅ | ✅ | ✅ | ❌ | ❌ |
| Manual Process | `transaction:manual` | ✅ | ✅ | ❌ | ❌ | ❌ |
| **User Management** |
| View Users | `user:read` | ✅ | ✅ | ✅ | ✅ | ✅ |
| Update User | `user:update` | ✅ | ✅ | ❌ | ✅ | ❌ |
| Suspend User | `user:suspend` | ✅ | ✅ | ❌ | ✅ | ❌ |
| Adjust Balance | `user:balance` | ✅ | ❌ | ✅ | ❌ | ❌ |
| **Promo Management** |
| View Promos | `promo:read` | ✅ | ✅ | ✅ | ✅ | ✅ |
| Create Promo | `promo:create` | ✅ | ✅ | ❌ | ❌ | ❌ |
| Update Promo | `promo:update` | ✅ | ✅ | ❌ | ❌ | ❌ |
| Delete Promo | `promo:delete` | ✅ | ✅ | ❌ | ❌ | ❌ |
| **Content Management** |
| View Content | `content:read` | ✅ | ✅ | ❌ | ❌ | ❌ |
| Manage Banners | `content:banner` | ✅ | ✅ | ❌ | ❌ | ❌ |
| Manage Popups | `content:popup` | ✅ | ✅ | ❌ | ❌ | ❌ |
| **Reports** |
| View Reports | `report:read` | ✅ | ✅ | ✅ | ✅ | ❌ |
| Export Reports | `report:export` | ✅ | ✅ | ✅ | ❌ | ❌ |
| **Audit Logs** |
| View Audit Logs | `audit:read` | ✅ | ❌ | ❌ | ❌ | ❌ |
| **Settings** |
| View Settings | `setting:read` | ✅ | ✅ | ❌ | ❌ | ❌ |
| Update Settings | `setting:update` | ✅ | ❌ | ❌ | ❌ | ❌ |
| **Region & Language** |
| View Regions | `setting:read` | ✅ | ✅ | ❌ | ❌ | ❌ |
| Manage Regions | `setting:update` | ✅ | ❌ | ❌ | ❌ | ❌ |
| View Languages | `setting:read` | ✅ | ✅ | ❌ | ❌ | ❌ |
| Manage Languages | `setting:update` | ✅ | ❌ | ❌ | ❌ | ❌ |
| **Category & Section** |
| View Categories | `product:read` | ✅ | ✅ | ✅ | ✅ | ✅ |
| Manage Categories | `product:create` | ✅ | ✅ | ❌ | ❌ | ❌ |
| View Sections | `product:read` | ✅ | ✅ | ✅ | ✅ | ✅ |
| Manage Sections | `product:create` | ✅ | ✅ | ❌ | ❌ | ❌ |
| **Payment Channel** |
| View Channels | `gateway:read` | ✅ | ✅ | ✅ | ❌ | ❌ |
| Create Channel | `gateway:create` | ✅ | ❌ | ❌ | ❌ | ❌ |
| Update Channel | `gateway:update` | ✅ | ✅ | ❌ | ❌ | ❌ |
| Delete Channel | `gateway:delete` | ✅ | ❌ | ❌ | ❌ | ❌ |
| **Deposit Management** |
| View Deposits | `transaction:read` | ✅ | ✅ | ✅ | ✅ | ✅ |
| Confirm Deposit | `transaction:manual` | ✅ | ✅ | ❌ | ❌ | ❌ |
| Cancel Deposit | `transaction:update` | ✅ | ✅ | ✅ | ✅ | ❌ |
| Refund Deposit | `transaction:refund` | ✅ | ✅ | ✅ | ❌ | ❌ |

### Permission Check in Response

Every admin endpoint response includes permission info:

```json
{
    "data": { ... },
    "_meta": {
        "requiredPermission": "product:create",
        "adminId": "adm_1a2b3c4d5e6f",
        "adminRole": "ADMIN"
    }
}
```

### Permission Denied Response

```json
{
    "error": {
        "code": "PERMISSION_DENIED",
        "message": "Anda tidak memiliki akses untuk melakukan aksi ini",
        "details": "Required permission: admin:create"
    }
}
```

---

## Environment Variables

### Provider API Credentials

API credentials are stored in `.env` file (NOT in database) for security:

```env
# ============================================
# PRODUCT PROVIDERS
# ============================================

# Digiflazz
DIGIFLAZZ_USERNAME=your_username
DIGIFLAZZ_API_KEY=your_api_key
DIGIFLAZZ_WEBHOOK_SECRET=your_webhook_secret

# VIP Reseller
VIPRESELLER_API_ID=your_api_id
VIPRESELLER_API_KEY=your_api_key

# BangJeff
BANGJEFF_MEMBER_ID=your_member_id
BANGJEFF_SECRET_KEY=your_secret_key
BANGJEFF_WEBHOOK_TOKEN=your_webhook_token

# ============================================
# PAYMENT GATEWAYS
# ============================================

# LinkQu (QRIS)
LINKQU_CLIENT_ID=your_client_id
LINKQU_CLIENT_SECRET=your_client_secret
LINKQU_USERNAME=your_username
LINKQU_PIN=your_pin

# BCA API (Virtual Account)
BCA_CLIENT_ID=your_client_id
BCA_CLIENT_SECRET=your_client_secret
BCA_API_KEY=your_api_key
BCA_API_SECRET=your_api_secret
BCA_CORPORATE_ID=your_corporate_id

# BRI API (Virtual Account)
BRI_CLIENT_ID=your_client_id
BRI_CLIENT_SECRET=your_client_secret
BRI_INSTITUTION_CODE=your_institution_code

# Xendit (Permata VA, etc)
XENDIT_SECRET_KEY=your_secret_key
XENDIT_CALLBACK_TOKEN=your_callback_token

# DANA
DANA_CLIENT_ID=your_client_id
DANA_CLIENT_SECRET=your_client_secret
DANA_PRIVATE_KEY_PATH=/path/to/private_key.pem

# OVO
OVO_APP_ID=your_app_id
OVO_API_KEY=your_api_key

# GoPay (via Midtrans)
MIDTRANS_SERVER_KEY=your_server_key
MIDTRANS_CLIENT_KEY=your_client_key
MIDTRANS_IS_PRODUCTION=true

# ShopeePay
SHOPEEPAY_CLIENT_ID=your_client_id
SHOPEEPAY_CLIENT_SECRET=your_client_secret
SHOPEEPAY_MERCHANT_ID=your_merchant_id
```

> **⚠️ IMPORTANT:** Never store API credentials in database. Always use environment variables.

---

## Admin Management

### 1. Get All Admins

**Endpoint:** `GET /admin/v2/admins`

**Permission Required:** `admin:read`

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| limit | integer | No | Items per page. Default: 10 |
| page | integer | No | Page number. Default: 1 |
| search | string | No | Search by name or email |
| role | string | No | Filter by role code |
| status | string | No | Filter: ACTIVE, INACTIVE, SUSPENDED |

**Response:**

```json
{
    "data": {
        "admins": [
            {
                "id": "adm_1a2b3c4d5e6f",
                "name": "John Superadmin",
                "email": "superadmin@gate.co.id",
                "role": {
                    "code": "SUPERADMIN",
                    "name": "Super Administrator",
                    "level": 1
                },
                "status": "ACTIVE",
                "mfaEnabled": true,
                "createdAt": "2025-01-01T00:00:00+07:00",
                "lastLoginAt": "2025-12-03T10:30:00+07:00"
            },
            {
                "id": "adm_2b3c4d5e6f7g",
                "name": "Jane Admin",
                "email": "admin@gate.co.id",
                "role": {
                    "code": "ADMIN",
                    "name": "Administrator",
                    "level": 2
                },
                "status": "ACTIVE",
                "mfaEnabled": true,
                "createdAt": "2025-06-01T00:00:00+07:00",
                "lastLoginAt": "2025-12-03T09:15:00+07:00"
            },
            {
                "id": "adm_3c4d5e6f7g8h",
                "name": "Bob CS",
                "email": "cs@gate.co.id",
                "role": {
                    "code": "CS",
                    "name": "Customer Service",
                    "level": 5
                },
                "status": "ACTIVE",
                "mfaEnabled": false,
                "createdAt": "2025-09-01T00:00:00+07:00",
                "lastLoginAt": "2025-12-03T08:00:00+07:00"
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

### 2. Get Admin Detail

**Endpoint:** `GET /admin/v2/admins/{adminId}`

**Permission Required:** `admin:read`

**Response:**

```json
{
    "data": {
        "id": "adm_1a2b3c4d5e6f",
        "name": "John Superadmin",
        "email": "superadmin@gate.co.id",
        "phoneNumber": "+628123456789",
        "role": {
            "code": "SUPERADMIN",
            "name": "Super Administrator",
            "level": 1,
            "permissions": [
                "admin:read", "admin:create", "admin:update", "admin:delete",
                "role:manage", "provider:read", "provider:create", "..."
            ]
        },
        "status": "ACTIVE",
        "mfaEnabled": true,
        "createdAt": "2025-01-01T00:00:00+07:00",
        "updatedAt": "2025-12-01T00:00:00+07:00",
        "lastLoginAt": "2025-12-03T10:30:00+07:00",
        "createdBy": null
    }
}
```

---

### 3. Create Admin

**Endpoint:** `POST /admin/v2/admins`

**Permission Required:** `admin:create`

**Request Body:**

```json
{
    "name": "New Admin",
    "email": "newadmin@gate.co.id",
    "phoneNumber": "+628123456789",
    "password": "SecureP@ssw0rd",
    "roleCode": "ADMIN",
    "status": "ACTIVE"
}
```

**Response:**

```json
{
    "data": {
        "id": "adm_4d5e6f7g8h9i",
        "name": "New Admin",
        "email": "newadmin@gate.co.id",
        "phoneNumber": "+628123456789",
        "role": {
            "code": "ADMIN",
            "name": "Administrator",
            "level": 2
        },
        "status": "ACTIVE",
        "mfaEnabled": false,
        "createdAt": "2025-12-03T11:00:00+07:00",
        "createdBy": {
            "id": "adm_1a2b3c4d5e6f",
            "name": "John Superadmin"
        }
    }
}
```

> **Note:** Only SUPERADMIN can create admins with role level <= their own level.

---

### 4. Update Admin

**Endpoint:** `PUT /admin/v2/admins/{adminId}`

**Permission Required:** `admin:update`

**Request Body:**

```json
{
    "name": "Updated Admin Name",
    "phoneNumber": "+628987654321",
    "roleCode": "CS_LEAD",
    "status": "ACTIVE"
}
```

---

### 5. Delete Admin

**Endpoint:** `DELETE /admin/v2/admins/{adminId}`

**Permission Required:** `admin:delete`

**Response:**

```json
{
    "data": {
        "message": "Admin deleted successfully"
    }
}
```

> **Note:** Cannot delete own account or admins with higher role level.

---

### 6. Get All Roles

**Endpoint:** `GET /admin/v2/roles`

**Permission Required:** `role:manage`

**Response:**

```json
{
    "data": [
        {
            "code": "SUPERADMIN",
            "name": "Super Administrator",
            "level": 1,
            "description": "Full system access",
            "permissions": ["admin:read", "admin:create", "..."],
            "adminCount": 2
        },
        {
            "code": "ADMIN",
            "name": "Administrator",
            "level": 2,
            "description": "Manage products, SKUs, promos, content",
            "permissions": ["product:read", "product:create", "..."],
            "adminCount": 5
        },
        {
            "code": "FINANCE",
            "name": "Finance",
            "level": 3,
            "description": "View transactions, reports, manage deposits",
            "permissions": ["transaction:read", "report:read", "..."],
            "adminCount": 3
        },
        {
            "code": "CS_LEAD",
            "name": "CS Lead",
            "level": 4,
            "description": "Handle escalations, manage CS team",
            "permissions": ["transaction:read", "user:read", "..."],
            "adminCount": 4
        },
        {
            "code": "CS",
            "name": "Customer Service",
            "level": 5,
            "description": "View transactions, handle user issues",
            "permissions": ["transaction:read", "user:read"],
            "adminCount": 10
        }
    ]
}
```

---

### 7. Update Role Permissions

**Endpoint:** `PUT /admin/v2/roles/{roleCode}/permissions`

**Permission Required:** `role:manage`

**Request Body:**

```json
{
    "permissions": [
        "product:read",
        "product:create",
        "product:update",
        "sku:read",
        "sku:create",
        "transaction:read"
    ]
}
```

**Response:**

```json
{
    "data": {
        "code": "ADMIN",
        "name": "Administrator",
        "permissions": [
            "product:read",
            "product:create",
            "product:update",
            "sku:read",
            "sku:create",
            "transaction:read"
        ],
        "updatedAt": "2025-12-03T11:30:00+07:00",
        "updatedBy": {
            "id": "adm_1a2b3c4d5e6f",
            "name": "John Superadmin"
        }
    }
}
```

> **Note:** Cannot modify SUPERADMIN permissions.

---

## Provider Management

Providers are external vendors that supply products/SKUs (Digiflazz, VIP Reseller, BangJeff, etc).

### Provider Configuration (Stored in Database)

| Field | Storage | Description |
|-------|---------|-------------|
| `code` | Database | Provider identifier |
| `name` | Database | Display name |
| `baseUrl` | Database | API base URL |
| `isActive` | Database | Enable/disable provider |
| `priority` | Database | Fallback order |
| `webhookUrl` | Database | Callback URL |
| `apiCredentials` | `.env` | API keys, secrets (NOT in DB) |

### 8. Get All Providers

**Endpoint:** `GET /admin/v2/providers`

**Permission Required:** `provider:read`

**Response:**

```json
{
    "data": [
        {
            "id": "prv_1a2b3c4d",
            "code": "DIGIFLAZZ",
            "name": "Digiflazz",
            "baseUrl": "https://api.digiflazz.com/v1",
            "webhookUrl": "https://gateway.gate.id/webhooks/digiflazz",
            "isActive": true,
            "priority": 1,
            "supportedTypes": ["PULSA", "DATA", "GAME", "EWALLET", "PLN"],
            "healthStatus": "HEALTHY",
            "lastHealthCheck": "2025-12-03T11:00:00+07:00",
            "stats": {
                "totalSkus": 1250,
                "activeSkus": 1180,
                "successRate": 98.5,
                "avgResponseTime": 1200
            },
            "createdAt": "2025-01-01T00:00:00+07:00",
            "updatedAt": "2025-12-01T00:00:00+07:00"
        },
        {
            "id": "prv_2b3c4d5e",
            "code": "VIPRESELLER",
            "name": "VIP Reseller",
            "baseUrl": "https://vip-reseller.co.id/api",
            "webhookUrl": "https://gateway.gate.id/webhooks/vipreseller",
            "isActive": true,
            "priority": 2,
            "supportedTypes": ["GAME", "VOUCHER"],
            "healthStatus": "HEALTHY",
            "lastHealthCheck": "2025-12-03T11:00:00+07:00",
            "stats": {
                "totalSkus": 450,
                "activeSkus": 420,
                "successRate": 97.8,
                "avgResponseTime": 1500
            },
            "createdAt": "2025-03-01T00:00:00+07:00",
            "updatedAt": "2025-11-15T00:00:00+07:00"
        },
        {
            "id": "prv_3c4d5e6f",
            "code": "BANGJEFF",
            "name": "BangJeff",
            "baseUrl": "https://api.bangjeff.com",
            "webhookUrl": "https://gateway.gate.id/webhooks/bangjeff",
            "isActive": true,
            "priority": 3,
            "supportedTypes": ["GAME", "STREAMING"],
            "healthStatus": "DEGRADED",
            "lastHealthCheck": "2025-12-03T11:00:00+07:00",
            "stats": {
                "totalSkus": 320,
                "activeSkus": 280,
                "successRate": 95.2,
                "avgResponseTime": 2100
            },
            "createdAt": "2025-06-01T00:00:00+07:00",
            "updatedAt": "2025-12-02T00:00:00+07:00"
        }
    ]
}
```

---

### 9. Get Provider Detail

**Endpoint:** `GET /admin/v2/providers/{providerId}`

**Permission Required:** `provider:read`

**Response:**

```json
{
    "data": {
        "id": "prv_1a2b3c4d",
        "code": "DIGIFLAZZ",
        "name": "Digiflazz",
        "baseUrl": "https://api.digiflazz.com/v1",
        "webhookUrl": "https://gateway.gate.id/webhooks/digiflazz",
        "isActive": true,
        "priority": 1,
        "supportedTypes": ["PULSA", "DATA", "GAME", "EWALLET", "PLN"],
        "healthStatus": "HEALTHY",
        "lastHealthCheck": "2025-12-03T11:00:00+07:00",
        "apiConfig": {
            "timeout": 30000,
            "retryAttempts": 3,
            "retryDelay": 1000
        },
        "mapping": {
            "statusSuccess": ["Sukses", "success"],
            "statusPending": ["Pending", "pending"],
            "statusFailed": ["Gagal", "failed"]
        },
        "stats": {
            "totalSkus": 1250,
            "activeSkus": 1180,
            "successRate": 98.5,
            "avgResponseTime": 1200,
            "todayTransactions": 5420,
            "todaySuccessRate": 98.8
        },
        "credentials": {
            "hasUsername": true,
            "hasApiKey": true,
            "hasWebhookSecret": true
        },
        "createdAt": "2025-01-01T00:00:00+07:00",
        "updatedAt": "2025-12-01T00:00:00+07:00"
    }
}
```

> **Note:** `credentials` only shows whether credentials exist (not the actual values).

---

### 10. Create Provider

**Endpoint:** `POST /admin/v2/providers`

**Permission Required:** `provider:create`

**Request Body:**

```json
{
    "code": "NEWPROVIDER",
    "name": "New Provider",
    "baseUrl": "https://api.newprovider.com/v1",
    "webhookUrl": "https://gateway.gate.id/webhooks/newprovider",
    "isActive": false,
    "priority": 4,
    "supportedTypes": ["GAME", "VOUCHER"],
    "apiConfig": {
        "timeout": 30000,
        "retryAttempts": 3,
        "retryDelay": 1000
    },
    "mapping": {
        "statusSuccess": ["Sukses", "success", "1"],
        "statusPending": ["Pending", "pending", "0"],
        "statusFailed": ["Gagal", "failed", "-1"]
    },
    "envCredentialKeys": {
        "username": "NEWPROVIDER_USERNAME",
        "apiKey": "NEWPROVIDER_API_KEY",
        "secretKey": "NEWPROVIDER_SECRET_KEY"
    }
}
```

**Response:**

```json
{
    "data": {
        "id": "prv_4d5e6f7g",
        "code": "NEWPROVIDER",
        "name": "New Provider",
        "message": "Provider created. Please add credentials to .env file",
        "requiredEnvVars": [
            "NEWPROVIDER_USERNAME",
            "NEWPROVIDER_API_KEY",
            "NEWPROVIDER_SECRET_KEY"
        ]
    }
}
```

---

### 11. Update Provider

**Endpoint:** `PUT /admin/v2/providers/{providerId}`

**Permission Required:** `provider:update`

**Request Body:**

```json
{
    "name": "Updated Provider Name",
    "baseUrl": "https://api.newprovider.com/v2",
    "isActive": true,
    "priority": 2,
    "apiConfig": {
        "timeout": 45000,
        "retryAttempts": 5,
        "retryDelay": 2000
    }
}
```

---

### 12. Delete Provider

**Endpoint:** `DELETE /admin/v2/providers/{providerId}`

**Permission Required:** `provider:delete`

> **Warning:** Cannot delete provider with active SKUs. Deactivate or reassign SKUs first.

---

### 13. Test Provider Connection

**Endpoint:** `POST /admin/v2/providers/{providerId}/test`

**Permission Required:** `provider:update`

**Response:**

```json
{
    "data": {
        "status": "SUCCESS",
        "responseTime": 850,
        "balance": 15000000,
        "message": "Connection successful"
    }
}
```

---

### 14. Sync Provider SKUs

**Endpoint:** `POST /admin/v2/providers/{providerId}/sync`

**Permission Required:** `sku:sync`

**Response:**

```json
{
    "data": {
        "status": "COMPLETED",
        "summary": {
            "totalFromProvider": 1300,
            "newSkus": 25,
            "updatedSkus": 150,
            "deactivatedSkus": 10,
            "unchanged": 1115
        },
        "syncedAt": "2025-12-03T11:30:00+07:00"
    }
}
```

---

## Payment Gateway Management

Payment gateways are providers that process payments (LinkQu, BCA, BRI, Xendit, etc).

### Gateway Configuration (Stored in Database)

| Field | Storage | Description |
|-------|---------|-------------|
| `code` | Database | Gateway identifier |
| `name` | Database | Display name |
| `baseUrl` | Database | API base URL |
| `isActive` | Database | Enable/disable gateway |
| `supportedMethods` | Database | Payment methods supported |
| `supportedTypes` | Database | `purchase`, `deposit`, or both |
| `feeConfig` | Database | Fee calculation settings |
| `apiCredentials` | `.env` | API keys, secrets (NOT in DB) |

### 15. Get All Payment Gateways

**Endpoint:** `GET /admin/v2/payment-gateways`

**Permission Required:** `gateway:read`

**Response:**

```json
{
    "data": [
        {
            "id": "gw_1a2b3c4d",
            "code": "LINKQU",
            "name": "LinkQu",
            "baseUrl": "https://api.linkqu.id",
            "isActive": true,
            "supportedMethods": ["QRIS"],
            "supportedTypes": ["purchase", "deposit"],
            "healthStatus": "HEALTHY",
            "lastHealthCheck": "2025-12-03T11:00:00+07:00",
            "stats": {
                "todayTransactions": 2500,
                "todayVolume": 125000000,
                "successRate": 99.2
            }
        },
        {
            "id": "gw_2b3c4d5e",
            "code": "BCA_DIRECT",
            "name": "BCA Direct API",
            "baseUrl": "https://sandbox.bca.co.id",
            "isActive": true,
            "supportedMethods": ["BCA_VA"],
            "supportedTypes": ["purchase", "deposit"],
            "healthStatus": "HEALTHY",
            "lastHealthCheck": "2025-12-03T11:00:00+07:00",
            "stats": {
                "todayTransactions": 1800,
                "todayVolume": 250000000,
                "successRate": 99.5
            }
        },
        {
            "id": "gw_3c4d5e6f",
            "code": "BRI_DIRECT",
            "name": "BRI Direct API",
            "baseUrl": "https://sandbox.bri.co.id",
            "isActive": true,
            "supportedMethods": ["BRI_VA"],
            "supportedTypes": ["purchase", "deposit"],
            "healthStatus": "HEALTHY",
            "lastHealthCheck": "2025-12-03T11:00:00+07:00",
            "stats": {
                "todayTransactions": 1200,
                "todayVolume": 180000000,
                "successRate": 99.1
            }
        },
        {
            "id": "gw_4d5e6f7g",
            "code": "XENDIT",
            "name": "Xendit",
            "baseUrl": "https://api.xendit.co",
            "isActive": true,
            "supportedMethods": ["PERMATA_VA", "MANDIRI_VA", "CARD"],
            "supportedTypes": ["purchase", "deposit"],
            "healthStatus": "HEALTHY",
            "lastHealthCheck": "2025-12-03T11:00:00+07:00",
            "stats": {
                "todayTransactions": 950,
                "todayVolume": 85000000,
                "successRate": 98.8
            }
        },
        {
            "id": "gw_5e6f7g8h",
            "code": "MIDTRANS",
            "name": "Midtrans",
            "baseUrl": "https://api.midtrans.com",
            "isActive": true,
            "supportedMethods": ["GOPAY", "SHOPEEPAY"],
            "supportedTypes": ["purchase"],
            "healthStatus": "HEALTHY",
            "lastHealthCheck": "2025-12-03T11:00:00+07:00",
            "stats": {
                "todayTransactions": 3200,
                "todayVolume": 95000000,
                "successRate": 98.5
            }
        },
        {
            "id": "gw_6f7g8h9i",
            "code": "DANA_DIRECT",
            "name": "DANA Direct",
            "baseUrl": "https://api.dana.id",
            "isActive": true,
            "supportedMethods": ["DANA"],
            "supportedTypes": ["purchase"],
            "healthStatus": "HEALTHY",
            "lastHealthCheck": "2025-12-03T11:00:00+07:00",
            "stats": {
                "todayTransactions": 2100,
                "todayVolume": 65000000,
                "successRate": 99.0
            }
        }
    ]
}
```

---

### 16. Get Payment Gateway Detail

**Endpoint:** `GET /admin/v2/payment-gateways/{gatewayId}`

**Permission Required:** `gateway:read`

**Response:**

```json
{
    "data": {
        "id": "gw_1a2b3c4d",
        "code": "LINKQU",
        "name": "LinkQu",
        "baseUrl": "https://api.linkqu.id",
        "callbackUrl": "https://gateway.gate.id/callbacks/linkqu",
        "isActive": true,
        "supportedMethods": ["QRIS"],
        "supportedTypes": ["purchase", "deposit"],
        "healthStatus": "HEALTHY",
        "lastHealthCheck": "2025-12-03T11:00:00+07:00",
        "apiConfig": {
            "timeout": 30000,
            "retryAttempts": 3
        },
        "feeConfig": {
            "QRIS": {
                "feeType": "PERCENTAGE",
                "feeAmount": 0,
                "feePercentage": 0.7,
                "minFee": 0,
                "maxFee": 0
            }
        },
        "mapping": {
            "statusSuccess": ["00", "success"],
            "statusPending": ["01", "pending"],
            "statusFailed": ["02", "failed"]
        },
        "credentials": {
            "hasClientId": true,
            "hasClientSecret": true,
            "hasUsername": true,
            "hasPin": true
        },
        "stats": {
            "todayTransactions": 2500,
            "todayVolume": 125000000,
            "successRate": 99.2,
            "avgResponseTime": 950
        },
        "createdAt": "2025-01-01T00:00:00+07:00",
        "updatedAt": "2025-12-01T00:00:00+07:00"
    }
}
```

---

### 17. Create Payment Gateway

**Endpoint:** `POST /admin/v2/payment-gateways`

**Permission Required:** `gateway:create`

**Request Body:**

```json
{
    "code": "NEWGATEWAY",
    "name": "New Payment Gateway",
    "baseUrl": "https://api.newgateway.com",
    "callbackUrl": "https://gateway.gate.id/callbacks/newgateway",
    "isActive": false,
    "supportedMethods": ["NEWGATEWAY_VA"],
    "supportedTypes": ["purchase", "deposit"],
    "apiConfig": {
        "timeout": 30000,
        "retryAttempts": 3
    },
    "feeConfig": {
        "NEWGATEWAY_VA": {
            "feeType": "FIXED",
            "feeAmount": 4000,
            "feePercentage": 0,
            "minFee": 4000,
            "maxFee": 4000
        }
    },
    "envCredentialKeys": {
        "clientId": "NEWGATEWAY_CLIENT_ID",
        "clientSecret": "NEWGATEWAY_CLIENT_SECRET",
        "apiKey": "NEWGATEWAY_API_KEY"
    }
}
```

---

### 18. Update Payment Gateway

**Endpoint:** `PUT /admin/v2/payment-gateways/{gatewayId}`

**Permission Required:** `gateway:update`

---

### 19. Delete Payment Gateway

**Endpoint:** `DELETE /admin/v2/payment-gateways/{gatewayId}`

**Permission Required:** `gateway:delete`

> **Warning:** Cannot delete gateway with active payment channels.

---

### 20. Test Gateway Connection

**Endpoint:** `POST /admin/v2/payment-gateways/{gatewayId}/test`

**Permission Required:** `gateway:update`

**Response:**

```json
{
    "data": {
        "status": "SUCCESS",
        "responseTime": 650,
        "balance": 50000000,
        "message": "Connection successful"
    }
}
```

---

### 21. Get Payment Channel Assignments

**Endpoint:** `GET /admin/v2/payment-channels/assignments`

**Permission Required:** `gateway:read`

**Response:**

```json
{
    "data": [
        {
            "paymentCode": "QRIS",
            "paymentName": "QRIS",
            "assignments": {
                "purchase": {
                    "gatewayCode": "LINKQU",
                    "gatewayName": "LinkQu",
                    "isActive": true
                },
                "deposit": {
                    "gatewayCode": "LINKQU",
                    "gatewayName": "LinkQu",
                    "isActive": true
                }
            }
        },
        {
            "paymentCode": "BCA_VA",
            "paymentName": "BCA Virtual Account",
            "assignments": {
                "purchase": {
                    "gatewayCode": "BCA_DIRECT",
                    "gatewayName": "BCA Direct API",
                    "isActive": true
                },
                "deposit": {
                    "gatewayCode": "BCA_DIRECT",
                    "gatewayName": "BCA Direct API",
                    "isActive": true
                }
            }
        },
        {
            "paymentCode": "PERMATA_VA",
            "paymentName": "Permata Virtual Account",
            "assignments": {
                "purchase": {
                    "gatewayCode": "XENDIT",
                    "gatewayName": "Xendit",
                    "isActive": true
                },
                "deposit": {
                    "gatewayCode": "XENDIT",
                    "gatewayName": "Xendit",
                    "isActive": true
                }
            }
        },
        {
            "paymentCode": "GOPAY",
            "paymentName": "GoPay",
            "assignments": {
                "purchase": {
                    "gatewayCode": "MIDTRANS",
                    "gatewayName": "Midtrans",
                    "isActive": true
                },
                "deposit": {
                    "gatewayCode": null,
                    "gatewayName": null,
                    "isActive": false
                }
            }
        },
        {
            "paymentCode": "DANA",
            "paymentName": "DANA",
            "assignments": {
                "purchase": {
                    "gatewayCode": "DANA_DIRECT",
                    "gatewayName": "DANA Direct",
                    "isActive": true
                },
                "deposit": {
                    "gatewayCode": null,
                    "gatewayName": null,
                    "isActive": false
                }
            }
        }
    ]
}
```

---

### 22. Update Payment Channel Assignment

**Endpoint:** `PUT /admin/v2/payment-channels/{paymentCode}/assignment`

**Permission Required:** `gateway:update`

**Request Body:**

```json
{
    "purchase": {
        "gatewayCode": "XENDIT",
        "isActive": true
    },
    "deposit": {
        "gatewayCode": "XENDIT",
        "isActive": true
    }
}
```

**Response:**

```json
{
    "data": {
        "paymentCode": "PERMATA_VA",
        "paymentName": "Permata Virtual Account",
        "assignments": {
            "purchase": {
                "gatewayCode": "XENDIT",
                "gatewayName": "Xendit",
                "isActive": true
            },
            "deposit": {
                "gatewayCode": "XENDIT",
                "gatewayName": "Xendit",
                "isActive": true
            }
        },
        "updatedAt": "2025-12-03T12:00:00+07:00",
        "updatedBy": {
            "id": "adm_1a2b3c4d5e6f",
            "name": "John Superadmin"
        }
    }
}
```

> **Note:** This allows easy switching of payment gateway vendors without code changes.

---

## Product Management

### 23. Get All Products

**Endpoint:** `GET /admin/v2/products`

**Permission Required:** `product:read`

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| limit | integer | No | Items per page. Default: 10 |
| page | integer | No | Page number. Default: 1 |
| search | string | No | Search by name or code |
| categoryCode | string | No | Filter by category |
| region | string | No | Filter by region |
| isActive | boolean | No | Filter by status |

**Response:**

```json
{
    "data": {
        "products": [
            {
                "id": "prd_1a2b3c4d",
                "code": "MLBB",
                "slug": "mobile-legends",
                "title": "Mobile Legends: Bang Bang",
                "subtitle": "Moonton",
                "publisher": "Moonton",
                "thumbnail": "https://nos.jkt-1.neo.id/gate/products/mlbb-icon.webp",
                "category": {
                    "code": "top-up-game",
                    "title": "Top Up Game"
                },
                "isActive": true,
                "isPopular": true,
                "regions": ["ID", "MY", "PH", "SG", "TH"],
                "skuCount": 15,
                "stats": {
                    "todayTransactions": 1250,
                    "todayRevenue": 125000000
                },
                "createdAt": "2025-01-01T00:00:00+07:00",
                "updatedAt": "2025-12-01T00:00:00+07:00"
            }
        ],
        "pagination": {
            "limit": 10,
            "page": 1,
            "totalRows": 50,
            "totalPages": 5
        }
    }
}
```

---

### 24. Get Product Detail

**Endpoint:** `GET /admin/v2/products/{productId}`

**Permission Required:** `product:read`

---

### 25. Create Product

**Endpoint:** `POST /admin/v2/products`

**Permission Required:** `product:create`

**Request Body (multipart/form-data):**

```json
{
    "code": "NEWGAME",
    "slug": "new-game",
    "title": "New Game",
    "subtitle": "Publisher",
    "description": "Product description",
    "publisher": "Game Publisher",
    "categoryCode": "top-up-game",
    "isActive": true,
    "isPopular": false,
    "regions": ["ID", "MY"],
    "thumbnail": "[FILE]",
    "banner": "[FILE]",
    "features": ["⚡ Proses Instan", "🔒 Aman"],
    "howToOrder": ["Step 1", "Step 2"],
    "tags": ["RPG", "Adventure"]
}
```

---

### 26. Update Product

**Endpoint:** `PUT /admin/v2/products/{productId}`

**Permission Required:** `product:update`

---

### 27. Delete Product

**Endpoint:** `DELETE /admin/v2/products/{productId}`

**Permission Required:** `product:delete`

> **Warning:** Cannot delete product with active SKUs or transactions.

---

### 28. Get Product Fields

**Endpoint:** `GET /admin/v2/products/{productId}/fields`

**Permission Required:** `product:read`

---

### 29. Update Product Fields

**Endpoint:** `PUT /admin/v2/products/{productId}/fields`

**Permission Required:** `product:update`

**Request Body:**

```json
{
    "fields": [
        {
            "name": "User ID",
            "key": "userId",
            "type": "number",
            "label": "Masukkan User ID",
            "required": true,
            "minLength": 1,
            "maxLength": 12,
            "placeholder": "123456789",
            "hint": "Cek di profil game"
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
            "hint": "Zone ID di samping User ID"
        }
    ]
}
```

---

## SKU Management

### 30. Get All SKUs

**Endpoint:** `GET /admin/v2/skus`

**Permission Required:** `sku:read`

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| limit | integer | No | Items per page. Default: 10 |
| page | integer | No | Page number. Default: 1 |
| search | string | No | Search by name or code |
| productCode | string | No | Filter by product |
| providerCode | string | No | Filter by provider |
| region | string | No | Filter by region |
| isActive | boolean | No | Filter by status |

**Response:**

```json
{
    "data": {
        "skus": [
            {
                "id": "sku_1a2b3c4d",
                "code": "MLBB_86",
                "providerSkuCode": "ml-86",
                "name": "86 Diamonds",
                "description": "86 (78+8) Diamonds",
                "product": {
                    "code": "MLBB",
                    "title": "Mobile Legends"
                },
                "provider": {
                    "code": "DIGIFLAZZ",
                    "name": "Digiflazz"
                },
                "pricing": {
                    "ID": {
                        "currency": "IDR",
                        "buyPrice": 22000,
                        "sellPrice": 24750,
                        "originalPrice": 25000,
                        "margin": 12.5,
                        "discount": 1.0
                    },
                    "MY": {
                        "currency": "MYR",
                        "buyPrice": 6.5,
                        "sellPrice": 7.5,
                        "originalPrice": 8.0,
                        "margin": 15.4,
                        "discount": 6.25
                    }
                },
                "section": {
                    "code": "topup-instant",
                    "title": "Topup Instan"
                },
                "isActive": true,
                "isFeatured": false,
                "processTime": 0,
                "stock": "AVAILABLE",
                "stats": {
                    "todaySold": 125,
                    "totalSold": 15000
                },
                "createdAt": "2025-01-01T00:00:00+07:00",
                "updatedAt": "2025-12-01T00:00:00+07:00"
            }
        ],
        "pagination": {
            "limit": 10,
            "page": 1,
            "totalRows": 500,
            "totalPages": 50
        }
    }
}
```

---

### 31. Get SKU Detail

**Endpoint:** `GET /admin/v2/skus/{skuId}`

**Permission Required:** `sku:read`

---

### 32. Create SKU

**Endpoint:** `POST /admin/v2/skus`

**Permission Required:** `sku:create`

**Request Body:**

```json
{
    "code": "MLBB_NEW",
    "providerSkuCode": "ml-new",
    "name": "New Diamond Pack",
    "description": "New Diamond Pack Description",
    "productCode": "MLBB",
    "providerCode": "DIGIFLAZZ",
    "sectionCode": "topup-instant",
    "isActive": true,
    "isFeatured": false,
    "processTime": 0,
    "image": "https://...",
    "info": "Bonus info",
    "pricing": {
        "ID": {
            "buyPrice": 22000,
            "sellPrice": 24750,
            "originalPrice": 25000
        },
        "MY": {
            "buyPrice": 6.5,
            "sellPrice": 7.5,
            "originalPrice": 8.0
        }
    },
    "badge": {
        "text": "NEW",
        "color": "#4CAF50"
    }
}
```

---

### 33. Update SKU

**Endpoint:** `PUT /admin/v2/skus/{skuId}`

**Permission Required:** `sku:update`

---

### 34. Delete SKU

**Endpoint:** `DELETE /admin/v2/skus/{skuId}`

**Permission Required:** `sku:delete`

---

### 35. Bulk Update SKU Prices

**Endpoint:** `PUT /admin/v2/skus/bulk-price`

**Permission Required:** `sku:update`

**Request Body:**

```json
{
    "skus": [
        {
            "code": "MLBB_86",
            "pricing": {
                "ID": {
                    "sellPrice": 25000,
                    "originalPrice": 26000
                }
            }
        },
        {
            "code": "MLBB_172",
            "pricing": {
                "ID": {
                    "sellPrice": 50000,
                    "originalPrice": 52000
                }
            }
        }
    ]
}
```

---

### 36. Sync SKUs from Provider

**Endpoint:** `POST /admin/v2/skus/sync`

**Permission Required:** `sku:sync`

**Request Body:**

```json
{
    "providerCode": "DIGIFLAZZ",
    "productCode": "MLBB",
    "autoActivate": false,
    "priceMargin": 10
}
```

**Response:**

```json
{
    "data": {
        "status": "COMPLETED",
        "summary": {
            "totalFromProvider": 20,
            "newSkus": 5,
            "updatedSkus": 10,
            "skippedSkus": 5
        },
        "newSkus": [
            {
                "providerSkuCode": "ml-new-1",
                "name": "New Pack 1",
                "buyPrice": 20000,
                "suggestedSellPrice": 22000
            }
        ],
        "syncedAt": "2025-12-03T12:00:00+07:00"
    }
}
```

---

## Transaction Management

### 37. Get All Transactions

**Endpoint:** `GET /admin/v2/transactions`

**Permission Required:** `transaction:read`

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| limit | integer | No | Items per page. Default: 10 |
| page | integer | No | Page number. Default: 1 |
| search | string | No | Search by invoice number |
| status | string | No | Filter: SUCCESS, PROCESSING, PENDING, FAILED |
| paymentStatus | string | No | Filter: PAID, UNPAID, EXPIRED |
| productCode | string | No | Filter by product |
| providerCode | string | No | Filter by provider |
| paymentCode | string | No | Filter by payment method |
| region | string | No | Filter by region |
| userId | string | No | Filter by user |
| startDate | string | No | Start date (YYYY-MM-DD) |
| endDate | string | No | End date (YYYY-MM-DD) |

**Response:**

```json
{
    "data": {
        "overview": {
            "totalTransactions": 15420,
            "totalRevenue": 1542000000,
            "totalProfit": 154200000,
            "successCount": 14850,
            "processingCount": 320,
            "pendingCount": 200,
            "failedCount": 50
        },
        "transactions": [
            {
                "id": "trx_1a2b3c4d",
                "invoiceNumber": "GATE1A11BB97DF88D56530993",
                "status": "SUCCESS",
                "paymentStatus": "PAID",
                "product": {
                    "code": "MLBB",
                    "name": "Mobile Legends"
                },
                "sku": {
                    "code": "MLBB_172",
                    "name": "172 Diamonds"
                },
                "provider": {
                    "code": "DIGIFLAZZ",
                    "name": "Digiflazz",
                    "refId": "DGF123456789"
                },
                "account": {
                    "nickname": "り い こ ✧",
                    "inputs": "656696292 - 8610"
                },
                "user": {
                    "id": "usr_1a2b3c4d5e6f",
                    "name": "John Doe",
                    "email": "john@example.com"
                },
                "pricing": {
                    "buyPrice": 44000,
                    "sellPrice": 49450,
                    "discount": 4945,
                    "paymentFee": 346,
                    "total": 44851,
                    "profit": 4851,
                    "currency": "IDR"
                },
                "payment": {
                    "code": "QRIS",
                    "name": "QRIS",
                    "gateway": "LINKQU",
                    "paidAt": "2025-12-03T10:30:45+07:00"
                },
                "promo": {
                    "code": "WELCOME10",
                    "discountAmount": 4945
                },
                "region": "ID",
                "ipAddress": "103.xxx.xxx.xxx",
                "userAgent": "Mozilla/5.0...",
                "createdAt": "2025-12-03T10:25:00+07:00",
                "completedAt": "2025-12-03T10:31:15+07:00"
            }
        ],
        "pagination": {
            "limit": 10,
            "page": 1,
            "totalRows": 15420,
            "totalPages": 1542
        }
    }
}
```

---

### 38. Get Transaction Detail

**Endpoint:** `GET /admin/v2/transactions/{transactionId}`

**Permission Required:** `transaction:read`

**Response:**

```json
{
    "data": {
        "id": "trx_1a2b3c4d",
        "invoiceNumber": "GATE1A11BB97DF88D56530993",
        "status": "SUCCESS",
        "paymentStatus": "PAID",
        "product": { ... },
        "sku": { ... },
        "provider": {
            "code": "DIGIFLAZZ",
            "name": "Digiflazz",
            "refId": "DGF123456789",
            "serialNumber": "SN123456789",
            "response": {
                "rc": "00",
                "message": "Sukses"
            }
        },
        "account": { ... },
        "user": { ... },
        "pricing": { ... },
        "payment": {
            "code": "QRIS",
            "name": "QRIS",
            "gateway": "LINKQU",
            "gatewayRefId": "LQ123456789",
            "paidAt": "2025-12-03T10:30:45+07:00"
        },
        "promo": { ... },
        "timeline": [
            {
                "status": "PENDING",
                "message": "Order created",
                "timestamp": "2025-12-03T10:25:00+07:00"
            },
            {
                "status": "PAID",
                "message": "Payment received via QRIS",
                "timestamp": "2025-12-03T10:30:45+07:00"
            },
            {
                "status": "PROCESSING",
                "message": "Sending to provider DIGIFLAZZ",
                "timestamp": "2025-12-03T10:30:50+07:00"
            },
            {
                "status": "SUCCESS",
                "message": "Transaction completed. SN: SN123456789",
                "timestamp": "2025-12-03T10:31:15+07:00"
            }
        ],
        "logs": [
            {
                "type": "PROVIDER_REQUEST",
                "data": { ... },
                "timestamp": "2025-12-03T10:30:50+07:00"
            },
            {
                "type": "PROVIDER_RESPONSE",
                "data": { ... },
                "timestamp": "2025-12-03T10:31:15+07:00"
            }
        ],
        "region": "ID",
        "ipAddress": "103.xxx.xxx.xxx",
        "userAgent": "Mozilla/5.0...",
        "createdAt": "2025-12-03T10:25:00+07:00",
        "completedAt": "2025-12-03T10:31:15+07:00"
    }
}
```

---

### 39. Update Transaction Status

**Endpoint:** `PUT /admin/v2/transactions/{transactionId}/status`

**Permission Required:** `transaction:update`

**Request Body:**

```json
{
    "status": "SUCCESS",
    "reason": "Manually verified by admin",
    "serialNumber": "SN123456789"
}
```

---

### 40. Refund Transaction

**Endpoint:** `POST /admin/v2/transactions/{transactionId}/refund`

**Permission Required:** `transaction:refund`

**Request Body:**

```json
{
    "reason": "Customer request - product not received",
    "refundTo": "BALANCE",
    "amount": 44851
}
```

**Response:**

```json
{
    "data": {
        "refundId": "ref_1a2b3c4d",
        "transactionId": "trx_1a2b3c4d",
        "invoiceNumber": "GATE1A11BB97DF88D56530993",
        "amount": 44851,
        "currency": "IDR",
        "refundTo": "BALANCE",
        "status": "SUCCESS",
        "reason": "Customer request - product not received",
        "processedBy": {
            "id": "adm_1a2b3c4d5e6f",
            "name": "John Admin"
        },
        "createdAt": "2025-12-03T12:00:00+07:00"
    }
}
```

---

### 41. Retry Transaction

**Endpoint:** `POST /admin/v2/transactions/{transactionId}/retry`

**Permission Required:** `transaction:manual`

**Request Body:**

```json
{
    "providerCode": "VIPRESELLER",
    "reason": "Retry with different provider"
}
```

---

### 42. Manual Process Transaction

**Endpoint:** `POST /admin/v2/transactions/{transactionId}/manual`

**Permission Required:** `transaction:manual`

**Request Body:**

```json
{
    "serialNumber": "SN123456789",
    "reason": "Manually processed outside system"
}
```

---

## User Management

### 43. Get All Users

**Endpoint:** `GET /admin/v2/users`

**Permission Required:** `user:read`

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| limit | integer | No | Items per page. Default: 10 |
| page | integer | No | Page number. Default: 1 |
| search | string | No | Search by name, email, phone |
| status | string | No | Filter: ACTIVE, INACTIVE, SUSPENDED |
| membership | string | No | Filter: CLASSIC, PRESTIGE, ROYAL |
| region | string | No | Filter by primary region |

**Response:**

```json
{
    "data": {
        "users": [
            {
                "id": "usr_1a2b3c4d5e6f",
                "firstName": "John",
                "lastName": "Doe",
                "email": "john@example.com",
                "phoneNumber": "+628123456789",
                "profilePicture": "https://...",
                "status": "ACTIVE",
                "primaryRegion": "ID",
                "membership": {
                    "level": "PRESTIGE",
                    "name": "Prestige"
                },
                "balance": {
                    "IDR": 150000,
                    "MYR": 500,
                    "PHP": 0,
                    "SGD": 0,
                    "THB": 0
                },
                "stats": {
                    "totalTransactions": 125,
                    "totalSpent": 5420000,
                    "lastTransactionAt": "2025-12-03T10:30:00+07:00"
                },
                "createdAt": "2025-11-01T00:00:00+07:00",
                "lastLoginAt": "2025-12-03T10:00:00+07:00"
            }
        ],
        "pagination": {
            "limit": 10,
            "page": 1,
            "totalRows": 25000,
            "totalPages": 2500
        }
    }
}
```

---

### 44. Get User Detail

**Endpoint:** `GET /admin/v2/users/{userId}`

**Permission Required:** `user:read`

---

### 45. Update User Status

**Endpoint:** `PUT /admin/v2/users/{userId}/status`

**Permission Required:** `user:suspend`

**Request Body:**

```json
{
    "status": "SUSPENDED",
    "reason": "Suspicious activity detected"
}
```

---

### 46. Adjust User Balance

**Endpoint:** `POST /admin/v2/users/{userId}/balance`

**Permission Required:** `user:balance`

**Request Body:**

```json
{
    "type": "CREDIT",
    "amount": 50000,
    "currency": "IDR",
    "reason": "Compensation for failed transaction #GATE123"
}
```

**Response:**

```json
{
    "data": {
        "userId": "usr_1a2b3c4d5e6f",
        "type": "CREDIT",
        "amount": 50000,
        "currency": "IDR",
        "balanceBefore": 150000,
        "balanceAfter": 200000,
        "reason": "Compensation for failed transaction #GATE123",
        "processedBy": {
            "id": "adm_1a2b3c4d5e6f",
            "name": "John Admin"
        },
        "createdAt": "2025-12-03T12:00:00+07:00"
    }
}
```

---

### 47. Get User Transactions

**Endpoint:** `GET /admin/v2/users/{userId}/transactions`

**Permission Required:** `user:read`

---

### 48. Get User Mutations

**Endpoint:** `GET /admin/v2/users/{userId}/mutations`

**Permission Required:** `user:read`

---

## Promo Management

### 49. Get All Promos

**Endpoint:** `GET /admin/v2/promos`

**Permission Required:** `promo:read`

---

### 50. Create Promo

**Endpoint:** `POST /admin/v2/promos`

**Permission Required:** `promo:create`

**Request Body:**

```json
{
    "code": "NEWYEAR2026",
    "title": "New Year 2026 Promo",
    "description": "Celebrate New Year with 25% discount",
    "products": ["MLBB", "FF", "PUBGM"],
    "paymentChannels": ["QRIS", "DANA"],
    "regions": ["ID", "MY"],
    "daysAvailable": ["SAT", "SUN"],
    "maxDailyUsage": 1000,
    "maxUsage": 10000,
    "maxUsagePerId": 3,
    "maxUsagePerDevice": 3,
    "maxUsagePerIp": 5,
    "startAt": "2025-12-31T00:00:00+07:00",
    "expiredAt": "2026-01-07T23:59:59+07:00",
    "minAmount": 50000,
    "maxPromoAmount": 25000,
    "promoFlat": 0,
    "promoPercentage": 25,
    "isActive": true,
    "note": "Valid for weekend only"
}
```

---

### 51. Update Promo

**Endpoint:** `PUT /admin/v2/promos/{promoId}`

**Permission Required:** `promo:update`

---

### 52. Delete Promo

**Endpoint:** `DELETE /admin/v2/promos/{promoId}`

**Permission Required:** `promo:delete`

---

### 53. Get Promo Usage Stats

**Endpoint:** `GET /admin/v2/promos/{promoId}/stats`

**Permission Required:** `promo:read`

**Response:**

```json
{
    "data": {
        "promoCode": "NEWYEAR2026",
        "totalUsage": 5421,
        "totalDiscount": 135525000,
        "todayUsage": 320,
        "todayDiscount": 8000000,
        "usageByProduct": [
            { "product": "MLBB", "count": 2500, "discount": 62500000 },
            { "product": "FF", "count": 1800, "discount": 45000000 },
            { "product": "PUBGM", "count": 1121, "discount": 28025000 }
        ],
        "usageByPayment": [
            { "payment": "QRIS", "count": 3200, "discount": 80000000 },
            { "payment": "DANA", "count": 2221, "discount": 55525000 }
        ],
        "usageByRegion": [
            { "region": "ID", "count": 4500, "discount": 112500000 },
            { "region": "MY", "count": 921, "discount": 23025000 }
        ]
    }
}
```

---

## Content Management

### 54. Get All Banners

**Endpoint:** `GET /admin/v2/banners`

**Permission Required:** `content:banner`

---

### 55. Create Banner

**Endpoint:** `POST /admin/v2/banners`

**Permission Required:** `content:banner`

**Request Body (multipart/form-data):**

```json
{
    "title": "New Banner",
    "description": "Banner description",
    "href": "/id-id/mobile-legends",
    "image": "[FILE]",
    "regions": ["ID", "MY"],
    "order": 1,
    "isActive": true,
    "startAt": "2025-12-01T00:00:00+07:00",
    "expiredAt": "2025-12-31T23:59:59+07:00"
}
```

---

### 56. Update Banner

**Endpoint:** `PUT /admin/v2/banners/{bannerId}`

**Permission Required:** `content:banner`

---

### 57. Delete Banner

**Endpoint:** `DELETE /admin/v2/banners/{bannerId}`

**Permission Required:** `content:banner`

---

### 58. Get Popups

**Endpoint:** `GET /admin/v2/popups`

**Permission Required:** `content:popup`

---

### 59. Update Popup

**Endpoint:** `PUT /admin/v2/popups/{region}`

**Permission Required:** `content:popup`

**Request Body:**

```json
{
    "title": "🔥 PROMO SPESIAL! 🔥",
    "content": "<p>Diskon hingga 50%!</p>",
    "image": "[FILE or URL]",
    "href": "/id-id/promo",
    "isActive": true
}
```

---

## Reports & Analytics

### 60. Get Dashboard Overview

**Endpoint:** `GET /admin/v2/reports/dashboard`

**Permission Required:** `report:read`

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| region | string | No | Filter by region |
| startDate | string | No | Start date |
| endDate | string | No | End date |

**Response:**

```json
{
    "data": {
        "summary": {
            "totalRevenue": 15420000000,
            "totalProfit": 1542000000,
            "totalTransactions": 154200,
            "totalUsers": 25000,
            "newUsers": 1250,
            "activeUsers": 8500
        },
        "revenueChart": [
            { "date": "2025-12-01", "revenue": 520000000, "profit": 52000000 },
            { "date": "2025-12-02", "revenue": 480000000, "profit": 48000000 },
            { "date": "2025-12-03", "revenue": 542000000, "profit": 54200000 }
        ],
        "topProducts": [
            { "code": "MLBB", "name": "Mobile Legends", "revenue": 5200000000, "transactions": 52000 },
            { "code": "FF", "name": "Free Fire", "revenue": 3800000000, "transactions": 38000 },
            { "code": "PUBGM", "name": "PUBG Mobile", "revenue": 2500000000, "transactions": 25000 }
        ],
        "topPayments": [
            { "code": "QRIS", "name": "QRIS", "revenue": 6500000000, "transactions": 65000 },
            { "code": "DANA", "name": "DANA", "revenue": 4200000000, "transactions": 42000 }
        ],
        "providerHealth": [
            { "code": "DIGIFLAZZ", "status": "HEALTHY", "successRate": 98.5 },
            { "code": "VIPRESELLER", "status": "HEALTHY", "successRate": 97.8 },
            { "code": "BANGJEFF", "status": "DEGRADED", "successRate": 95.2 }
        ]
    }
}
```

---

### 61. Get Revenue Report

**Endpoint:** `GET /admin/v2/reports/revenue`

**Permission Required:** `report:read`

---

### 62. Get Transaction Report

**Endpoint:** `GET /admin/v2/reports/transactions`

**Permission Required:** `report:read`

---

### 63. Get Product Report

**Endpoint:** `GET /admin/v2/reports/products`

**Permission Required:** `report:read`

---

### 64. Get Provider Report

**Endpoint:** `GET /admin/v2/reports/providers`

**Permission Required:** `report:read`

---

### 65. Export Report

**Endpoint:** `POST /admin/v2/reports/export`

**Permission Required:** `report:export`

**Request Body:**

```json
{
    "reportType": "transactions",
    "format": "xlsx",
    "filters": {
        "startDate": "2025-12-01",
        "endDate": "2025-12-31",
        "region": "ID"
    }
}
```

**Response:**

```json
{
    "data": {
        "exportId": "exp_1a2b3c4d",
        "status": "PROCESSING",
        "downloadUrl": null,
        "expiresAt": null
    }
}
```

---

### 66. Get Export Status

**Endpoint:** `GET /admin/v2/reports/export/{exportId}`

**Permission Required:** `report:export`

**Response:**

```json
{
    "data": {
        "exportId": "exp_1a2b3c4d",
        "status": "COMPLETED",
        "downloadUrl": "https://nos.jkt-1.neo.id/gate/exports/report_123.xlsx",
        "expiresAt": "2025-12-04T12:00:00+07:00"
    }
}
```

---

## Audit Logs

### 67. Get Audit Logs

**Endpoint:** `GET /admin/v2/audit-logs`

**Permission Required:** `audit:read`

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| limit | integer | No | Items per page. Default: 10 |
| page | integer | No | Page number. Default: 1 |
| adminId | string | No | Filter by admin |
| action | string | No | Filter by action type |
| resource | string | No | Filter by resource type |
| startDate | string | No | Start date |
| endDate | string | No | End date |

**Response:**

```json
{
    "data": {
        "logs": [
            {
                "id": "log_1a2b3c4d",
                "admin": {
                    "id": "adm_1a2b3c4d5e6f",
                    "name": "John Superadmin",
                    "email": "superadmin@gate.co.id"
                },
                "action": "UPDATE",
                "resource": "SKU",
                "resourceId": "sku_1a2b3c4d",
                "description": "Updated SKU MLBB_86 price from 24000 to 24750",
                "changes": {
                    "before": { "sellPrice": 24000 },
                    "after": { "sellPrice": 24750 }
                },
                "ipAddress": "103.xxx.xxx.xxx",
                "userAgent": "Mozilla/5.0...",
                "createdAt": "2025-12-03T12:00:00+07:00"
            },
            {
                "id": "log_2b3c4d5e",
                "admin": {
                    "id": "adm_2b3c4d5e6f7g",
                    "name": "Jane Admin",
                    "email": "admin@gate.co.id"
                },
                "action": "CREATE",
                "resource": "PROMO",
                "resourceId": "prm_1a2b3c4d",
                "description": "Created new promo NEWYEAR2026",
                "changes": {
                    "before": null,
                    "after": { "code": "NEWYEAR2026", "promoPercentage": 25 }
                },
                "ipAddress": "103.xxx.xxx.xxx",
                "userAgent": "Mozilla/5.0...",
                "createdAt": "2025-12-03T11:30:00+07:00"
            }
        ],
        "pagination": {
            "limit": 10,
            "page": 1,
            "totalRows": 5000,
            "totalPages": 500
        }
    }
}
```

---

## Settings

### 68. Get All Settings

**Endpoint:** `GET /admin/v2/settings`

**Permission Required:** `setting:read`

**Response:**

```json
{
    "data": {
        "general": {
            "siteName": "Gate.co.id",
            "siteDescription": "Top Up Game & Voucher Digital Terpercaya",
            "maintenanceMode": false,
            "maintenanceMessage": null
        },
        "transaction": {
            "orderExpiry": 3600,
            "autoRefundOnFail": true,
            "maxRetryAttempts": 3
        },
        "notification": {
            "emailEnabled": true,
            "whatsappEnabled": true,
            "telegramEnabled": false
        },
        "security": {
            "maxLoginAttempts": 5,
            "lockoutDuration": 900,
            "sessionTimeout": 3600,
            "mfaRequired": true
        }
    }
}
```

---

### 69. Update Settings

**Endpoint:** `PUT /admin/v2/settings/{category}`

**Permission Required:** `setting:update`

**Request Body:**

```json
{
    "orderExpiry": 7200,
    "autoRefundOnFail": true,
    "maxRetryAttempts": 5
}
```

---

### 70. Get Contacts Settings

**Endpoint:** `GET /admin/v2/settings/contacts`

**Permission Required:** `setting:read`

---

### 71. Update Contacts Settings

**Endpoint:** `PUT /admin/v2/settings/contacts`

**Permission Required:** `setting:update`

**Request Body:**

```json
{
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
```

---

## Region Management

### 72. Get All Regions

**Endpoint:** `GET /admin/v2/regions`

**Permission Required:** `setting:read`

**Response:**

```json
{
    "data": [
        {
            "id": "reg_1a2b3c4d",
            "code": "ID",
            "country": "Indonesia",
            "currency": "IDR",
            "currencySymbol": "Rp",
            "image": "https://nos.jkt-1.neo.id/gate/flags/id.svg",
            "isDefault": true,
            "isActive": true,
            "order": 1,
            "stats": {
                "totalUsers": 20000,
                "totalTransactions": 150000,
                "totalRevenue": 15000000000
            },
            "createdAt": "2025-01-01T00:00:00+07:00",
            "updatedAt": "2025-12-01T00:00:00+07:00"
        },
        {
            "id": "reg_2b3c4d5e",
            "code": "MY",
            "country": "Malaysia",
            "currency": "MYR",
            "currencySymbol": "RM",
            "image": "https://nos.jkt-1.neo.id/gate/flags/my.svg",
            "isDefault": false,
            "isActive": true,
            "order": 2,
            "stats": {
                "totalUsers": 5000,
                "totalTransactions": 35000,
                "totalRevenue": 1500000
            },
            "createdAt": "2025-01-01T00:00:00+07:00",
            "updatedAt": "2025-12-01T00:00:00+07:00"
        }
    ]
}
```

---

### 73. Create Region

**Endpoint:** `POST /admin/v2/regions`

**Permission Required:** `setting:update`

**Request Body (multipart/form-data):**

```json
{
    "code": "VN",
    "country": "Vietnam",
    "currency": "VND",
    "currencySymbol": "₫",
    "image": "[FILE]",
    "isDefault": false,
    "isActive": true,
    "order": 6
}
```

---

### 74. Update Region

**Endpoint:** `PUT /admin/v2/regions/{regionId}`

**Permission Required:** `setting:update`

**Request Body:**

```json
{
    "country": "Updated Country Name",
    "isActive": true,
    "order": 3
}
```

---

### 75. Delete Region

**Endpoint:** `DELETE /admin/v2/regions/{regionId}`

**Permission Required:** `setting:update`

> **Warning:** Cannot delete region with active users or transactions.

---

## Language Management

### 76. Get All Languages

**Endpoint:** `GET /admin/v2/languages`

**Permission Required:** `setting:read`

**Response:**

```json
{
    "data": [
        {
            "id": "lang_1a2b3c4d",
            "code": "id",
            "name": "Bahasa Indonesia",
            "country": "Indonesia",
            "image": "https://nos.jkt-1.neo.id/gate/flags/id.svg",
            "isDefault": true,
            "isActive": true,
            "order": 1,
            "createdAt": "2025-01-01T00:00:00+07:00"
        },
        {
            "id": "lang_2b3c4d5e",
            "code": "en",
            "name": "English",
            "country": "United States",
            "image": "https://nos.jkt-1.neo.id/gate/flags/us.svg",
            "isDefault": false,
            "isActive": true,
            "order": 2,
            "createdAt": "2025-01-01T00:00:00+07:00"
        }
    ]
}
```

---

### 77. Create Language

**Endpoint:** `POST /admin/v2/languages`

**Permission Required:** `setting:update`

**Request Body:**

```json
{
    "code": "ms",
    "name": "Bahasa Melayu",
    "country": "Malaysia",
    "image": "[FILE or URL]",
    "isDefault": false,
    "isActive": true,
    "order": 3
}
```

---

### 78. Update Language

**Endpoint:** `PUT /admin/v2/languages/{languageId}`

**Permission Required:** `setting:update`

---

### 79. Delete Language

**Endpoint:** `DELETE /admin/v2/languages/{languageId}`

**Permission Required:** `setting:update`

> **Warning:** Cannot delete default language.

---

## Category Management

### 80. Get All Categories

**Endpoint:** `GET /admin/v2/categories`

**Permission Required:** `product:read`

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| region | string | No | Filter by region |
| isActive | boolean | No | Filter by status |

**Response:**

```json
{
    "data": [
        {
            "id": "cat_1a2b3c4d",
            "code": "top-up-game",
            "title": "Top Up Game",
            "description": "Top up diamond, UC, dan in-game currency lainnya",
            "icon": "https://nos.jkt-1.neo.id/gate/icons/game-controller.svg",
            "isActive": true,
            "order": 1,
            "regions": ["ID", "MY", "PH", "SG", "TH"],
            "productCount": 25,
            "stats": {
                "totalTransactions": 100000,
                "totalRevenue": 10000000000
            },
            "createdAt": "2025-01-01T00:00:00+07:00",
            "updatedAt": "2025-12-01T00:00:00+07:00"
        },
        {
            "id": "cat_2b3c4d5e",
            "code": "voucher",
            "title": "Voucher",
            "description": "Voucher game dan digital content",
            "icon": "https://nos.jkt-1.neo.id/gate/icons/ticket.svg",
            "isActive": true,
            "order": 2,
            "regions": ["ID", "MY"],
            "productCount": 15,
            "stats": {
                "totalTransactions": 25000,
                "totalRevenue": 2500000000
            },
            "createdAt": "2025-01-01T00:00:00+07:00",
            "updatedAt": "2025-12-01T00:00:00+07:00"
        },
        {
            "id": "cat_3c4d5e6f",
            "code": "e-money",
            "title": "E-Money",
            "description": "Top up saldo e-wallet dan e-money",
            "icon": "https://nos.jkt-1.neo.id/gate/icons/wallet.svg",
            "isActive": true,
            "order": 3,
            "regions": ["ID"],
            "productCount": 10,
            "stats": {
                "totalTransactions": 50000,
                "totalRevenue": 5000000000
            },
            "createdAt": "2025-01-01T00:00:00+07:00",
            "updatedAt": "2025-12-01T00:00:00+07:00"
        }
    ]
}
```

---

### 81. Create Category

**Endpoint:** `POST /admin/v2/categories`

**Permission Required:** `product:create`

**Request Body (multipart/form-data):**

```json
{
    "code": "new-category",
    "title": "New Category",
    "description": "New category description",
    "icon": "[FILE]",
    "isActive": true,
    "order": 5,
    "regions": ["ID", "MY", "SG"]
}
```

---

### 82. Update Category

**Endpoint:** `PUT /admin/v2/categories/{categoryId}`

**Permission Required:** `product:update`

**Request Body:**

```json
{
    "title": "Updated Category Name",
    "description": "Updated description",
    "isActive": true,
    "order": 3,
    "regions": ["ID", "MY", "PH", "SG", "TH"]
}
```

---

### 83. Delete Category

**Endpoint:** `DELETE /admin/v2/categories/{categoryId}`

**Permission Required:** `product:delete`

> **Warning:** Cannot delete category with active products.

---

## Section Management

Sections are used to group SKUs within a product (e.g., "Spesial Item", "Topup Instan").

### 84. Get All Sections

**Endpoint:** `GET /admin/v2/sections`

**Permission Required:** `product:read`

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| productCode | string | No | Filter by product |

**Response:**

```json
{
    "data": [
        {
            "id": "sec_1a2b3c4d",
            "code": "special-item",
            "title": "Spesial Item",
            "icon": "⭐",
            "isActive": true,
            "order": 1,
            "products": ["MLBB", "FF", "PUBGM"],
            "skuCount": 25,
            "createdAt": "2025-01-01T00:00:00+07:00"
        },
        {
            "id": "sec_2b3c4d5e",
            "code": "topup-instant",
            "title": "Topup Instan",
            "icon": "⚡",
            "isActive": true,
            "order": 2,
            "products": ["MLBB", "FF", "PUBGM", "GENSHIN"],
            "skuCount": 150,
            "createdAt": "2025-01-01T00:00:00+07:00"
        },
        {
            "id": "sec_3c4d5e6f",
            "code": "weekly-pass",
            "title": "Weekly Pass",
            "icon": "📅",
            "isActive": true,
            "order": 3,
            "products": ["MLBB", "GENSHIN"],
            "skuCount": 10,
            "createdAt": "2025-01-01T00:00:00+07:00"
        }
    ]
}
```

---

### 85. Create Section

**Endpoint:** `POST /admin/v2/sections`

**Permission Required:** `product:create`

**Request Body:**

```json
{
    "code": "new-section",
    "title": "New Section",
    "icon": "🎁",
    "isActive": true,
    "order": 4,
    "products": ["MLBB", "FF"]
}
```

---

### 86. Update Section

**Endpoint:** `PUT /admin/v2/sections/{sectionId}`

**Permission Required:** `product:update`

---

### 87. Delete Section

**Endpoint:** `DELETE /admin/v2/sections/{sectionId}`

**Permission Required:** `product:delete`

> **Warning:** Cannot delete section with active SKUs.

---

### 88. Assign Section to Products

**Endpoint:** `PUT /admin/v2/sections/{sectionId}/products`

**Permission Required:** `product:update`

**Request Body:**

```json
{
    "products": ["MLBB", "FF", "PUBGM", "GENSHIN", "VALORANT"]
}
```

---

## Payment Channel Management

Payment channels are the user-facing payment methods (QRIS, DANA, BCA VA, etc).

### 89. Get All Payment Channels

**Endpoint:** `GET /admin/v2/payment-channels`

**Permission Required:** `gateway:read`

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| categoryCode | string | No | Filter by category |
| region | string | No | Filter by region |
| isActive | boolean | No | Filter by status |

**Response:**

```json
{
    "data": [
        {
            "id": "pc_1a2b3c4d",
            "code": "QRIS",
            "name": "QRIS",
            "description": "Bayar menggunakan QRIS dari semua aplikasi e-wallet dan mobile banking",
            "image": "https://nos.jkt-1.neo.id/gate/payment/qris.webp",
            "category": {
                "code": "E_WALLET",
                "title": "E-Wallet"
            },
            "gateway": {
                "purchase": { "code": "LINKQU", "name": "LinkQu" },
                "deposit": { "code": "LINKQU", "name": "LinkQu" }
            },
            "fee": {
                "feeType": "PERCENTAGE",
                "feeAmount": 0,
                "feePercentage": 0.7
            },
            "limits": {
                "minAmount": 1000,
                "maxAmount": 10000000
            },
            "regions": ["ID"],
            "supportedTypes": ["purchase", "deposit"],
            "isActive": true,
            "isFeatured": true,
            "order": 1,
            "instruction": "<p>Gunakan E-wallet atau aplikasi mobile banking...</p>",
            "stats": {
                "todayTransactions": 2500,
                "todayVolume": 125000000
            },
            "createdAt": "2025-01-01T00:00:00+07:00",
            "updatedAt": "2025-12-01T00:00:00+07:00"
        },
        {
            "id": "pc_2b3c4d5e",
            "code": "DANA",
            "name": "DANA",
            "description": "Bayar menggunakan DANA",
            "image": "https://nos.jkt-1.neo.id/gate/payment/dana.webp",
            "category": {
                "code": "E_WALLET",
                "title": "E-Wallet"
            },
            "gateway": {
                "purchase": { "code": "DANA_DIRECT", "name": "DANA Direct" },
                "deposit": null
            },
            "fee": {
                "feeType": "PERCENTAGE",
                "feeAmount": 0,
                "feePercentage": 1.5
            },
            "limits": {
                "minAmount": 1000,
                "maxAmount": 5000000
            },
            "regions": ["ID"],
            "supportedTypes": ["purchase"],
            "isActive": true,
            "isFeatured": true,
            "order": 2,
            "instruction": "<ol><li>Setelah klik bayar...</li></ol>",
            "stats": {
                "todayTransactions": 2100,
                "todayVolume": 65000000
            },
            "createdAt": "2025-01-01T00:00:00+07:00",
            "updatedAt": "2025-12-01T00:00:00+07:00"
        }
    ]
}
```

---

### 90. Get Payment Channel Detail

**Endpoint:** `GET /admin/v2/payment-channels/{channelId}`

**Permission Required:** `gateway:read`

---

### 91. Create Payment Channel

**Endpoint:** `POST /admin/v2/payment-channels`

**Permission Required:** `gateway:create`

**Request Body (multipart/form-data):**

```json
{
    "code": "NEW_PAYMENT",
    "name": "New Payment Method",
    "description": "Description of payment method",
    "image": "[FILE]",
    "categoryCode": "E_WALLET",
    "gatewayAssignment": {
        "purchase": "XENDIT",
        "deposit": "XENDIT"
    },
    "fee": {
        "feeType": "PERCENTAGE",
        "feeAmount": 0,
        "feePercentage": 2.0
    },
    "limits": {
        "minAmount": 10000,
        "maxAmount": 5000000
    },
    "regions": ["ID", "MY"],
    "supportedTypes": ["purchase", "deposit"],
    "isActive": true,
    "isFeatured": false,
    "order": 10,
    "instruction": "<ol><li>Step 1</li><li>Step 2</li></ol>"
}
```

---

### 92. Update Payment Channel

**Endpoint:** `PUT /admin/v2/payment-channels/{channelId}`

**Permission Required:** `gateway:update`

**Request Body:**

```json
{
    "name": "Updated Payment Name",
    "description": "Updated description",
    "fee": {
        "feeType": "FIXED",
        "feeAmount": 2500,
        "feePercentage": 0
    },
    "limits": {
        "minAmount": 5000,
        "maxAmount": 10000000
    },
    "isActive": true,
    "isFeatured": true,
    "order": 1
}
```

---

### 93. Delete Payment Channel

**Endpoint:** `DELETE /admin/v2/payment-channels/{channelId}`

**Permission Required:** `gateway:delete`

> **Warning:** Cannot delete payment channel with pending transactions.

---

### 94. Get Payment Channel Categories

**Endpoint:** `GET /admin/v2/payment-channel-categories`

**Permission Required:** `gateway:read`

**Response:**

```json
{
    "data": [
        {
            "id": "pcc_1a2b3c4d",
            "code": "E_WALLET",
            "title": "E-Wallet",
            "icon": "https://nos.jkt-1.neo.id/gate/icons/wallet.svg",
            "isActive": true,
            "order": 1,
            "channelCount": 6
        },
        {
            "id": "pcc_2b3c4d5e",
            "code": "VIRTUAL_ACCOUNT",
            "title": "Virtual Account",
            "icon": "https://nos.jkt-1.neo.id/gate/icons/bank.svg",
            "isActive": true,
            "order": 2,
            "channelCount": 5
        },
        {
            "id": "pcc_3c4d5e6f",
            "code": "RETAIL",
            "title": "Convenience Store",
            "icon": "https://nos.jkt-1.neo.id/gate/icons/store.svg",
            "isActive": true,
            "order": 3,
            "channelCount": 2
        },
        {
            "id": "pcc_4d5e6f7g",
            "code": "CARD",
            "title": "Credit or Debit Card",
            "icon": "https://nos.jkt-1.neo.id/gate/icons/card.svg",
            "isActive": true,
            "order": 4,
            "channelCount": 1
        }
    ]
}
```

---

### 95. Create Payment Channel Category

**Endpoint:** `POST /admin/v2/payment-channel-categories`

**Permission Required:** `gateway:create`

**Request Body:**

```json
{
    "code": "CRYPTO",
    "title": "Cryptocurrency",
    "icon": "[FILE or URL]",
    "isActive": true,
    "order": 5
}
```

---

### 96. Update Payment Channel Category

**Endpoint:** `PUT /admin/v2/payment-channel-categories/{categoryId}`

**Permission Required:** `gateway:update`

---

### 97. Delete Payment Channel Category

**Endpoint:** `DELETE /admin/v2/payment-channel-categories/{categoryId}`

**Permission Required:** `gateway:delete`

> **Warning:** Cannot delete category with active payment channels.

---

## Deposit Management

Admin can view and manage all user deposits.

### 98. Get All Deposits

**Endpoint:** `GET /admin/v2/deposits`

**Permission Required:** `transaction:read`

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| limit | integer | No | Items per page. Default: 10 |
| page | integer | No | Page number. Default: 1 |
| search | string | No | Search by invoice number |
| status | string | No | Filter: SUCCESS, PENDING, EXPIRED, FAILED |
| paymentCode | string | No | Filter by payment method |
| gatewayCode | string | No | Filter by gateway |
| region | string | No | Filter by region |
| userId | string | No | Filter by user |
| startDate | string | No | Start date (YYYY-MM-DD) |
| endDate | string | No | End date (YYYY-MM-DD) |

**Response:**

```json
{
    "data": {
        "overview": {
            "totalDeposits": 15000,
            "totalAmount": 7500000000,
            "successCount": 14200,
            "pendingCount": 500,
            "expiredCount": 250,
            "failedCount": 50
        },
        "deposits": [
            {
                "id": "dep_1a2b3c4d",
                "invoiceNumber": "DEP5E55FF11IJ22H90974337",
                "user": {
                    "id": "usr_1a2b3c4d5e6f",
                    "name": "John Doe",
                    "email": "john@example.com"
                },
                "amount": 200000,
                "currency": "IDR",
                "status": "SUCCESS",
                "payment": {
                    "code": "QRIS",
                    "name": "QRIS",
                    "gateway": "LINKQU",
                    "gatewayRefId": "LQ123456789"
                },
                "region": "ID",
                "ipAddress": "103.xxx.xxx.xxx",
                "createdAt": "2025-12-03T10:00:00+07:00",
                "paidAt": "2025-12-03T10:01:30+07:00"
            },
            {
                "id": "dep_2b3c4d5e",
                "invoiceNumber": "DEP6F66GG22JK33I01085448",
                "user": {
                    "id": "usr_2b3c4d5e6f7g",
                    "name": "Jane Smith",
                    "email": "jane@example.com"
                },
                "amount": 100000,
                "currency": "IDR",
                "status": "PENDING",
                "payment": {
                    "code": "BCA_VA",
                    "name": "BCA Virtual Account",
                    "gateway": "BCA_DIRECT",
                    "accountNumber": "80777123456789012"
                },
                "region": "ID",
                "ipAddress": "103.xxx.xxx.xxx",
                "createdAt": "2025-12-03T09:30:00+07:00",
                "expiredAt": "2025-12-04T09:30:00+07:00"
            }
        ],
        "pagination": {
            "limit": 10,
            "page": 1,
            "totalRows": 15000,
            "totalPages": 1500
        }
    }
}
```

---

### 99. Get Deposit Detail

**Endpoint:** `GET /admin/v2/deposits/{depositId}`

**Permission Required:** `transaction:read`

**Response:**

```json
{
    "data": {
        "id": "dep_1a2b3c4d",
        "invoiceNumber": "DEP5E55FF11IJ22H90974337",
        "user": {
            "id": "usr_1a2b3c4d5e6f",
            "name": "John Doe",
            "email": "john@example.com",
            "phoneNumber": "+628123456789"
        },
        "amount": 200000,
        "pricing": {
            "subtotal": 200000,
            "paymentFee": 1400,
            "total": 201400,
            "currency": "IDR"
        },
        "status": "SUCCESS",
        "payment": {
            "code": "QRIS",
            "name": "QRIS",
            "gateway": "LINKQU",
            "gatewayRefId": "LQ123456789",
            "qrCode": "00020101021226660016ID.CO.QRIS.WWW..."
        },
        "timeline": [
            {
                "status": "PENDING",
                "message": "Deposit created",
                "timestamp": "2025-12-03T10:00:00+07:00"
            },
            {
                "status": "SUCCESS",
                "message": "Payment received via QRIS",
                "timestamp": "2025-12-03T10:01:30+07:00"
            }
        ],
        "balanceChange": {
            "before": 0,
            "after": 200000,
            "currency": "IDR"
        },
        "region": "ID",
        "ipAddress": "103.xxx.xxx.xxx",
        "userAgent": "Mozilla/5.0...",
        "createdAt": "2025-12-03T10:00:00+07:00",
        "paidAt": "2025-12-03T10:01:30+07:00"
    }
}
```

---

### 100. Manual Confirm Deposit

Manually confirm a pending deposit (for cases where callback failed).

**Endpoint:** `POST /admin/v2/deposits/{depositId}/confirm`

**Permission Required:** `transaction:manual`

**Request Body:**

```json
{
    "reason": "Payment confirmed manually - callback not received",
    "gatewayRefId": "MANUAL123456"
}
```

**Response:**

```json
{
    "data": {
        "id": "dep_2b3c4d5e",
        "invoiceNumber": "DEP6F66GG22JK33I01085448",
        "status": "SUCCESS",
        "amount": 100000,
        "currency": "IDR",
        "balanceChange": {
            "before": 50000,
            "after": 150000
        },
        "confirmedBy": {
            "id": "adm_1a2b3c4d5e6f",
            "name": "John Admin"
        },
        "confirmedAt": "2025-12-03T12:00:00+07:00"
    }
}
```

---

### 101. Cancel/Expire Deposit

Cancel or expire a pending deposit.

**Endpoint:** `POST /admin/v2/deposits/{depositId}/cancel`

**Permission Required:** `transaction:update`

**Request Body:**

```json
{
    "reason": "User requested cancellation"
}
```

---

### 102. Refund Deposit

Refund a completed deposit back to user's payment method.

**Endpoint:** `POST /admin/v2/deposits/{depositId}/refund`

**Permission Required:** `transaction:refund`

**Request Body:**

```json
{
    "reason": "User requested refund",
    "refundTo": "ORIGINAL_METHOD",
    "amount": 200000
}
```

**Response:**

```json
{
    "data": {
        "refundId": "ref_1a2b3c4d",
        "depositId": "dep_1a2b3c4d",
        "invoiceNumber": "DEP5E55FF11IJ22H90974337",
        "amount": 200000,
        "currency": "IDR",
        "refundTo": "ORIGINAL_METHOD",
        "status": "PROCESSING",
        "reason": "User requested refund",
        "balanceChange": {
            "before": 200000,
            "after": 0
        },
        "processedBy": {
            "id": "adm_1a2b3c4d5e6f",
            "name": "John Admin"
        },
        "createdAt": "2025-12-03T12:00:00+07:00"
    }
}
```

---

## Invoice/Order Management (Additional)

### 103. Get All Orders/Invoices

**Endpoint:** `GET /admin/v2/invoices`

**Permission Required:** `transaction:read`

> **Note:** This is an alias for `GET /admin/v2/transactions` with invoice-focused response.

---

### 104. Search Invoice

**Endpoint:** `GET /admin/v2/invoices/search`

**Permission Required:** `transaction:read`

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| q | string | Yes | Search query (invoice number, email, phone) |

**Response:**

```json
{
    "data": [
        {
            "invoiceNumber": "GATE1A11BB97DF88D56530993",
            "type": "PURCHASE",
            "status": "SUCCESS",
            "user": {
                "name": "John Doe",
                "email": "john@example.com"
            },
            "product": "Mobile Legends - 172 Diamonds",
            "total": 44851,
            "currency": "IDR",
            "createdAt": "2025-12-03T10:25:00+07:00"
        },
        {
            "invoiceNumber": "DEP5E55FF11IJ22H90974337",
            "type": "DEPOSIT",
            "status": "SUCCESS",
            "user": {
                "name": "John Doe",
                "email": "john@example.com"
            },
            "product": "Balance Top Up",
            "total": 201400,
            "currency": "IDR",
            "createdAt": "2025-12-03T10:00:00+07:00"
        }
    ]
}
```

---

### 105. Send Invoice Email

**Endpoint:** `POST /admin/v2/invoices/{invoiceNumber}/send-email`

**Permission Required:** `transaction:update`

**Request Body:**

```json
{
    "email": "customer@example.com",
    "type": "RECEIPT"
}
```

**Response:**

```json
{
    "data": {
        "message": "Invoice email sent successfully",
        "sentTo": "customer@example.com",
        "sentAt": "2025-12-03T12:00:00+07:00"
    }
}
```

---

## Error Codes

### Admin-Specific Error Codes

| Code | Description |
|------|-------------|
| `PERMISSION_DENIED` | Insufficient permissions |
| `ADMIN_NOT_FOUND` | Admin account not found |
| `ROLE_NOT_FOUND` | Role not found |
| `INVALID_ROLE_LEVEL` | Cannot modify role with higher level |
| `PROVIDER_NOT_FOUND` | Provider not found |
| `PROVIDER_HAS_SKUS` | Cannot delete provider with active SKUs |
| `GATEWAY_NOT_FOUND` | Payment gateway not found |
| `GATEWAY_HAS_CHANNELS` | Cannot delete gateway with active channels |
| `SKU_HAS_TRANSACTIONS` | Cannot delete SKU with transactions |
| `PRODUCT_HAS_SKUS` | Cannot delete product with active SKUs |
| `PROMO_NOT_FOUND` | Promo not found |
| `EXPORT_NOT_READY` | Export still processing |
| `EXPORT_EXPIRED` | Export download link expired |
| `REGION_NOT_FOUND` | Region not found |
| `REGION_HAS_USERS` | Cannot delete region with active users |
| `REGION_HAS_TRANSACTIONS` | Cannot delete region with transactions |
| `LANGUAGE_NOT_FOUND` | Language not found |
| `CANNOT_DELETE_DEFAULT_LANGUAGE` | Cannot delete default language |
| `CATEGORY_NOT_FOUND` | Category not found |
| `CATEGORY_HAS_PRODUCTS` | Cannot delete category with products |
| `SECTION_NOT_FOUND` | Section not found |
| `SECTION_HAS_SKUS` | Cannot delete section with active SKUs |
| `PAYMENT_CHANNEL_NOT_FOUND` | Payment channel not found |
| `CHANNEL_HAS_TRANSACTIONS` | Cannot delete channel with transactions |
| `DEPOSIT_NOT_FOUND` | Deposit not found |
| `DEPOSIT_ALREADY_CONFIRMED` | Deposit already confirmed |
| `DEPOSIT_EXPIRED` | Deposit has expired |
| `DEPOSIT_CANNOT_REFUND` | Cannot refund this deposit |
| `INVOICE_NOT_FOUND` | Invoice not found |

---

## Summary

### Total Admin Endpoints: 105

| Category | Count | Endpoints |
|----------|-------|-----------|
| Authentication | 4 | Login, Verify MFA, Refresh, Logout |
| Admin Management | 7 | CRUD Admins, Roles, Permissions |
| Provider Management | 7 | CRUD Providers, Test, Sync |
| Payment Gateway | 8 | CRUD Gateways, Test, Assignments |
| Product Management | 7 | CRUD Products, Fields |
| SKU Management | 7 | CRUD SKUs, Bulk Price, Sync |
| Transaction Management | 6 | List, Detail, Update, Refund, Retry, Manual |
| User Management | 6 | List, Detail, Status, Balance, History |
| Promo Management | 5 | CRUD Promos, Stats |
| Content Management | 6 | Banners, Popups |
| Reports | 7 | Dashboard, Revenue, Export |
| Audit Logs | 1 | List Logs |
| Settings | 4 | General, Contacts |
| Region Management | 4 | CRUD Regions |
| Language Management | 4 | CRUD Languages |
| Category Management | 4 | CRUD Categories |
| Section Management | 5 | CRUD Sections, Assign Products |
| Payment Channel Mgmt | 9 | CRUD Channels, Categories |
| Deposit Management | 5 | List, Detail, Confirm, Cancel, Refund |
| Invoice Management | 3 | List, Search, Send Email |

---

**Last Updated:** December 3, 2025 | **API Version:** v2.0 | **Document Version:** 1.0

