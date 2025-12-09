package domain

import (
	"time"

	"github.com/google/uuid"
)

type TransactionStatus string

const (
	TransactionStatusPending    TransactionStatus = "PENDING"
	TransactionStatusPaid       TransactionStatus = "PAID"
	TransactionStatusProcessing TransactionStatus = "PROCESSING"
	TransactionStatusSuccess    TransactionStatus = "SUCCESS"
	TransactionStatusFailed     TransactionStatus = "FAILED"
	TransactionStatusRefunded   TransactionStatus = "REFUNDED"
	TransactionStatusExpired    TransactionStatus = "EXPIRED"
)

type PaymentStatus string

const (
	PaymentStatusUnpaid  PaymentStatus = "UNPAID"
	PaymentStatusPaid    PaymentStatus = "PAID"
	PaymentStatusExpired PaymentStatus = "EXPIRED"
)

type Transaction struct {
	ID              uuid.UUID         `json:"id" db:"id"`
	InvoiceNumber   string            `json:"invoiceNumber" db:"invoice_number"`
	UserID          *uuid.UUID        `json:"userId" db:"user_id"` // nullable for guest checkout
	ProductID       uuid.UUID         `json:"productId" db:"product_id"`
	SKUID           uuid.UUID         `json:"skuId" db:"sku_id"`
	ProviderID      uuid.UUID         `json:"providerId" db:"provider_id"`
	ChannelID       uuid.UUID         `json:"channelId" db:"channel_id"`
	GatewayID       uuid.UUID         `json:"gatewayId" db:"gateway_id"`
	PromoID         *uuid.UUID        `json:"promoId" db:"promo_id"`
	Status          TransactionStatus `json:"status" db:"status"`
	PaymentStatus   PaymentStatus     `json:"paymentStatus" db:"payment_status"`
	Quantity        int               `json:"quantity" db:"quantity"`
	AccountNickname *string           `json:"accountNickname" db:"account_nickname"`
	AccountInputs   string            `json:"accountInputs" db:"account_inputs"` // JSON stored as string
	BuyPrice        float64           `json:"buyPrice" db:"buy_price"`
	SellPrice       float64           `json:"sellPrice" db:"sell_price"`
	Discount        float64           `json:"discount" db:"discount"`
	PaymentFee      float64           `json:"paymentFee" db:"payment_fee"`
	Total           float64           `json:"total" db:"total"`
	Currency        string            `json:"currency" db:"currency"`
	Region          string            `json:"region" db:"region"`
	Email           *string           `json:"email" db:"email"`
	PhoneNumber     *string           `json:"phoneNumber" db:"phone_number"`
	IPAddress       string            `json:"ipAddress" db:"ip_address"`
	UserAgent       string            `json:"userAgent" db:"user_agent"`
	ProviderRefID   *string           `json:"providerRefId" db:"provider_ref_id"`
	SerialNumber    *string           `json:"serialNumber" db:"serial_number"`
	GatewayRefID    *string           `json:"gatewayRefId" db:"gateway_ref_id"`
	PaymentData     *string           `json:"paymentData" db:"payment_data"` // JSON for QR, VA, etc
	PaidAt          *time.Time        `json:"paidAt" db:"paid_at"`
	ProcessedAt     *time.Time        `json:"processedAt" db:"processed_at"`
	CompletedAt     *time.Time        `json:"completedAt" db:"completed_at"`
	ExpiredAt       time.Time         `json:"expiredAt" db:"expired_at"`
	CreatedAt       time.Time         `json:"createdAt" db:"created_at"`
	UpdatedAt       time.Time         `json:"updatedAt" db:"updated_at"`
}

type TransactionTimeline struct {
	ID            uuid.UUID `json:"id" db:"id"`
	TransactionID uuid.UUID `json:"transactionId" db:"transaction_id"`
	Status        string    `json:"status" db:"status"`
	Message       string    `json:"message" db:"message"`
	CreatedAt     time.Time `json:"createdAt" db:"created_at"`
}

type TransactionLog struct {
	ID            uuid.UUID              `json:"id" db:"id"`
	TransactionID uuid.UUID              `json:"transactionId" db:"transaction_id"`
	Type          string                 `json:"type" db:"type"` // PROVIDER_REQUEST, PROVIDER_RESPONSE, etc
	Data          map[string]interface{} `json:"data" db:"data"`
	CreatedAt     time.Time              `json:"createdAt" db:"created_at"`
}

