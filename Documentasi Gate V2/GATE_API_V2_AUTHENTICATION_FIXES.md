# Gate API v2.0 - Authentication Changes & Clarifications

## 🔧 Perubahan yang Dilakukan

### 1. ✅ Phone Number Format - E.164 Standard

**FIXED: Sekarang pakai `+` di awal**

```json
// ❌ SEBELUM (Salah)
{
    "phoneNumber": "628123456789"
}

// ✅ SEKARANG (Benar)
{
    "phoneNumber": "+628123456789"
}
```

**Format yang Benar:**
- Indonesia: `+628123456789`
- Malaysia: `+60123456789`
- Philippines: `+63912345678`
- Singapore: `+6591234567`
- Thailand: `+66812345678`

**Kenapa pakai `+`?**
- Ini adalah E.164 standard yang benar
- Digunakan oleh WhatsApp, Telegram, dan semua messaging apps
- Mudah di-parse oleh library international phone number
- Konsisten dengan standar internasional

---

### 2. ✅ Register dengan Google - Tidak Perlu Phone Number

**FIXED: Hapus phoneNumber dari request body**

**Endpoint:** `POST /v2/auth/register/google`

**Request Body SEBELUM:**
```json
{
    "idToken": "eyJhbGciOiJSUzI1NiIsImtpZCI6IjZmODI4...",
    "phoneNumber": "628123456789"  // ❌ Tidak perlu!
}
```

