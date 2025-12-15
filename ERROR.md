# Error Responses Documentation

Dokumentasi lengkap untuk semua error responses pada endpoint-endpoint berikut:
- `POST /v2/account/inquiries`
- `POST /v2/promos/validate`
- `POST /v2/orders/inquiries`
- `POST /v2/orders`

---

## POST /v2/account/inquiries

Endpoint untuk validasi account sebelum transaksi.

### Error Responses

#### 1. Invalid Request Body
**Status Code:** `400 Bad Request`

```json
{
  "error": {
    "code": "BAD_REQUEST",
    "message": "Invalid request body"
  }
}
```

**Kondisi:** Request body tidak valid atau tidak dapat di-parse sebagai JSON.

---

#### 2. Missing Required Fields - Product Code
**Status Code:** `400 Bad Request`

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "fields": {
      "productCode": "Product code is required"
    }
  }
}
```

**Kondisi:** Field `productCode` tidak disediakan dalam request body.

---

#### 3. Missing Required Fields - User ID
**Status Code:** `400 Bad Request`

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "fields": {
      "userId": "User ID is required"
    }
  }
}
```

**Kondisi:** Field `userId` tidak disediakan dalam request body.

---

#### 4. Product Not Found
**Status Code:** `404 Not Found`

```json
{
  "error": {
    "code": "PRODUCT_NOT_FOUND",
    "message": "Product not found",
    "details": ""
  }
}
```

**Kondisi:** Product dengan `productCode` yang diberikan tidak ditemukan atau tidak aktif.

---

#### 5. Inquiry Not Configured
**Status Code:** `400 Bad Request`

```json
{
  "error": {
    "code": "INQUIRY_NOT_CONFIGURED",
    "message": "Inquiry slug is not configured for this product",
    "details": ""
  }
}
```

**Kondisi:** Product tidak memiliki `inquiry_slug` yang dikonfigurasi, sehingga tidak dapat melakukan account inquiry.

---

#### 6. Inquiry Service Connection Error
**Status Code:** `500 Internal Server Error`

```json
{
  "error": {
    "code": "INQUIRY_SERVICE_ERROR",
    "message": "Failed to connect to inquiry service",
    "details": ""
  }
}
```

**Kondisi:** Gagal terhubung ke external inquiry service.

---

#### 7. Inquiry Response Parse Error
**Status Code:** `500 Internal Server Error`

```json
{
  "error": {
    "code": "INQUIRY_RESPONSE_ERROR",
    "message": "Failed to parse inquiry response",
    "details": ""
  }
}
```

**Kondisi:** Response dari inquiry service tidak dapat di-parse.

---

#### 8. Account Not Found
**Status Code:** `404 Not Found`

```json
{
  "error": {
    "code": "ACCOUNT_NOT_FOUND",
    "message": "Account not found",
    "details": "The provided User ID and Zone ID combination does not exist"
  }
}
```

**Kondisi:** Account dengan User ID dan Zone ID yang diberikan tidak ditemukan.

---

#### 9. Bad Request from Inquiry Service
**Status Code:** `400 Bad Request`

```json
{
  "error": {
    "code": "BAD_REQUEST",
    "message": "Invalid request",
    "details": "The request data is invalid. Please check your input."
  }
}
```

**Kondisi:** Inquiry service mengembalikan error `BAD_REQUEST` karena data request tidak valid.

---

#### 10. Too Many Requests
**Status Code:** `429 Too Many Requests`

```json
{
  "error": {
    "code": "TOO_MANY_REQUESTS",
    "message": "Too many requests",
    "details": "Please try again later."
  }
}
```

**Kondisi:** Terlalu banyak request ke inquiry service dalam waktu singkat.

---

#### 11. Internal Server Error from Inquiry Service
**Status Code:** `500 Internal Server Error`

```json
{
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "Internal server error",
    "details": "An error occurred on the inquiry server. Please try again later."
  }
}
```

**Kondisi:** Inquiry service mengalami internal server error.

---

#### 12. Internal Server Error
**Status Code:** `500 Internal Server Error`

```json
{
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "Terjadi kesalahan internal server"
  }
}
```

**Kondisi:** Terjadi kesalahan internal server yang tidak terduga.

---

## POST /v2/promos/validate

Endpoint untuk validasi promo code sebelum order.

### Error Responses

#### 1. Invalid Request Body
**Status Code:** `400 Bad Request`

