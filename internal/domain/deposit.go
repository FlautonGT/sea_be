package domain

import (
	"time"

	"github.com/google/uuid"
)

type DepositStatus string

const (
	DepositStatusPending   DepositStatus = "PENDING"
	DepositStatusSuccess   DepositStatus = "SUCCESS"
	DepositStatusExpired   DepositStatus = "EXPIRED"
	DepositStatusFailed    DepositStatus = "FAILED"
	DepositStatusCancelled DepositStatus = "CANCELLED"
	DepositStatusRefunded  DepositStatus = "REFUNDED"
)

type Deposit struct {
	ID            uuid.UUID     `json:"id" db:"id"`
	InvoiceNumber string        `json:"invoiceNumber" db:"invoice_number"`
	UserID        uuid.UUID     `json:"userId" db:"user_id"`
	ChannelID     uuid.UUID     `json:"channelId" db:"channel_id"`
	GatewayID     uuid.UUID     `json:"gatewayId" db:"gateway_id"`
	Amount        float64       `json:"amount" db:"amount"`
	PaymentFee    float64       `json:"paymentFee" db:"payment_fee"`
	Total         float64       `json:"total" db:"total"`
	Currency      string        `json:"currency" db:"currency"`
	Region        string        `json:"region" db:"region"`
	Status        DepositStatus `json:"status" db:"status"`
	GatewayRefID  *string       `json:"gatewayRefId" db:"gateway_ref_id"`
	PaymentData   *string       `json:"paymentData" db:"payment_data"` // JSON
	IPAddress     string        `json:"ipAddress" db:"ip_address"`
	UserAgent     string        `json:"userAgent" db:"user_agent"`
	PaidAt        *time.Time    `json:"paidAt" db:"paid_at"`
	ExpiredAt     time.Time     `json:"expiredAt" db:"expired_at"`
	CreatedAt     time.Time     `json:"createdAt" db:"created_at"`
	UpdatedAt     time.Time     `json:"updatedAt" db:"updated_at"`
}

