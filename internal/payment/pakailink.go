package payment

import (
	"bytes"
	"context"
	"crypto"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// PakaiLinkGatewayConfig holds the configuration for PakaiLink gateway
type PakaiLinkGatewayConfig struct {
	ClientKey      string
	ClientSecret   string
	PartnerID      string
	PrivateKeyPath string
	BaseURL        string
	CallbackURL    string
	IsProduction   bool
}

// PakaiLinkGateway implements the Gateway interface for PakaiLink
type PakaiLinkGateway struct {
	config      PakaiLinkGatewayConfig
	httpClient  *http.Client
	privateKey  *rsa.PrivateKey
	accessToken string
	tokenExpiry time.Time
	tokenMutex  sync.RWMutex
}

// Bank code mapping from payment channel to PakaiLink bank code
var pakaiLinkBankCodes = map[string]string{
	"BCA_VA":      "014",
	"BNI_VA":      "009",
	"BRI_VA":      "002",
	"BSI_VA":      "451",
	"CIMB_VA":     "022",
	"DANAMON_VA":  "011",
	"MANDIRI_VA":  "008",
	"BMI_VA":      "147",
	"BNC_VA":      "490",
	"OCBC_VA":     "028",
	"PERMATA_VA":  "013",
	"SINARMAS_VA": "153",
}

// NewPakaiLinkGateway creates a new PakaiLink gateway instance
func NewPakaiLinkGateway(cfg PakaiLinkGatewayConfig) (*PakaiLinkGateway, error) {
	if cfg.ClientKey == "" {
		return nil, fmt.Errorf("pakailink client key is required")
	}
	if cfg.ClientSecret == "" {
		return nil, fmt.Errorf("pakailink client secret is required")
	}
	if cfg.PartnerID == "" {
		return nil, fmt.Errorf("pakailink partner id is required")
	}

	// Set default base URL
	if cfg.BaseURL == "" {
		if cfg.IsProduction {
			cfg.BaseURL = "https://api.pakailink.id"
		} else {
			cfg.BaseURL = "https://sandbox.pakailink.id"
		}
	}

	// Load private key
	privateKey, err := loadPrivateKey(cfg.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	return &PakaiLinkGateway{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		privateKey: privateKey,
	}, nil
}

// loadPrivateKey loads RSA private key from PEM file
func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	// Try PKCS8 first, then PKCS1
	var privateKey *rsa.PrivateKey

	// Try PKCS8
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err == nil {
		var ok bool
		privateKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("private key is not RSA")
		}
		return privateKey, nil
	}

	// Try PKCS1
	privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return privateKey, nil
}

// GetName returns the gateway name
func (g *PakaiLinkGateway) GetName() string {
	return "PAKAILINK"
}

// GetSupportedChannels returns supported payment channels
func (g *PakaiLinkGateway) GetSupportedChannels() []string {
	channels := make([]string, 0, len(pakaiLinkBankCodes))
	for channel := range pakaiLinkBankCodes {
		channels = append(channels, channel)
	}
	return channels
}

// CreatePayment creates a new payment via PakaiLink VA
func (g *PakaiLinkGateway) CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	// Get bank code from channel
	bankCode, ok := pakaiLinkBankCodes[strings.ToUpper(req.Channel)]
	if !ok {
		return nil, fmt.Errorf("unsupported payment channel: %s", req.Channel)
	}

	// Get access token
	token, err := g.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// Create VA
	return g.createVA(ctx, token, req, bankCode)
}

// getAccessToken gets or refreshes the B2B access token
func (g *PakaiLinkGateway) getAccessToken(ctx context.Context) (string, error) {
	g.tokenMutex.RLock()
	if g.accessToken != "" && time.Now().Before(g.tokenExpiry) {
		token := g.accessToken
		g.tokenMutex.RUnlock()
		return token, nil
	}
	g.tokenMutex.RUnlock()

	g.tokenMutex.Lock()
	defer g.tokenMutex.Unlock()

	// Double check after acquiring write lock
	if g.accessToken != "" && time.Now().Before(g.tokenExpiry) {
		return g.accessToken, nil
	}

	// Generate timestamp
	timestamp := time.Now().Format("2006-01-02T15:04:05+07:00")

	// Create asymmetric signature
	stringToSign := g.config.ClientKey + "|" + timestamp
	signature, err := g.createAsymmetricSignature(stringToSign)
	if err != nil {
		return "", fmt.Errorf("failed to create signature: %w", err)
	}

	// Prepare request body
	reqBody := map[string]string{
		"grantType": "client_credentials",
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", g.config.BaseURL+"/snap/v1.0/access-token/b2b", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-TIMESTAMP", timestamp)
	httpReq.Header.Set("X-CLIENT-KEY", g.config.ClientKey)
	httpReq.Header.Set("X-SIGNATURE", signature)

	log.Debug().
		Str("gateway", "pakailink").
		Str("url", httpReq.URL.String()).
		Str("timestamp", timestamp).
		Msg("PakaiLink access token request")

	// Send request
	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	log.Debug().
		Str("gateway", "pakailink").
		Int("status_code", resp.StatusCode).
		RawJSON("response", respBody).
		Msg("PakaiLink access token response")

	// Parse response
	var tokenResp PakaiLinkTokenResponse
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if tokenResp.ResponseCode != "2007300" {
		return "", fmt.Errorf("pakailink error: %s - %s", tokenResp.ResponseCode, tokenResp.ResponseMessage)
	}

	// Store token with expiry
	g.accessToken = tokenResp.AccessToken
	expiresIn := 900 // Default 15 minutes
	if tokenResp.ExpiresIn != "" {
		fmt.Sscanf(tokenResp.ExpiresIn, "%d", &expiresIn)
	}
	g.tokenExpiry = time.Now().Add(time.Duration(expiresIn-60) * time.Second) // Refresh 1 minute early

	return g.accessToken, nil
}

