# Gate API v2.0 - Multi-Region & Multi-Currency Support

## 🌍 Overview

Gate.co.id adalah platform multi-region dan multi-currency yang support 5 negara:
- **Indonesia (ID)** - Currency: IDR
- **Malaysia (MY)** - Currency: MYR
- **Philippines (PH)** - Currency: PHP
- **Singapore (SG)** - Currency: SGD
- **Thailand (TH)** - Currency: THB

Setiap user memiliki:
1. **Primary Region** - Region saat pertama kali register
2. **Current Region** - Region yang sedang aktif (bisa ganti-ganti)
3. **Multi-Currency Balance** - Saldo dalam semua currencies

---

## 📋 Updated Authentication Endpoints

### 1. Register User (with Region)

**Endpoint:** `POST /v2/auth/register?region={region_code}`

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

**Example Requests:**
```bash
# Register as Indonesian user
POST /v2/auth/register?region=ID

# Register as Malaysian user
POST /v2/auth/register?region=MY

# Register as Singaporean user
POST /v2/auth/register?region=SG
```

**Response Example:**
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

**Notes:**
- `primaryRegion` is set based on query parameter (or default ID)
- `currentRegion` initially same as `primaryRegion`
- `currency` is set based on region (IDR for ID, MYR for MY, etc.)
- User gets balance in ALL currencies (initialized to 0)
- Primary region cannot be changed later (permanent)

---

### 2. Register with Google (with Region)

**Endpoint:** `POST /v2/auth/register/google?region={region_code}`

**Query Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| region | string | No | Region code (ID, MY, PH, SG, TH). Default: ID |

**Request Body:**
```json
{
    "idToken": "eyJhbGciOiJSUzI1NiIsImtpZCI6IjZmODI4..."
}
```

**Example Request:**
```bash
POST /v2/auth/register/google?region=MY
```

**Response Example:**
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
            "primaryRegion": "MY",
            "currentRegion": "MY",
            "currency": "MYR",
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

### 3. Login (with Region)

**Endpoint:** `POST /v2/auth/login?region={region_code}`

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

**Example Requests:**
```bash
# Login and set current region to Indonesia
POST /v2/auth/login?region=ID

# Login and set current region to Malaysia
POST /v2/auth/login?region=MY

# Login without region (uses primaryRegion)
POST /v2/auth/login
```

**Response Example:**
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
            "currentRegion": "MY",
            "currency": "MYR",
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
            "mfaStatus": "ACTIVE",
            "createdAt": "2025-11-01T10:00:00+07:00",
            "lastLoginAt": "2025-12-03T10:30:00+07:00"
        }
    }
}
```

**Notes:**
- User can login from any region
- `region` query parameter sets the `currentRegion`
- `currency` changes based on `currentRegion`
- `balance` always shows all currencies (regardless of current region)
- This allows user to browse products in different regions

---

### 4. Login with Google (with Region)

**Endpoint:** `POST /v2/auth/login/google?region={region_code}`

**Query Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| region | string | No | Region to set as current. Default: user's primaryRegion |

**Request Body:**
```json
{
    "idToken": "eyJhbGciOiJSUzI1NiIsImtpZCI6IjZmODI4..."
}
```

**Example Request:**
```bash
POST /v2/auth/login/google?region=SG
```

Response sama dengan regular login (dengan region yang dipilih).

---

### 5. Get User Profile (with Region)

**Endpoint:** `GET /v2/user/profile?region={region_code}`

**Headers:**
```
Authorization: Bearer {access_token}
```

**Query Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| region | string | No | Region to switch to. If not provided, returns current region |

**Example Requests:**
```bash
# Get profile with current region
GET /v2/user/profile

# Switch to Indonesia and get profile
GET /v2/user/profile?region=ID

# Switch to Singapore and get profile
GET /v2/user/profile?region=SG
```

**Response Example:**
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
        "currentRegion": "SG",
        "currency": "SGD",
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

**Notes:**
- Calling this endpoint with `region` parameter updates user's `currentRegion`
- `currency` changes based on new `currentRegion`
- `balance` always shows all currencies
- Membership progress is always calculated in primary currency (IDR in this case)
- This endpoint is called when user changes region in UI

---

## 🔄 Region Switching Flow

### Frontend Implementation

```javascript
// User Profile State
const [user, setUser] = useState(null);
const [currentRegion, setCurrentRegion] = useState('ID');

