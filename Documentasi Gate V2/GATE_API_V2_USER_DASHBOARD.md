# Gate API v2.0 - User Dashboard Endpoints

## 📊 Overview

User Dashboard endpoints menyediakan akses ke:
1. **Transactions** - Riwayat transaksi pembelian produk
2. **Mutations** - Riwayat mutasi saldo (debit/credit)
3. **Reports** - Laporan agregat transaksi per hari
4. **Deposits** - Riwayat dan proses top-up saldo

Semua endpoint ini **require authentication** dan support multi-region & multi-currency.

---

## 🔐 Authentication

All endpoints require Bearer token:

```
Authorization: Bearer {access_token}
```

---

## 1. Get Transactions

Retrieve transaction history with filtering and pagination.

**Endpoint:** `GET /v2/transactions`

**Headers:**
```
Authorization: Bearer {access_token}
```

**Query Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| region | string | No | Region code (ID, MY, PH, SG, TH). Default: user's currentRegion |
| limit | integer | No | Items per page. Default: 10, Max: 100 |
| page | integer | No | Page number. Default: 1 |
| search | string | No | Search by invoice number |
| status | string | No | Filter by transaction status: ALL, SUCCESS, PROCESSING, PENDING, FAILED |
| paymentStatus | string | No | Filter by payment status: PAID, UNPAID, EXPIRED |
| startDate | string | No | Start date (YYYY-MM-DD) |
| endDate | string | No | End date (YYYY-MM-DD) |

**Example Requests:**

```bash
# Get all transactions
GET /v2/transactions?region=ID&limit=10&page=1

# Get successful transactions
GET /v2/transactions?status=SUCCESS

# Get pending payments
GET /v2/transactions?paymentStatus=UNPAID

# Search by invoice
GET /v2/transactions?search=GATE1A11BB

# Filter by date range
GET /v2/transactions?startDate=2025-02-01&endDate=2025-12-31
```