**Request Body SEKARANG:**
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
        "user": {
            "phoneNumber": null,  // ✅ Null dulu, bisa diisi nanti
            "status": "ACTIVE",
            "googleId": "117562748392847562"
        }
    }
}
```

**Alasan:**
- Google register harusnya quick & simple
- User bisa tambah phone number nanti di profile settings
- Tidak semua user mau kasih phone number saat register
- Phone number optional untuk transaksi (bisa isi saat checkout)

---

### 3. ✅ Verify Email - Tidak Perlu Email di Body

**FIXED: Token sudah encode email-nya**

**Email Link:**
```
https://gate.co.id/verify-email/{token}
```

**Token Contains:**
```json
{
    "email": "john.doe@example.com",
    "userId": "usr_1a2b3c4d5e6f",
    "exp": 1733280000
}
```

**Endpoint:** `POST /v2/auth/verify-email`

**Request Body SEBELUM:**
```json
{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "email": "john.doe@example.com"  // ❌ Redundant!
}
```

**Request Body SEKARANG:**
```json
{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Backend Process:**
1. User clicks link di email
2. Frontend extract token dari URL
3. Frontend call `POST /v2/auth/verify-email` dengan token
4. Backend decode token untuk dapat email
5. Verify email dan activate account

**Alasan:**
- Token JWT sudah berisi email (encrypted & signed)
- Tidak perlu user input email lagi
- Lebih secure (user tidak bisa verify email orang lain)
- UX lebih baik (one-click verification)

---

### 4. ✅ MFA Endpoints - Perlu Authentication

**FIXED: Tambah penjelasan bahwa perlu login**

**Semua MFA endpoints ini HARUS authenticated:**

#### Enable MFA
```
POST /v2/auth/mfa/enable
Authorization: Bearer {access_token}  // ✅ REQUIRED
```

#### Verify MFA Setup
```
POST /v2/auth/mfa/verify-setup
Authorization: Bearer {access_token}  // ✅ REQUIRED
```

#### Disable MFA
```
POST /v2/auth/mfa/disable
Authorization: Bearer {access_token}  // ✅ REQUIRED
```

**Kenapa perlu authentication?**
- Setup MFA = modify user security settings
- Hanya user yang login bisa enable/disable MFA sendiri
- Prevent unauthorized MFA changes
- Standard security practice

**User Flow:**
```
1. User login → dapat access_token
2. User masuk Settings → Security
3. Click "Enable MFA"
4. Frontend call: POST /v2/auth/mfa/enable (dengan Bearer token)
5. Backend return QR code + backup codes
6. User scan QR dengan Google Authenticator
7. User input 6-digit code
8. Frontend call: POST /v2/auth/mfa/verify-setup (dengan Bearer token)
9. MFA enabled! ✅
```

**Disable MFA Flow:**
```
1. User sudah login (MFA enabled)
2. User masuk Settings → Security
3. Click "Disable MFA"
4. User input: password + current MFA code
5. Frontend call: POST /v2/auth/mfa/disable (dengan Bearer token)
6. MFA disabled ✅
```

---

## 📋 Complete Authentication Flow

### Flow 1: Regular Registration

```mermaid
User Register
    ↓
POST /v2/auth/register
    ↓
step: EMAIL_VERIFICATION
    ↓
Email sent to user
    ↓
User clicks: https://gate.co.id/verify-email/{token}
    ↓
Frontend extract token from URL
    ↓
POST /v2/auth/verify-email { "token": "..." }
    ↓
step: SUCCESS + access_token
    ↓
User logged in ✅
```

### Flow 2: Google Registration

```mermaid
User clicks "Sign up with Google"
    ↓
Google OAuth popup
    ↓
POST /v2/auth/register/google { "idToken": "..." }
    ↓
step: SUCCESS + access_token
    ↓
User logged in ✅ (no email verification needed)
    ↓
phoneNumber: null (can add later in profile)
```

### Flow 3: Regular Login (No MFA)

```mermaid
User login
    ↓
POST /v2/auth/login
    ↓
step: SUCCESS + access_token
    ↓
User logged in ✅
```

### Flow 4: Regular Login (With MFA)

```mermaid
User login
    ↓
POST /v2/auth/login
    ↓
step: MFA_VERIFICATION + mfaToken
    ↓
User input MFA code
    ↓
POST /v2/auth/verify-mfa
    ↓
step: SUCCESS + access_token
    ↓
User logged in ✅
```

### Flow 5: Google Login

```mermaid
User clicks "Sign in with Google"
    ↓
Google OAuth popup
    ↓
POST /v2/auth/login/google
    ↓
step: SUCCESS + access_token
    ↓
User logged in ✅ (no MFA needed)
```

### Flow 6: Enable MFA

```mermaid
User logged in (has access_token)
    ↓
POST /v2/auth/mfa/enable
Authorization: Bearer {access_token}
    ↓
Get QR code + secret + backup codes
    ↓
User scan QR with Google Authenticator
    ↓
User input 6-digit code
    ↓
POST /v2/auth/mfa/verify-setup
Authorization: Bearer {access_token}
    ↓
MFA enabled ✅
    ↓
Next login requires MFA code
```

---

## 🔐 Security Best Practices

### 1. Token Management

**Access Token:**
- Expiry: 1 hour
- Storage: Memory (not localStorage for security)
- Usage: All authenticated endpoints

**Refresh Token:**
- Expiry: 30 days
- Storage: HttpOnly cookie (secure)
- Usage: Get new access token

**MFA Token:**
- Expiry: 5 minutes
- Storage: Memory only
- Usage: Verify MFA during login

### 2. Phone Number Validation

**Frontend:**
```javascript
// Use international phone number library
import { parsePhoneNumber } from 'libphonenumber-js'

const phoneNumber = parsePhoneNumber(input, 'ID')
if (phoneNumber && phoneNumber.isValid()) {
    const formatted = phoneNumber.format('E.164') // "+628123456789"
    // Send to API
}
```

**Backend:**
```go
// Validate E.164 format
func validatePhoneNumber(phone string) bool {
    // Must start with +
    // Must have country code (1-3 digits)
    // Must have national number (4-14 digits)
    regex := `^\+[1-9]\d{1,14}$`
    return regexp.MustCompile(regex).MatchString(phone)
}
```

### 3. Email Verification Token

**Token Structure:**
```json
{
    "email": "john.doe@example.com",
    "userId": "usr_1a2b3c4d5e6f",
    "type": "EMAIL_VERIFICATION",
    "exp": 1733280000  // 24 hours from now
}
```

**Email Template:**
```html
<p>Click the button below to verify your email:</p>
<a href="https://gate.co.id/verify-email/eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...">
    Verify Email
</a>
```

**Frontend Handling:**
```javascript
// Extract token from URL
const params = new URLSearchParams(window.location.search);
const token = window.location.pathname.split('/').pop(); // Get token from path

// Call API
const response = await fetch('/v2/auth/verify-email', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ token })
});
```

---

## 📊 Endpoint Summary

### Public Endpoints (No Auth Required)

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/v2/auth/register` | POST | Register dengan email |
| `/v2/auth/register/google` | POST | Register dengan Google |
| `/v2/auth/verify-email` | POST | Verify email setelah register |
| `/v2/auth/resend-verification` | POST | Kirim ulang email verifikasi |
| `/v2/auth/login` | POST | Login dengan email/password |
| `/v2/auth/login/google` | POST | Login dengan Google |
| `/v2/auth/verify-mfa` | POST | Verify MFA code saat login |
| `/v2/auth/forgot-password` | POST | Request reset password |
| `/v2/auth/reset-password` | POST | Reset password dengan token |
| `/v2/auth/refresh` | POST | Refresh access token |

### Protected Endpoints (Auth Required)

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/v2/auth/mfa/enable` | POST | Enable MFA (get QR code) |
| `/v2/auth/mfa/verify-setup` | POST | Verify MFA setup |
| `/v2/auth/mfa/disable` | POST | Disable MFA |
| `/v2/auth/logout` | POST | Logout user |
| `/v2/user/profile` | GET | Get user profile |
| `/v2/user/profile` | PUT | Update user profile |
| `/v2/user/change-password` | POST | Change password |

---

## 🎯 Key Takeaways

### ✅ BENAR (After Fixes)

1. **Phone Number:** `+628123456789` (dengan +)
2. **Google Register:** Tidak perlu phoneNumber
3. **Verify Email:** Hanya perlu token (email sudah di dalam token)
4. **MFA Setup:** Perlu authentication (user harus login)

### ❌ SALAH (Before Fixes)

1. **Phone Number:** `628123456789` (tanpa +)
2. **Google Register:** Perlu phoneNumber di request
3. **Verify Email:** Perlu email di request body
4. **MFA Setup:** Dokumentasi tidak jelas perlu auth atau tidak

---

## 💡 Implementation Tips

### Frontend Tips

```javascript
// 1. Phone Number Input with Country Selector
import PhoneInput from 'react-phone-number-input'

<PhoneInput
    defaultCountry="ID"
    value={phoneNumber}
    onChange={setPhoneNumber}  // Auto formats to E.164
/>
// Output: "+628123456789" ✅

// 2. Email Verification Link Handler
useEffect(() => {
    const token = window.location.pathname.split('/verify-email/')[1];
    if (token) {
        verifyEmail(token);
    }
}, []);

// 3. MFA Setup
const enableMFA = async () => {
    const response = await fetch('/v2/auth/mfa/enable', {
        headers: {
            'Authorization': `Bearer ${accessToken}`  // ✅ Required
        }
    });
    const { qrCode, secret, backupCodes } = await response.json();
    // Show QR code to user
};

// 4. Google OAuth
const handleGoogleLogin = async (googleResponse) => {
    const response = await fetch('/v2/auth/login/google', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            idToken: googleResponse.credential
        })
    });
    // No phoneNumber needed! ✅
};
```

### Backend Tips

```go
// 1. Validate Phone Number
func ValidatePhoneNumber(phone string) error {
    if !strings.HasPrefix(phone, "+") {
        return errors.New("phone number must start with +")
    }
    
    // Use go library: github.com/nyaruka/phonenumbers
    num, err := phonenumbers.Parse(phone, "")
    if err != nil {
        return err
    }
    
    if !phonenumbers.IsValidNumber(num) {
        return errors.New("invalid phone number")
    }
    
    return nil
}

// 2. Generate Email Verification Token
func GenerateEmailVerificationToken(email, userID string) (string, error) {
    claims := jwt.MapClaims{
        "email":  email,
        "userId": userID,
        "type":   "EMAIL_VERIFICATION",
        "exp":    time.Now().Add(24 * time.Hour).Unix(),
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(secret))
}

// 3. Verify Email Handler
func VerifyEmail(token string) error {
    claims, err := jwt.Parse(token)
    if err != nil {
        return err
    }
    
    email := claims["email"].(string)
    // Email is already in token, no need to get from request body!
    
    // Update user status to ACTIVE
    return db.UpdateUserStatus(email, "ACTIVE")
}

// 4. MFA Middleware
func RequireAuth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "Unauthorized", 401)
            return
        }
        
        // Validate token
        claims, err := ValidateToken(token)
        if err != nil {
            http.Error(w, "Invalid token", 401)
            return
        }
        
        // Add user to context
        ctx := context.WithValue(r.Context(), "userId", claims["userId"])
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Apply middleware to MFA routes
mux.Handle("/v2/auth/mfa/enable", RequireAuth(enableMFAHandler))
mux.Handle("/v2/auth/mfa/verify-setup", RequireAuth(verifyMFASetupHandler))
mux.Handle("/v2/auth/mfa/disable", RequireAuth(disableMFAHandler))
```

---

## 🚀 Ready for Production!

Dokumentasi API sudah diperbaiki dan siap digunakan untuk development. Semua endpoint sudah jelas, konsisten, dan mengikuti best practices.

**File Location:** `/mnt/user-data/outputs/GATE_API_DOCUMENTATION_V2_COMPLETE.md`