// Initial Load - Get user profile
useEffect(() => {
    const loadProfile = async () => {
        const response = await fetch('/v2/user/profile', {
            headers: {
                'Authorization': `Bearer ${accessToken}`
            }
        });
        const { data } = await response.json();
        setUser(data);
        setCurrentRegion(data.currentRegion);
    };
    
    if (accessToken) {
        loadProfile();
    }
}, [accessToken]);

// Region Selector Change
const handleRegionChange = async (newRegion) => {
    // Update UI immediately for better UX
    setCurrentRegion(newRegion);
    
    // Call API to persist region change
    const response = await fetch(`/v2/user/profile?region=${newRegion}`, {
        headers: {
            'Authorization': `Bearer ${accessToken}`
        }
    });
    
    const { data } = await response.json();
    setUser(data);
    
    // Update currency display
    setCurrency(data.currency);
    
    // Reload products for new region
    reloadProducts(newRegion);
};

// Region Selector Component
<select value={currentRegion} onChange={(e) => handleRegionChange(e.target.value)}>
    <option value="ID">🇮🇩 Indonesia (IDR)</option>
    <option value="MY">🇲🇾 Malaysia (MYR)</option>
    <option value="PH">🇵🇭 Philippines (PHP)</option>
    <option value="SG">🇸🇬 Singapore (SGD)</option>
    <option value="TH">🇹🇭 Thailand (THB)</option>
</select>
```

---

## 💰 Currency & Balance Management

### Balance Display

```javascript
// Display balance for current region
const displayBalance = (user) => {
    return user.balance[user.currency];
};

// Example:
// If currentRegion = "ID", currency = "IDR"
// Display: Rp 150.000

// If currentRegion = "MY", currency = "MYR"
// Display: RM 500

// Display all balances
const displayAllBalances = (user) => {
    return (
        <div>
            <div>🇮🇩 IDR: {formatCurrency(user.balance.IDR, 'IDR')}</div>
            <div>🇲🇾 MYR: {formatCurrency(user.balance.MYR, 'MYR')}</div>
            <div>🇵🇭 PHP: {formatCurrency(user.balance.PHP, 'PHP')}</div>
            <div>🇸🇬 SGD: {formatCurrency(user.balance.SGD, 'SGD')}</div>
            <div>🇹🇭 THB: {formatCurrency(user.balance.THB, 'THB')}</div>
        </div>
    );
};
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

## 🎯 User Object Structure

```typescript
interface User {
    id: string;
    firstName: string;
    lastName: string;
    email: string;
    phoneNumber: string | null;
    profilePicture: string | null;
    status: 'ACTIVE' | 'INACTIVE' | 'SUSPENDED';
    
    // Region & Currency
    primaryRegion: 'ID' | 'MY' | 'PH' | 'SG' | 'TH';  // Set at registration, permanent
    currentRegion: 'ID' | 'MY' | 'PH' | 'SG' | 'TH';  // Can change, used for browsing
    currency: 'IDR' | 'MYR' | 'PHP' | 'SGD' | 'THB';  // Derived from currentRegion
    
    // Multi-Currency Balance
    balance: {
        IDR: number;
        MYR: number;
        PHP: number;
        SGD: number;
        THB: number;
    };
    
    // Membership
    membership: {
        level: 'CLASSIC' | 'PRESTIGE' | 'ROYAL';
        name: string;
        benefits: string[];
        progress: {
            current: number;        // Transaction value in primary currency
            target: number;         // Target for next level
            percentage: number;     // Progress percentage
            nextLevel: string;      // Next membership level
            currency: string;       // Currency for calculation (always primary)
        };
    };
    
    // Security
    mfaStatus: 'ACTIVE' | 'INACTIVE';
    googleId?: string;
    
    // Timestamps
    createdAt: string;
    lastLoginAt: string;
    emailVerifiedAt: string;
    updatedAt: string;
}
```

---

## 📊 Backend Implementation Guide

### Database Schema