**Response Example:**

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
            },
            {
                "invoiceNumber": "GATE2B22CC88EF99E67641004",
                "status": "PENDING",
                "productCode": "PUBGM",
                "productName": "PUBG Mobile",
                "skuCode": "PUBGM_60",
                "skuName": "60 UC",
                "account": {
                    "nickname": "Test",
                    "inputs": "123456789"
                },
                "pricing": {
                    "subtotal": 16000,
                    "discount": 0,
                    "paymentFee": 112,
                    "total": 16112,
                    "currency": "IDR"
                },
                "payment": {
                    "code": "DANA",
                    "name": "DANA"
                },
                "createdAt": "2025-09-25T11:15:00+07:00"
            },
            {
                "invoiceNumber": "GATE3C33DD99FG00F78752115",
                "status": "PROCESSING",
                "productCode": "FF",
                "productName": "Free Fire",
                "skuCode": "FF_100",
                "skuName": "100 Diamonds",
                "account": {
                    "nickname": "ProGamer",
                    "inputs": "987654321"
                },
                "pricing": {
                    "subtotal": 14000,
                    "discount": 1400,
                    "paymentFee": 88,
                    "total": 12688,
                    "currency": "IDR"
                },
                "payment": {
                    "code": "GOPAY",
                    "name": "GoPay"
                },
                "createdAt": "2025-09-25T12:00:00+07:00"
            },
            {
                "invoiceNumber": "GATE4D44EE00GH11G89863226",
                "status": "FAILED",
                "productCode": "DANA",
                "productName": "Dana",
                "skuCode": "DANA_100",
                "skuName": "Dana 100K",
                "account": {
                    "nickname": "DNID JXXX DXXX",
                    "inputs": "+628123456789"
                },
                "pricing": {
                    "subtotal": 100000,
                    "discount": 0,
                    "paymentFee": 700,
                    "total": 100700,
                    "currency": "IDR"
                },
                "payment": {
                    "code": "BCA_VA",
                    "name": "BCA Virtual Account"
                },
                "createdAt": "2025-09-25T13:30:00+07:00"
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

**Status Values:**

| Status | Description | Color (UI) |
|--------|-------------|------------|
| SUCCESS | Transaction completed successfully | Green |
| PROCESSING | Being processed by provider | Yellow |
| PENDING | Waiting for payment | Orange |
| FAILED | Transaction failed | Red |
| EXPIRED | Payment deadline exceeded | Gray |

**Payment Status Values:**

| Status | Description |
|--------|-------------|
| PAID | Payment received |
| UNPAID | Waiting for payment |
| EXPIRED | Payment time expired |

**Notes:**
- `overview` provides summary statistics for filtered transactions
- Transactions are sorted by `createdAt` (newest first)
- `totalPurchase` is sum of all transaction amounts in selected currency
- Empty `discount` means no promo was applied
- `account.inputs` format varies by product (game ID, phone number, etc.)

---

## 2. Get Mutations

Retrieve balance mutation history (debits and credits).

**Endpoint:** `GET /v2/mutations`

**Headers:**
```
Authorization: Bearer {access_token}
```

**Query Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| region | string | No | Region code (ID, MY, PH, SG, TH). Default: user's currentRegion |
| limit | integer | No | Items per page. Default: 10, Max: 100 |
| page | integer | No | Page number. Default: 1 |
| search | string | No | Search by invoice number |
| type | string | No | Filter by type: ALL, DEBIT, CREDIT |
| startDate | string | No | Start date (YYYY-MM-DD) |
| endDate | string | No | End date (YYYY-MM-DD) |

**Example Requests:**

```bash
# Get all mutations
GET /v2/mutations?region=ID&limit=10&page=1

# Get only debits (purchases)
GET /v2/mutations?type=DEBIT

# Get only credits (top-ups)
GET /v2/mutations?type=CREDIT

# Filter by date range
GET /v2/mutations?startDate=2025-11-01&endDate=2025-11-30
```

**Response Example:**

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
            },
            {
                "invoiceNumber": "GATE2B22CC88EF99E67641004",
                "description": "Pembelian PUBG Mobile - 60 UC",
                "amount": 16112,
                "type": "DEBIT",
                "balanceBefore": 155149,
                "balanceAfter": 139037,
                "currency": "IDR",
                "createdAt": "2025-09-25T11:15:00+07:00"
            },
            {
                "invoiceNumber": "REF3C33DD99FG00F78752115",
                "description": "Refund - Free Fire 100 Diamonds",
                "amount": 12688,
                "type": "CREDIT",
                "balanceBefore": 139037,
                "balanceAfter": 151725,
                "currency": "IDR",
                "createdAt": "2025-09-25T14:00:00+07:00"
            },
            {
                "invoiceNumber": "GATE4D44EE00GH11G89863226",
                "description": "Pembelian Dana - 100K",
                "amount": 100700,
                "type": "DEBIT",
                "balanceBefore": 151725,
                "balanceAfter": 51025,
                "currency": "IDR",
                "createdAt": "2025-09-25T13:30:00+07:00"
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

| Type | Description | Icon (UI) |
|------|-------------|-----------|
| DEBIT | Money out (purchases) | ↓ Red |
| CREDIT | Money in (top-ups, refunds) | ↑ Green |

**Description Patterns:**

| Pattern | Example |
|---------|---------|
| Purchase | `Pembelian {ProductName} - {SkuName}` |
| Top-up | `Isi Ulang Saldo via {PaymentMethod}` |
| Refund | `Refund - {ProductName} {SkuName}` |
| Bonus | `Bonus {EventName}` |
| Cashback | `Cashback {PromoName}` |

**Notes:**
- `overview.netBalance` = totalCredit - totalDebit
- `balanceBefore` and `balanceAfter` help track balance changes
- Mutations are sorted by `createdAt` (newest first)
- All amounts are in the currency specified by region
- Balance is currency-specific (IDR balance for ID region, MYR for MY, etc.)

---

## 3. Get Reports

Retrieve daily transaction reports (aggregated by date).

**Endpoint:** `GET /v2/reports`

**Headers:**
```
Authorization: Bearer {access_token}
```

**Query Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| region | string | No | Region code (ID, MY, PH, SG, TH). Default: user's currentRegion |
| limit | integer | No | Items per page. Default: 10, Max: 100 |
| page | integer | No | Page number. Default: 1 |
| startDate | string | No | Start date (YYYY-MM-DD) |
| endDate | string | No | End date (YYYY-MM-DD) |

**Example Requests:**

```bash
# Get last 30 days reports
GET /v2/reports?region=ID&limit=30&page=1

# Get specific date range
GET /v2/reports?startDate=2025-09-01&endDate=2025-09-30
```

**Response Example:**

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
            },
            {
                "date": "2025-09-23",
                "totalTransactions": 92,
                "totalAmount": 135000,
                "currency": "IDR"
            },
            {
                "date": "2025-09-22",
                "totalTransactions": 78,
                "totalAmount": 95000,
                "currency": "IDR"
            },
            {
                "date": "2025-09-21",
                "totalTransactions": 105,
                "totalAmount": 162000,
                "currency": "IDR"
            },
            {
                "date": "2025-09-20",
                "totalTransactions": 88,
                "totalAmount": 110000,
                "currency": "IDR"
            },
            {
                "date": "2025-09-19",
                "totalTransactions": 95,
                "totalAmount": 125000,
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

**Notes:**
- Each report represents one day's aggregated data
- Reports are sorted by date (newest first)
- `totalAmount` is sum of all successful transactions on that date
- `totalTransactions` counts all transactions (successful + failed)
- Useful for displaying charts and graphs
- `overview` provides summary statistics across all dates in range

**UI Implementation Suggestions:**

```javascript
// Display as line chart
const chartData = reports.map(report => ({
    date: report.date,
    amount: report.totalAmount,
    count: report.totalTransactions
}));

// Display as table
<table>
    <thead>
        <tr>
            <th>Tanggal</th>
            <th>Total Transaksi</th>
            <th>Jumlah</th>
        </tr>
    </thead>
    <tbody>
        {reports.map(report => (
            <tr key={report.date}>
                <td>{formatDate(report.date)}</td>
                <td>{report.totalTransactions}</td>
                <td>{formatCurrency(report.totalAmount, report.currency)}</td>
            </tr>
        ))}
    </tbody>
</table>
```

---

## 4. Get Deposits

Retrieve deposit/top-up transaction history.

**Endpoint:** `GET /v2/deposits`

**Headers:**
```
Authorization: Bearer {access_token}
```

**Query Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| region | string | No | Region code (ID, MY, PH, SG, TH). Default: user's currentRegion |
| limit | integer | No | Items per page. Default: 10, Max: 100 |
| page | integer | No | Page number. Default: 1 |
| search | string | No | Search by invoice number |
| status | string | No | Filter by status: ALL, SUCCESS, PENDING, EXPIRED, FAILED |
| startDate | string | No | Start date (YYYY-MM-DD) |
| endDate | string | No | End date (YYYY-MM-DD) |

**Example Requests:**

```bash
# Get all deposits
GET /v2/deposits?region=ID&limit=10&page=1

# Get successful deposits only
GET /v2/deposits?status=SUCCESS

# Get pending payments
GET /v2/deposits?status=PENDING

# Search by invoice
GET /v2/deposits?search=DEP5E55FF
```

**Response Example:**

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
            },
            {
                "invoiceNumber": "DEP7G77HH33KL44J12196559",
                "status": "SUCCESS",
                "amount": 500000,
                "payment": {
                    "code": "DANA",
                    "name": "DANA"
                },
                "currency": "IDR",
                "createdAt": "2025-09-24T15:20:00+07:00",
                "paidAt": "2025-09-24T15:21:00+07:00"
            },
            {
                "invoiceNumber": "DEP8H88II44LM55K23207660",
                "status": "EXPIRED",
                "amount": 150000,
                "payment": {
                    "code": "MANDIRI_VA",
                    "name": "Mandiri Virtual Account"
                },
                "currency": "IDR",
                "createdAt": "2025-09-23T10:00:00+07:00",
                "expiredAt": "2025-09-24T10:00:00+07:00"
            },
            {
                "invoiceNumber": "DEP9I99JJ55MN66L34318771",
                "status": "FAILED",
                "amount": 300000,
                "payment": {
                    "code": "GOPAY",
                    "name": "GoPay"
                },
                "currency": "IDR",
                "createdAt": "2025-09-22T18:45:00+07:00"
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

**Deposit Status Values:**

| Status | Description | Action |
|--------|-------------|--------|
| SUCCESS | Payment received, balance added | View receipt |
| PENDING | Waiting for payment | Pay now |
| EXPIRED | Payment deadline passed | Create new deposit |
| FAILED | Payment failed | Try again |

**Notes:**
- Deposits are sorted by `createdAt` (newest first)
- `paidAt` only present for SUCCESS status
- `expiredAt` shows payment deadline for PENDING status
- Failed deposits don't affect balance
- Each successful deposit creates a corresponding credit mutation

---

## 5. Create Deposit Inquiry

Pre-validate deposit request before creating actual deposit.

**Endpoint:** `POST /v2/deposits/inquirys`

**Headers:**
```
Authorization: Bearer {access_token}
Content-Type: application/json
```

**Query Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| region | string | No | Region code (ID, MY, PH, SG, TH). Default: user's currentRegion |

**Request Body:**

```json
{
    "amount": 100000,
    "paymentCode": "QRIS"
}
```

**Request Fields:**
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| amount | integer | Yes | Deposit amount (min based on payment method) |
| paymentCode | string | Yes | Payment method code (from /v2/payment-channels) |

**Example Request:**

```bash
POST /v2/deposits/inquirys?region=ID
Content-Type: application/json

{
    "amount": 100000,
    "paymentCode": "QRIS"
}
```

**Response Example:**

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

**Error Responses:**

```json
// Amount below minimum
{
    "error": {
        "code": "AMOUNT_TOO_LOW",
        "message": "Jumlah deposit terlalu kecil",
        "details": "Minimum deposit untuk QRIS adalah Rp 1.000"
    }
}

// Amount above maximum
{
    "error": {
        "code": "AMOUNT_TOO_HIGH",
        "message": "Jumlah deposit terlalu besar",
        "details": "Maximum deposit untuk QRIS adalah Rp 10.000.000"
    }
}

// Invalid payment method
{
    "error": {
        "code": "INVALID_PAYMENT_METHOD",
        "message": "Metode pembayaran tidak valid",
        "details": "Payment method 'INVALID' not found"
    }
}
```

**Notes:**
- `validationToken` valid for 5 minutes
- Shows exact fees and total amount before user confirms
- Payment fee calculated: `feeAmount + (amount * feePercentage / 100)`
- Different payment methods have different min/max limits

---

## 6. Create Deposit

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

**Response Example (QRIS):**

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

**Response Example (Virtual Account):**

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
                "instruction": "<ol><li>Pilih m-Transfer > BCA Virtual Account</li><li>Masukkan nomor Virtual Account</li><li>Periksa informasi yang tertera di layar</li><li>Masukkan PIN m-BCA</li><li>Transaksi selesai</li></ol>",
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

**Response Example (E-Wallet Redirect):**

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
                "instruction": "<ol><li>Kamu akan diarahkan ke aplikasi DANA</li><li>Pastikan saldo DANA mencukupi</li><li>Konfirmasi pembayaran dengan PIN DANA</li><li>Transaksi selesai</li></ol>",
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

**Error Responses:**

```json
// Token expired
{
    "error": {
        "code": "TOKEN_EXPIRED",
        "message": "Validation token telah kadaluarsa",
        "details": "Please create a new deposit inquiry"
    }
}

// Token invalid
{
    "error": {
        "code": "INVALID_TOKEN",
        "message": "Validation token tidak valid",
        "details": "The provided validation token is invalid or has been tampered with"
    }
}
```

**Notes:**
- After successful creation, status is PENDING until payment received
- Different payment methods return different payment data structures
- QRIS: includes QR code string and image URL
- Virtual Account: includes account number and bank details
- E-Wallet: includes redirect URL for payment
- Payment expiry varies by method (QRIS: 1 hour, VA: 24 hours)

---

## 📱 Frontend Implementation Guide

### 1. Transaction History Page

```javascript
import { useState, useEffect } from 'react';

const TransactionsPage = () => {
    const [transactions, setTransactions] = useState([]);
    const [overview, setOverview] = useState(null);
    const [filters, setFilters] = useState({
        status: 'ALL',
        page: 1,
        limit: 10
    });

    useEffect(() => {
        fetchTransactions();
    }, [filters]);

    const fetchTransactions = async () => {
        const params = new URLSearchParams(filters);
        const response = await fetch(`/v2/transactions?${params}`, {
            headers: {
                'Authorization': `Bearer ${accessToken}`
            }
        });
        
        const { data } = await response.json();
        setTransactions(data.transactions);
        setOverview(data.overview);
    };

    return (
        <div>
            {/* Overview Cards */}
            <div className="grid grid-cols-4 gap-4">
                <StatCard 
                    title="Total Transaksi" 
                    value={overview?.totalTransaction}
                    icon="📊"
                />
                <StatCard 
                    title="Sukses" 
                    value={overview?.success}
                    icon="✅"
                    color="green"
                />
                <StatCard 
                    title="Proses" 
                    value={overview?.processing}
                    icon="⏳"
                    color="yellow"
                />
                <StatCard 
                    title="Pending" 
                    value={overview?.pending}
                    icon="⏸️"
                    color="orange"
                />
            </div>

            {/* Filters */}
            <div className="filters">
                <select 
                    value={filters.status}
                    onChange={(e) => setFilters({...filters, status: e.target.value})}
                >
                    <option value="ALL">Semua Status</option>
                    <option value="SUCCESS">Sukses</option>
                    <option value="PROCESSING">Proses</option>
                    <option value="PENDING">Pending</option>
                    <option value="FAILED">Gagal</option>
                </select>
            </div>

            {/* Transaction List */}
            <table>
                <thead>
                    <tr>
                        <th>No. Invoice</th>
                        <th>ID/TRX</th>
                        <th>Item</th>
                        <th>User Input</th>
                        <th>Harga</th>
                        <th>Tanggal</th>
                        <th>Status</th>
                    </tr>
                </thead>
                <tbody>
                    {transactions.map(tx => (
                        <tr key={tx.invoiceNumber}>
                            <td>{tx.invoiceNumber}</td>
                            <td>{tx.skuCode}</td>
                            <td>
                                <div>{tx.productName}</div>
                                <small>{tx.skuName}</small>
                            </td>
                            <td>{tx.account.inputs}</td>
                            <td>{formatCurrency(tx.pricing.total, tx.pricing.currency)}</td>
                            <td>{formatDate(tx.createdAt)}</td>
                            <td>
                                <StatusBadge status={tx.status} />
                            </td>
                        </tr>
                    ))}
                </tbody>
            </table>

            {/* Pagination */}
            <Pagination {...pagination} onChange={handlePageChange} />
        </div>
    );
};
```

### 2. Balance Page with Top-Up

```javascript
const BalancePage = () => {
    const [deposits, setDeposits] = useState([]);
    const [showTopUpModal, setShowTopUpModal] = useState(false);

    const handleTopUp = async (amount, paymentCode) => {
        try {
            // Step 1: Create inquiry
            const inquiryRes = await fetch(`/v2/deposits/inquirys?region=${currentRegion}`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${accessToken}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ amount, paymentCode })
            });

            const { data: inquiryData } = await inquiryRes.json();
            
            // Show confirmation with fees
            const confirmed = await showConfirmation(
                `Total pembayaran: ${formatCurrency(inquiryData.deposit.pricing.total, 'IDR')}`
            );

            if (!confirmed) return;

            // Step 2: Create deposit
            const depositRes = await fetch('/v2/deposits', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${accessToken}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    validationToken: inquiryData.validationToken
                })
            });

            const { data: depositData } = await depositRes.json();

            // Handle payment based on method
            if (depositData.deposit.payment.qrCode) {
                // Show QR Code
                showQRCode(depositData.deposit.payment.qrCodeImage);
            } else if (depositData.deposit.payment.redirectUrl) {
                // Redirect to e-wallet
                window.location.href = depositData.deposit.payment.redirectUrl;
            } else if (depositData.deposit.payment.accountNumber) {
                // Show VA number
                showVADetails(depositData.deposit.payment);
            }

        } catch (error) {
            showError(error.message);
        }
    };

    return (
        <div>
            {/* Balance Card */}
            <div className="balance-card">
                <h2>Saldo Anda Saat Ini: {currentRegion}</h2>
                <h1>{formatCurrency(balance[currency], currency)}</h1>
                <button onClick={() => setShowTopUpModal(true)}>
                    Topup Balance
                </button>
            </div>

            {/* Deposit History */}
            <table>
                <thead>
                    <tr>
                        <th>No. Invoice</th>
                        <th>Tanggal</th>
                        <th>Harga</th>
                        <th>Metode Pembayaran</th>
                        <th>Status</th>
                    </tr>
                </thead>
                <tbody>
                    {deposits.map(deposit => (
                        <tr key={deposit.invoiceNumber}>
                            <td>{deposit.invoiceNumber}</td>
                            <td>{formatDate(deposit.createdAt)}</td>
                            <td>{formatCurrency(deposit.amount, deposit.currency)}</td>
                            <td>{deposit.payment.name}</td>
                            <td>
                                <StatusBadge status={deposit.status} />
                            </td>
                        </tr>
                    ))}
                </tbody>
            </table>

            {/* Top-Up Modal */}
            {showTopUpModal && (
                <TopUpModal 
                    onSubmit={handleTopUp}
                    onClose={() => setShowTopUpModal(false)}
                />
            )}
        </div>
    );
};
```

### 3. Reports with Chart

```javascript
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip } from 'recharts';

const ReportsPage = () => {
    const [reports, setReports] = useState([]);
    const [dateRange, setDateRange] = useState({
        startDate: '2025-09-01',
        endDate: '2025-09-30'
    });

    const fetchReports = async () => {
        const params = new URLSearchParams(dateRange);
        const response = await fetch(`/v2/reports?${params}`, {
            headers: {
                'Authorization': `Bearer ${accessToken}`
            }
        });
        
        const { data } = await response.json();
        setReports(data.reports);
    };

    return (
        <div>
            {/* Date Range Picker */}
            <DateRangePicker 
                value={dateRange}
                onChange={setDateRange}
            />

            {/* Chart */}
            <LineChart width={800} height={400} data={reports}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="date" />
                <YAxis />
                <Tooltip />
                <Line 
                    type="monotone" 
                    dataKey="totalAmount" 
                    stroke="#8884d8"
                    name="Total Transaksi"
                />
            </LineChart>

            {/* Table */}
            <table>
                <thead>
                    <tr>
                        <th>Tanggal</th>
                        <th>Total Transaksi</th>
                        <th>Jumlah</th>
                    </tr>
                </thead>
                <tbody>
                    {reports.map(report => (
                        <tr key={report.date}>
                            <td>{formatDate(report.date)}</td>
                            <td>{report.totalTransactions}</td>
                            <td>{formatCurrency(report.totalAmount, report.currency)}</td>
                        </tr>
                    ))}
                </tbody>
            </table>
        </div>
    );
};
```

---

## 🔄 Complete User Flow Examples

### Flow 1: User Checks Transaction History

```
1. User clicks "Transaksi" menu
2. GET /v2/transactions?region=ID&limit=10&page=1
3. Display overview cards and transaction table
4. User filters by status "SUCCESS"
5. GET /v2/transactions?status=SUCCESS
6. Display filtered results
```

### Flow 2: User Top-Up Balance

```
1. User clicks "Topup Balance" button
2. User selects amount: Rp 100.000
3. User selects payment: QRIS
4. POST /v2/deposits/inquirys
   → Shows: Subtotal + Fee = Total
5. User confirms
6. POST /v2/deposits
   → Returns QR code
7. User scans QR with mobile banking
8. Payment received (webhook updates status)
9. Balance updated automatically
10. User sees success notification
```

### Flow 3: User Views Reports

```
1. User clicks "Laporan" menu
2. GET /v2/reports?startDate=2025-09-01&endDate=2025-09-30
3. Display chart and table
4. User changes date range
5. GET /v2/reports with new dates
6. Chart updates automatically
```

---

## ✅ Summary

### Endpoints Overview

| Endpoint | Method | Purpose | Auth Required |
|----------|--------|---------|---------------|
| `/v2/transactions` | GET | Get transaction history | ✅ Yes |
| `/v2/mutations` | GET | Get balance mutations | ✅ Yes |
| `/v2/reports` | GET | Get daily reports | ✅ Yes |
| `/v2/deposits` | GET | Get deposit history | ✅ Yes |
| `/v2/deposits/inquirys` | POST | Validate deposit request | ✅ Yes |
| `/v2/deposits` | POST | Create deposit | ✅ Yes |

### Key Features

1. ✅ **Multi-Region Support** - All endpoints support region switching
2. ✅ **Multi-Currency** - Amounts displayed in region's currency
3. ✅ **Pagination** - All list endpoints support pagination
4. ✅ **Filtering** - Status, date range, search by invoice
5. ✅ **Overview Statistics** - Summary cards for quick insights
6. ✅ **Export Data** - CSV/XLSX export buttons (client-side)

### Response Patterns

```typescript
// Standard list response
{
    "data": {
        "overview": { /* summary stats */ },
        "items": [ /* array of items */ ],
        "pagination": { /* pagination info */ }
    }
}

// Standard create response
{
    "data": {
        "step": "SUCCESS",
        "item": { /* created item */ }
    }
}
```

---

**Ready for Production!** 🚀

All dashboard endpoints are complete with:
- ✅ Comprehensive filtering
- ✅ Pagination support
- ✅ Multi-region/currency
- ✅ Overview statistics
- ✅ Clear status indicators
- ✅ Export capabilities
- ✅ Real-time updates via webhooks
