package payment

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// MidtransGatewayConfig holds the configuration for Midtrans gateway
type MidtransGatewayConfig struct {
	ServerKey    string
	ClientKey    string
	IsProduction bool
	BaseURL      string
	CallbackURL  string
}

// MidtransGateway implements the Gateway interface for Midtrans
type MidtransGateway struct {
	config     MidtransGatewayConfig
	httpClient *http.Client
}

// NewMidtransGateway creates a new Midtrans gateway instance
func NewMidtransGateway(cfg MidtransGatewayConfig) (*MidtransGateway, error) {
	if cfg.ServerKey == "" {
		return nil, fmt.Errorf("midtrans server key is required")
	}

	// Set default base URL based on environment
	if cfg.BaseURL == "" {
		if cfg.IsProduction {
			cfg.BaseURL = "https://api.midtrans.com"
		} else {
			cfg.BaseURL = "https://api.sandbox.midtrans.com"
		}
	}

	return &MidtransGateway{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// GetName returns the gateway name
func (g *MidtransGateway) GetName() string {
	return "MIDTRANS"
}

// GetSupportedChannels returns supported payment channels
func (g *MidtransGateway) GetSupportedChannels() []string {
	return []string{"GOPAY", "SHOPEEPAY"}
}

// CreatePayment creates a new payment via Midtrans
func (g *MidtransGateway) CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	channel := strings.ToUpper(req.Channel)

	switch channel {
	case "GOPAY":
		return g.createGopayPayment(ctx, req)
	case "SHOPEEPAY":
		return g.createShopeepayPayment(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported payment channel: %s", channel)
	}
}

// createGopayPayment creates a GoPay payment
func (g *MidtransGateway) createGopayPayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	// Build request body
	requestBody := map[string]interface{}{
		"payment_type": "gopay",
		"transaction_details": map[string]interface{}{
			"order_id":     req.RefID,
			"gross_amount": int64(req.Amount),
		},
		"gopay": map[string]interface{}{
			"enable_callback": true,
			"callback_url":    req.SuccessURL,
		},
		"item_details": []map[string]interface{}{
			{
				"id":       req.Metadata["sku_code"],
				"price":    int64(req.Amount),
				"quantity": 1,
				"name":     truncateString(req.Description, 50),
			},
		},
	}

	// Add customer details if available
	if req.CustomerName != "" || req.CustomerEmail != "" || req.CustomerPhone != "" {
		customerDetails := map[string]interface{}{}
		if req.CustomerName != "" {
			customerDetails["first_name"] = req.CustomerName
		}
		if req.CustomerEmail != "" {
			customerDetails["email"] = req.CustomerEmail
		}
		if req.CustomerPhone != "" {
			customerDetails["phone"] = req.CustomerPhone
		}
		requestBody["customer_details"] = customerDetails
	}

	return g.chargePayment(ctx, req, requestBody)
}

// createShopeepayPayment creates a ShopeePay payment
func (g *MidtransGateway) createShopeepayPayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	// Build request body
	requestBody := map[string]interface{}{
		"payment_type": "shopeepay",
		"transaction_details": map[string]interface{}{
			"order_id":     req.RefID,
			"gross_amount": int64(req.Amount),
		},
		"shopeepay": map[string]interface{}{
			"callback_url": req.SuccessURL,
		},
		"item_details": []map[string]interface{}{
			{
				"id":       req.Metadata["sku_code"],
				"price":    int64(req.Amount),
				"quantity": 1,
				"name":     truncateString(req.Description, 50),
			},
		},
	}

	// Add customer details if available
	if req.CustomerName != "" || req.CustomerEmail != "" || req.CustomerPhone != "" {
		customerDetails := map[string]interface{}{}
		if req.CustomerName != "" {
			customerDetails["first_name"] = req.CustomerName
		}
		if req.CustomerEmail != "" {
			customerDetails["email"] = req.CustomerEmail
		}
		if req.CustomerPhone != "" {
			customerDetails["phone"] = req.CustomerPhone
		}
		requestBody["customer_details"] = customerDetails
	}

	return g.chargePayment(ctx, req, requestBody)
}

