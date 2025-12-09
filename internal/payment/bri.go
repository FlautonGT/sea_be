package payment

import (
	"bytes"
	"context"
	"crypto"
	"crypto/hmac"
	"crypto/rand"
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

// BRIGateway implements the Gateway interface for BRI SNAP API
type BRIGateway struct {
	clientID       string
	clientSecret   string
	partnerID      string // X-PARTNER-ID
	privateKeyPath string
	baseURL        string
	callbackURL    string
	client         *http.Client
	privateKey     *rsa.PrivateKey

	// Token caching
	accessToken string
	tokenExpiry time.Time
	tokenMutex  sync.RWMutex
}

// NewBRIGateway creates a new BRI SNAP gateway instance
func NewBRIGateway(clientID, clientSecret, partnerID, privateKeyPath, baseURL, callbackURL string) (*BRIGateway, error) {
	if baseURL == "" {
		baseURL = "https://sandbox.partner.api.bri.co.id"
	}

	log.Info().
		Str("client_id", clientID).
		Str("partner_id", partnerID).
		Str("private_key_path", privateKeyPath).
		Str("base_url", baseURL).
		Msg("[BRI] Initializing BRI gateway")

	gateway := &BRIGateway{
		clientID:       clientID,
		clientSecret:   clientSecret,
		partnerID:      partnerID,
		privateKeyPath: privateKeyPath,
		baseURL:        strings.TrimSuffix(baseURL, "/"),
		callbackURL:    callbackURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Load private key
	if err := gateway.loadPrivateKey(); err != nil {
		log.Error().
			Err(err).
			Str("private_key_path", privateKeyPath).
			Msg("[BRI] CRITICAL: Failed to load private key - BRI payments will NOT work")
		// Return error to prevent gateway from being registered if key is missing
		return gateway, fmt.Errorf("failed to load BRI private key: %w", err)
	}

	log.Info().Msg("[BRI] Private key loaded successfully")
	return gateway, nil
}

// loadPrivateKey loads RSA private key from file
func (b *BRIGateway) loadPrivateKey() error {
	log.Debug().
		Str("path", b.privateKeyPath).
		Msg("[BRI] Loading private key")

	// Check if file exists
	if _, err := os.Stat(b.privateKeyPath); os.IsNotExist(err) {
		log.Error().
			Str("path", b.privateKeyPath).
			Msg("[BRI] Private key file does not exist")
		return fmt.Errorf("private key file does not exist: %s", b.privateKeyPath)
	}

	keyData, err := os.ReadFile(b.privateKeyPath)
	if err != nil {
		log.Error().
			Err(err).
			Str("path", b.privateKeyPath).
			Msg("[BRI] Failed to read private key file")
		return fmt.Errorf("failed to read private key file %s: %w", b.privateKeyPath, err)
	}

	log.Debug().
		Int("file_size", len(keyData)).
		Msg("[BRI] Private key file read successfully")

	block, _ := pem.Decode(keyData)
	if block == nil {
		log.Error().
			Str("path", b.privateKeyPath).
			Str("content_preview", string(keyData[:min(100, len(keyData))])).
			Msg("[BRI] Failed to decode PEM block - file may not be in PEM format")
		return fmt.Errorf("failed to decode PEM block from %s - ensure file is in PEM format", b.privateKeyPath)
	}

	log.Debug().
		Str("block_type", block.Type).
		Msg("[BRI] PEM block decoded")

	// Try parsing as PKCS8 first
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		log.Debug().
			Err(err).
			Msg("[BRI] PKCS8 parsing failed, trying PKCS1")

		// Try PKCS1
		rsaKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			log.Error().
				Err(err).
				Str("path", b.privateKeyPath).
				Msg("[BRI] Failed to parse private key as PKCS1 or PKCS8")
			return fmt.Errorf("failed to parse private key from %s: %w", b.privateKeyPath, err)
		}
		b.privateKey = rsaKey
		log.Info().Msg("[BRI] Private key parsed as PKCS1 RSA")
		return nil
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		log.Error().
			Str("path", b.privateKeyPath).
			Msg("[BRI] Private key is not RSA type")
		return fmt.Errorf("private key in %s is not RSA type", b.privateKeyPath)
	}
	b.privateKey = rsaKey
	log.Info().Msg("[BRI] Private key parsed as PKCS8 RSA")

	return nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetName returns the gateway name
func (b *BRIGateway) GetName() string {
	return "BRI_DIRECT"
}

// GetSupportedChannels returns supported payment channels
func (b *BRIGateway) GetSupportedChannels() []string {
	return []string{"BRI_VA", "VA_BRI"}
}

// generateTimestamp generates ISO8601 timestamp
func (b *BRIGateway) generateTimestamp() string {
	return time.Now().Format("2006-01-02T15:04:05.000+07:00")
}

// generateExternalID generates unique external ID
func (b *BRIGateway) generateExternalID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())[:9]
}

