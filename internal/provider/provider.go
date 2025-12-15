package provider

import (
	"context"
	"encoding/json"
	"time"
)

// Provider represents a product provider interface
type Provider interface {
	// GetName returns the provider name
	GetName() string

	// GetProducts fetches all available products from the provider
	GetProducts(ctx context.Context) ([]Product, error)

	// CheckPrice checks the current price for a product
	CheckPrice(ctx context.Context, sku string) (*PriceInfo, error)

	// CreateOrder creates an order with the provider
	CreateOrder(ctx context.Context, req *OrderRequest) (*OrderResponse, error)

	// CheckStatus checks the status of an order
	CheckStatus(ctx context.Context, refID string) (*OrderStatus, error)

	// GetBalance returns the current balance with the provider
	GetBalance(ctx context.Context) (*Balance, error)

	// HealthCheck checks if the provider is accessible
	HealthCheck(ctx context.Context) error
}

// Product represents a product from a provider
type Product struct {
	SKU           string  `json:"sku"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	Category      string  `json:"category"`
	Brand         string  `json:"brand"`
	Type          string  `json:"type"`
	SellerPrice   float64 `json:"seller_price"`
	Price         float64 `json:"price"`
	BuyerSKUCode  string  `json:"buyer_sku_code"`
	IsActive      bool    `json:"is_active"`
	IsAvailable   bool    `json:"is_available"`
	Stock         int     `json:"stock"`
	MultiStock    bool    `json:"multi_stock"`
	StartCutOff   string  `json:"start_cut_off"`
	EndCutOff     string  `json:"end_cut_off"`
	Unlimited     bool    `json:"unlimited"`
}

// PriceInfo represents price information for a product
type PriceInfo struct {
	SKU         string  `json:"sku"`
	Price       float64 `json:"price"`
	SellerPrice float64 `json:"seller_price"`
	IsAvailable bool    `json:"is_available"`
	Stock       int     `json:"stock"`
}

// OrderRequest represents an order request to the provider
type OrderRequest struct {
	RefID        string `json:"ref_id"`
	SKU          string `json:"sku"`
	CustomerNo   string `json:"customer_no"`
	CustomerData map[string]string `json:"customer_data,omitempty"`
}

// OrderResponse represents the response from creating an order
type OrderResponse struct {
	RefID         string    `json:"ref_id"`
	ProviderRefID string    `json:"provider_ref_id"`
	SKU           string    `json:"sku"`
	CustomerNo    string    `json:"customer_no"`
	Price         float64   `json:"price"`
	SellingPrice  float64   `json:"selling_price"`
	Status        string    `json:"status"`
	Message       string    `json:"message"`
	SN            string    `json:"sn"`
	CreatedAt     time.Time `json:"created_at"`
	// Raw request/response for logging purposes
	RawRequest  json.RawMessage `json:"raw_request,omitempty"`
	RawResponse json.RawMessage `json:"raw_response,omitempty"`
}

// OrderStatus represents the status of an order
type OrderStatus struct {
	RefID         string    `json:"ref_id"`
	ProviderRefID string    `json:"provider_ref_id"`
	Status        string    `json:"status"`
	Message       string    `json:"message"`
	SN            string    `json:"sn"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Balance represents the account balance with a provider
type Balance struct {
	Balance    float64   `json:"balance"`
	Currency   string    `json:"currency"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ProviderConfig represents common provider configuration
type ProviderConfig struct {
	Name          string
	BaseURL       string
	Timeout       time.Duration
	MaxRetries    int
	RetryDelay    time.Duration
}

// Common status constants
const (
	StatusPending    = "PENDING"
	StatusProcessing = "PROCESSING"
	StatusSuccess    = "SUCCESS"
	StatusFailed     = "FAILED"
)

