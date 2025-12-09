package domain

import (
	"time"

	"github.com/google/uuid"
)

type MutationType string

const (
	MutationTypeDebit  MutationType = "DEBIT"
	MutationTypeCredit MutationType = "CREDIT"
)

type Mutation struct {
	ID            uuid.UUID    `json:"id" db:"id"`
	UserID        uuid.UUID    `json:"userId" db:"user_id"`
	InvoiceNumber string       `json:"invoiceNumber" db:"invoice_number"`
	Type          MutationType `json:"type" db:"type"`
	Amount        float64      `json:"amount" db:"amount"`
	BalanceBefore float64      `json:"balanceBefore" db:"balance_before"`
	BalanceAfter  float64      `json:"balanceAfter" db:"balance_after"`
	Currency      string       `json:"currency" db:"currency"`
	Description   string       `json:"description" db:"description"`
	ReferenceType string       `json:"referenceType" db:"reference_type"` // TRANSACTION, DEPOSIT, REFUND, ADJUSTMENT
	ReferenceID   uuid.UUID    `json:"referenceId" db:"reference_id"`
	CreatedAt     time.Time    `json:"createdAt" db:"created_at"`
}

// Response DTOs
type MutationResponse struct {
	InvoiceNumber string       `json:"invoiceNumber"`
	Description   string       `json:"description"`
	Amount        float64      `json:"amount"`
	Type          MutationType `json:"type"`
	BalanceBefore float64      `json:"balanceBefore"`
	BalanceAfter  float64      `json:"balanceAfter"`
	Currency      string       `json:"currency"`
	CreatedAt     time.Time    `json:"createdAt"`
}

type MutationOverview struct {
	TotalDebit       float64 `json:"totalDebit"`
	TotalCredit      float64 `json:"totalCredit"`
	NetBalance       float64 `json:"netBalance"`
	TransactionCount int     `json:"transactionCount"`
}

// Report Models
type DailyReport struct {
	Date              time.Time `json:"date" db:"date"`
	TotalTransactions int       `json:"totalTransactions" db:"total_transactions"`
	TotalAmount       float64   `json:"totalAmount" db:"total_amount"`
	Currency          string    `json:"currency" db:"currency"`
}

type ReportOverview struct {
	TotalDays         int              `json:"totalDays"`
	TotalTransactions int              `json:"totalTransactions"`
	TotalAmount       float64          `json:"totalAmount"`
	AveragePerDay     float64          `json:"averagePerDay"`
	HighestDay        *DayStats        `json:"highestDay,omitempty"`
	LowestDay         *DayStats        `json:"lowestDay,omitempty"`
}

type DayStats struct {
	Date   string  `json:"date"`
	Amount float64 `json:"amount"`
}

type ReportResponse struct {
	Date              string  `json:"date"`
	TotalTransactions int     `json:"totalTransactions"`
	TotalAmount       float64 `json:"totalAmount"`
	Currency          string  `json:"currency"`
}

// Admin Report Models
type DashboardOverview struct {
	Summary      DashboardSummary      `json:"summary"`
	RevenueChart []RevenueChartItem    `json:"revenueChart"`
	TopProducts  []TopProductItem      `json:"topProducts"`
	TopPayments  []TopPaymentItem      `json:"topPayments"`
	ProviderHealth []ProviderHealthItem `json:"providerHealth"`
}

type DashboardSummary struct {
	TotalRevenue      float64 `json:"totalRevenue"`
	TotalProfit       float64 `json:"totalProfit"`
	TotalTransactions int     `json:"totalTransactions"`
	TotalUsers        int     `json:"totalUsers"`
	NewUsers          int     `json:"newUsers"`
	ActiveUsers       int     `json:"activeUsers"`
}

type RevenueChartItem struct {
	Date    string  `json:"date"`
	Revenue float64 `json:"revenue"`
	Profit  float64 `json:"profit"`
}

type TopProductItem struct {
	Code         string  `json:"code"`
	Name         string  `json:"name"`
	Revenue      float64 `json:"revenue"`
	Transactions int     `json:"transactions"`
}

type TopPaymentItem struct {
	Code         string  `json:"code"`
	Name         string  `json:"name"`
	Revenue      float64 `json:"revenue"`
	Transactions int     `json:"transactions"`
}

type ProviderHealthItem struct {
	Code        string       `json:"code"`
	Status      HealthStatus `json:"status"`
	SuccessRate float64      `json:"successRate"`
}

// Export Models
type ExportStatus string

const (
	ExportStatusProcessing ExportStatus = "PROCESSING"
	ExportStatusCompleted  ExportStatus = "COMPLETED"
	ExportStatusFailed     ExportStatus = "FAILED"
)

type Export struct {
	ID          uuid.UUID    `json:"id" db:"id"`
	AdminID     uuid.UUID    `json:"adminId" db:"admin_id"`
	ReportType  string       `json:"reportType" db:"report_type"`
	Format      string       `json:"format" db:"format"` // xlsx, csv
	Filters     string       `json:"filters" db:"filters"` // JSON
	Status      ExportStatus `json:"status" db:"status"`
	FilePath    *string      `json:"filePath" db:"file_path"`
	DownloadURL *string      `json:"downloadUrl" db:"download_url"`
	ExpiresAt   *time.Time   `json:"expiresAt" db:"expires_at"`
	CreatedAt   time.Time    `json:"createdAt" db:"created_at"`
	CompletedAt *time.Time   `json:"completedAt" db:"completed_at"`
}

type ExportRequest struct {
	ReportType string              `json:"reportType" validate:"required,oneof=transactions revenue products providers users"`
	Format     string              `json:"format" validate:"required,oneof=xlsx csv"`
	Filters    map[string]string   `json:"filters"`
}

type ExportResponse struct {
	ExportID    string     `json:"exportId"`
	Status      ExportStatus `json:"status"`
	DownloadURL *string    `json:"downloadUrl,omitempty"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
}

