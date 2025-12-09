package payment

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// BCAGateway implements the Gateway interface for BCA Direct API
type BCAGateway struct {
	clientID     string
	clientSecret string
	apiKey       string
	apiSecret    string
	corporateID  string
	baseURL      string
	client       *http.Client
	isProduction bool
	accessToken  string
	tokenExpiry  time.Time
}

// NewBCAGateway creates a new BCA gateway instance
func NewBCAGateway(clientID, clientSecret, apiKey, apiSecret, corporateID string, isProduction bool) *BCAGateway {
	baseURL := "https://sandbox.bca.co.id"
	if isProduction {
		baseURL = "https://api.bca.co.id"
	}

	return &BCAGateway{
		clientID:     clientID,
		clientSecret: clientSecret,
		apiKey:       apiKey,
		apiSecret:    apiSecret,
		corporateID:  corporateID,
		baseURL:      baseURL,
		isProduction: isProduction,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetName returns the gateway name
func (b *BCAGateway) GetName() string {
	return "BCA_DIRECT"
}

// GetSupportedChannels returns supported payment channels
func (b *BCAGateway) GetSupportedChannels() []string {
	return []string{"VA_BCA"}
}

// getAccessToken retrieves or refreshes the OAuth2 access token
func (b *BCAGateway) getAccessToken(ctx context.Context) (string, error) {
	// Return cached token if still valid
	if b.accessToken != "" && time.Now().Before(b.tokenExpiry) {
		return b.accessToken, nil
	}

	// Request new token
	data := "grant_type=client_credentials"
	req, err := http.NewRequestWithContext(ctx, "POST", b.baseURL+"/api/oauth/token", strings.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}

	auth := base64.StdEncoding.EncodeToString([]byte(b.clientID + ":" + b.clientSecret))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := b.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to request token: %w", err)
	}
	defer resp.Body.Close()

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	b.accessToken = tokenResp.AccessToken
	b.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second)

	return b.accessToken, nil
}

// generateSignature generates HMAC-SHA256 signature for BCA API
func (b *BCAGateway) generateSignature(httpMethod, relativePath, accessToken, requestBody, timestamp string) string {
	stringToSign := fmt.Sprintf("%s:%s:%s:%s:%s",
		httpMethod,
		relativePath,
		accessToken,
		sha256Hash(requestBody),
		timestamp,
	)

	h := hmac.New(sha256.New, []byte(b.apiSecret))
	h.Write([]byte(stringToSign))
	return hex.EncodeToString(h.Sum(nil))
}

func sha256Hash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// BCAVARequest represents a BCA VA creation request
type BCAVARequest struct {
	CompanyCode           string `json:"CompanyCode"`
	CustomerPhone         string `json:"CustomerPhone"`
	PrimaryID             string `json:"PrimaryID"`
	SecondaryID           string `json:"SecondaryID"`
	CustomerName          string `json:"CustomerName"`
	CurrencyCode          string `json:"CurrencyCode"`
	TotalAmount           string `json:"TotalAmount"`
	AdditionalInfo        string `json:"AdditionalInfo,omitempty"`
	RequestDate           string `json:"RequestDate"`
	ExpiredDate           string `json:"ExpiredDate"`
}

// BCAVAResponse represents a BCA VA creation response
type BCAVAResponse struct {
	ErrorCode    string `json:"ErrorCode"`
	ErrorMessage struct {
		Indonesian string `json:"Indonesian"`
		English    string `json:"English"`
	} `json:"ErrorMessage"`
	VirtualAccountData struct {
		VirtualAccountNumber string `json:"VirtualAccountNumber"`
		VirtualAccountName   string `json:"VirtualAccountName"`
	} `json:"VirtualAccountData,omitempty"`
}