```sql
CREATE TABLE users (
    id VARCHAR(50) PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100),
    email VARCHAR(255) UNIQUE NOT NULL,
    phone_number VARCHAR(20),
    password_hash VARCHAR(255),
    profile_picture TEXT,
    status ENUM('ACTIVE', 'INACTIVE', 'SUSPENDED') DEFAULT 'INACTIVE',
    
    -- Region & Currency
    primary_region ENUM('ID', 'MY', 'PH', 'SG', 'TH') NOT NULL DEFAULT 'ID',
    current_region ENUM('ID', 'MY', 'PH', 'SG', 'TH') NOT NULL DEFAULT 'ID',
    
    -- Multi-Currency Balances (stored as cents/sen/centavos)
    balance_idr BIGINT DEFAULT 0,
    balance_myr BIGINT DEFAULT 0,
    balance_php BIGINT DEFAULT 0,
    balance_sgd BIGINT DEFAULT 0,
    balance_thb BIGINT DEFAULT 0,
    
    -- Membership
    membership_level ENUM('CLASSIC', 'PRESTIGE', 'ROYAL') DEFAULT 'CLASSIC',
    
    -- Security
    mfa_status ENUM('ACTIVE', 'INACTIVE') DEFAULT 'INACTIVE',
    mfa_secret VARCHAR(255),
    google_id VARCHAR(255),
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP,
    email_verified_at TIMESTAMP,
    
    INDEX idx_email (email),
    INDEX idx_google_id (google_id),
    INDEX idx_primary_region (primary_region),
    INDEX idx_membership_level (membership_level)
);
```

### Go Implementation

```go
type User struct {
    ID             string    `json:"id"`
    FirstName      string    `json:"firstName"`
    LastName       string    `json:"lastName"`
    Email          string    `json:"email"`
    PhoneNumber    *string   `json:"phoneNumber"`
    ProfilePicture *string   `json:"profilePicture"`
    Status         string    `json:"status"`
    
    // Region & Currency
    PrimaryRegion  string    `json:"primaryRegion"`
    CurrentRegion  string    `json:"currentRegion"`
    Currency       string    `json:"currency"`
    
    // Multi-Currency Balance
    Balance        Balance   `json:"balance"`
    
    // Membership
    Membership     Membership `json:"membership"`
    
    // Security
    MFAStatus      string    `json:"mfaStatus"`
    GoogleID       *string   `json:"googleId,omitempty"`
    
    // Timestamps
    CreatedAt      time.Time `json:"createdAt"`
    UpdatedAt      time.Time `json:"updatedAt"`
    LastLoginAt    time.Time `json:"lastLoginAt"`
    EmailVerifiedAt *time.Time `json:"emailVerifiedAt,omitempty"`
}

type Balance struct {
    IDR int64 `json:"IDR"`
    MYR int64 `json:"MYR"`
    PHP int64 `json:"PHP"`
    SGD int64 `json:"SGD"`
    THB int64 `json:"THB"`
}

// Get currency from region
func GetCurrencyFromRegion(region string) string {
    currencies := map[string]string{
        "ID": "IDR",
        "MY": "MYR",
        "PH": "PHP",
        "SG": "SGD",
        "TH": "THB",
    }
    return currencies[region]
}

// Register Handler
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
    region := r.URL.Query().Get("region")
    if region == "" {
        region = "ID" // Default
    }
    
    // Validate region
    validRegions := []string{"ID", "MY", "PH", "SG", "TH"}
    if !contains(validRegions, region) {
        region = "ID"
    }
    
    // Create user
    user := User{
        ID:            generateID(),
        FirstName:     req.FirstName,
        LastName:      req.LastName,
        Email:         req.Email,
        PhoneNumber:   req.PhoneNumber,
        Status:        "INACTIVE",
        PrimaryRegion: region,
        CurrentRegion: region,
        Currency:      GetCurrencyFromRegion(region),
        Balance: Balance{
            IDR: 0,
            MYR: 0,
            PHP: 0,
            SGD: 0,
            THB: 0,
        },
        // ... other fields
    }
    
    // Save to database
    db.Save(&user)
    
    // Send response
    respondJSON(w, user)
}

// Login Handler
func LoginHandler(w http.ResponseWriter, r *http.Request) {
    region := r.URL.Query().Get("region")
    
    // Authenticate user
    user, err := authenticateUser(email, password)
    if err != nil {
        respondError(w, "Invalid credentials")
        return
    }
    
    // Update current region if provided
    if region != "" && isValidRegion(region) {
        user.CurrentRegion = region
        user.Currency = GetCurrencyFromRegion(region)
        db.Save(&user)
    }
    
    // Generate tokens
    accessToken, refreshToken := generateTokens(user)
    
    // Send response
    respondJSON(w, LoginResponse{
        Token: Token{
            AccessToken:  accessToken,
            RefreshToken: refreshToken,
            ExpiresIn:    3600,
        },
        User: user,
    })
}

// Get Profile Handler
func GetProfileHandler(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("userId").(string)
    region := r.URL.Query().Get("region")
    
    // Get user from database
    user, err := db.GetUser(userID)
    if err != nil {
        respondError(w, "User not found")
        return
    }
    
    // Update current region if provided
    if region != "" && isValidRegion(region) {
        user.CurrentRegion = region
        user.Currency = GetCurrencyFromRegion(region)
        user.UpdatedAt = time.Now()
        db.Save(&user)
    }
    
    // Send response
    respondJSON(w, user)
}
```