```json
{
  "error": {
    "code": "BAD_REQUEST",
    "message": "Invalid request body"
  }
}
```

**Kondisi:** Request body tidak valid atau tidak dapat di-parse sebagai JSON.

---

#### 2. Missing Required Fields
**Status Code:** `400 Bad Request`

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "fields": {
      "promoCode": "Promo code is required",
      "productCode": "Product code is required",
      "skuCode": "SKU code is required",
      "paymentCode": "Payment code is required",
      "region": "Region is required"
    }
  }
}
```

**Kondisi:** Salah satu atau lebih field required tidak disediakan.

---

#### 3. Product Not Found
**Status Code:** `404 Not Found`

```json
{
  "error": {
    "code": "PRODUCT_NOT_FOUND",
    "message": "Product not found",
    "details": "The product code does not exist or is inactive"
  }
}
```

**Kondisi:** Product dengan `productCode` yang diberikan tidak ditemukan atau tidak aktif.

---

#### 4. Missing Required Account Fields
**Status Code:** `400 Bad Request`

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "fields": {
      "account": "Missing required account fields: userId, zoneId"
    }
  }
}
```

**Kondisi:** Account object tidak mengandung field-field yang required sesuai product fields.

---

#### 5. Invalid Account Field
**Status Code:** `400 Bad Request`

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "fields": {
      "account": "Invalid account field: invalidKey. Valid fields are: userId, zoneId"
    }
  }
}
```

**Kondisi:** Account object mengandung field yang tidak valid (tidak sesuai dengan product fields).

---

#### 6. SKU Not Found
**Status Code:** `404 Not Found`

```json
{
  "error": {
    "code": "SKU_NOT_FOUND",
    "message": "SKU not found",
    "details": "The SKU code does not exist or is inactive"
  }
}
```

**Kondisi:** SKU dengan `skuCode` yang diberikan tidak ditemukan atau tidak aktif.

---

#### 7. Promo Not Found
**Status Code:** `200 OK` (Success response dengan valid: false)

```json
{
  "data": {
    "valid": false,
    "reason": "PROMO_NOT_FOUND"
  }
}
```

**Kondisi:** Promo code tidak ditemukan di database.

---

#### 8. Promo Not Active
**Status Code:** `200 OK` (Success response dengan valid: false)

```json
{
  "data": {
    "valid": false,
    "reason": "PROMO_NOT_ACTIVE"
  }
}
```

**Kondisi:** Promo code tidak aktif.

---

#### 9. Promo Not Started
**Status Code:** `200 OK` (Success response dengan valid: false)

```json
{
  "data": {
    "valid": false,
    "reason": "PROMO_NOT_STARTED"
  }
}
```

**Kondisi:** Promo code belum dimulai (masih dalam periode `start_at` di masa depan).

---

#### 10. Promo Expired
**Status Code:** `400 Bad Request`

```json
{
  "error": {
    "code": "PROMO_EXPIRED",
    "message": "Kode promo telah kadaluarsa",
    "details": "Promo ini berakhir pada 30 November 2025"
  }
}
```

**Kondisi:** Promo code sudah kadaluarsa (melewati `expired_at`).

---

#### 11. Product Not Applicable
**Status Code:** `400 Bad Request`

```json
{
  "error": {
    "code": "PRODUCT_NOT_APPLICABLE",
    "message": "Promo code is not applicable to this product",
    "details": ""
  }
}
```

**Kondisi:** Promo code memiliki product restrictions dan product yang diberikan tidak ada dalam daftar.

---

#### 12. Payment Not Applicable
**Status Code:** `400 Bad Request`

```json
{
  "error": {
    "code": "PAYMENT_NOT_APPLICABLE",
    "message": "Promo code is not applicable to this payment method",
    "details": ""
  }
}
```

**Kondisi:** Promo code memiliki payment channel restrictions dan payment method yang diberikan tidak ada dalam daftar.

---

#### 13. Region Not Applicable
**Status Code:** `200 OK` (Success response dengan valid: false)

```json
{
  "data": {
    "valid": false,
    "reason": "REGION_NOT_APPLICABLE"
  }
}
```

**Kondisi:** Promo code tidak berlaku untuk region yang diberikan.

---

#### 14. Day Not Applicable
**Status Code:** `200 OK` (Success response dengan valid: false)

```json
{
  "data": {
    "valid": false,
    "reason": "DAY_NOT_APPLICABLE"
  }
}
```

**Kondisi:** Promo code hanya berlaku pada hari-hari tertentu dan hari ini tidak termasuk.

---

#### 15. Minimum Amount Not Met
**Status Code:** `400 Bad Request`

```json
{
  "error": {
    "code": "MIN_AMOUNT_NOT_MET",
    "message": "Minimum amount required: 50000",
    "details": "Current amount: 30000"
  }
}
```

**Kondisi:** Total amount (SKU price × quantity) kurang dari minimum amount yang disyaratkan promo.

---

#### 16. Usage Limit Exceeded
**Status Code:** `200 OK` (Success response dengan valid: false)

```json
{
  "data": {
    "valid": false,
    "reason": "USAGE_LIMIT_EXCEEDED"
  }
}
```

**Kondisi:** Total penggunaan promo sudah mencapai batas maksimum (`max_usage`).

---

#### 17. Daily Usage Limit Exceeded
**Status Code:** `200 OK` (Success response dengan valid: false)

```json
{
  "data": {
    "valid": false,
    "reason": "DAILY_USAGE_LIMIT_EXCEEDED"
  }
}
```

**Kondisi:** Penggunaan promo hari ini sudah mencapai batas maksimum harian (`max_daily_usage`).

---

#### 18. User Usage Limit Exceeded
**Status Code:** `200 OK` (Success response dengan valid: false)

```json
{
  "data": {
    "valid": false,
    "reason": "USER_USAGE_LIMIT_EXCEEDED"
  }
}
```

**Kondisi:** User yang terautentikasi sudah menggunakan promo ini melebihi batas per user (`max_usage_per_id`).

---

#### 19. Device Usage Limit Exceeded
**Status Code:** `200 OK` (Success response dengan valid: false)

```json
{
  "data": {
    "valid": false,
    "reason": "DEVICE_USAGE_LIMIT_EXCEEDED"
  }
}
```

**Kondisi:** Device ID sudah menggunakan promo ini melebihi batas per device (`max_usage_per_device`).

---

#### 20. IP Usage Limit Exceeded
**Status Code:** `200 OK` (Success response dengan valid: false)

```json
{
  "data": {
    "valid": false,
    "reason": "IP_USAGE_LIMIT_EXCEEDED"
  }
}
```

**Kondisi:** IP address sudah menggunakan promo ini melebihi batas per IP (`max_usage_per_ip`).

---

#### 21. Internal Server Error
**Status Code:** `500 Internal Server Error`

```json
{
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "Terjadi kesalahan internal server"
  }
}
```

**Kondisi:** Terjadi kesalahan internal server yang tidak terduga.

---

## POST /v2/orders/inquiries

Endpoint untuk pre-validate order sebelum creation.

### Error Responses

#### 1. Invalid Request Body
**Status Code:** `400 Bad Request`

```json
{
  "error": {
    "code": "BAD_REQUEST",
    "message": "Invalid request body"
  }
}
```

**Kondisi:** Request body tidak valid atau tidak dapat di-parse sebagai JSON.

---

#### 2. Product Code Required
**Status Code:** `400 Bad Request`

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "fields": {
      "productCode": "Product code is required"
    }
  }
}
```

