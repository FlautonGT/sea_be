package domain

import (
	"time"

	"github.com/google/uuid"
)

type StockStatus string

const (
	StockAvailable   StockStatus = "AVAILABLE"
	StockEmpty       StockStatus = "EMPTY"
	StockMaintenance StockStatus = "MAINTENANCE"
)

type SKU struct {
	ID              uuid.UUID   `json:"id" db:"id"`
	Code            string      `json:"code" db:"code"`
	ProviderSKUCode string      `json:"providerSkuCode" db:"provider_sku_code"`
	Name            string      `json:"name" db:"name"`
	Description     *string     `json:"description" db:"description"`
	ProductID       uuid.UUID   `json:"productId" db:"product_id"`
	ProviderID      uuid.UUID   `json:"providerId" db:"provider_id"`
	SectionID       *uuid.UUID  `json:"sectionId" db:"section_id"`
	Image           *string     `json:"image" db:"image"`
	Info            *string     `json:"info" db:"info"`
	ProcessTime     int         `json:"processTime" db:"process_time"` // in minutes
	IsActive        bool        `json:"isActive" db:"is_active"`
	IsFeatured      bool        `json:"isFeatured" db:"is_featured"`
	Stock           StockStatus `json:"stock" db:"stock"`
	CreatedAt       time.Time   `json:"createdAt" db:"created_at"`
	UpdatedAt       time.Time   `json:"updatedAt" db:"updated_at"`
}

type SKUPricing struct {
	ID            uuid.UUID `json:"id" db:"id"`
	SKUID         uuid.UUID `json:"skuId" db:"sku_id"`
	RegionCode    string    `json:"regionCode" db:"region_code"`
	Currency      string    `json:"currency" db:"currency"`
	BuyPrice      float64   `json:"buyPrice" db:"buy_price"`
	SellPrice     float64   `json:"sellPrice" db:"sell_price"`
	OriginalPrice float64   `json:"originalPrice" db:"original_price"`
	CreatedAt     time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt     time.Time `json:"updatedAt" db:"updated_at"`
}