// generateTokenSignature generates X-SIGNATURE for access token request using RSA SHA256
// stringToSign = client_ID + "|" + X-TIMESTAMP
func (b *BRIGateway) generateTokenSignature(timestamp string) (string, error) {
	if b.privateKey == nil {
		log.Error().
			Str("client_id", b.clientID).
			Str("private_key_path", b.privateKeyPath).
			Msg("[BRI] Cannot generate signature - private key not loaded")
		return "", fmt.Errorf("private key not loaded - check BRI_PRIVATE_KEY_PATH configuration")
	}

	stringToSign := b.clientID + "|" + timestamp

	log.Debug().
		Str("string_to_sign", stringToSign).
		Msg("[BRI] Generating token signature")

	hashed := sha256.Sum256([]byte(stringToSign))
	signature, err := rsa.SignPKCS1v15(rand.Reader, b.privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		log.Error().
			Err(err).
			Msg("[BRI] Failed to sign with RSA")
		return "", fmt.Errorf("failed to sign: %w", err)
	}

	// Return base64 encoded signature
	return bytesToBase64(signature), nil
}

// generateAPISignature generates X-SIGNATURE for API requests using HMAC SHA512
// stringToSign = HTTPMethod+":"+ EndpointUrl +":"+ AccessToken+":"+ Lowercase(HexEncode(SHA-256(minify(RequestBody))))+ ":" +TimeStamp
func (b *BRIGateway) generateAPISignature(method, path, token, timestamp string, body []byte) string {
	// Calculate body hash
	var bodyHash string
	if len(body) > 0 {
		hash := sha256.Sum256(body)
		bodyHash = strings.ToLower(hex.EncodeToString(hash[:]))
	}

	stringToSign := method + ":" + path + ":" + token + ":" + bodyHash + ":" + timestamp

	h := hmac.New(sha512.New, []byte(b.clientSecret))
	h.Write([]byte(stringToSign))
	return bytesToBase64(h.Sum(nil))
}

// getAccessToken retrieves or refreshes the OAuth2 access token using SNAP method
func (b *BRIGateway) getAccessToken(ctx context.Context) (string, error) {
	b.tokenMutex.RLock()
	if b.accessToken != "" && time.Now().Before(b.tokenExpiry) {
		token := b.accessToken
		b.tokenMutex.RUnlock()
		log.Debug().
			Str("token_preview", token[:20]+"...").
			Time("expiry", b.tokenExpiry).
			Msg("[BRI] Using cached token")
		return token, nil
	}
	b.tokenMutex.RUnlock()

	b.tokenMutex.Lock()
	defer b.tokenMutex.Unlock()

	// Double check after acquiring write lock
	if b.accessToken != "" && time.Now().Before(b.tokenExpiry) {
		log.Debug().Msg("[BRI] Using cached token (after lock)")
		return b.accessToken, nil
	}

	log.Info().Msg("[BRI] Fetching new access token...")

	timestamp := b.generateTimestamp()

	// Generate signature
	signature, err := b.generateTokenSignature(timestamp)
	if err != nil {
		return "", fmt.Errorf("failed to generate signature: %w", err)
	}

	// Request body
	reqBody := map[string]string{
		"grantType": "client_credentials",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", b.baseURL+"/snap/v1.0/access-token/b2b", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CLIENT-KEY", b.clientID)
	req.Header.Set("X-TIMESTAMP", timestamp)
	req.Header.Set("X-SIGNATURE", signature)

	stringToSignToken := b.clientID + "|" + timestamp
	log.Debug().
		Str("url", req.URL.String()).
		Str("x-client-key", b.clientID).
		Str("x-timestamp", timestamp).
		Str("string_to_sign", stringToSignToken).
		Str("x-signature", signature[:30]+"...").
		Msg("[BRI] Getting access token")

	// Log full request
	log.Info().
		Str("method", "POST").
		Str("url", req.URL.String()).
		Str("header_content_type", req.Header.Get("Content-Type")).
		Str("header_x_client_key", req.Header.Get("X-CLIENT-KEY")).
		Str("header_x_timestamp", req.Header.Get("X-TIMESTAMP")).
		Str("header_x_signature", req.Header.Get("X-SIGNATURE")).
		Str("body", string(bodyBytes)).
		Msg("[BRI] Token Request")

	resp, err := b.client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("[BRI] Token request failed")
		return "", fmt.Errorf("failed to request token: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	// Log full response
	log.Info().
		Int("status_code", resp.StatusCode).
		Str("response_body", string(respBody)).
		Msg("[BRI] Token Response")

	var tokenResp struct {
		ResponseCode    string `json:"responseCode"`
		ResponseMessage string `json:"responseMessage"`
		AccessToken     string `json:"accessToken"`
		TokenType       string `json:"tokenType"`
		ExpiresIn       string `json:"expiresIn"`
	}

	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("failed to get token: %s - %s", tokenResp.ResponseCode, tokenResp.ResponseMessage)
	}

	b.accessToken = tokenResp.AccessToken
	// Parse expires_in (default 899 seconds = ~15 minutes)
	expiresIn := 899
	fmt.Sscanf(tokenResp.ExpiresIn, "%d", &expiresIn)
	b.tokenExpiry = time.Now().Add(time.Duration(expiresIn-60) * time.Second) // 1 minute buffer

	return b.accessToken, nil
}

