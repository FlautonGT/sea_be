package domain

import "time"

// Pagination
type Pagination struct {
	Limit      int `json:"limit"`
	Page       int `json:"page"`
	TotalRows  int `json:"totalRows"`
	TotalPages int `json:"totalPages"`
}

type PaginationParams struct {
	Limit  int
	Page   int
	Offset int
}

func NewPaginationParams(limit, page int) PaginationParams {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if page <= 0 {
		page = 1
	}
	return PaginationParams{
		Limit:  limit,
		Page:   page,
		Offset: (page - 1) * limit,
	}
}

func NewPagination(limit, page, totalRows int) Pagination {
	totalPages := totalRows / limit
	if totalRows%limit > 0 {
		totalPages++
	}
	return Pagination{
		Limit:      limit,
		Page:       page,
		TotalRows:  totalRows,
		TotalPages: totalPages,
	}
}

// API Response Wrappers
type SuccessResponse struct {
	Data interface{} `json:"data"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details string            `json:"details,omitempty"`
	Fields  map[string]string `json:"fields,omitempty"`
}

type ListResponse struct {
	Data       interface{} `json:"data"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

type AdminMeta struct {
	RequiredPermission string `json:"requiredPermission"`
	AdminID            string `json:"adminId"`
	AdminRole          string `json:"adminRole"`
}

type AdminAPIResponse struct {
	Data interface{} `json:"data"`
	Meta *AdminMeta  `json:"_meta,omitempty"`
}

// Token Response
type TokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int    `json:"expiresIn"`
	TokenType    string `json:"tokenType"`
}

type AuthResponse struct {
	Step  string         `json:"step"`
	Token *TokenResponse `json:"token,omitempty"`
	User  interface{}    `json:"user,omitempty"`
}

type MFARequiredResponse struct {
	Step      string    `json:"step"`
	MFAToken  string    `json:"mfaToken"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type MFASetupResponse struct {
	Step        string   `json:"step"`
	QRCode      string   `json:"qrCode"`
	QRCodeImage string   `json:"qrCodeImage"`
	Secret      string   `json:"secret"`
	BackupCodes []string `json:"backupCodes"`
}

// Settings
type GeneralSettings struct {
	SiteName           string `json:"siteName"`
	SiteDescription    string `json:"siteDescription"`
	MaintenanceMode    bool   `json:"maintenanceMode"`
	MaintenanceMessage string `json:"maintenanceMessage,omitempty"`
}

type TransactionSettings struct {
	OrderExpiry       int  `json:"orderExpiry"` // seconds
	AutoRefundOnFail  bool `json:"autoRefundOnFail"`
	MaxRetryAttempts  int  `json:"maxRetryAttempts"`
}

type NotificationSettings struct {
	EmailEnabled    bool `json:"emailEnabled"`
	WhatsappEnabled bool `json:"whatsappEnabled"`
	TelegramEnabled bool `json:"telegramEnabled"`
}

type SecuritySettings struct {
	MaxLoginAttempts int  `json:"maxLoginAttempts"`
	LockoutDuration  int  `json:"lockoutDuration"` // seconds
	SessionTimeout   int  `json:"sessionTimeout"`  // seconds
	MFARequired      bool `json:"mfaRequired"`
}

type AllSettings struct {
	General      GeneralSettings      `json:"general"`
	Transaction  TransactionSettings  `json:"transaction"`
	Notification NotificationSettings `json:"notification"`
	Security     SecuritySettings     `json:"security"`
}

// Common Filters
type DateRangeFilter struct {
	StartDate *time.Time
	EndDate   *time.Time
}

// Currency mapping
var CurrencyByRegion = map[string]string{
	"ID": "IDR",
	"MY": "MYR",
	"PH": "PHP",
	"SG": "SGD",
	"TH": "THB",
}

// Membership benefits
var MembershipBenefits = map[MembershipLevel][]string{
	MembershipClassic: {
		"Standard transactions",
		"24/7 support",
		"1% bonus points",
	},
	MembershipPrestige: {
		"Diskon eksklusif hingga 5%",
		"Priority customer support",
		"Bonus poin 3%",
		"Akses promo premium",
	},
	MembershipRoyal: {
		"Diskon eksklusif hingga 10%",
		"Dedicated manager",
		"Bonus poin 5%",
		"VIP promos",
		"Priority transactions",
	},
}

// Membership thresholds (in IDR)
var MembershipThresholds = map[MembershipLevel]float64{
	MembershipClassic:  0,
	MembershipPrestige: 5000000,
	MembershipRoyal:    10000000,
}

func GetMembershipName(level MembershipLevel) string {
	names := map[MembershipLevel]string{
		MembershipClassic:  "Classic",
		MembershipPrestige: "Prestige",
		MembershipRoyal:    "Royal",
	}
	return names[level]
}

func GetNextMembershipLevel(level MembershipLevel) MembershipLevel {
	switch level {
	case MembershipClassic:
		return MembershipPrestige
	case MembershipPrestige:
		return MembershipRoyal
	default:
		return MembershipRoyal
	}
}

