package domain

import (
	"time"

	"github.com/google/uuid"
)

type Promo struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	Code              string     `json:"code" db:"code"`
	Title             string     `json:"title" db:"title"`
	Description       *string    `json:"description" db:"description"`
	MaxDailyUsage     int        `json:"maxDailyUsage" db:"max_daily_usage"`
	MaxUsage          int        `json:"maxUsage" db:"max_usage"`
	MaxUsagePerID     int        `json:"maxUsagePerId" db:"max_usage_per_id"`
	MaxUsagePerDevice int        `json:"maxUsagePerDevice" db:"max_usage_per_device"`
	MaxUsagePerIP     int        `json:"maxUsagePerIp" db:"max_usage_per_ip"`
	StartAt           *time.Time `json:"startAt" db:"start_at"`
	ExpiredAt         *time.Time `json:"expiredAt" db:"expired_at"`
	MinAmount         float64    `json:"minAmount" db:"min_amount"`
	MaxPromoAmount    float64    `json:"maxPromoAmount" db:"max_promo_amount"`
	PromoFlat         float64    `json:"promoFlat" db:"promo_flat"`
	PromoPercentage   float64    `json:"promoPercentage" db:"promo_percentage"`
	IsActive          bool       `json:"isActive" db:"is_active"`
	Note              *string    `json:"note" db:"note"`
	TotalUsage        int        `json:"totalUsage" db:"total_usage"`
	CreatedAt         time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt         time.Time  `json:"updatedAt" db:"updated_at"`
}