// BRISNAPCreateVARequest represents SNAP VA creation request
type BRISNAPCreateVARequest struct {
	PartnerServiceID   string                  `json:"partnerServiceId"`
	CustomerNo         string                  `json:"customerNo"`
	VirtualAccountNo   string                  `json:"virtualAccountNo"`
	VirtualAccountName string                  `json:"virtualAccountName"`
	TrxID              string                  `json:"trxId"`
	TotalAmount        BRISNAPAmount           `json:"totalAmount"`
	ExpiredDate        string                  `json:"expiredDate"`
	AdditionalInfo     BRISNAPVAAdditionalInfo `json:"additionalInfo"`
}

type BRISNAPAmount struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

type BRISNAPVAAdditionalInfo struct {
	Description string `json:"description"`
}

// BRISNAPCreateVAResponse represents SNAP VA creation response
type BRISNAPCreateVAResponse struct {
	ResponseCode       string `json:"responseCode"`
	ResponseMessage    string `json:"responseMessage"`
	VirtualAccountData *struct {
		PartnerServiceID   string                  `json:"partnerServiceId"`
		CustomerNo         string                  `json:"customerNo"`
		VirtualAccountNo   string                  `json:"virtualAccountNo"`
		VirtualAccountName string                  `json:"virtualAccountName"`
		TrxID              string                  `json:"trxId"`
		TotalAmount        BRISNAPAmount           `json:"totalAmount"`
		ExpiredDate        string                  `json:"expiredDate"`
		AdditionalInfo     BRISNAPVAAdditionalInfo `json:"additionalInfo"`
	} `json:"virtualAccountData,omitempty"`
}