type Refund struct {
	ID            uuid.UUID      `json:"id" db:"id"`
	TransactionID uuid.UUID      `json:"transactionId" db:"transaction_id"`
	Amount        float64        `json:"amount" db:"amount"`
	Currency      string         `json:"currency" db:"currency"`
	RefundTo      string         `json:"refundTo" db:"refund_to"` // BALANCE, ORIGINAL_METHOD
	Status        string         `json:"status" db:"status"` // PROCESSING, SUCCESS, FAILED
	Reason        string         `json:"reason" db:"reason"`
	ProcessedBy   uuid.UUID      `json:"processedBy" db:"processed_by"`
	CreatedAt     time.Time      `json:"createdAt" db:"created_at"`
	CompletedAt   *time.Time     `json:"completedAt" db:"completed_at"`
}

// Response DTOs
type TransactionResponse struct {
	InvoiceNumber string            `json:"invoiceNumber"`
	Status        TransactionStatus `json:"status"`
	ProductCode   string            `json:"productCode"`
	ProductName   string            `json:"productName"`
	SKUCode       string            `json:"skuCode"`
	SKUName       string            `json:"skuName"`
	Quantity      int               `json:"quantity"`
	Account       AccountInfo       `json:"account"`
	Pricing       PricingResponse   `json:"pricing"`
	Payment       PaymentInfo       `json:"payment"`
	Promo         *PromoInfo        `json:"promo,omitempty"`
	Contact       *ContactInfo      `json:"contact,omitempty"`
	Timeline      []TimelineItem    `json:"timeline,omitempty"`
	CreatedAt     time.Time         `json:"createdAt"`
	ExpiredAt     *time.Time        `json:"expiredAt,omitempty"`
	CompletedAt   *time.Time        `json:"completedAt,omitempty"`
}

type AccountInfo struct {
	Nickname *string `json:"nickname,omitempty"`
	Inputs   string  `json:"inputs"`
}

type PricingResponse struct {
	Subtotal   float64 `json:"subtotal"`
	Discount   float64 `json:"discount"`
	PaymentFee float64 `json:"paymentFee"`
	Total      float64 `json:"total"`
	Currency   string  `json:"currency"`
}

type PaymentInfo struct {
	Code        string     `json:"code"`
	Name        string     `json:"name"`
	Instruction *string    `json:"instruction,omitempty"`
	// For QRIS
	QRCode      *string    `json:"qrCode,omitempty"`
	QRCodeImage *string    `json:"qrCodeImage,omitempty"`
	// For Virtual Account
	AccountNumber *string  `json:"accountNumber,omitempty"`
	BankName      *string  `json:"bankName,omitempty"`
	AccountName   *string  `json:"accountName,omitempty"`
	// For E-Wallet redirect
	RedirectURL   *string  `json:"redirectUrl,omitempty"`
	Deeplink      *string  `json:"deeplink,omitempty"`
	// Common
	PaidAt        *time.Time `json:"paidAt,omitempty"`
	ExpiredAt     *time.Time `json:"expiredAt,omitempty"`
}

type PromoInfo struct {
	Code           string  `json:"code"`
	DiscountAmount float64 `json:"discountAmount"`
}

type ContactInfo struct {
	Email       *string `json:"email,omitempty"`
	PhoneNumber *string `json:"phoneNumber,omitempty"`
}

type TimelineItem struct {
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

type AdminTransactionResponse struct {
	ID              string            `json:"id"`
	InvoiceNumber   string            `json:"invoiceNumber"`
	Status          TransactionStatus `json:"status"`
	PaymentStatus   PaymentStatus     `json:"paymentStatus"`
	Product         ProductInfo       `json:"product"`
	SKU             SKUInfo           `json:"sku"`
	Provider        ProviderTrxInfo   `json:"provider"`
	Account         AccountInfo       `json:"account"`
	User            *UserTrxInfo      `json:"user,omitempty"`
	Pricing         AdminPricingInfo  `json:"pricing"`
	Payment         AdminPaymentInfo  `json:"payment"`
	Promo           *PromoInfo        `json:"promo,omitempty"`
	Region          string            `json:"region"`
	IPAddress       string            `json:"ipAddress"`
	UserAgent       string            `json:"userAgent"`
	CreatedAt       time.Time         `json:"createdAt"`
	CompletedAt     *time.Time        `json:"completedAt,omitempty"`
}

type SKUInfo struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type ProviderTrxInfo struct {
	Code   string  `json:"code"`
	Name   string  `json:"name"`
	RefID  *string `json:"refId,omitempty"`
}

type UserTrxInfo struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Email string  `json:"email"`
}