type PromoProduct struct {
	PromoID   uuid.UUID `json:"promoId" db:"promo_id"`
	ProductID uuid.UUID `json:"productId" db:"product_id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

type PromoPaymentChannel struct {
	PromoID   uuid.UUID `json:"promoId" db:"promo_id"`
	ChannelID uuid.UUID `json:"channelId" db:"channel_id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

type PromoRegion struct {
	PromoID    uuid.UUID `json:"promoId" db:"promo_id"`
	RegionCode string    `json:"regionCode" db:"region_code"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
}

type PromoDayAvailable struct {
	PromoID   uuid.UUID `json:"promoId" db:"promo_id"`
	Day       string    `json:"day" db:"day"` // MON, TUE, WED, THU, FRI, SAT, SUN
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

type PromoUsage struct {
	ID            uuid.UUID `json:"id" db:"id"`
	PromoID       uuid.UUID `json:"promoId" db:"promo_id"`
	TransactionID uuid.UUID `json:"transactionId" db:"transaction_id"`
	UserID        *uuid.UUID `json:"userId" db:"user_id"`
	DeviceID      *string   `json:"deviceId" db:"device_id"`
	IPAddress     string    `json:"ipAddress" db:"ip_address"`
	DiscountAmount float64  `json:"discountAmount" db:"discount_amount"`
	CreatedAt     time.Time `json:"createdAt" db:"created_at"`
}

// Response DTOs
type PromoResponse struct {
	Code              string              `json:"code"`
	Title             string              `json:"title"`
	Description       *string             `json:"description,omitempty"`
	Products          []PromoItemInfo     `json:"products"`
	PaymentChannels   []PromoItemInfo     `json:"paymentChannels"`
	DaysAvailable     []string            `json:"daysAvailable"`
	MaxDailyUsage     int                 `json:"maxDailyUsage"`
	MaxUsage          int                 `json:"maxUsage"`
	MaxUsagePerID     int                 `json:"maxUsagePerId"`
	MaxUsagePerDevice int                 `json:"maxUsagePerDevice"`
	MaxUsagePerIP     int                 `json:"maxUsagePerIp"`
	ExpiredAt         *time.Time          `json:"expiredAt,omitempty"`
	MinAmount         float64             `json:"minAmount"`
	MaxPromoAmount    float64             `json:"maxPromoAmount"`
	PromoFlat         float64             `json:"promoFlat"`
	PromoPercentage   float64             `json:"promoPercentage"`
	IsAvailable       bool                `json:"isAvailable"`
	Note              *string             `json:"note,omitempty"`
	TotalUsage        int                 `json:"totalUsage"`
	TotalDailyUsage   int                 `json:"totalDailyUsage"`
}

type PromoItemInfo struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type PromoValidateResponse struct {
	PromoCode       string          `json:"promoCode"`
	DiscountAmount  float64         `json:"discountAmount"`
	OriginalAmount  float64         `json:"originalAmount"`
	FinalAmount     float64         `json:"finalAmount"`
	PromoDetails    PromoDetails    `json:"promoDetails"`
}

type PromoDetails struct {
	Title           string  `json:"title"`
	PromoPercentage float64 `json:"promoPercentage,omitempty"`
	PromoFlat       float64 `json:"promoFlat,omitempty"`
	MaxPromoAmount  float64 `json:"maxPromoAmount"`
}

type AdminPromoResponse struct {
	ID                string          `json:"id"`
	Code              string          `json:"code"`
	Title             string          `json:"title"`
	Description       *string         `json:"description,omitempty"`
	Products          []PromoItemInfo `json:"products"`
	PaymentChannels   []PromoItemInfo `json:"paymentChannels"`
	Regions           []string        `json:"regions"`
	DaysAvailable     []string        `json:"daysAvailable"`
	MaxDailyUsage     int             `json:"maxDailyUsage"`
	MaxUsage          int             `json:"maxUsage"`
	MaxUsagePerID     int             `json:"maxUsagePerId"`
	MaxUsagePerDevice int             `json:"maxUsagePerDevice"`
	MaxUsagePerIP     int             `json:"maxUsagePerIp"`
	StartAt           *time.Time      `json:"startAt,omitempty"`
	ExpiredAt         *time.Time      `json:"expiredAt,omitempty"`
	MinAmount         float64         `json:"minAmount"`
	MaxPromoAmount    float64         `json:"maxPromoAmount"`
	PromoFlat         float64         `json:"promoFlat"`
	PromoPercentage   float64         `json:"promoPercentage"`
	IsActive          bool            `json:"isActive"`
	Note              *string         `json:"note,omitempty"`
	Stats             *PromoStats     `json:"stats,omitempty"`
	CreatedAt         time.Time       `json:"createdAt"`
	UpdatedAt         time.Time       `json:"updatedAt"`
}

type PromoStats struct {
	TotalUsage      int     `json:"totalUsage"`
	TotalDiscount   float64 `json:"totalDiscount"`
	TodayUsage      int     `json:"todayUsage"`
	TodayDiscount   float64 `json:"todayDiscount"`
}

type PromoUsageStats struct {
	PromoCode       string              `json:"promoCode"`
	TotalUsage      int                 `json:"totalUsage"`
	TotalDiscount   float64             `json:"totalDiscount"`
	TodayUsage      int                 `json:"todayUsage"`
	TodayDiscount   float64             `json:"todayDiscount"`
	UsageByProduct  []UsageByItem       `json:"usageByProduct"`
	UsageByPayment  []UsageByItem       `json:"usageByPayment"`
	UsageByRegion   []UsageByItem       `json:"usageByRegion"`
}

type UsageByItem struct {
	Name     string  `json:"product,omitempty"`
	Payment  string  `json:"payment,omitempty"`
	Region   string  `json:"region,omitempty"`
	Count    int     `json:"count"`
	Discount float64 `json:"discount"`
}

// Request DTOs
type ValidatePromoRequest struct {
	PromoCode   string  `json:"promoCode" validate:"required"`
	ProductCode string  `json:"productCode" validate:"required"`
	SKUCode     string  `json:"skuCode" validate:"required"`
	PaymentCode string  `json:"paymentCode" validate:"required"`
	Region      string  `json:"region" validate:"required"`
	Amount      float64 `json:"amount" validate:"required,gt=0"`
}

type CreatePromoRequest struct {
	Code              string    `json:"code" validate:"required,min=1,max=50"`
	Title             string    `json:"title" validate:"required,min=1,max=100"`
	Description       *string   `json:"description"`
	Products          []string  `json:"products"`
	PaymentChannels   []string  `json:"paymentChannels"`
	Regions           []string  `json:"regions"`
	DaysAvailable     []string  `json:"daysAvailable" validate:"omitempty,dive,oneof=MON TUE WED THU FRI SAT SUN"`
	MaxDailyUsage     int       `json:"maxDailyUsage" validate:"min=0"`
	MaxUsage          int       `json:"maxUsage" validate:"min=0"`
	MaxUsagePerID     int       `json:"maxUsagePerId" validate:"min=0"`
	MaxUsagePerDevice int       `json:"maxUsagePerDevice" validate:"min=0"`
	MaxUsagePerIP     int       `json:"maxUsagePerIp" validate:"min=0"`
	StartAt           *time.Time `json:"startAt"`
	ExpiredAt         *time.Time `json:"expiredAt"`
	MinAmount         float64   `json:"minAmount" validate:"min=0"`
	MaxPromoAmount    float64   `json:"maxPromoAmount" validate:"min=0"`
	PromoFlat         float64   `json:"promoFlat" validate:"min=0"`
	PromoPercentage   float64   `json:"promoPercentage" validate:"min=0,max=100"`
	IsActive          bool      `json:"isActive"`
	Note              *string   `json:"note"`
}

type UpdatePromoRequest struct {
	Title             *string    `json:"title" validate:"omitempty,min=1,max=100"`
	Description       *string    `json:"description"`
	Products          []string   `json:"products"`
	PaymentChannels   []string   `json:"paymentChannels"`
	Regions           []string   `json:"regions"`
	DaysAvailable     []string   `json:"daysAvailable"`
	MaxDailyUsage     *int       `json:"maxDailyUsage" validate:"omitempty,min=0"`
	MaxUsage          *int       `json:"maxUsage" validate:"omitempty,min=0"`
	MaxUsagePerID     *int       `json:"maxUsagePerId" validate:"omitempty,min=0"`
	MaxUsagePerDevice *int       `json:"maxUsagePerDevice" validate:"omitempty,min=0"`
	MaxUsagePerIP     *int       `json:"maxUsagePerIp" validate:"omitempty,min=0"`
	StartAt           *time.Time `json:"startAt"`
	ExpiredAt         *time.Time `json:"expiredAt"`
	MinAmount         *float64   `json:"minAmount" validate:"omitempty,min=0"`
	MaxPromoAmount    *float64   `json:"maxPromoAmount" validate:"omitempty,min=0"`
	PromoFlat         *float64   `json:"promoFlat" validate:"omitempty,min=0"`
	PromoPercentage   *float64   `json:"promoPercentage" validate:"omitempty,min=0,max=100"`
	IsActive          *bool      `json:"isActive"`
	Note              *string    `json:"note"`
}

