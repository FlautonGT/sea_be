package domain

import (
	"time"

	"github.com/google/uuid"
)

type FeeType string

const (
	FeeTypeFixed      FeeType = "FIXED"
	FeeTypePercentage FeeType = "PERCENTAGE"
	FeeTypeMixed      FeeType = "MIXED"
)

type PaymentGateway struct {
	ID               uuid.UUID              `json:"id" db:"id"`
	Code             string                 `json:"code" db:"code"`
	Name             string                 `json:"name" db:"name"`
	BaseURL          string                 `json:"baseUrl" db:"base_url"`
	CallbackURL      *string                `json:"callbackUrl" db:"callback_url"`
	IsActive         bool                   `json:"isActive" db:"is_active"`
	SupportedMethods []string               `json:"supportedMethods" db:"supported_methods"`
	SupportedTypes   []string               `json:"supportedTypes" db:"supported_types"` // purchase, deposit
	HealthStatus     HealthStatus           `json:"healthStatus" db:"health_status"`
	LastHealthCheck  *time.Time             `json:"lastHealthCheck" db:"last_health_check"`
	APIConfig        GatewayAPIConfig       `json:"apiConfig" db:"api_config"`
	FeeConfig        map[string]FeeConfig   `json:"feeConfig" db:"fee_config"` // per method
	Mapping          GatewayMapping         `json:"mapping" db:"mapping"`
	EnvCredKeys      map[string]string      `json:"envCredKeys" db:"env_cred_keys"`
	CreatedAt        time.Time              `json:"createdAt" db:"created_at"`
	UpdatedAt        time.Time              `json:"updatedAt" db:"updated_at"`
}

type GatewayAPIConfig struct {
	Timeout       int `json:"timeout"`
	RetryAttempts int `json:"retryAttempts"`
}

type FeeConfig struct {
	FeeType       FeeType `json:"feeType"`
	FeeAmount     float64 `json:"feeAmount"`
	FeePercentage float64 `json:"feePercentage"`
	MinFee        float64 `json:"minFee"`
	MaxFee        float64 `json:"maxFee"`
}

type GatewayMapping struct {
	StatusSuccess []string `json:"statusSuccess"`
	StatusPending []string `json:"statusPending"`
	StatusFailed  []string `json:"statusFailed"`
}

type PaymentChannel struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	Code           string     `json:"code" db:"code"`
	Name           string     `json:"name" db:"name"`
	Description    *string    `json:"description" db:"description"`
	Image          *string    `json:"image" db:"image"`
	CategoryID     uuid.UUID  `json:"categoryId" db:"category_id"`
	IsActive       bool       `json:"isActive" db:"is_active"`
	IsFeatured     bool       `json:"isFeatured" db:"is_featured"`
	Order          int        `json:"order" db:"order"`
	Instruction    *string    `json:"instruction" db:"instruction"`
	FeeType        FeeType    `json:"feeType" db:"fee_type"`
	FeeAmount      float64    `json:"feeAmount" db:"fee_amount"`
	FeePercentage  float64    `json:"feePercentage" db:"fee_percentage"`
	MinAmount      float64    `json:"minAmount" db:"min_amount"`
	MaxAmount      float64    `json:"maxAmount" db:"max_amount"`
	SupportedTypes []string   `json:"supportedTypes" db:"supported_types"` // purchase, deposit
	CreatedAt      time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt      time.Time  `json:"updatedAt" db:"updated_at"`
}

