package domain

import (
	"time"

	"github.com/google/uuid"
)

type Region struct {
	ID             uuid.UUID `json:"id" db:"id"`
	Code           string    `json:"code" db:"code"`
	Country        string    `json:"country" db:"country"`
	Currency       string    `json:"currency" db:"currency"`
	CurrencySymbol string    `json:"currencySymbol" db:"currency_symbol"`
	Image          *string   `json:"image" db:"image"`
	IsDefault      bool      `json:"isDefault" db:"is_default"`
	IsActive       bool      `json:"isActive" db:"is_active"`
	Order          int       `json:"order" db:"order"`
	CreatedAt      time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt      time.Time `json:"updatedAt" db:"updated_at"`
}

type Language struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Code      string    `json:"code" db:"code"`
	Name      string    `json:"name" db:"name"`
	Country   string    `json:"country" db:"country"`
	Image     *string   `json:"image" db:"image"`
	IsDefault bool      `json:"isDefault" db:"is_default"`
	IsActive  bool      `json:"isActive" db:"is_active"`
	Order     int       `json:"order" db:"order"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

// Response DTOs
type RegionResponse struct {
	Country   string  `json:"country"`
	Code      string  `json:"code"`
	Currency  string  `json:"currency"`
	Image     *string `json:"image,omitempty"`
	IsDefault bool    `json:"isDefault"`
}

type LanguageResponse struct {
	Country   string  `json:"country"`
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	Image     *string `json:"image,omitempty"`
	IsDefault bool    `json:"isDefault"`
}

type AdminRegionResponse struct {
	ID             string       `json:"id"`
	Code           string       `json:"code"`
	Country        string       `json:"country"`
	Currency       string       `json:"currency"`
	CurrencySymbol string       `json:"currencySymbol"`
	Image          *string      `json:"image,omitempty"`
	IsDefault      bool         `json:"isDefault"`
	IsActive       bool         `json:"isActive"`
	Order          int          `json:"order"`
	Stats          *RegionStats `json:"stats,omitempty"`
	CreatedAt      time.Time    `json:"createdAt"`
	UpdatedAt      time.Time    `json:"updatedAt"`
}

type RegionStats struct {
	TotalUsers        int     `json:"totalUsers"`
	TotalTransactions int     `json:"totalTransactions"`
	TotalRevenue      float64 `json:"totalRevenue"`
}

type AdminLanguageResponse struct {
	ID        string    `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Country   string    `json:"country"`
	Image     *string   `json:"image,omitempty"`
	IsDefault bool      `json:"isDefault"`
	IsActive  bool      `json:"isActive"`
	Order     int       `json:"order"`
	CreatedAt time.Time `json:"createdAt"`
}

// Request DTOs
type CreateRegionRequest struct {
	Code           string `json:"code" validate:"required,len=2"`
	Country        string `json:"country" validate:"required,min=1,max=100"`
	Currency       string `json:"currency" validate:"required,len=3"`
	CurrencySymbol string `json:"currencySymbol" validate:"required,min=1,max=5"`
	IsDefault      bool   `json:"isDefault"`
	IsActive       bool   `json:"isActive"`
	Order          int    `json:"order" validate:"min=0"`
}

type UpdateRegionRequest struct {
	Country        *string `json:"country" validate:"omitempty,min=1,max=100"`
	CurrencySymbol *string `json:"currencySymbol" validate:"omitempty,min=1,max=5"`
	IsActive       *bool   `json:"isActive"`
	Order          *int    `json:"order" validate:"omitempty,min=0"`
}

type CreateLanguageRequest struct {
	Code      string `json:"code" validate:"required,len=2"`
	Name      string `json:"name" validate:"required,min=1,max=100"`
	Country   string `json:"country" validate:"required,min=1,max=100"`
	IsDefault bool   `json:"isDefault"`
	IsActive  bool   `json:"isActive"`
	Order     int    `json:"order" validate:"min=0"`
}

type UpdateLanguageRequest struct {
	Name      *string `json:"name" validate:"omitempty,min=1,max=100"`
	Country   *string `json:"country" validate:"omitempty,min=1,max=100"`
	IsActive  *bool   `json:"isActive"`
	Order     *int    `json:"order" validate:"omitempty,min=0"`
}

