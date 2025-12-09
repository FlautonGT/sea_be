package payment

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// XenditGateway implements the Gateway interface for Xendit
type XenditGateway struct {
	secretKey     string
	callbackToken string
	baseURL       string
	client        *http.Client
	isProduction  bool
}

// NewXenditGateway creates a new Xendit gateway instance
func NewXenditGateway(secretKey, callbackToken string, isProduction bool) *XenditGateway {
	return &XenditGateway{
		secretKey:     secretKey,
		callbackToken: callbackToken,
		baseURL:       "https://api.xendit.co",
		isProduction:  isProduction,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetName returns the gateway name
func (x *XenditGateway) GetName() string {
	return "XENDIT"
}

// GetSupportedChannels returns supported payment channels
func (x *XenditGateway) GetSupportedChannels() []string {
	return []string{"VA_PERMATA", "VA_MANDIRI", "VA_BNI", "VA_BRI", "BRI_VA", "VA_BCA", "BCA_VA", "ALFAMART", "INDOMARET"}
}

// XenditVARequest represents a Xendit VA creation request
type XenditVARequest struct {
	ExternalID     string `json:"external_id"`
	BankCode       string `json:"bank_code"`
	Name           string `json:"name"`
	ExpectedAmount int64  `json:"expected_amount"`
	ExpirationDate string `json:"expiration_date,omitempty"`
	IsClosed       bool   `json:"is_closed"`
	IsSingleUse    bool   `json:"is_single_use"`
}

// XenditVAResponse represents a Xendit VA creation response
type XenditVAResponse struct {
	ID             string `json:"id"`
	ExternalID     string `json:"external_id"`
	OwnerID        string `json:"owner_id"`
	BankCode       string `json:"bank_code"`
	MerchantCode   string `json:"merchant_code"`
	AccountNumber  string `json:"account_number"`
	Name           string `json:"name"`
	ExpectedAmount int64  `json:"expected_amount"`
	ExpirationDate string `json:"expiration_date"`
	IsClosed       bool   `json:"is_closed"`
	IsSingleUse    bool   `json:"is_single_use"`
	Status         string `json:"status"`
	ErrorCode      string `json:"error_code,omitempty"`
	Message        string `json:"message,omitempty"`
}

// CreatePayment creates a payment with Xendit
func (x *XenditGateway) CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	channel := strings.ToUpper(req.Channel)

	// Check if it's a retail payment (Alfamart/Indomaret)
	if channel == "ALFAMART" || channel == "INDOMARET" {
		return x.createRetailPayment(ctx, req)
	}

	// Otherwise, create Virtual Account payment
	return x.createVAPayment(ctx, req)
}

// createRetailPayment creates a retail payment (Alfamart/Indomaret) using Payment Requests API v3
func (x *XenditGateway) createRetailPayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	channel := strings.ToUpper(req.Channel)

	// Build request body for Payment Requests API v3
	requestBody := map[string]interface{}{
		"reference_id":   req.RefID,
		"type":           "PAY",
		"country":        "ID",
		"currency":       "IDR",
		"request_amount": int64(req.Amount),
		"channel_code":   channel,
		"channel_properties": map[string]interface{}{
			"payer_name": "Customer Gate",
		},
		"description": truncateString(req.Description, 100),
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	log.Debug().
		Str("gateway", "xendit").
		Str("url", x.baseURL+"/v3/payment_requests").
		RawJSON("body", jsonBody).
		Msg("Xendit retail payment request")

	httpReq, err := http.NewRequestWithContext(ctx, "POST", x.baseURL+"/v3/payment_requests", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	auth := base64.StdEncoding.EncodeToString([]byte(x.secretKey + ":"))
	httpReq.Header.Set("Authorization", "Basic "+auth)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("api-version", "2024-11-11")

	resp, err := x.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	log.Debug().
		Str("gateway", "xendit").
		Int("status_code", resp.StatusCode).
		RawJSON("response", respBody).
		Msg("Xendit retail payment response")

	// Parse response
	var retailResp XenditPaymentRequestResponse
	if err := json.Unmarshal(respBody, &retailResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check for error
	if retailResp.ErrorCode != "" {
		return nil, fmt.Errorf("API error: %s - %s", retailResp.ErrorCode, retailResp.Message)
	}

	// Extract payment code from actions
	var paymentCode string
	for _, action := range retailResp.Actions {
		if action.Type == "PRESENT_TO_CUSTOMER" && action.Descriptor == "PAYMENT_CODE" {
			paymentCode = action.Value
			break
		}
	}

	// Parse expiry time
	var expiryTime time.Time
	if retailResp.ChannelProperties.ExpiresAt != "" {
		expiryTime, _ = time.Parse(time.RFC3339, retailResp.ChannelProperties.ExpiresAt)
	} else {
		expiryTime = time.Now().Add(req.ExpiryDuration)
	}

	return &PaymentResponse{
		RefID:        req.RefID,
		GatewayRefID: retailResp.PaymentRequestID,
		Channel:      req.Channel,
		Amount:       req.Amount,
		Fee:          0,
		TotalAmount:  req.Amount,
		Currency:     "IDR",
		Status:       PaymentStatusPending,
		PaymentCode:  paymentCode, // Retail payment code for barcode
		ExpiresAt:    expiryTime,
		CreatedAt:    time.Now(),
		Instructions: x.getRetailInstructions(channel, paymentCode),
	}, nil
}

// createVAPayment creates a Virtual Account payment
func (x *XenditGateway) createVAPayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	// Map channel to bank code
	bankCode := x.mapChannelToBankCode(req.Channel)
	if bankCode == "" {
		return nil, fmt.Errorf("unsupported channel: %s", req.Channel)
	}

	expiryTime := time.Now().Add(req.ExpiryDuration)

	vaReq := XenditVARequest{
		ExternalID:     req.RefID,
		BankCode:       bankCode,
		Name:           req.CustomerName,
		ExpectedAmount: int64(req.Amount),
		ExpirationDate: expiryTime.Format(time.RFC3339),
		IsClosed:       true,
		IsSingleUse:    true,
	}

	body, err := json.Marshal(vaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", x.baseURL+"/callback_virtual_accounts", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	auth := base64.StdEncoding.EncodeToString([]byte(x.secretKey + ":"))
	httpReq.Header.Set("Authorization", "Basic "+auth)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := x.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var vaResp XenditVAResponse
	if err := json.Unmarshal(respBody, &vaResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if vaResp.ErrorCode != "" {
		return nil, fmt.Errorf("API error: %s - %s", vaResp.ErrorCode, vaResp.Message)
	}

	vaNumber := vaResp.MerchantCode + vaResp.AccountNumber

	return &PaymentResponse{
		RefID:          req.RefID,
		GatewayRefID:   vaResp.ID,
		Channel:        req.Channel,
		Amount:         req.Amount,
		Fee:            0,
		TotalAmount:    req.Amount,
		Currency:       "IDR",
		Status:         PaymentStatusPending,
		VirtualAccount: vaNumber,
		AccountName:    vaResp.Name,
		BankCode:       vaResp.BankCode,
		ExpiresAt:      expiryTime,
		CreatedAt:      time.Now(),
		Instructions:   x.getInstructions(bankCode, vaNumber),
	}, nil
}

// CheckStatus checks the status of a Xendit payment
func (x *XenditGateway) CheckStatus(ctx context.Context, paymentID string) (*PaymentStatus, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", x.baseURL+"/callback_virtual_accounts/"+paymentID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	auth := base64.StdEncoding.EncodeToString([]byte(x.secretKey + ":"))
	httpReq.Header.Set("Authorization", "Basic "+auth)

	resp, err := x.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var vaResp XenditVAResponse
	if err := json.Unmarshal(respBody, &vaResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if vaResp.ErrorCode != "" {
		return nil, fmt.Errorf("API error: %s - %s", vaResp.ErrorCode, vaResp.Message)
	}

	status := x.mapStatus(vaResp.Status)

	return &PaymentStatus{
		RefID:        vaResp.ExternalID,
		GatewayRefID: vaResp.ID,
		Status:       status,
		Amount:       float64(vaResp.ExpectedAmount),
		Fee:          0,
		UpdatedAt:    time.Now(),
	}, nil
}

// HealthCheck checks if Xendit API is accessible
func (x *XenditGateway) HealthCheck(ctx context.Context) error {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", x.baseURL+"/balance", nil)
	if err != nil {
		return err
	}

	auth := base64.StdEncoding.EncodeToString([]byte(x.secretKey + ":"))
	httpReq.Header.Set("Authorization", "Basic "+auth)

	resp, err := x.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: status %d", resp.StatusCode)
	}

	return nil
}

// mapChannelToBankCode maps channel code to Xendit bank code
func (x *XenditGateway) mapChannelToBankCode(channel string) string {
	switch channel {
	case "VA_BRI", "BRI_VA":
		return "BRI"
	case "VA_BCA", "BCA_VA":
		return "BCA"
	case "VA_BNI", "BNI_VA":
		return "BNI"
	case "VA_PERMATA", "PERMATA_VA":
		return "PERMATA"
	case "VA_MANDIRI", "MANDIRI_VA":
		return "MANDIRI"
	case "VA_CIMB", "CIMB_VA":
		return "CIMB"
	default:
		return ""
	}
}

// mapStatus maps Xendit status to common status
func (x *XenditGateway) mapStatus(status string) string {
	switch status {
	case "ACTIVE":
		return PaymentStatusPending
	case "INACTIVE":
		return PaymentStatusExpired
	default:
		return PaymentStatusPending
	}
}

// getInstructions returns payment instructions for a bank
func (x *XenditGateway) getInstructions(bankCode, vaNumber string) []string {
	switch bankCode {
	case "PERMATA":
		return []string{
			"1. Login ke PermataNet atau PermataMobile",
			"2. Pilih menu Pembayaran > Virtual Account",
			"3. Masukkan nomor Virtual Account: " + vaNumber,
			"4. Periksa detail pembayaran dan konfirmasi",
			"5. Masukkan PIN untuk menyelesaikan pembayaran",
		}
	case "MANDIRI":
		return []string{
			"1. Login ke Livin by Mandiri atau Internet Banking Mandiri",
			"2. Pilih menu Bayar > Multipayment",
			"3. Masukkan nomor Virtual Account: " + vaNumber,
			"4. Periksa detail pembayaran dan konfirmasi",
			"5. Masukkan PIN untuk menyelesaikan pembayaran",
		}
	case "BNI":
		return []string{
			"1. Login ke BNI Mobile atau Internet Banking BNI",
			"2. Pilih menu Pembayaran > Virtual Account",
			"3. Masukkan nomor Virtual Account: " + vaNumber,
			"4. Periksa detail pembayaran dan konfirmasi",
			"5. Masukkan PIN untuk menyelesaikan pembayaran",
		}
	default:
		return []string{}
	}
}

// getRetailInstructions returns payment instructions for retail
func (x *XenditGateway) getRetailInstructions(channel, paymentCode string) []string {
	storeName := channel
	if channel == "ALFAMART" {
		storeName = "Alfamart"
	} else if channel == "INDOMARET" {
		storeName = "Indomaret"
	}

	return []string{
		"1. Kunjungi gerai " + storeName + " terdekat",
		"2. Informasikan ke kasir untuk pembayaran",
		"3. Tunjukkan kode pembayaran: " + paymentCode,
		"4. Bayar sesuai nominal yang tertera",
		"5. Simpan struk sebagai bukti pembayaran",
	}
}

// ValidateCallback validates a callback from Xendit
func (x *XenditGateway) ValidateCallback(token string, data []byte) (*XenditCallback, error) {
	if x.callbackToken != "" && token != x.callbackToken {
		return nil, fmt.Errorf("invalid callback token")
	}

	var callback XenditCallback
	if err := json.Unmarshal(data, &callback); err != nil {
		return nil, fmt.Errorf("failed to unmarshal callback: %w", err)
	}

	return &callback, nil
}

// XenditCallback represents a Xendit callback payload (for VA)
type XenditCallback struct {
	ID                       string  `json:"id"`
	ExternalID               string  `json:"external_id"`
	BankCode                 string  `json:"bank_code"`
	MerchantCode             string  `json:"merchant_code"`
	AccountNumber            string  `json:"account_number"`
	CallbackVirtualAccountID string  `json:"callback_virtual_account_id"`
	TransactionTimestamp     string  `json:"transaction_timestamp"`
	Amount                   float64 `json:"amount"`
	SenderName               string  `json:"sender_name,omitempty"`
}

// XenditPaymentRequestResponse represents Payment Requests API v3 response
type XenditPaymentRequestResponse struct {
	PaymentRequestID  string `json:"payment_request_id"`
	Country           string `json:"country"`
	Currency          string `json:"currency"`
	BusinessID        string `json:"business_id"`
	ReferenceID       string `json:"reference_id"`
	Description       string `json:"description"`
	Created           string `json:"created"`
	Updated           string `json:"updated"`
	Status            string `json:"status"`
	CaptureMethod     string `json:"capture_method"`
	ChannelCode       string `json:"channel_code"`
	RequestAmount     int64  `json:"request_amount"`
	Type              string `json:"type"`
	ChannelProperties struct {
		PayerName string `json:"payer_name"`
		ExpiresAt string `json:"expires_at"`
	} `json:"channel_properties"`
	Actions []struct {
		Type       string `json:"type"`
		Descriptor string `json:"descriptor"`
		Value      string `json:"value"`
	} `json:"actions"`
	ErrorCode string `json:"error_code,omitempty"`
	Message   string `json:"message,omitempty"`
}

// XenditWebhookEvent represents Xendit webhook event for Payment Requests API
type XenditWebhookEvent struct {
	Event      string `json:"event"`
	BusinessID string `json:"business_id"`
	Created    string `json:"created"`
	Data       struct {
		PaymentID        string  `json:"payment_id"`
		BusinessID       string  `json:"business_id"`
		Status           string  `json:"status"`
		PaymentRequestID string  `json:"payment_request_id"`
		RequestAmount    float64 `json:"request_amount"`
		CustomerID       string  `json:"customer_id"`
		ChannelCode      string  `json:"channel_code"`
		Country          string  `json:"country"`
		Currency         string  `json:"currency"`
		ReferenceID      string  `json:"reference_id"`
		Description      string  `json:"description"`
		FailureCode      string  `json:"failure_code,omitempty"`
		Type             string  `json:"type"`
		Created          string  `json:"created"`
		Updated          string  `json:"updated"`
	} `json:"data"`
}