// CreatePayment creates a BCA Virtual Account payment
func (b *BCAGateway) CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	token, err := b.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	timestamp := time.Now().Format("2006-01-02T15:04:05.000-07:00")
	expiryTime := time.Now().Add(req.ExpiryDuration)

	vaReq := BCAVARequest{
		CompanyCode:    b.corporateID,
		CustomerPhone:  req.CustomerPhone,
		PrimaryID:      req.RefID,
		SecondaryID:    "",
		CustomerName:   req.CustomerName,
		CurrencyCode:   "IDR",
		TotalAmount:    fmt.Sprintf("%.2f", req.Amount),
		AdditionalInfo: req.Description,
		RequestDate:    time.Now().Format("2006-01-02"),
		ExpiredDate:    expiryTime.Format("2006-01-02"),
	}

	body, err := json.Marshal(vaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	relativePath := "/va/payments"
	signature := b.generateSignature("POST", relativePath, token, string(body), timestamp)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", b.baseURL+relativePath, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-BCA-Key", b.apiKey)
	httpReq.Header.Set("X-BCA-Timestamp", timestamp)
	httpReq.Header.Set("X-BCA-Signature", signature)

	resp, err := b.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var vaResp BCAVAResponse
	if err := json.Unmarshal(respBody, &vaResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if vaResp.ErrorCode != "0" && vaResp.ErrorCode != "" {
		return nil, fmt.Errorf("API error: %s - %s", vaResp.ErrorCode, vaResp.ErrorMessage.English)
	}

	return &PaymentResponse{
		RefID:          req.RefID,
		GatewayRefID:   vaResp.VirtualAccountData.VirtualAccountNumber,
		Channel:        "VA_BCA",
		Amount:         req.Amount,
		Fee:            0,
		TotalAmount:    req.Amount,
		Currency:       "IDR",
		Status:         PaymentStatusPending,
		VirtualAccount: vaResp.VirtualAccountData.VirtualAccountNumber,
		AccountName:    vaResp.VirtualAccountData.VirtualAccountName,
		BankCode:       "BCA",
		ExpiresAt:      expiryTime,
		CreatedAt:      time.Now(),
		Instructions: []string{
			"1. Login ke BCA Mobile atau KlikBCA",
			"2. Pilih menu m-Transfer > Transfer BCA Virtual Account",
			"3. Masukkan nomor Virtual Account: " + vaResp.VirtualAccountData.VirtualAccountNumber,
			"4. Periksa detail pembayaran dan pastikan sudah benar",
			"5. Masukkan PIN untuk menyelesaikan pembayaran",
		},
	}, nil
}

// CheckStatus checks the status of a BCA VA payment
func (b *BCAGateway) CheckStatus(ctx context.Context, paymentID string) (*PaymentStatus, error) {
	token, err := b.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	timestamp := time.Now().Format("2006-01-02T15:04:05.000-07:00")
	relativePath := fmt.Sprintf("/va/payments/%s", paymentID)
	signature := b.generateSignature("GET", relativePath, token, "", timestamp)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", b.baseURL+relativePath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-BCA-Key", b.apiKey)
	httpReq.Header.Set("X-BCA-Timestamp", timestamp)
	httpReq.Header.Set("X-BCA-Signature", signature)

	resp, err := b.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var statusResp struct {
		ErrorCode    string `json:"ErrorCode"`
		ErrorMessage struct {
			English string `json:"English"`
		} `json:"ErrorMessage"`
		PaymentFlagStatus string `json:"PaymentFlagStatus"`
		TransactionDate   string `json:"TransactionDate"`
		TotalAmount       string `json:"TotalAmount"`
	}

	if err := json.Unmarshal(respBody, &statusResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if statusResp.ErrorCode != "0" && statusResp.ErrorCode != "" {
		return nil, fmt.Errorf("API error: %s - %s", statusResp.ErrorCode, statusResp.ErrorMessage.English)
	}

	status := b.mapStatus(statusResp.PaymentFlagStatus)
	var paidAt time.Time
	if statusResp.TransactionDate != "" {
		paidAt, _ = time.Parse("2006-01-02T15:04:05", statusResp.TransactionDate)
	}

	amount := 0.0
	fmt.Sscanf(statusResp.TotalAmount, "%f", &amount)

	return &PaymentStatus{
		RefID:        paymentID,
		GatewayRefID: paymentID,
		Status:       status,
		Amount:       amount,
		Fee:          0,
		PaidAt:       paidAt,
		UpdatedAt:    time.Now(),
	}, nil
}

// HealthCheck checks if BCA API is accessible
func (b *BCAGateway) HealthCheck(ctx context.Context) error {
	_, err := b.getAccessToken(ctx)
	return err
}

// mapStatus maps BCA status to common status
func (b *BCAGateway) mapStatus(status string) string {
	switch status {
	case "Y":
		return PaymentStatusPaid
	case "E":
		return PaymentStatusExpired
	default:
		return PaymentStatusPending
	}
}