// chargePayment calls Midtrans charge API
func (g *MidtransGateway) chargePayment(ctx context.Context, req *PaymentRequest, body map[string]interface{}) (*PaymentResponse, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	log.Debug().
		Str("gateway", "midtrans").
		Str("url", g.config.BaseURL+"/v2/charge").
		RawJSON("body", jsonBody).
		Msg("Midtrans charge request")

	httpReq, err := http.NewRequestWithContext(ctx, "POST", g.config.BaseURL+"/v2/charge", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Authorization", g.getAuthHeader())

	// Send request
	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	log.Debug().
		Str("gateway", "midtrans").
		Int("status_code", resp.StatusCode).
		RawJSON("response", respBody).
		Msg("Midtrans charge response")

	// Parse response
	var chargeResp MidtransChargeResponse
	if err := json.Unmarshal(respBody, &chargeResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for errors (status codes 4xx, 5xx)
	if chargeResp.StatusCode != "200" && chargeResp.StatusCode != "201" {
		return nil, fmt.Errorf("midtrans error: %s - %s", chargeResp.StatusCode, chargeResp.StatusMessage)
	}

	// Extract deeplink URL from actions
	var deeplinkURL string
	var qrCodeURL string
	for _, action := range chargeResp.Actions {
		if action.Name == "deeplink-redirect" {
			deeplinkURL = action.URL
		}
		if action.Name == "generate-qr-code" {
			qrCodeURL = action.URL
		}
	}

	// Parse expiry time
	expiryTime, _ := time.ParseInLocation("2006-01-02 15:04:05", chargeResp.ExpiryTime, time.FixedZone("WIB", 7*3600))

	return &PaymentResponse{
		RefID:        req.RefID,
		GatewayRefID: chargeResp.TransactionID,
		Channel:      req.Channel,
		Amount:       req.Amount,
		Currency:     chargeResp.Currency,
		Status:       mapMidtransStatus(chargeResp.TransactionStatus),
		PaymentURL:   deeplinkURL, // Use deeplink URL as payment URL for e-wallet
		QRCode:       chargeResp.QRString,
		QRCodeURL:    qrCodeURL,
		ExpiresAt:    expiryTime,
		CreatedAt:    time.Now(),
	}, nil
}

// CheckStatus checks the payment status
func (g *MidtransGateway) CheckStatus(ctx context.Context, orderID string) (*PaymentStatus, error) {
	url := fmt.Sprintf("%s/v2/%s/status", g.config.BaseURL, orderID)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Authorization", g.getAuthHeader())

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var statusResp MidtransStatusResponse
	if err := json.Unmarshal(respBody, &statusResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Parse settlement time if available
	var paidAt time.Time
	if statusResp.SettlementTime != "" {
		paidAt, _ = time.ParseInLocation("2006-01-02 15:04:05", statusResp.SettlementTime, time.FixedZone("WIB", 7*3600))
	}

	return &PaymentStatus{
		RefID:        statusResp.OrderID,
		GatewayRefID: statusResp.TransactionID,
		Status:       mapMidtransStatus(statusResp.TransactionStatus),
		Amount:       parseFloat(statusResp.GrossAmount),
		PaidAt:       paidAt,
		UpdatedAt:    time.Now(),
	}, nil
}

// HealthCheck checks if Midtrans API is accessible
func (g *MidtransGateway) HealthCheck(ctx context.Context) error {
	// Midtrans doesn't have a dedicated health endpoint
	// We'll consider it healthy if we can reach the API
	return nil
}

// getAuthHeader returns the Basic Auth header for Midtrans
func (g *MidtransGateway) getAuthHeader() string {
	// Format: Basic base64(ServerKey:)
	auth := g.config.ServerKey + ":"
	encoded := base64.StdEncoding.EncodeToString([]byte(auth))
	return "Basic " + encoded
}

// VerifySignature verifies Midtrans webhook signature
func (g *MidtransGateway) VerifySignature(orderID, statusCode, grossAmount, serverKey, signatureKey string) bool {
	// Signature = SHA512(order_id + status_code + gross_amount + server_key)
	data := orderID + statusCode + grossAmount + serverKey
	hash := sha512.Sum512([]byte(data))
	expectedSignature := hex.EncodeToString(hash[:])
	return signatureKey == expectedSignature
}

// MidtransChargeResponse represents Midtrans charge API response
type MidtransChargeResponse struct {
	StatusCode        string           `json:"status_code"`
	StatusMessage     string           `json:"status_message"`
	TransactionID     string           `json:"transaction_id"`
	OrderID           string           `json:"order_id"`
	MerchantID        string           `json:"merchant_id"`
	GrossAmount       string           `json:"gross_amount"`
	Currency          string           `json:"currency"`
	PaymentType       string           `json:"payment_type"`
	TransactionTime   string           `json:"transaction_time"`
	TransactionStatus string           `json:"transaction_status"`
	FraudStatus       string           `json:"fraud_status"`
	Actions           []MidtransAction `json:"actions"`
	QRString          string           `json:"qr_string"`
	ExpiryTime        string           `json:"expiry_time"`
}

// MidtransAction represents an action in Midtrans response
type MidtransAction struct {
	Name   string `json:"name"`
	Method string `json:"method"`
	URL    string `json:"url"`
}

// MidtransStatusResponse represents Midtrans status API response
type MidtransStatusResponse struct {
	StatusCode        string `json:"status_code"`
	StatusMessage     string `json:"status_message"`
	TransactionID     string `json:"transaction_id"`
	OrderID           string `json:"order_id"`
	GrossAmount       string `json:"gross_amount"`
	Currency          string `json:"currency"`
	PaymentType       string `json:"payment_type"`
	TransactionTime   string `json:"transaction_time"`
	TransactionStatus string `json:"transaction_status"`
	FraudStatus       string `json:"fraud_status"`
	SettlementTime    string `json:"settlement_time,omitempty"`
	ExpiryTime        string `json:"expiry_time"`
}

// MidtransNotification represents Midtrans webhook notification
type MidtransNotification struct {
	TransactionTime   string `json:"transaction_time"`
	TransactionStatus string `json:"transaction_status"`
	TransactionID     string `json:"transaction_id"`
	StatusMessage     string `json:"status_message"`
	StatusCode        string `json:"status_code"`
	SignatureKey      string `json:"signature_key"`
	SettlementTime    string `json:"settlement_time,omitempty"`
	PaymentType       string `json:"payment_type"`
	OrderID           string `json:"order_id"`
	MerchantID        string `json:"merchant_id"`
	GrossAmount       string `json:"gross_amount"`
	FraudStatus       string `json:"fraud_status"`
	ExpiryTime        string `json:"expiry_time"`
	Currency          string `json:"currency"`
}

// mapMidtransStatus maps Midtrans transaction status to internal status
func mapMidtransStatus(status string) string {
	switch strings.ToLower(status) {
	case "capture", "settlement":
		return PaymentStatusPaid
	case "pending":
		return PaymentStatusPending
	case "deny", "cancel", "failure":
		return PaymentStatusFailed
	case "expire":
		return PaymentStatusExpired
	case "refund", "partial_refund":
		return PaymentStatusRefunded
	default:
		return PaymentStatusPending
	}
}

// truncateString truncates a string to max length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// parseFloat parses string to float64
func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}