// CreatePayment creates a BRI Virtual Account payment using SNAP API
func (b *BRIGateway) CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	log.Info().
		Str("ref_id", req.RefID).
		Str("channel", req.Channel).
		Float64("amount", req.Amount).
		Str("customer_name", req.CustomerName).
		Msg("[BRI] Creating payment")

	token, err := b.getAccessToken(ctx)
	if err != nil {
		log.Error().
			Err(err).
			Str("ref_id", req.RefID).
			Msg("[BRI] Failed to get access token")
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	timestamp := b.generateTimestamp()
	externalID := b.generateExternalID()
	expiryTime := time.Now().Add(req.ExpiryDuration)

	// Format partnerServiceId with left padding spaces (8 digits)
	partnerServiceID := fmt.Sprintf("%8s", b.partnerID)
	if len(partnerServiceID) > 8 {
		partnerServiceID = partnerServiceID[:8]
	}

	// Generate customerNo (max 20 digits, use timestamp + random)
	customerNo := fmt.Sprintf("%d", time.Now().UnixNano()%100000000000000)
	if len(customerNo) > 13 {
		customerNo = customerNo[:13]
	}

	// virtualAccountNo = partnerServiceId + customerNo
	virtualAccountNo := partnerServiceID + customerNo

	// Format amount with 2 decimal places
	amountStr := fmt.Sprintf("%.2f", req.Amount)

	vaReq := BRISNAPCreateVARequest{
		PartnerServiceID:   partnerServiceID,
		CustomerNo:         customerNo,
		VirtualAccountNo:   virtualAccountNo,
		VirtualAccountName: req.CustomerName,
		TrxID:              req.RefID,
		TotalAmount: BRISNAPAmount{
			Value:    amountStr,
			Currency: "IDR",
		},
		ExpiredDate: expiryTime.Format("2006-01-02T15:04:05+07:00"),
		AdditionalInfo: BRISNAPVAAdditionalInfo{
			Description: req.Description,
		},
	}

	body, err := json.Marshal(vaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	path := "/snap/v1.0/transfer-va/create-va"
	signature := b.generateAPISignature("POST", path, token, timestamp, body)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", b.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("X-TIMESTAMP", timestamp)
	httpReq.Header.Set("X-SIGNATURE", signature)
	httpReq.Header.Set("X-PARTNER-ID", b.partnerID)
	httpReq.Header.Set("CHANNEL-ID", "95221")
	httpReq.Header.Set("X-EXTERNAL-ID", externalID)

	// Debug: Log signature calculation details
	var bodyHashDebug string
	if len(body) > 0 {
		hash := sha256.Sum256(body)
		bodyHashDebug = strings.ToLower(hex.EncodeToString(hash[:]))
	}
	stringToSignDebug := "POST:" + path + ":" + token + ":" + bodyHashDebug + ":" + timestamp

	// Log full request details
	log.Info().
		Str("method", "POST").
		Str("url", httpReq.URL.String()).
		Str("header_authorization", "Bearer "+token[:20]+"...").
		Str("header_x_timestamp", timestamp).
		Str("header_x_signature", signature).
		Str("header_x_partner_id", b.partnerID).
		Str("header_channel_id", "95231").
		Str("header_x_external_id", externalID).
		Str("string_to_sign", stringToSignDebug).
		Str("body_hash", bodyHashDebug).
		Str("request_body", string(body)).
		Msg("[BRI] Create VA Request")

	resp, err := b.client.Do(httpReq)
	if err != nil {
		log.Error().Err(err).Msg("[BRI] Create VA request failed")
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("[BRI] Failed to read response body")
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Log full response
	log.Info().
		Int("status_code", resp.StatusCode).
		Str("response_body", string(respBody)).
		Msg("[BRI] Create VA Response")

	var vaResp BRISNAPCreateVAResponse
	if err := json.Unmarshal(respBody, &vaResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check response code - 2002700 is success
	if vaResp.ResponseCode != "2002700" {
		return nil, fmt.Errorf("BRI API error: %s - %s", vaResp.ResponseCode, vaResp.ResponseMessage)
	}

	if vaResp.VirtualAccountData == nil {
		return nil, fmt.Errorf("BRI API returned empty VA data")
	}

	// Extract VA number without leading spaces for display
	vaNumber := strings.TrimLeft(vaResp.VirtualAccountData.VirtualAccountNo, " ")

	return &PaymentResponse{
		RefID:          req.RefID,
		GatewayRefID:   vaResp.VirtualAccountData.VirtualAccountNo,
		Channel:        req.Channel, // Use the original channel code (BRI_VA or VA_BRI)
		Amount:         req.Amount,
		Fee:            0,
		TotalAmount:    req.Amount,
		Currency:       "IDR",
		Status:         PaymentStatusPending,
		VirtualAccount: vaNumber,
		AccountName:    vaResp.VirtualAccountData.VirtualAccountName,
		BankCode:       "BRI",
		ExpiresAt:      expiryTime,
		CreatedAt:      time.Now(),
		Instructions: []string{
			"1. Login ke BRI Mobile atau Internet Banking BRI",
			"2. Pilih menu Pembayaran > BRIVA",
			"3. Masukkan nomor BRIVA: " + vaNumber,
			"4. Periksa detail pembayaran dan pastikan sudah benar",
			"5. Masukkan PIN untuk menyelesaikan pembayaran",
		},
	}, nil
}

// BRISNAPStatusRequest represents SNAP VA status inquiry request
type BRISNAPStatusRequest struct {
	PartnerServiceID string `json:"partnerServiceId"`
	CustomerNo       string `json:"customerNo"`
	VirtualAccountNo string `json:"virtualAccountNo"`
	InquiryRequestID string `json:"inquiryRequestId"`
}

// BRISNAPStatusResponse represents SNAP VA status inquiry response
type BRISNAPStatusResponse struct {
	ResponseCode       string `json:"responseCode"`
	ResponseMessage    string `json:"responseMessage"`
	VirtualAccountData *struct {
		PartnerServiceID string `json:"partnerServiceId"`
		CustomerNo       string `json:"customerNo"`
		VirtualAccountNo string `json:"virtualAccountNo"`
		InquiryRequestID string `json:"inquiryRequestId"`
	} `json:"virtualAccountData,omitempty"`
	AdditionalInfo *struct {
		PaidStatus string `json:"paidStatus"`
	} `json:"additionalInfo,omitempty"`
}

// CheckStatus checks the status of a BRI VA payment using SNAP API
func (b *BRIGateway) CheckStatus(ctx context.Context, paymentID string) (*PaymentStatus, error) {
	token, err := b.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	timestamp := b.generateTimestamp()
	externalID := b.generateExternalID()

	// Parse VA number to extract partnerServiceId and customerNo
	partnerServiceID := fmt.Sprintf("%8s", b.partnerID)
	customerNo := paymentID
	virtualAccountNo := paymentID

	// If paymentID contains partnerServiceId, extract customerNo
	if len(paymentID) > 8 {
		partnerServiceID = paymentID[:8]
		customerNo = strings.TrimLeft(paymentID[8:], "0")
		if customerNo == "" {
			customerNo = "0"
		}
	}

	statusReq := BRISNAPStatusRequest{
		PartnerServiceID: partnerServiceID,
		CustomerNo:       customerNo,
		VirtualAccountNo: virtualAccountNo,
		InquiryRequestID: fmt.Sprintf("%d", time.Now().UnixNano()),
	}

	body, err := json.Marshal(statusReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	path := "/snap/v1.0/transfer-va/status"
	signature := b.generateAPISignature("POST", path, token, timestamp, body)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", b.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("X-TIMESTAMP", timestamp)
	httpReq.Header.Set("X-SIGNATURE", signature)
	httpReq.Header.Set("X-PARTNER-ID", b.partnerID)
	httpReq.Header.Set("CHANNEL-ID", "95231")
	httpReq.Header.Set("X-EXTERNAL-ID", externalID)

	// Log full request
	log.Info().
		Str("method", "POST").
		Str("url", httpReq.URL.String()).
		Str("header_authorization", "Bearer "+token[:20]+"...").
		Str("header_x_timestamp", timestamp).
		Str("header_x_signature", signature).
		Str("header_x_partner_id", b.partnerID).
		Str("request_body", string(body)).
		Msg("[BRI] Check Status Request")

	resp, err := b.client.Do(httpReq)
	if err != nil {
		log.Error().Err(err).Msg("[BRI] Check status request failed")
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("[BRI] Failed to read response body")
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Log full response
	log.Info().
		Int("status_code", resp.StatusCode).
		Str("response_body", string(respBody)).
		Msg("[BRI] Check Status Response")

	var statusResp BRISNAPStatusResponse
	if err := json.Unmarshal(respBody, &statusResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check response code - 2002600 is success
	if statusResp.ResponseCode != "2002600" {
		return nil, fmt.Errorf("BRI API error: %s - %s", statusResp.ResponseCode, statusResp.ResponseMessage)
	}

	// Determine payment status based on paidStatus
	status := PaymentStatusPending
	if statusResp.AdditionalInfo != nil && statusResp.AdditionalInfo.PaidStatus == "Y" {
		status = PaymentStatusPaid
	}

	return &PaymentStatus{
		RefID:        paymentID,
		GatewayRefID: paymentID,
		Status:       status,
		UpdatedAt:    time.Now(),
	}, nil
}

// BRISNAPDeleteVARequest represents SNAP VA deletion request
type BRISNAPDeleteVARequest struct {
	PartnerServiceID string `json:"partnerServiceId"`
	CustomerNo       string `json:"customerNo"`
	VirtualAccountNo string `json:"virtualAccountNo"`
	TrxID            string `json:"trxId,omitempty"`
}

// DeleteVA deletes a BRI Virtual Account using SNAP API
func (b *BRIGateway) DeleteVA(ctx context.Context, virtualAccountNo, trxID string) error {
	token, err := b.getAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	timestamp := b.generateTimestamp()
	externalID := b.generateExternalID()

	// Parse VA number
	partnerServiceID := fmt.Sprintf("%8s", b.partnerID)
	customerNo := virtualAccountNo
	if len(virtualAccountNo) > 8 {
		partnerServiceID = virtualAccountNo[:8]
		customerNo = virtualAccountNo[8:]
	}

	deleteReq := BRISNAPDeleteVARequest{
		PartnerServiceID: partnerServiceID,
		CustomerNo:       customerNo,
		VirtualAccountNo: virtualAccountNo,
		TrxID:            trxID,
	}

	body, err := json.Marshal(deleteReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	path := "/snap/v1.0/transfer-va/delete-va"
	signature := b.generateAPISignature("DELETE", path, token, timestamp, body)

	httpReq, err := http.NewRequestWithContext(ctx, "DELETE", b.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("X-TIMESTAMP", timestamp)
	httpReq.Header.Set("X-SIGNATURE", signature)
	httpReq.Header.Set("X-PARTNER-ID", b.partnerID)
	httpReq.Header.Set("CHANNEL-ID", "95231")
	httpReq.Header.Set("X-EXTERNAL-ID", externalID)

	// Log full request
	log.Info().
		Str("method", "DELETE").
		Str("url", httpReq.URL.String()).
		Str("header_authorization", "Bearer "+token[:20]+"...").
		Str("header_x_timestamp", timestamp).
		Str("header_x_signature", signature).
		Str("header_x_partner_id", b.partnerID).
		Str("request_body", string(body)).
		Msg("[BRI] Delete VA Request")

	resp, err := b.client.Do(httpReq)
	if err != nil {
		log.Error().Err(err).Msg("[BRI] Delete VA request failed")
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("[BRI] Failed to read response body")
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Log full response
	log.Info().
		Int("status_code", resp.StatusCode).
		Str("response_body", string(respBody)).
		Msg("[BRI] Delete VA Response")

	var deleteResp struct {
		ResponseCode    string `json:"responseCode"`
		ResponseMessage string `json:"responseMessage"`
	}

	if err := json.Unmarshal(respBody, &deleteResp); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check response code - 2003100 is success
	if deleteResp.ResponseCode != "2003100" {
		return fmt.Errorf("BRI API error: %s - %s", deleteResp.ResponseCode, deleteResp.ResponseMessage)
	}

	return nil
}

// HealthCheck checks if BRI API is accessible
func (b *BRIGateway) HealthCheck(ctx context.Context) error {
	_, err := b.getAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("BRI health check failed: %w", err)
	}
	return nil
}

// VerifySignature verifies BRI webhook signature using HMAC SHA512
func (b *BRIGateway) VerifySignature(signature, method, path, token, timestamp string, body []byte) bool {
	expected := b.generateAPISignature(method, path, token, timestamp, body)
	return hmac.Equal([]byte(signature), []byte(expected))
}

// BRIWebhookPayload represents the webhook payload from BRI
type BRIWebhookPayload struct {
	PartnerServiceID string `json:"partnerServiceId"`
	CustomerNo       string `json:"customerNo"`
	VirtualAccountNo string `json:"virtualAccountNo"`
	PaymentRequestID string `json:"paymentRequestId"`
	TrxDateTime      string `json:"trxDateTime"`
	AdditionalInfo   *struct {
		IDApp         string `json:"idApp"`
		PassApp       string `json:"passApp"`
		PaymentAmount string `json:"paymentAmount"`
		TerminalID    string `json:"terminalId"`
		BankID        string `json:"bankId"`
	} `json:"additionalInfo,omitempty"`
}

// ParseWebhookPayload parses and validates BRI webhook payload
func (b *BRIGateway) ParseWebhookPayload(body []byte) (*BRIWebhookPayload, error) {
	var payload BRIWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse webhook payload: %w", err)
	}
	return &payload, nil
}

// bytesToBase64 encodes bytes to standard base64 string
func bytesToBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}