type SKUBadge struct {
	ID        uuid.UUID `json:"id" db:"id"`
	SKUID     uuid.UUID `json:"skuId" db:"sku_id"`
	Text      string    `json:"text" db:"text"`
	Color     string    `json:"color" db:"color"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

// Response DTOs
type SKUResponse struct {
	Code          string       `json:"code"`
	Name          string       `json:"name"`
	Description   *string      `json:"description,omitempty"`
	Currency      string       `json:"currency"`
	Price         float64      `json:"price"`
	OriginalPrice float64      `json:"originalPrice"`
	Discount      float64      `json:"discount"` // percentage
	Image         *string      `json:"image,omitempty"`
	Info          *string      `json:"info,omitempty"`
	ProcessTime   int          `json:"processTime"`
	IsAvailable   bool         `json:"isAvailable"`
	IsFeatured    bool         `json:"isFeatured"`
	Section       *SectionInfo `json:"section,omitempty"`
	Badge         *BadgeInfo   `json:"badge,omitempty"`
}

type SectionInfo struct {
	Title string `json:"title"`
	Code  string `json:"code"`
}

type BadgeInfo struct {
	Text  string `json:"text"`
	Color string `json:"color"`
}

// Admin Response DTOs
type AdminSKUResponse struct {
	ID              string             `json:"id"`
	Code            string             `json:"code"`
	ProviderSKUCode string             `json:"providerSkuCode"`
	Name            string             `json:"name"`
	Description     *string            `json:"description,omitempty"`
	Product         ProductInfo        `json:"product"`
	Provider        ProviderInfo       `json:"provider"`
	Pricing         map[string]PricingInfo `json:"pricing"`
	Section         *SectionInfo       `json:"section,omitempty"`
	IsActive        bool               `json:"isActive"`
	IsFeatured      bool               `json:"isFeatured"`
	ProcessTime     int                `json:"processTime"`
	Stock           StockStatus        `json:"stock"`
	Stats           *SKUStats          `json:"stats,omitempty"`
	CreatedAt       time.Time          `json:"createdAt"`
	UpdatedAt       time.Time          `json:"updatedAt"`
}

type ProductInfo struct {
	Code  string `json:"code"`
	Title string `json:"title"`
}

type ProviderInfo struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type PricingInfo struct {
	Currency      string  `json:"currency"`
	BuyPrice      float64 `json:"buyPrice"`
	SellPrice     float64 `json:"sellPrice"`
	OriginalPrice float64 `json:"originalPrice"`
	Margin        float64 `json:"margin"` // percentage
	Discount      float64 `json:"discount"` // percentage
}

type SKUStats struct {
	TodaySold int `json:"todaySold"`
	TotalSold int `json:"totalSold"`
}

// Request DTOs
type CreateSKURequest struct {
	Code            string              `json:"code" validate:"required,min=1,max=50"`
	ProviderSKUCode string              `json:"providerSkuCode" validate:"required"`
	Name            string              `json:"name" validate:"required,min=1,max=200"`
	Description     *string             `json:"description"`
	ProductCode     string              `json:"productCode" validate:"required"`
	ProviderCode    string              `json:"providerCode" validate:"required"`
	SectionCode     *string             `json:"sectionCode"`
	Image           *string             `json:"image"`
	Info            *string             `json:"info"`
	ProcessTime     int                 `json:"processTime"`
	IsActive        bool                `json:"isActive"`
	IsFeatured      bool                `json:"isFeatured"`
	Pricing         map[string]PricingRequest `json:"pricing" validate:"required"`
	Badge           *BadgeRequest       `json:"badge"`
}

type PricingRequest struct {
	BuyPrice      float64 `json:"buyPrice" validate:"required,min=0"`
	SellPrice     float64 `json:"sellPrice" validate:"required,min=0"`
	OriginalPrice float64 `json:"originalPrice" validate:"required,min=0"`
}

type BadgeRequest struct {
	Text  string `json:"text" validate:"required,max=20"`
	Color string `json:"color" validate:"required,hexcolor"`
}

type UpdateSKURequest struct {
	Name        *string             `json:"name" validate:"omitempty,min=1,max=200"`
	Description *string             `json:"description"`
	SectionCode *string             `json:"sectionCode"`
	Image       *string             `json:"image"`
	Info        *string             `json:"info"`
	ProcessTime *int                `json:"processTime"`
	IsActive    *bool               `json:"isActive"`
	IsFeatured  *bool               `json:"isFeatured"`
	Pricing     map[string]PricingRequest `json:"pricing"`
	Badge       *BadgeRequest       `json:"badge"`
}

type BulkUpdatePriceRequest struct {
	SKUs []BulkSKUPrice `json:"skus" validate:"required,dive"`
}

type BulkSKUPrice struct {
	Code    string                     `json:"code" validate:"required"`
	Pricing map[string]BulkPricingRequest `json:"pricing" validate:"required"`
}

type BulkPricingRequest struct {
	SellPrice     *float64 `json:"sellPrice"`
	OriginalPrice *float64 `json:"originalPrice"`
}

type SyncSKURequest struct {
	ProviderCode string `json:"providerCode" validate:"required"`
	ProductCode  string `json:"productCode" validate:"required"`
	AutoActivate bool   `json:"autoActivate"`
	PriceMargin  int    `json:"priceMargin"` // percentage
}

type SyncSKUResponse struct {
	Status  string       `json:"status"`
	Summary SyncSummary  `json:"summary"`
	NewSKUs []NewSKUInfo `json:"newSkus,omitempty"`
	SyncedAt time.Time   `json:"syncedAt"`
}

type SyncSummary struct {
	TotalFromProvider int `json:"totalFromProvider"`
	NewSKUs           int `json:"newSkus"`
	UpdatedSKUs       int `json:"updatedSkus"`
	SkippedSKUs       int `json:"skippedSkus"`
}

type NewSKUInfo struct {
	ProviderSKUCode    string  `json:"providerSkuCode"`
	Name               string  `json:"name"`
	BuyPrice           float64 `json:"buyPrice"`
	SuggestedSellPrice float64 `json:"suggestedSellPrice"`
}