type AdminPricingInfo struct {
	BuyPrice   float64 `json:"buyPrice"`
	SellPrice  float64 `json:"sellPrice"`
	Discount   float64 `json:"discount"`
	PaymentFee float64 `json:"paymentFee"`
	Total      float64 `json:"total"`
	Profit     float64 `json:"profit"`
	Currency   string  `json:"currency"`
}

type AdminPaymentInfo struct {
	Code         string     `json:"code"`
	Name         string     `json:"name"`
	Gateway      string     `json:"gateway"`
	GatewayRefID *string    `json:"gatewayRefId,omitempty"`
	PaidAt       *time.Time `json:"paidAt,omitempty"`
}

type TransactionDetailResponse struct {
	AdminTransactionResponse
	Timeline []TimelineItem   `json:"timeline"`
	Logs     []TransactionLogInfo `json:"logs,omitempty"`
}

type TransactionLogInfo struct {
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

type TransactionOverview struct {
	TotalTransactions int     `json:"totalTransactions"`
	TotalRevenue      float64 `json:"totalRevenue"`
	TotalProfit       float64 `json:"totalProfit"`
	SuccessCount      int     `json:"successCount"`
	ProcessingCount   int     `json:"processingCount"`
	PendingCount      int     `json:"pendingCount"`
	FailedCount       int     `json:"failedCount"`
}

type RefundResponse struct {
	RefundID      string     `json:"refundId"`
	TransactionID string     `json:"transactionId"`
	InvoiceNumber string     `json:"invoiceNumber"`
	Amount        float64    `json:"amount"`
	Currency      string     `json:"currency"`
	RefundTo      string     `json:"refundTo"`
	Status        string     `json:"status"`
	Reason        string     `json:"reason"`
	ProcessedBy   CreatedBy  `json:"processedBy"`
	CreatedAt     time.Time  `json:"createdAt"`
}

// Request DTOs
type AccountInquiryRequest struct {
	ProductCode string `json:"productCode" validate:"required"`
	// Dynamic fields based on product
}

type AccountInquiryResponse struct {
	Product ProductBasic `json:"product"`
	Account AccountData  `json:"account"`
}

type ProductBasic struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type AccountData struct {
	Region   string `json:"region"`
	Nickname string `json:"nickname"`
}

type OrderInquiryRequest struct {
	ProductCode string  `json:"productCode" validate:"required"`
	SKUCode     string  `json:"skuCode" validate:"required"`
	Quantity    int     `json:"quantity" validate:"required,min=1"`
	PaymentCode string  `json:"paymentCode" validate:"required"`
	PromoCode   *string `json:"promoCode"`
	Email       *string `json:"email" validate:"omitempty,email"`
	PhoneNumber *string `json:"phoneNumber"`
	// Dynamic account fields
}

type OrderInquiryResponse struct {
	ValidationToken string           `json:"validationToken"`
	ExpiresAt       time.Time        `json:"expiresAt"`
	Order           OrderPreview     `json:"order"`
}

type OrderPreview struct {
	Product  ProductBasic     `json:"product"`
	SKU      SKUBasic         `json:"sku"`
	Account  AccountData      `json:"account"`
	Payment  PaymentBasic     `json:"payment"`
	Pricing  PricingResponse  `json:"pricing"`
	Promo    *PromoInfo       `json:"promo,omitempty"`
	Contact  *ContactInfo     `json:"contact,omitempty"`
}

type SKUBasic struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
}

type PaymentBasic struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Currency string `json:"currency"`
}

type CreateOrderRequest struct {
	ValidationToken string `json:"validationToken" validate:"required"`
}

type CreateOrderResponse struct {
	Step  string              `json:"step"`
	Order TransactionResponse `json:"order"`
}

type UpdateTransactionStatusRequest struct {
	Status       TransactionStatus `json:"status" validate:"required"`
	Reason       string            `json:"reason" validate:"required"`
	SerialNumber *string           `json:"serialNumber"`
}

type RefundTransactionRequest struct {
	Reason   string  `json:"reason" validate:"required"`
	RefundTo string  `json:"refundTo" validate:"required,oneof=BALANCE ORIGINAL_METHOD"`
	Amount   float64 `json:"amount" validate:"required,min=0"`
}

type RetryTransactionRequest struct {
	ProviderCode string `json:"providerCode" validate:"required"`
	Reason       string `json:"reason" validate:"required"`
}

type ManualProcessRequest struct {
	SerialNumber string `json:"serialNumber" validate:"required"`
	Reason       string `json:"reason" validate:"required"`
}