**Kondisi:** Field `productCode` tidak disediakan.

---

#### 3. SKU Code Required
**Status Code:** `400 Bad Request`

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "fields": {
      "skuCode": "SKU code is required"
    }
  }
}
```

**Kondisi:** Field `skuCode` tidak disediakan.

---

#### 4. Product Not Found
**Status Code:** `404 Not Found`

```json
{
  "error": {
    "code": "PRODUCT_NOT_FOUND",
    "message": "Product not found",
    "details": ""
  }
}
```

**Kondisi:** Product dengan `productCode` yang diberikan tidak ditemukan atau tidak aktif.

---

#### 5. SKU Not Found
**Status Code:** `404 Not Found`

```json
{
  "error": {
    "code": "SKU_NOT_FOUND",
    "message": "SKU not found",
    "details": ""
  }
}
```

**Kondisi:** SKU dengan `skuCode` yang diberikan tidak ditemukan atau tidak aktif.

---

#### 6. Zone ID / Server ID Required
**Status Code:** `400 Bad Request`

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "fields": {
      "zoneId": "Zone ID / Server ID is required for this product",
      "serverId": "Zone ID / Server ID is required for this product"
    }
  }
}
```

**Kondisi:** Product memerlukan zone ID atau server ID, tetapi tidak disediakan.

---

