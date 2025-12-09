package payment

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// LinkQuGateway implements the Gateway interface for LinkQu (QRIS)
type LinkQuGateway struct {
	clientID     string
	clientSecret string
	username     string
	pin          string
	baseURL      string
	client       *http.Client
	isProduction bool
}

// NewLinkQuGateway creates a new LinkQu gateway instance
func NewLinkQuGateway(clientID, clientSecret, username, pin string, isProduction bool) *LinkQuGateway {
	baseURL := "https://sandbox-api.linkqu.id"
	if isProduction {
		baseURL = "https://api.linkqu.id"
	}

	return &LinkQuGateway{
		clientID:     clientID,
		clientSecret: clientSecret,
		username:     username,
		pin:          pin,
		baseURL:      baseURL,
		isProduction: isProduction,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetName returns the gateway name
func (l *LinkQuGateway) GetName() string {
	return "LINKQU"
}

// GetSupportedChannels returns supported payment channels
func (l *LinkQuGateway) GetSupportedChannels() []string {
	return []string{"QRIS"}
}

// generateSignature generates signature for LinkQu API
func (l *LinkQuGateway) generateSignature(data string) string {
	hash := sha256.Sum256([]byte(data + l.clientSecret))
	return hex.EncodeToString(hash[:])
}

// LinkQuRequest represents a LinkQu API request
type LinkQuRequest struct {
	Username      string `json:"username"`
	Pin           string `json:"pin"`
	ClientID      string `json:"client_id"`
	RequestID     string `json:"request_id"`
	Amount        int64  `json:"amount"`
	ExpiredTime   int64  `json:"expired_time"`
	PartnerReff   string `json:"partner_reff"`
	CustomerName  string `json:"customer_name,omitempty"`
	CustomerEmail string `json:"customer_email,omitempty"`
	CustomerPhone string `json:"customer_phone,omitempty"`
	Signature     string `json:"signature"`
}

// LinkQuResponse represents a LinkQu API response
type LinkQuResponse struct {
	Success     bool   `json:"success"`
	Code        string `json:"code"`
	Message     string `json:"message"`
	Data        *LinkQuData `json:"data,omitempty"`
}

// LinkQuData represents the data in LinkQu response
type LinkQuData struct {
	RequestID     string `json:"request_id"`
	PartnerReff   string `json:"partner_reff"`
	TransactionID string `json:"transaction_id"`
	Amount        int64  `json:"amount"`
	QRString      string `json:"qr_string"`
	QRURL         string `json:"qr_url"`
	ExpiredTime   string `json:"expired_time"`
	Status        string `json:"status"`
	PaidAt        string `json:"paid_at,omitempty"`
}

// CreatePayment creates a QRIS payment with LinkQu
func (l *LinkQuGateway) CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	requestID := fmt.Sprintf("LQ%d", time.Now().UnixNano())
	expiredTime := time.Now().Add(req.ExpiryDuration).Unix()
	
	signData := fmt.Sprintf("%s%s%s%d%d", l.username, l.pin, l.clientID, int64(req.Amount), expiredTime)
	signature := l.generateSignature(signData)

	lqReq := LinkQuRequest{
		Username:      l.username,
		Pin:           l.pin,
		ClientID:      l.clientID,
		RequestID:     requestID,
		Amount:        int64(req.Amount),
		ExpiredTime:   expiredTime,
		PartnerReff:   req.RefID,
		CustomerName:  req.CustomerName,
		CustomerEmail: req.CustomerEmail,
		CustomerPhone: req.CustomerPhone,
		Signature:     signature,
	}

	body, err := json.Marshal(lqReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", l.baseURL+"/transaction/create-qris", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := l.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var lqResp LinkQuResponse
	if err := json.Unmarshal(respBody, &lqResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !lqResp.Success || lqResp.Data == nil {
		return nil, fmt.Errorf("API error: %s - %s", lqResp.Code, lqResp.Message)
	}

	expiresAt, _ := time.Parse("2006-01-02 15:04:05", lqResp.Data.ExpiredTime)

	return &PaymentResponse{
		RefID:        req.RefID,
		GatewayRefID: lqResp.Data.TransactionID,
		Channel:      "QRIS",
		Amount:       req.Amount,
		Fee:          0, // LinkQu typically charges merchant fee
		TotalAmount:  req.Amount,
		Currency:     "IDR",
		Status:       PaymentStatusPending,
		QRCode:       lqResp.Data.QRString,
		QRCodeURL:    lqResp.Data.QRURL,
		ExpiresAt:    expiresAt,
		CreatedAt:    time.Now(),
		Instructions: []string{
			"1. Buka aplikasi mobile banking atau e-wallet Anda",
			"2. Pilih menu Scan QR atau QRIS",
			"3. Scan QR Code yang ditampilkan",
			"4. Periksa detail pembayaran dan konfirmasi",
			"5. Masukkan PIN untuk menyelesaikan pembayaran",
		},
	}, nil
}

// CheckStatus checks the status of a QRIS payment
func (l *LinkQuGateway) CheckStatus(ctx context.Context, paymentID string) (*PaymentStatus, error) {
	signData := fmt.Sprintf("%s%s%s%s", l.username, l.pin, l.clientID, paymentID)
	signature := l.generateSignature(signData)

	reqBody := map[string]string{
		"username":       l.username,
		"pin":            l.pin,
		"client_id":      l.clientID,
		"transaction_id": paymentID,
		"signature":      signature,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", l.baseURL+"/transaction/check-status", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := l.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var lqResp LinkQuResponse
	if err := json.Unmarshal(respBody, &lqResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !lqResp.Success || lqResp.Data == nil {
		return nil, fmt.Errorf("API error: %s - %s", lqResp.Code, lqResp.Message)
	}

	status := l.mapStatus(lqResp.Data.Status)
	var paidAt time.Time
	if lqResp.Data.PaidAt != "" {
		paidAt, _ = time.Parse("2006-01-02 15:04:05", lqResp.Data.PaidAt)
	}

	return &PaymentStatus{
		RefID:        lqResp.Data.PartnerReff,
		GatewayRefID: lqResp.Data.TransactionID,
		Status:       status,
		Amount:       float64(lqResp.Data.Amount),
		Fee:          0,
		PaidAt:       paidAt,
		UpdatedAt:    time.Now(),
	}, nil
}

// HealthCheck checks if LinkQu is accessible
func (l *LinkQuGateway) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", l.baseURL+"/health", nil)
	if err != nil {
		return err
	}
	
	resp, err := l.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: status %d", resp.StatusCode)
	}
	
	return nil
}

// mapStatus maps LinkQu status to common status
func (l *LinkQuGateway) mapStatus(status string) string {
	switch status {
	case "PAID", "SUCCESS":
		return PaymentStatusPaid
	case "EXPIRED":
		return PaymentStatusExpired
	case "FAILED":
		return PaymentStatusFailed
	case "REFUNDED":
		return PaymentStatusRefunded
	default:
		return PaymentStatusPending
	}
}

// ValidateCallback validates a callback from LinkQu
func (l *LinkQuGateway) ValidateCallback(data []byte, signature string) (*LinkQuData, error) {
	var callback LinkQuData
	if err := json.Unmarshal(data, &callback); err != nil {
		return nil, fmt.Errorf("failed to unmarshal callback: %w", err)
	}

	// Validate signature
	expectedSign := l.generateSignature(fmt.Sprintf("%s%s%d", callback.TransactionID, callback.PartnerReff, callback.Amount))
	if signature != expectedSign {
		return nil, fmt.Errorf("invalid callback signature")
	}

	return &callback, nil
}

