package domain

import (
	"time"

	"github.com/google/uuid"
)

type HealthStatus string

const (
	HealthStatusHealthy  HealthStatus = "HEALTHY"
	HealthStatusDegraded HealthStatus = "DEGRADED"
	HealthStatusDown     HealthStatus = "DOWN"
)

type Provider struct {
	ID              uuid.UUID         `json:"id" db:"id"`
	Code            string            `json:"code" db:"code"`
	Name            string            `json:"name" db:"name"`
	BaseURL         string            `json:"baseUrl" db:"base_url"`
	WebhookURL      *string           `json:"webhookUrl" db:"webhook_url"`
	IsActive        bool              `json:"isActive" db:"is_active"`
	Priority        int               `json:"priority" db:"priority"`
	SupportedTypes  []string          `json:"supportedTypes" db:"supported_types"` // PULSA, DATA, GAME, etc
	HealthStatus    HealthStatus      `json:"healthStatus" db:"health_status"`
	LastHealthCheck *time.Time        `json:"lastHealthCheck" db:"last_health_check"`
	APIConfig       ProviderAPIConfig `json:"apiConfig" db:"api_config"`
	Mapping         ProviderMapping   `json:"mapping" db:"mapping"`
	EnvCredKeys     map[string]string `json:"envCredKeys" db:"env_cred_keys"` // Maps to env vars
	CreatedAt       time.Time         `json:"createdAt" db:"created_at"`
	UpdatedAt       time.Time         `json:"updatedAt" db:"updated_at"`
}

type ProviderAPIConfig struct {
	Timeout       int `json:"timeout"`       // milliseconds
	RetryAttempts int `json:"retryAttempts"`
	RetryDelay    int `json:"retryDelay"`    // milliseconds
}

type ProviderMapping struct {
	StatusSuccess []string `json:"statusSuccess"`
	StatusPending []string `json:"statusPending"`
	StatusFailed  []string `json:"statusFailed"`
}

// Response DTOs
type ProviderResponse struct {
	ID              string              `json:"id"`
	Code            string              `json:"code"`
	Name            string              `json:"name"`
	BaseURL         string              `json:"baseUrl"`
	WebhookURL      *string             `json:"webhookUrl,omitempty"`
	IsActive        bool                `json:"isActive"`
	Priority        int                 `json:"priority"`
	SupportedTypes  []string            `json:"supportedTypes"`
	HealthStatus    HealthStatus        `json:"healthStatus"`
	LastHealthCheck *time.Time          `json:"lastHealthCheck,omitempty"`
	Stats           *ProviderStats      `json:"stats,omitempty"`
	CreatedAt       time.Time           `json:"createdAt"`
	UpdatedAt       time.Time           `json:"updatedAt"`
}

type ProviderDetailResponse struct {
	ProviderResponse
	APIConfig   ProviderAPIConfig       `json:"apiConfig"`
	Mapping     ProviderMapping         `json:"mapping"`
	Credentials map[string]bool         `json:"credentials"` // Shows which creds exist
}

type ProviderStats struct {
	TotalSKUs        int     `json:"totalSkus"`
	ActiveSKUs       int     `json:"activeSkus"`
	SuccessRate      float64 `json:"successRate"`
	AvgResponseTime  int     `json:"avgResponseTime"` // milliseconds
	TodayTransactions int    `json:"todayTransactions,omitempty"`
	TodaySuccessRate float64 `json:"todaySuccessRate,omitempty"`
}

type TestConnectionResponse struct {
	Status       string  `json:"status"`
	ResponseTime int     `json:"responseTime"` // milliseconds
	Balance      float64 `json:"balance,omitempty"`
	Message      string  `json:"message"`
}

// Request DTOs
type CreateProviderRequest struct {
	Code           string              `json:"code" validate:"required,min=1,max=50"`
	Name           string              `json:"name" validate:"required,min=1,max=100"`
	BaseURL        string              `json:"baseUrl" validate:"required,url"`
	WebhookURL     *string             `json:"webhookUrl" validate:"omitempty,url"`
	IsActive       bool                `json:"isActive"`
	Priority       int                 `json:"priority" validate:"min=1"`
	SupportedTypes []string            `json:"supportedTypes" validate:"required,min=1"`
	APIConfig      ProviderAPIConfig   `json:"apiConfig"`
	Mapping        ProviderMapping     `json:"mapping"`
	EnvCredKeys    map[string]string   `json:"envCredentialKeys"`
}

type UpdateProviderRequest struct {
	Name           *string             `json:"name" validate:"omitempty,min=1,max=100"`
	BaseURL        *string             `json:"baseUrl" validate:"omitempty,url"`
	WebhookURL     *string             `json:"webhookUrl" validate:"omitempty,url"`
	IsActive       *bool               `json:"isActive"`
	Priority       *int                `json:"priority" validate:"omitempty,min=1"`
	SupportedTypes []string            `json:"supportedTypes"`
	APIConfig      *ProviderAPIConfig  `json:"apiConfig"`
	Mapping        *ProviderMapping    `json:"mapping"`
}