#### 7. Invalid Email Format
**Status Code:** `400 Bad Request`

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "fields": {
      "email": "Invalid email format"
    }
  }
}
```

**Kondisi:** Format email tidak valid.

---

#### 8. Invalid Phone Number Format
**Status Code:** `400 Bad Request`

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "fields": {
      "phoneNumber": "Invalid phone number format. Please use international format (e.g., +628123456789)"
    }
  }
}
```

**Kondisi:** Format nomor telepon tidak valid. Harus menggunakan format internasional.

---

#### 9. User ID or Phone Number Required
**Status Code:** `400 Bad Request`

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "fields": {
      "userId": "User ID or phone number is required",
      "phoneNumber": "User ID or phone number is required"
    }
  }
}
```

**Kondisi:** Baik `userId` maupun `phoneNumber` tidak disediakan.

---

#### 10. Inquiry Service Connection Error
**Status Code:** `500 Internal Server Error`

```json
{
  "error": {
    "code": "INQUIRY_SERVICE_ERROR",
    "message": "Failed to connect to inquiry service",
    "details": ""
  }
}
```

**Kondisi:** Gagal terhubung ke external inquiry service.

---

#### 11. Inquiry Response Parse Error
**Status Code:** `500 Internal Server Error`

```json
{
  "error": {
    "code": "INQUIRY_RESPONSE_ERROR",
    "message": "Failed to parse inquiry response",
    "details": ""
  }
}
```

**Kondisi:** Response dari inquiry service tidak dapat di-parse.

---

#### 12. Account Not Found
**Status Code:** `404 Not Found`

```json
{
  "error": {
    "code": "ACCOUNT_NOT_FOUND",
    "message": "Account not found",
    "details": "The provided User ID and Zone ID combination does not exist"
  }
}
```

**Kondisi:** Account dengan User ID dan Zone ID yang diberikan tidak ditemukan.

---

#### 13. Promo Code Not Found
**Status Code:** `400 Bad Request`

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "fields": {
      "promoCode": "Promo code not found"
    }
  }
}
```

**Kondisi:** Promo code yang diberikan tidak ditemukan di database.

---

#### 14. Promo Code Not Active
**Status Code:** `400 Bad Request`

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "fields": {
      "promoCode": "Promo code is not active"
    }
  }
}
```

**Kondisi:** Promo code tidak aktif.

---

#### 15. Promo Code Not Started
**Status Code:** `400 Bad Request`

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "fields": {
      "promoCode": "Promo code has not started yet"
    }
  }
}
```

**Kondisi:** Promo code belum dimulai.

---

#### 16. Promo Code Expired
**Status Code:** `400 Bad Request`

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "fields": {
      "promoCode": "Promo code has expired"
    }
  }
}
```

**Kondisi:** Promo code sudah kadaluarsa.

---

#### 17. Promo Not Applicable to Product
**Status Code:** `400 Bad Request`

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "fields": {
      "promoCode": "Promo code is not applicable to this product"
    }
  }
}
```

**Kondisi:** Promo code tidak berlaku untuk product yang dipilih.

---

#### 18. Promo Not Applicable to Payment Method
**Status Code:** `400 Bad Request`

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "fields": {
      "promoCode": "Promo code is not applicable to this payment method"
    }
  }
}
```

**Kondisi:** Promo code tidak berlaku untuk payment method yang dipilih.

---

#### 19. Payment Method Not Found
**Status Code:** `400 Bad Request`

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "fields": {
      "paymentCode": "Payment method not found"
    }
  }
}
```

**Kondisi:** Payment channel dengan `paymentCode` yang diberikan tidak ditemukan.

---

#### 20. Payment Method Not Active
**Status Code:** `400 Bad Request`

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "fields": {
      "paymentCode": "Payment method is not active"
    }
  }
}
```

**Kondisi:** Payment channel tidak aktif.

---

#### 21. Internal Server Error
**Status Code:** `500 Internal Server Error`

```json
{
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "Terjadi kesalahan internal server"
  }
}
```

**Kondisi:** Terjadi kesalahan internal server yang tidak terduga.

---

## POST /v2/orders

Endpoint untuk membuat order dengan validation token.

### Error Responses

#### 1. Invalid Request Body
**Status Code:** `400 Bad Request`

```json
{
  "error": {
    "code": "BAD_REQUEST",
    "message": "Invalid request body"
  }
}
```

**Kondisi:** Request body tidak valid atau tidak dapat di-parse sebagai JSON.

---

#### 2. Validation Token Required
**Status Code:** `400 Bad Request`

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "fields": {
      "validationToken": "Validation token is required"
    }
  }
}
```