type PaymentChannelRegion struct {
	ChannelID  uuid.UUID `json:"channelId" db:"channel_id"`
	RegionCode string    `json:"regionCode" db:"region_code"`
	IsActive   bool      `json:"isActive" db:"is_active"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
}

type PaymentChannelGateway struct {
	ChannelID   uuid.UUID  `json:"channelId" db:"channel_id"`
	GatewayID   uuid.UUID  `json:"gatewayId" db:"gateway_id"`
	PaymentType string     `json:"paymentType" db:"payment_type"` // purchase or deposit
	IsActive    bool       `json:"isActive" db:"is_active"`
	CreatedAt   time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time  `json:"updatedAt" db:"updated_at"`
}

type PaymentChannelCategory struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Code      string    `json:"code" db:"code"`
	Title     string    `json:"title" db:"title"`
	Icon      *string   `json:"icon" db:"icon"`
	Order     int       `json:"order" db:"order"`
	IsActive  bool      `json:"isActive" db:"is_active"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

// Response DTOs
type PaymentGatewayResponse struct {
	ID               string             `json:"id"`
	Code             string             `json:"code"`
	Name             string             `json:"name"`
	BaseURL          string             `json:"baseUrl"`
	IsActive         bool               `json:"isActive"`
	SupportedMethods []string           `json:"supportedMethods"`
	SupportedTypes   []string           `json:"supportedTypes"`
	HealthStatus     HealthStatus       `json:"healthStatus"`
	LastHealthCheck  *time.Time         `json:"lastHealthCheck,omitempty"`
	Stats            *PaymentGatewayStats `json:"stats,omitempty"`
}

type PaymentGatewayDetailResponse struct {
	PaymentGatewayResponse
	CallbackURL *string                `json:"callbackUrl,omitempty"`
	APIConfig   GatewayAPIConfig       `json:"apiConfig"`
	FeeConfig   map[string]FeeConfig   `json:"feeConfig"`
	Mapping     GatewayMapping         `json:"mapping"`
	Credentials map[string]bool        `json:"credentials"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
}

type PaymentGatewayStats struct {
	TodayTransactions int     `json:"todayTransactions"`
	TodayVolume       float64 `json:"todayVolume"`
	SuccessRate       float64 `json:"successRate"`
	AvgResponseTime   int     `json:"avgResponseTime,omitempty"`
}

type PaymentChannelResponse struct {
	Code          string                  `json:"code"`
	Name          string                  `json:"name"`
	Description   *string                 `json:"description,omitempty"`
	Image         *string                 `json:"image,omitempty"`
	Currency      string                  `json:"currency"`
	FeeAmount     float64                 `json:"feeAmount"`
	FeePercentage float64                 `json:"feePercentage"`
	MinAmount     float64                 `json:"minAmount"`
	MaxAmount     float64                 `json:"maxAmount"`
	Featured      bool                    `json:"featured"`
	Instruction   *string                 `json:"instruction,omitempty"`
	Category      PaymentCategoryInfo     `json:"category"`
}

type PaymentCategoryInfo struct {
	Title string `json:"title"`
	Code  string `json:"code"`
}

type PaymentCategoryResponse struct {
	Title string  `json:"title"`
	Code  string  `json:"code"`
	Icon  *string `json:"icon,omitempty"`
	Order int     `json:"order"`
}

type AdminPaymentChannelResponse struct {
	ID             string                  `json:"id"`
	Code           string                  `json:"code"`
	Name           string                  `json:"name"`
	Description    *string                 `json:"description,omitempty"`
	Image          *string                 `json:"image,omitempty"`
	Category       PaymentCategoryInfo     `json:"category"`
	Gateway        GatewayAssignment       `json:"gateway"`
	Fee            FeeInfo                 `json:"fee"`
	Limits         LimitInfo               `json:"limits"`
	Regions        []string                `json:"regions"`
	SupportedTypes []string                `json:"supportedTypes"`
	IsActive       bool                    `json:"isActive"`
	IsFeatured     bool                    `json:"isFeatured"`
	Order          int                     `json:"order"`
	Instruction    *string                 `json:"instruction,omitempty"`
	Stats          *PaymentChannelStats    `json:"stats,omitempty"`
	CreatedAt      time.Time               `json:"createdAt"`
	UpdatedAt      time.Time               `json:"updatedAt"`
}

type GatewayAssignment struct {
	Purchase *GatewayInfo `json:"purchase"`
	Deposit  *GatewayInfo `json:"deposit"`
}

type GatewayInfo struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type FeeInfo struct {
	FeeType       FeeType `json:"feeType"`
	FeeAmount     float64 `json:"feeAmount"`
	FeePercentage float64 `json:"feePercentage"`
}

type LimitInfo struct {
	MinAmount float64 `json:"minAmount"`
	MaxAmount float64 `json:"maxAmount"`
}

type PaymentChannelStats struct {
	TodayTransactions int     `json:"todayTransactions"`
	TodayVolume       float64 `json:"todayVolume"`
}

type ChannelAssignmentResponse struct {
	PaymentCode  string                 `json:"paymentCode"`
	PaymentName  string                 `json:"paymentName"`
	Assignments  AssignmentDetails      `json:"assignments"`
}

type AssignmentDetails struct {
	Purchase AssignmentInfo `json:"purchase"`
	Deposit  AssignmentInfo `json:"deposit"`
}

type AssignmentInfo struct {
	GatewayCode *string `json:"gatewayCode"`
	GatewayName *string `json:"gatewayName"`
	IsActive    bool    `json:"isActive"`
}

type AdminPaymentCategoryResponse struct {
	ID           string    `json:"id"`
	Code         string    `json:"code"`
	Title        string    `json:"title"`
	Icon         *string   `json:"icon,omitempty"`
	IsActive     bool      `json:"isActive"`
	Order        int       `json:"order"`
	ChannelCount int       `json:"channelCount"`
}

// Request DTOs
type CreatePaymentGatewayRequest struct {
	Code             string              `json:"code" validate:"required,min=1,max=50"`
	Name             string              `json:"name" validate:"required,min=1,max=100"`
	BaseURL          string              `json:"baseUrl" validate:"required,url"`
	CallbackURL      *string             `json:"callbackUrl" validate:"omitempty,url"`
	IsActive         bool                `json:"isActive"`
	SupportedMethods []string            `json:"supportedMethods" validate:"required,min=1"`
	SupportedTypes   []string            `json:"supportedTypes" validate:"required,min=1"`
	APIConfig        GatewayAPIConfig    `json:"apiConfig"`
	FeeConfig        map[string]FeeConfig `json:"feeConfig"`
	EnvCredKeys      map[string]string   `json:"envCredentialKeys"`
}

type UpdatePaymentGatewayRequest struct {
	Name             *string             `json:"name" validate:"omitempty,min=1,max=100"`
	BaseURL          *string             `json:"baseUrl" validate:"omitempty,url"`
	CallbackURL      *string             `json:"callbackUrl" validate:"omitempty,url"`
	IsActive         *bool               `json:"isActive"`
	SupportedMethods []string            `json:"supportedMethods"`
	SupportedTypes   []string            `json:"supportedTypes"`
	APIConfig        *GatewayAPIConfig   `json:"apiConfig"`
	FeeConfig        map[string]FeeConfig `json:"feeConfig"`
}

type CreatePaymentChannelRequest struct {
	Code              string             `json:"code" validate:"required,min=1,max=50"`
	Name              string             `json:"name" validate:"required,min=1,max=100"`
	Description       *string            `json:"description"`
	CategoryCode      string             `json:"categoryCode" validate:"required"`
	GatewayAssignment GatewayAssignReq   `json:"gatewayAssignment"`
	Fee               FeeInfo            `json:"fee"`
	Limits            LimitInfo          `json:"limits"`
	Regions           []string           `json:"regions" validate:"required,min=1"`
	SupportedTypes    []string           `json:"supportedTypes" validate:"required,min=1"`
	IsActive          bool               `json:"isActive"`
	IsFeatured        bool               `json:"isFeatured"`
	Order             int                `json:"order"`
	Instruction       *string            `json:"instruction"`
}

type GatewayAssignReq struct {
	Purchase *string `json:"purchase"`
	Deposit  *string `json:"deposit"`
}

type UpdatePaymentChannelRequest struct {
	Name              *string            `json:"name" validate:"omitempty,min=1,max=100"`
	Description       *string            `json:"description"`
	Fee               *FeeInfo           `json:"fee"`
	Limits            *LimitInfo         `json:"limits"`
	IsActive          *bool              `json:"isActive"`
	IsFeatured        *bool              `json:"isFeatured"`
	Order             *int               `json:"order"`
	Instruction       *string            `json:"instruction"`
}

type UpdateChannelAssignmentRequest struct {
	Purchase *ChannelAssignReq `json:"purchase"`
	Deposit  *ChannelAssignReq `json:"deposit"`
}

type ChannelAssignReq struct {
	GatewayCode string `json:"gatewayCode"`
	IsActive    bool   `json:"isActive"`
}

type CreatePaymentCategoryRequest struct {
	Code     string  `json:"code" validate:"required,min=1,max=50"`
	Title    string  `json:"title" validate:"required,min=1,max=100"`
	IsActive bool    `json:"isActive"`
	Order    int     `json:"order"`
}

type UpdatePaymentCategoryRequest struct {
	Title    *string `json:"title" validate:"omitempty,min=1,max=100"`
	IsActive *bool   `json:"isActive"`
	Order    *int    `json:"order"`
}