---

## 🎨 UI/UX Considerations

### 1. Region Selector Placement

```
┌────────────────────────────────────┐
│  🌍 ID   John Doe   💰 Rp 150.000 │  ← Header with region selector
└────────────────────────────────────┘
```

### 2. Balance Display

**Option A: Show Current Region Balance Only**
```
💰 Saldo: Rp 150.000
```

**Option B: Show All Balances (Recommended)**
```
💰 Saldo:
🇮🇩 Rp 150.000
🇲🇾 RM 500
🇵🇭 ₱ 0
🇸🇬 S$ 100
🇹🇭 ฿ 0
```

### 3. Region Switch Confirmation

When user switches region during checkout:
```
⚠️ Anda akan menggunakan MYR untuk transaksi ini.
   Saldo MYR: RM 500
   
   [Lanjutkan]  [Batal]
```

### 4. Insufficient Balance Across Regions

```
❌ Saldo IDR tidak mencukupi (Rp 150.000)
   Harga: Rp 200.000
   
💡 Anda memiliki saldo di region lain:
   🇲🇾 RM 500 (≈ Rp 1.700.000)
   
   [Top Up IDR]  [Bayar dengan MYR]
```

---

## 🔄 Complete User Journey

### Scenario 1: Indonesian User Shopping in Malaysia

```
1. User registers from Indonesia
   POST /v2/auth/register?region=ID
   → primaryRegion: ID
   → currentRegion: ID
   → currency: IDR

2. User adds balance
   → balance.IDR: 150000

3. User switches to Malaysia to browse cheaper products
   GET /v2/user/profile?region=MY
   → currentRegion: MY
   → currency: MYR
   → balance.IDR: 150000 (unchanged)

4. User buys MLBB diamonds (priced in MYR)
   POST /v2/orders/inquirys
   → Uses MYR balance (if available)
   → Or shows currency conversion option

5. User switches back to Indonesia
   GET /v2/user/profile?region=ID
   → currentRegion: ID
   → currency: IDR
```

### Scenario 2: Malaysian User Travels to Singapore

```
1. User registered in Malaysia (primaryRegion: MY)
2. User travels to Singapore
3. User changes region to SG in app
   GET /v2/user/profile?region=SG
   → currentRegion: SG
   → currency: SGD
4. Can browse and buy products in SGD
5. Can use existing MYR balance or add SGD balance
```

---

## ✅ Summary

### Key Points

1. **Registration**: User sets `primaryRegion` at registration (permanent)
2. **Login**: User can set `currentRegion` at login (flexible)
3. **Profile**: User can switch `currentRegion` anytime via profile endpoint
4. **Currency**: Auto-derived from `currentRegion`
5. **Balance**: Multi-currency (all 5 currencies stored)
6. **Membership**: Calculated in primary currency only

### Region vs Currency Mapping

| Region | Currency | Symbol | Example |
|--------|----------|--------|---------|
| ID | IDR | Rp | Rp 150.000 |
| MY | MYR | RM | RM 500 |
| PH | PHP | ₱ | ₱ 5.000 |
| SG | SGD | S$ | S$ 100 |
| TH | THB | ฿ | ฿ 3.000 |

### Updated Endpoints Summary

| Endpoint | Query Param | Purpose |
|----------|-------------|---------|
| `POST /v2/auth/register` | `?region=` | Set primary & current region |
| `POST /v2/auth/register/google` | `?region=` | Set primary & current region |
| `POST /v2/auth/login` | `?region=` | Set current region |
| `POST /v2/auth/login/google` | `?region=` | Set current region |
| `GET /v2/user/profile` | `?region=` | Switch current region |

---

**Ready for Implementation!** 🚀