**Kondisi:** Field `validationToken` tidak disediakan.

---

#### 3. Invalid or Expired Token
**Status Code:** `400 Bad Request`

```json
{
  "error": {
    "code": "INVALID_TOKEN",
    "message": "Invalid or expired validation token",
    "details": "Please create a new order inquiry"
  }
}
```

**Kondisi:** Validation token tidak valid atau sudah kadaluarsa.

---

#### 4. Token Already Used
**Status Code:** `400 Bad Request`

```json
{
  "error": {
    "code": "TOKEN_ALREADY_USED",
    "message": "Validation token has already been used",
    "details": "Please create a new order inquiry"
  }
}
```

**Kondisi:** Validation token sudah pernah digunakan untuk membuat order.

---

#### 5. Product Not Found
**Status Code:** `404 Not Found`

```json
{
  "error": {
    "code": "PRODUCT_NOT_FOUND",
    "message": "Product not found or not available in your region",
    "details": ""
  }
}
```

**Kondisi:** Product tidak ditemukan atau tidak tersedia di region yang dipilih.

---

#### 6. SKU Not Found
**Status Code:** `404 Not Found`

```json
{
  "error": {
    "code": "SKU_NOT_FOUND",
    "message": "SKU not found",
    "details": ""
  }
}
```

**Kondisi:** SKU tidak ditemukan atau tidak aktif.

---

#### 7. Payment Channel Not Found
**Status Code:** `404 Not Found`

```json
{
  "error": {
    "code": "PAYMENT_CHANNEL_NOT_FOUND",
    "message": "Payment channel not found",
    "details": ""
  }
}
```

**Kondisi:** Payment channel dengan `paymentCode` yang diberikan tidak ditemukan.

---

#### 8. Authentication Required (Balance Payment)
**Status Code:** `401 Unauthorized`

```json
{
  "error": {
    "code": "AUTHENTICATION_REQUIRED",
    "message": "Balance payment requires authentication",
    "details": ""
  }
}
```

**Kondisi:** Payment method adalah BALANCE tetapi user tidak terautentikasi.

---

#### 9. Insufficient Balance
**Status Code:** `400 Bad Request`

```json
{
  "error": {
    "code": "INSUFFICIENT_BALANCE",
    "message": "Insufficient balance",
    "details": "Please top up your balance or use another payment method"
  }
}
```

**Kondisi:** Saldo user tidak cukup untuk melakukan pembayaran dengan balance.

---

#### 10. Payment Gateway Unavailable
**Status Code:** `503 Service Unavailable`

```json
{
  "error": {
    "code": "PAYMENT_GATEWAY_UNAVAILABLE",
    "message": "Payment gateway is not available",
    "details": "Please try again later or use a different payment method"
  }
}
```

**Kondisi:** Payment gateway tidak tersedia atau tidak dapat diakses.

---

#### 11. Payment Gateway Error
**Status Code:** `502 Bad Gateway`

```json
{
  "error": {
    "code": "PAYMENT_GATEWAY_ERROR",
    "message": "Failed to create payment",
    "details": "Error message from payment gateway"
  }
}
```

**Kondisi:** Gagal membuat payment di payment gateway (error dari gateway).

---

#### 12. Internal Server Error
**Status Code:** `500 Internal Server Error`

```json
{
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "Terjadi kesalahan internal server"
  }
}
```

**Kondisi:** Terjadi kesalahan internal server yang tidak terduga (database error, transaction error, dll).

---

## Catatan

1. **Status Code 200 dengan valid: false**: Beberapa error pada `/v2/promos/validate` mengembalikan status code 200 dengan response success yang berisi `valid: false` dan `reason`. Ini adalah design choice untuk membedakan antara error yang fatal (400/404/500) dengan validasi promo yang gagal.

2. **Validation Error Format**: Error dengan code `VALIDATION_ERROR` selalu mengembalikan object `fields` yang berisi field-field yang error beserta pesan errornya.

3. **Details Field**: Field `details` pada error response bersifat optional dan dapat berisi informasi tambahan yang membantu user memahami error.

4. **Internal Server Error**: Error dengan code `INTERNAL_ERROR` biasanya terjadi karena masalah di server (database connection, unexpected error, dll) dan tidak seharusnya terjadi dalam kondisi normal.