type DepositTimeline struct {
	ID        uuid.UUID `json:"id" db:"id"`
	DepositID uuid.UUID `json:"depositId" db:"deposit_id"`
	Status    string    `json:"status" db:"status"`
	Message   string    `json:"message" db:"message"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

type DepositRefund struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	DepositID   uuid.UUID  `json:"depositId" db:"deposit_id"`
	Amount      float64    `json:"amount" db:"amount"`
	Currency    string     `json:"currency" db:"currency"`
	RefundTo    string     `json:"refundTo" db:"refund_to"`
	Status      string     `json:"status" db:"status"`
	Reason      string     `json:"reason" db:"reason"`
	ProcessedBy uuid.UUID  `json:"processedBy" db:"processed_by"`
	CreatedAt   time.Time  `json:"createdAt" db:"created_at"`
	CompletedAt *time.Time `json:"completedAt" db:"completed_at"`
}

// Response DTOs
type DepositResponse struct {
	InvoiceNumber string        `json:"invoiceNumber"`
	Status        DepositStatus `json:"status"`
	Amount        float64       `json:"amount"`
	Pricing       PricingResponse `json:"pricing"`
	Payment       PaymentInfo   `json:"payment"`
	CreatedAt     time.Time     `json:"createdAt"`
	ExpiredAt     *time.Time    `json:"expiredAt,omitempty"`
	PaidAt        *time.Time    `json:"paidAt,omitempty"`
}

type DepositListResponse struct {
	InvoiceNumber string        `json:"invoiceNumber"`
	Status        DepositStatus `json:"status"`
	Amount        float64       `json:"amount"`
	Payment       PaymentBasic  `json:"payment"`
	Currency      string        `json:"currency"`
	CreatedAt     time.Time     `json:"createdAt"`
	PaidAt        *time.Time    `json:"paidAt,omitempty"`
	ExpiredAt     *time.Time    `json:"expiredAt,omitempty"`
}

type DepositOverview struct {
	TotalDeposits int     `json:"totalDeposits"`
	TotalAmount   float64 `json:"totalAmount"`
	SuccessCount  int     `json:"successCount"`
	PendingCount  int     `json:"pendingCount"`
	FailedCount   int     `json:"failedCount"`
}

type AdminDepositResponse struct {
	ID            string        `json:"id"`
	InvoiceNumber string        `json:"invoiceNumber"`
	User          UserTrxInfo   `json:"user"`
	Amount        float64       `json:"amount"`
	Currency      string        `json:"currency"`
	Status        DepositStatus `json:"status"`
	Payment       AdminDepositPayment `json:"payment"`
	Region        string        `json:"region"`
	IPAddress     string        `json:"ipAddress"`
	CreatedAt     time.Time     `json:"createdAt"`
	PaidAt        *time.Time    `json:"paidAt,omitempty"`
}

type AdminDepositPayment struct {
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	Gateway       string  `json:"gateway"`
	GatewayRefID  *string `json:"gatewayRefId,omitempty"`
	AccountNumber *string `json:"accountNumber,omitempty"`
}

type AdminDepositDetailResponse struct {
	AdminDepositResponse
	Pricing       PricingResponse `json:"pricing"`
	Timeline      []TimelineItem  `json:"timeline"`
	BalanceChange *BalanceChange  `json:"balanceChange,omitempty"`
	UserAgent     string          `json:"userAgent"`
}

type BalanceChange struct {
	Before   float64 `json:"before"`
	After    float64 `json:"after"`
	Currency string  `json:"currency"`
}

type DepositConfirmResponse struct {
	ID            string        `json:"id"`
	InvoiceNumber string        `json:"invoiceNumber"`
	Status        DepositStatus `json:"status"`
	Amount        float64       `json:"amount"`
	Currency      string        `json:"currency"`
	BalanceChange BalanceChange `json:"balanceChange"`
	ConfirmedBy   CreatedBy     `json:"confirmedBy"`
	ConfirmedAt   time.Time     `json:"confirmedAt"`
}

type DepositRefundResponse struct {
	RefundID      string        `json:"refundId"`
	DepositID     string        `json:"depositId"`
	InvoiceNumber string        `json:"invoiceNumber"`
	Amount        float64       `json:"amount"`
	Currency      string        `json:"currency"`
	RefundTo      string        `json:"refundTo"`
	Status        string        `json:"status"`
	Reason        string        `json:"reason"`
	BalanceChange *BalanceChange `json:"balanceChange,omitempty"`
	ProcessedBy   CreatedBy     `json:"processedBy"`
	CreatedAt     time.Time     `json:"createdAt"`
}

// Request DTOs
type DepositInquiryRequest struct {
	Amount      float64 `json:"amount" validate:"required,min=1000"`
	PaymentCode string  `json:"paymentCode" validate:"required"`
}

type DepositInquiryResponse struct {
	ValidationToken string           `json:"validationToken"`
	ExpiresAt       time.Time        `json:"expiresAt"`
	Deposit         DepositPreview   `json:"deposit"`
}

type DepositPreview struct {
	Amount  float64         `json:"amount"`
	Pricing PricingResponse `json:"pricing"`
	Payment DepositPaymentPreview `json:"payment"`
}

type DepositPaymentPreview struct {
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	Currency      string  `json:"currency"`
	MinAmount     float64 `json:"minAmount"`
	MaxAmount     float64 `json:"maxAmount"`
	FeeAmount     float64 `json:"feeAmount"`
	FeePercentage float64 `json:"feePercentage"`
}

type CreateDepositRequest struct {
	ValidationToken string `json:"validationToken" validate:"required"`
}

type CreateDepositResponse struct {
	Step    string          `json:"step"`
	Deposit DepositResponse `json:"deposit"`
}

type ConfirmDepositRequest struct {
	Reason       string  `json:"reason" validate:"required"`
	GatewayRefID *string `json:"gatewayRefId"`
}

type CancelDepositRequest struct {
	Reason string `json:"reason" validate:"required"`
}

type RefundDepositRequest struct {
	Reason   string  `json:"reason" validate:"required"`
	RefundTo string  `json:"refundTo" validate:"required,oneof=BALANCE ORIGINAL_METHOD"`
	Amount   float64 `json:"amount" validate:"required,min=0"`
}

