package provider

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// DigiflazzProvider implements the Provider interface for Digiflazz
type DigiflazzProvider struct {
	username      string
	apiKey        string
	webhookSecret string
	baseURL       string
	client        *http.Client
	isProduction  bool
}

// NewDigiflazzProvider creates a new Digiflazz provider instance
func NewDigiflazzProvider(username, apiKey, webhookSecret string, isProduction bool) *DigiflazzProvider {
	baseURL := "https://api.digiflazz.com/v1"
	if !isProduction {
		baseURL = "https://api.digiflazz.com/v1" // Digiflazz uses same URL, test via dev credentials
	}

	return &DigiflazzProvider{
		username:      username,
		apiKey:        apiKey,
		webhookSecret: webhookSecret,
		baseURL:       baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		isProduction: isProduction,
	}
}

// GetName returns the provider name
func (d *DigiflazzProvider) GetName() string {
	return "digiflazz"
}

// generateSign generates MD5 signature for Digiflazz API
func (d *DigiflazzProvider) generateSign(refID string) string {
	data := d.username + d.apiKey + refID
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

// DigiflazzRequest represents a generic Digiflazz API request
type DigiflazzRequest struct {
	Username   string `json:"username"`
	Sign       string `json:"sign"`
	RefID      string `json:"ref_id,omitempty"`
	BuyerSKU   string `json:"buyer_sku_code,omitempty"`
	CustomerNo string `json:"customer_no,omitempty"`
	Command    string `json:"cmd,omitempty"`
	Testing    bool   `json:"testing,omitempty"`
}

// DigiflazzResponse represents a generic Digiflazz API response
type DigiflazzResponse struct {
	Data json.RawMessage `json:"data"`
}

// DigiflazzProduct represents a product from Digiflazz
type DigiflazzProduct struct {
	ProductName   string  `json:"product_name"`
	Category      string  `json:"category"`
	Brand         string  `json:"brand"`
	Type          string  `json:"type"`
	SellerName    string  `json:"seller_name"`
	Price         float64 `json:"price"`
	BuyerSKUCode  string  `json:"buyer_sku_code"`
	BuyerProduct  bool    `json:"buyer_product_status"`
	SellerProduct bool    `json:"seller_product_status"`
	Unlimited     bool    `json:"unlimited_stock"`
	Stock         int     `json:"stock"`
	Multi         bool    `json:"multi"`
	StartCutOff   string  `json:"start_cut_off"`
	EndCutOff     string  `json:"end_cut_off"`
	Description   string  `json:"desc"`
}

// DigiflazzTransaction represents a transaction from Digiflazz
type DigiflazzTransaction struct {
	RefID          string  `json:"ref_id"`
	CustomerNo     string  `json:"customer_no"`
	BuyerSKUCode   string  `json:"buyer_sku_code"`
	Message        string  `json:"message"`
	Status         string  `json:"status"`
	RC             string  `json:"rc"`
	SN             string  `json:"sn"`
	BuyerLastSaldo float64 `json:"buyer_last_saldo"`
	Price          float64 `json:"price"`
	SellingPrice   float64 `json:"selling_price"`
	Tele           string  `json:"tele,omitempty"`
	Wa             string  `json:"wa,omitempty"`
}

// DigiflazzBalance represents balance info from Digiflazz
type DigiflazzBalance struct {
	Deposit float64 `json:"deposit"`
}

// GetProducts fetches all available products from Digiflazz
func (d *DigiflazzProvider) GetProducts(ctx context.Context) ([]Product, error) {
	sign := d.generateSign("pricelist")

	reqBody := DigiflazzRequest{
		Username: d.username,
		Sign:     sign,
		Command:  "pricelist",
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", d.baseURL+"/price-list", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var digiResp DigiflazzResponse
	if err := json.Unmarshal(respBody, &digiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	var digiProducts []DigiflazzProduct
	if err := json.Unmarshal(digiResp.Data, &digiProducts); err != nil {
		return nil, fmt.Errorf("failed to unmarshal products: %w", err)
	}

	products := make([]Product, 0, len(digiProducts))
	for _, dp := range digiProducts {
		products = append(products, Product{
			SKU:          dp.BuyerSKUCode,
			Name:         dp.ProductName,
			Description:  dp.Description,
			Category:     dp.Category,
			Brand:        dp.Brand,
			Type:         dp.Type,
			SellerPrice:  dp.Price,
			Price:        dp.Price,
			BuyerSKUCode: dp.BuyerSKUCode,
			IsActive:     dp.BuyerProduct && dp.SellerProduct,
			IsAvailable:  dp.BuyerProduct && dp.SellerProduct,
			Stock:        dp.Stock,
			MultiStock:   dp.Multi,
			StartCutOff:  dp.StartCutOff,
			EndCutOff:    dp.EndCutOff,
			Unlimited:    dp.Unlimited,
		})
	}

	return products, nil
}

// CheckPrice checks the current price for a product
func (d *DigiflazzProvider) CheckPrice(ctx context.Context, sku string) (*PriceInfo, error) {
	products, err := d.GetProducts(ctx)
	if err != nil {
		return nil, err
	}

	for _, p := range products {
		if p.SKU == sku {
			return &PriceInfo{
				SKU:         p.SKU,
				Price:       p.Price,
				SellerPrice: p.SellerPrice,
				IsAvailable: p.IsAvailable,
				Stock:       p.Stock,
			}, nil
		}
	}

	return nil, fmt.Errorf("product not found: %s", sku)
}

// CreateOrder creates an order with Digiflazz
func (d *DigiflazzProvider) CreateOrder(ctx context.Context, req *OrderRequest) (*OrderResponse, error) {
	sign := d.generateSign(req.RefID)

	digiReq := DigiflazzRequest{
		Username:   d.username,
		Sign:       sign,
		RefID:      req.RefID,
		BuyerSKU:   req.SKU,
		CustomerNo: req.CustomerNo,
		Testing:    !d.isProduction,
	}

	body, err := json.Marshal(digiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Log the request
	log.Info().
		Str("provider", "digiflazz").
		Str("ref_id", req.RefID).
		Str("sku", req.SKU).
		Str("customer_no", req.CustomerNo).
		Bool("testing", digiReq.Testing).
		Msg("Sending CreateOrder request to Digiflazz")

	httpReq, err := http.NewRequestWithContext(ctx, "POST", d.baseURL+"/transaction", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Log the raw response
	log.Info().
		Str("provider", "digiflazz").
		Str("ref_id", req.RefID).
		Str("status_code", resp.Status).
		Str("response_body", string(respBody)).
		Msg("Received CreateOrder response from Digiflazz")

	var digiResp DigiflazzResponse
	if err := json.Unmarshal(respBody, &digiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %s, body: %s", err, string(respBody))
	}

	var trx DigiflazzTransaction
	if err := json.Unmarshal(digiResp.Data, &trx); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction: %w, data: %s", err, string(digiResp.Data))
	}

	status := d.mapStatus(trx.Status)

	return &OrderResponse{
		RefID:         trx.RefID,
		ProviderRefID: trx.RefID,
		SKU:           trx.BuyerSKUCode,
		CustomerNo:    trx.CustomerNo,
		Price:         trx.Price,
		SellingPrice:  trx.SellingPrice,
		Status:        status,
		Message:       trx.Message + " (RC: " + trx.RC + ")", // Append RC for clearer debugging
		SN:            trx.SN,
		CreatedAt:     time.Now(),
		RawRequest:    body,
		RawResponse:   respBody,
	}, nil
}

// CheckStatus checks the status of an order
func (d *DigiflazzProvider) CheckStatus(ctx context.Context, refID string) (*OrderStatus, error) {
	sign := d.generateSign(refID)

	digiReq := DigiflazzRequest{
		Username: d.username,
		Sign:     sign,
		RefID:    refID,
	}

	body, err := json.Marshal(digiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", d.baseURL+"/transaction", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var digiResp DigiflazzResponse
	if err := json.Unmarshal(respBody, &digiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	var trx DigiflazzTransaction
	if err := json.Unmarshal(digiResp.Data, &trx); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction: %w", err)
	}

	return &OrderStatus{
		RefID:         trx.RefID,
		ProviderRefID: trx.RefID,
		Status:        d.mapStatus(trx.Status),
		Message:       trx.Message,
		SN:            trx.SN,
		UpdatedAt:     time.Now(),
	}, nil
}

// GetBalance returns the current balance with Digiflazz
func (d *DigiflazzProvider) GetBalance(ctx context.Context) (*Balance, error) {
	sign := d.generateSign("depo")

	reqBody := DigiflazzRequest{
		Username: d.username,
		Sign:     sign,
		Command:  "deposit",
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", d.baseURL+"/cek-saldo", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var digiResp DigiflazzResponse
	if err := json.Unmarshal(respBody, &digiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	var balance DigiflazzBalance
	if err := json.Unmarshal(digiResp.Data, &balance); err != nil {
		return nil, fmt.Errorf("failed to unmarshal balance: %w", err)
	}

	return &Balance{
		Balance:   balance.Deposit,
		Currency:  "IDR",
		UpdatedAt: time.Now(),
	}, nil
}

// HealthCheck checks if Digiflazz is accessible
func (d *DigiflazzProvider) HealthCheck(ctx context.Context) error {
	_, err := d.GetBalance(ctx)
	return err
}

// mapStatus maps Digiflazz status to common status
func (d *DigiflazzProvider) mapStatus(status string) string {
	switch status {
	case "Sukses":
		return StatusSuccess
	case "Gagal":
		return StatusFailed
	case "Pending":
		return StatusPending
	default:
		return StatusProcessing
	}
}

// ValidateWebhook validates a webhook callback from Digiflazz
func (d *DigiflazzProvider) ValidateWebhook(body []byte, signature string) (*DigiflazzTransaction, error) {
	// Validate signature
	if d.webhookSecret != "" {
		if signature == "" {
			return nil, fmt.Errorf("missing signature")
		}

		// Calculate HMAC-SHA1
		mac := hmac.New(sha1.New, []byte(d.webhookSecret))
		mac.Write(body)
		expectedSignature := "sha1=" + hex.EncodeToString(mac.Sum(nil))

		if signature != expectedSignature {
			return nil, fmt.Errorf("invalid signature: expected %s, got %s", expectedSignature, signature)
		}
	}

	var callback struct {
		Data DigiflazzTransaction `json:"data"`
	}

	if err := json.Unmarshal(body, &callback); err != nil {
		return nil, fmt.Errorf("failed to unmarshal webhook: %w", err)
	}

	return &callback.Data, nil
}