// createVA creates a virtual account
func (g *PakaiLinkGateway) createVA(ctx context.Context, token string, req *PaymentRequest, bankCode string) (*PaymentResponse, error) {
	timestamp := time.Now().Format("2006-01-02T15:04:05+07:00")
	externalID := generateExternalID()

	// Build VA name
	vaName := fmt.Sprintf("Seaply #%s", req.RefID)

	// Calculate expiry time
	expiryTime := time.Now().Add(req.ExpiryDuration)
	if req.ExpiryDuration == 0 {
		expiryTime = time.Now().Add(24 * time.Hour)
	}

	// Build request body
	reqBody := map[string]interface{}{
		"partnerReferenceNo":  req.RefID,
		"customerNo":          generateCustomerNo(req.RefID),
		"virtualAccountName":  vaName,
		"virtualAccountPhone": req.CustomerPhone,
		"totalAmount": map[string]string{
			"value":    fmt.Sprintf("%.2f", req.Amount),
			"currency": "IDR",
		},
		"expiredDate": expiryTime.Format("2006-01-02T15:04:05+07:00"),
		"additionalInfo": map[string]string{
			"callbackUrl": g.config.CallbackURL,
			"bankCode":    bankCode,
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create symmetric signature
	signature, err := g.createSymmetricSignature("POST", "/snap/v1.0/transfer-va/create-va", token, jsonBody, timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to create signature: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", g.config.BaseURL+"/snap/v1.0/transfer-va/create-va", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("X-TIMESTAMP", timestamp)
	httpReq.Header.Set("X-PARTNER-ID", g.config.PartnerID)
	httpReq.Header.Set("X-EXTERNAL-ID", externalID)
	httpReq.Header.Set("X-SIGNATURE", signature)
	httpReq.Header.Set("CHANNEL-ID", "95221")

	log.Debug().
		Str("gateway", "pakailink").
		Str("url", httpReq.URL.String()).
		Str("timestamp", timestamp).
		Str("external_id", externalID).
		RawJSON("body", jsonBody).
		Msg("PakaiLink create VA request")

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
		Str("gateway", "pakailink").
		Int("status_code", resp.StatusCode).
		RawJSON("response", respBody).
		Msg("PakaiLink create VA response")

	// Parse response
	var vaResp PakaiLinkCreateVAResponse
	if err := json.Unmarshal(respBody, &vaResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if vaResp.ResponseCode != "2002700" {
		return nil, fmt.Errorf("pakailink error: %s - %s", vaResp.ResponseCode, vaResp.ResponseMessage)
	}

	// Parse expiry time
	vaExpiryTime, _ := time.Parse("2006-01-02T15:04:05+07:00", vaResp.VirtualAccountData.ExpiredDate)

	return &PaymentResponse{
		RefID:          req.RefID,
		GatewayRefID:   vaResp.VirtualAccountData.AdditionalInfo.ReferenceNo,
		Channel:        req.Channel,
		Amount:         req.Amount,
		Currency:       "IDR",
		Status:         PaymentStatusPending,
		VirtualAccount: vaResp.VirtualAccountData.VirtualAccountNo,
		AccountName:    vaName,
		BankCode:       bankCode,
		ExpiresAt:      vaExpiryTime,
		CreatedAt:      time.Now(),
	}, nil
}

// CheckStatus checks the payment status
func (g *PakaiLinkGateway) CheckStatus(ctx context.Context, partnerRefNo string) (*PaymentStatus, error) {
	// Get access token
	token, err := g.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	timestamp := time.Now().Format("2006-01-02T15:04:05+07:00")
	externalID := generateExternalID()

	// Build request body
	reqBody := map[string]string{
		"originalPartnerReferenceNo": partnerRefNo,
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Create symmetric signature
	signature, err := g.createSymmetricSignature("POST", "/snap/v1.0/transfer-va/create-va-status", token, jsonBody, timestamp)
	if err != nil {
		return nil, fmt.Errorf("failed to create signature: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", g.config.BaseURL+"/snap/v1.0/transfer-va/create-va-status", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("X-TIMESTAMP", timestamp)
	httpReq.Header.Set("X-PARTNER-ID", g.config.PartnerID)
	httpReq.Header.Set("X-EXTERNAL-ID", externalID)
	httpReq.Header.Set("X-SIGNATURE", signature)
	httpReq.Header.Set("CHANNEL-ID", "95221")

	log.Debug().
		Str("gateway", "pakailink").
		Str("url", httpReq.URL.String()).
		Str("partner_ref_no", partnerRefNo).
		Msg("PakaiLink check status request")

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
		Str("gateway", "pakailink").
		Int("status_code", resp.StatusCode).
		RawJSON("response", respBody).
		Msg("PakaiLink check status response")

	// Parse response
	var statusResp PakaiLinkStatusResponse
	if err := json.Unmarshal(respBody, &statusResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if statusResp.ResponseCode != "2003300" {
		return nil, fmt.Errorf("pakailink error: %s - %s", statusResp.ResponseCode, statusResp.ResponseMessage)
	}

	// Map status
	status := mapPakaiLinkStatus(statusResp.LatestTransactionStatus)

	// Parse paid time
	var paidAt time.Time
	if statusResp.PaidTime != "" {
		paidAt, _ = time.Parse("2006-01-02T15:04:05+07:00", statusResp.PaidTime)
	}

	// Parse amount
	var amount float64
	if statusResp.Amount.Value != "" {
		fmt.Sscanf(statusResp.Amount.Value, "%f", &amount)
	}

	return &PaymentStatus{
		RefID:        statusResp.OriginalPartnerReferenceNo,
		GatewayRefID: statusResp.OriginalReferenceNo,
		Status:       status,
		Amount:       amount,
		PaidAt:       paidAt,
		UpdatedAt:    time.Now(),
	}, nil
}

// HealthCheck checks if PakaiLink API is accessible
func (g *PakaiLinkGateway) HealthCheck(ctx context.Context) error {
	_, err := g.getAccessToken(ctx)
	return err
}

// createAsymmetricSignature creates RSA-SHA256 signature for access token
func (g *PakaiLinkGateway) createAsymmetricSignature(stringToSign string) (string, error) {
	hash := sha256.Sum256([]byte(stringToSign))
	signature, err := rsa.SignPKCS1v15(nil, g.privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", fmt.Errorf("failed to sign: %w", err)
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

// createSymmetricSignature creates HMAC-SHA512 signature for service requests
func (g *PakaiLinkGateway) createSymmetricSignature(method, path, token string, body []byte, timestamp string) (string, error) {
	// Minify body and hash it
	var minifiedBody bytes.Buffer
	if err := json.Compact(&minifiedBody, body); err != nil {
		minifiedBody.Write(body)
	}

	bodyHash := sha256.Sum256(minifiedBody.Bytes())
	bodyHashHex := strings.ToLower(hex.EncodeToString(bodyHash[:]))

	// Compose string to sign: METHOD:PATH:TOKEN:BODY_HASH:TIMESTAMP
	stringToSign := fmt.Sprintf("%s:%s:%s:%s:%s", method, path, token, bodyHashHex, timestamp)

	// Create HMAC-SHA512
	h := hmac.New(sha512.New, []byte(g.config.ClientSecret))
	h.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return signature, nil
}

// VerifyCallbackSignature verifies the callback signature from PakaiLink
func (g *PakaiLinkGateway) VerifyCallbackSignature(callbackURL string, body []byte, timestamp, signature string) bool {
	// Minify body and hash it
	var minifiedBody bytes.Buffer
	if err := json.Compact(&minifiedBody, body); err != nil {
		minifiedBody.Write(body)
	}

	bodyHash := sha256.Sum256(minifiedBody.Bytes())
	bodyHashHex := strings.ToLower(hex.EncodeToString(bodyHash[:]))

	// Compose string to verify: POST:CALLBACK_URL:BODY_HASH:TIMESTAMP
	stringToVerify := fmt.Sprintf("POST:%s:%s:%s", callbackURL, bodyHashHex, timestamp)

	// Verify using RSA-SHA256 with PakaiLink's public key
	// For now, we'll use HMAC-SHA512 verification since we have client secret
	h := hmac.New(sha512.New, []byte(g.config.ClientSecret))
	h.Write([]byte(stringToVerify))
	expectedSignature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return signature == expectedSignature
}

// Helper functions
func generateExternalID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func generateCustomerNo(refID string) string {
	// Generate a numeric customer number from refID hash
	h := sha256.Sum256([]byte(refID))
	// Use hash bytes to generate a numeric ID
	var num uint64
	for i := 0; i < 8; i++ {
		num = (num << 8) | uint64(h[i])
	}
	return fmt.Sprintf("%d", num%1000000000000)
}

func mapPakaiLinkStatus(status string) string {
	switch status {
	case "00":
		return PaymentStatusPaid
	case "01":
		return PaymentStatusPending
	case "02":
		return PaymentStatusPending // Paying
	case "03":
		return PaymentStatusExpired // Cancelled
	default:
		return PaymentStatusPending
	}
}

// Response types

type PakaiLinkTokenResponse struct {
	ResponseCode    string `json:"responseCode"`
	ResponseMessage string `json:"responseMessage"`
	AccessToken     string `json:"accessToken"`
	TokenType       string `json:"tokenType"`
	ExpiresIn       string `json:"expiresIn"`
}

type PakaiLinkCreateVAResponse struct {
	ResponseCode       string `json:"responseCode"`
	ResponseMessage    string `json:"responseMessage"`
	VirtualAccountData struct {
		PartnerReferenceNo string `json:"partnerReferenceNo"`
		CustomerNo         string `json:"customerNo"`
		VirtualAccountNo   string `json:"virtualAccountNo"`
		ExpiredDate        string `json:"expiredDate"`
		TotalAmount        struct {
			Value    string `json:"value"`
			Currency string `json:"currency"`
		} `json:"totalAmount"`
		AdditionalInfo struct {
			BankCode    string `json:"bankCode"`
			CallbackUrl string `json:"callbackUrl"`
			ReferenceNo string `json:"referenceNo"`
		} `json:"additionalInfo"`
	} `json:"virtualAccountData"`
}

type PakaiLinkStatusResponse struct {
	ResponseCode                string `json:"responseCode"`
	ResponseMessage             string `json:"responseMessage"`
	OriginalPartnerReferenceNo  string `json:"originalPartnerReferenceNo"`
	OriginalReferenceNo         string `json:"originalReferenceNo"`
	OriginalExternalId          string `json:"originalExternalId"`
	ServiceCode                 string `json:"serviceCode"`
	TransactionDate             string `json:"transactionDate"`
	LatestTransactionStatus     string `json:"latestTransactionStatus"`
	TransactionStatusDesc       string `json:"transactionStatusDesc"`
	PaidTime                    string `json:"paidTime,omitempty"`
	Amount                      struct {
		Value    string `json:"value"`
		Currency string `json:"currency"`
	} `json:"amount"`
	AdditionalInfo struct {
		Callback     string `json:"callback"`
		CustomerData string `json:"customerData"`
		NominalPaid  struct {
			Value    string `json:"value"`
			Currency string `json:"currency"`
		} `json:"nominalPaid"`
		ServiceFee struct {
			Value    string `json:"value"`
			Currency string `json:"currency"`
		} `json:"serviceFee"`
		TotalPaid struct {
			Value    string `json:"value"`
			Currency string `json:"currency"`
		} `json:"totalPaid"`
		TotalReceive struct {
			Value    string `json:"value"`
			Currency string `json:"currency"`
		} `json:"totalReceive"`
	} `json:"additionalInfo"`
}

// PakaiLinkCallback represents the callback payload from PakaiLink
type PakaiLinkCallback struct {
	TransactionData struct {
		PaymentFlagStatus string `json:"paymentFlagStatus"`
		PaymentFlagReason struct {
			English   string `json:"english"`
			Indonesia string `json:"indonesia"`
		} `json:"paymentFlagReason"`
		CustomerNo           string `json:"customerNo"`
		VirtualAccountNo     string `json:"virtualAccountNo"`
		VirtualAccountName   string `json:"virtualAccountName"`
		PartnerReferenceNo   string `json:"partnerReferenceNo"`
		CallbackType         string `json:"callbackType"`
		VirtualAccountTrxType string `json:"virtualAccountTrxType,omitempty"`
		PaidAmount           struct {
			Value    string `json:"value"`
			Currency string `json:"currency"`
		} `json:"paidAmount"`
		FeeAmount struct {
			Value    string `json:"value"`
			Currency string `json:"currency"`
		} `json:"feeAmount"`
		CreditBalance struct {
			Value    string `json:"value"`
			Currency string `json:"currency"`
		} `json:"creditBalance"`
		AdditionalInfo struct {
			CallbackUrl string `json:"callbackUrl"`
			Balance     struct {
				Value    string `json:"value"`
				Currency string `json:"currency"`
			} `json:"balance,omitempty"`
		} `json:"additionalInfo"`
	} `json:"transactionData"`
}

