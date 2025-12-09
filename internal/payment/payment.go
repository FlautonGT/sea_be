package payment

import (
	"context"
	"time"
)

// Gateway represents a payment gateway interface
type Gateway interface {
	// GetName returns the gateway name
	GetName() string

	// GetSupportedChannels returns the list of supported payment channels
	GetSupportedChannels() []string

	// CreatePayment creates a new payment
	CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error)

	// CheckStatus checks the status of a payment
	CheckStatus(ctx context.Context, paymentID string) (*PaymentStatus, error)

	// HealthCheck checks if the gateway is accessible
	HealthCheck(ctx context.Context) error
}

// PaymentRequest represents a payment request
type PaymentRequest struct {
	RefID          string            `json:"ref_id"`
	Amount         float64           `json:"amount"`
	Currency       string            `json:"currency"`
	Channel        string            `json:"channel"`        // Payment channel code (e.g., QRIS, BCA_VA)
	GatewayName    string            `json:"gateway_name"`   // Gateway to use (e.g., dana, xendit, midtrans)
	GatewayCode    string            `json:"gateway_code"`   // Gateway-specific code (e.g., 002 for BRI)
	Description    string            `json:"description"`
	CustomerName   string            `json:"customer_name"`
	CustomerEmail  string            `json:"customer_email"`
	CustomerPhone  string            `json:"customer_phone"`
	ExpiryDuration time.Duration     `json:"expiry_duration"`
	CallbackURL    string            `json:"callback_url"`
	SuccessURL     string            `json:"success_url"`
	FailureURL     string            `json:"failure_url"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

// PaymentResponse represents the response from creating a payment
type PaymentResponse struct {
	RefID           string    `json:"ref_id"`
	GatewayRefID    string    `json:"gateway_ref_id"`
	Channel         string    `json:"channel"`
	Amount          float64   `json:"amount"`
	Fee             float64   `json:"fee"`
	TotalAmount     float64   `json:"total_amount"`
	Currency        string    `json:"currency"`
	Status          string    `json:"status"`
	PaymentCode     string    `json:"payment_code,omitempty"`
	PaymentURL      string    `json:"payment_url,omitempty"`
	QRCode          string    `json:"qr_code,omitempty"`
	QRCodeURL       string    `json:"qr_code_url,omitempty"`
	VirtualAccount  string    `json:"virtual_account,omitempty"`
	AccountName     string    `json:"account_name,omitempty"`
	BankCode        string    `json:"bank_code,omitempty"`
	ExpiresAt       time.Time `json:"expires_at"`
	CreatedAt       time.Time `json:"created_at"`
	Instructions    []string  `json:"instructions,omitempty"`
}

// PaymentStatus represents the status of a payment
type PaymentStatus struct {
	RefID        string    `json:"ref_id"`
	GatewayRefID string    `json:"gateway_ref_id"`
	Status       string    `json:"status"`
	Amount       float64   `json:"amount"`
	Fee          float64   `json:"fee"`
	PaidAt       time.Time `json:"paid_at,omitempty"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// PaymentChannel represents a payment channel configuration
type PaymentChannel struct {
	Code          string   `json:"code"`
	Name          string   `json:"name"`
	Type          string   `json:"type"`
	Gateway       string   `json:"gateway"`
	IsActive      bool     `json:"is_active"`
	MinAmount     float64  `json:"min_amount"`
	MaxAmount     float64  `json:"max_amount"`
	FeeFixed      float64  `json:"fee_fixed"`
	FeePercentage float64  `json:"fee_percentage"`
	Icon          string   `json:"icon"`
	Instructions  []string `json:"instructions"`
}

// Common payment status constants
const (
	PaymentStatusPending  = "PENDING"
	PaymentStatusPaid     = "PAID"
	PaymentStatusExpired  = "EXPIRED"
	PaymentStatusFailed   = "FAILED"
	PaymentStatusRefunded = "REFUNDED"
)

// Payment channel types
const (
	ChannelTypeQRIS           = "QRIS"
	ChannelTypeVirtualAccount = "VIRTUAL_ACCOUNT"
	ChannelTypeEWallet        = "E_WALLET"
	ChannelTypeRetail         = "RETAIL"
	ChannelTypeCreditCard     = "CREDIT_CARD"
)

// GatewayConfig represents common gateway configuration
type GatewayConfig struct {
	Name         string
	BaseURL      string
	Timeout      time.Duration
	IsProduction bool
}

