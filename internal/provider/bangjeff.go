package provider

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// BangJeffProvider implements the Provider interface for BangJeff
type BangJeffProvider struct {
	memberID     string
	secretKey    string
	webhookToken string
	baseURL      string
	client       *http.Client
}

// NewBangJeffProvider creates a new BangJeff provider instance
func NewBangJeffProvider(memberID, secretKey, webhookToken string) *BangJeffProvider {
	return &BangJeffProvider{
		memberID:     memberID,
		secretKey:    secretKey,
		webhookToken: webhookToken,
		baseURL:      "https://api.bangjeff.id/v1",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetName returns the provider name
func (b *BangJeffProvider) GetName() string {
	return "bangjeff"
}

// generateSign generates HMAC-SHA256 signature for BangJeff API
func (b *BangJeffProvider) generateSign(data string) string {
	h := hmac.New(sha256.New, []byte(b.secretKey))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// BangJeffRequest represents a generic BangJeff API request
type BangJeffRequest struct {
	MemberID   string `json:"member_id"`
	Signature  string `json:"signature"`
	Timestamp  int64  `json:"timestamp"`
	ProductID  string `json:"product_id,omitempty"`
	RefID      string `json:"ref_id,omitempty"`
	CustomerNo string `json:"customer_no,omitempty"`
}

// BangJeffResponse represents a generic BangJeff API response
type BangJeffResponse struct {
	Success bool            `json:"success"`
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// BangJeffProduct represents a product from BangJeff
type BangJeffProduct struct {
	ProductID   string  `json:"product_id"`
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	Brand       string  `json:"brand"`
	Type        string  `json:"type"`
	Price       float64 `json:"price"`
	SellerPrice float64 `json:"seller_price"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	Stock       int     `json:"stock"`
}

// BangJeffTransaction represents a transaction from BangJeff
type BangJeffTransaction struct {
	TransactionID string  `json:"transaction_id"`
	RefID         string  `json:"ref_id"`
	ProductID     string  `json:"product_id"`
	CustomerNo    string  `json:"customer_no"`
	Price         float64 `json:"price"`
	SellingPrice  float64 `json:"selling_price"`
	Status        string  `json:"status"`
	Message       string  `json:"message"`
	SN            string  `json:"sn"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

// BangJeffBalance represents balance info from BangJeff
type BangJeffBalance struct {
	Balance float64 `json:"balance"`
}

// GetProducts fetches all available products from BangJeff
func (b *BangJeffProvider) GetProducts(ctx context.Context) ([]Product, error) {
	timestamp := time.Now().Unix()
	signData := fmt.Sprintf("%s%d", b.memberID, timestamp)
	signature := b.generateSign(signData)

	reqBody := map[string]interface{}{
		"member_id": b.memberID,
		"signature": signature,
		"timestamp": timestamp,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", b.baseURL+"/products", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var bjResp BangJeffResponse
	if err := json.Unmarshal(respBody, &bjResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !bjResp.Success {
		return nil, fmt.Errorf("API error: %s", bjResp.Message)
	}

	var bjProducts []BangJeffProduct
	if err := json.Unmarshal(bjResp.Data, &bjProducts); err != nil {
		return nil, fmt.Errorf("failed to unmarshal products: %w", err)
	}

	products := make([]Product, 0, len(bjProducts))
	for _, bp := range bjProducts {
		isActive := bp.Status == "active" || bp.Status == "available"
		products = append(products, Product{
			SKU:          bp.ProductID,
			Name:         bp.Name,
			Description:  bp.Description,
			Category:     bp.Category,
			Brand:        bp.Brand,
			Type:         bp.Type,
			SellerPrice:  bp.SellerPrice,
			Price:        bp.Price,
			BuyerSKUCode: bp.ProductID,
			IsActive:     isActive,
			IsAvailable:  isActive,
			Stock:        bp.Stock,
			Unlimited:    bp.Stock == -1,
		})
	}

	return products, nil
}

// CheckPrice checks the current price for a product
func (b *BangJeffProvider) CheckPrice(ctx context.Context, sku string) (*PriceInfo, error) {
	timestamp := time.Now().Unix()
	signData := fmt.Sprintf("%s%s%d", b.memberID, sku, timestamp)
	signature := b.generateSign(signData)

	reqBody := map[string]interface{}{
		"member_id":  b.memberID,
		"signature":  signature,
		"timestamp":  timestamp,
		"product_id": sku,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", b.baseURL+"/products/price", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var bjResp BangJeffResponse
	if err := json.Unmarshal(respBody, &bjResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !bjResp.Success {
		return nil, fmt.Errorf("API error: %s", bjResp.Message)
	}

	var product BangJeffProduct
	if err := json.Unmarshal(bjResp.Data, &product); err != nil {
		return nil, fmt.Errorf("failed to unmarshal product: %w", err)
	}

	return &PriceInfo{
		SKU:         product.ProductID,
		Price:       product.Price,
		SellerPrice: product.SellerPrice,
		IsAvailable: product.Status == "active" || product.Status == "available",
		Stock:       product.Stock,
	}, nil
}

// CreateOrder creates an order with BangJeff
func (b *BangJeffProvider) CreateOrder(ctx context.Context, req *OrderRequest) (*OrderResponse, error) {
	timestamp := time.Now().Unix()
	signData := fmt.Sprintf("%s%s%s%s%d", b.memberID, req.RefID, req.SKU, req.CustomerNo, timestamp)
	signature := b.generateSign(signData)

	bjReq := map[string]interface{}{
		"member_id":   b.memberID,
		"signature":   signature,
		"timestamp":   timestamp,
		"ref_id":      req.RefID,
		"product_id":  req.SKU,
		"customer_no": req.CustomerNo,
	}

	body, err := json.Marshal(bjReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", b.baseURL+"/transaction", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := b.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var bjResp BangJeffResponse
	if err := json.Unmarshal(respBody, &bjResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !bjResp.Success {
		return nil, fmt.Errorf("API error: %s", bjResp.Message)
	}

	var trx BangJeffTransaction
	if err := json.Unmarshal(bjResp.Data, &trx); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction: %w", err)
	}

	status := b.mapStatus(trx.Status)

	return &OrderResponse{
		RefID:         trx.RefID,
		ProviderRefID: trx.TransactionID,
		SKU:           trx.ProductID,
		CustomerNo:    trx.CustomerNo,
		Price:         trx.Price,
		SellingPrice:  trx.SellingPrice,
		Status:        status,
		Message:       trx.Message,
		SN:            trx.SN,
		CreatedAt:     time.Now(),
	}, nil
}

// CheckStatus checks the status of an order
func (b *BangJeffProvider) CheckStatus(ctx context.Context, refID string) (*OrderStatus, error) {
	timestamp := time.Now().Unix()
	signData := fmt.Sprintf("%s%s%d", b.memberID, refID, timestamp)
	signature := b.generateSign(signData)

	bjReq := map[string]interface{}{
		"member_id": b.memberID,
		"signature": signature,
		"timestamp": timestamp,
		"ref_id":    refID,
	}

	body, err := json.Marshal(bjReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", b.baseURL+"/transaction/status", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := b.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var bjResp BangJeffResponse
	if err := json.Unmarshal(respBody, &bjResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !bjResp.Success {
		return nil, fmt.Errorf("API error: %s", bjResp.Message)
	}

	var trx BangJeffTransaction
	if err := json.Unmarshal(bjResp.Data, &trx); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction: %w", err)
	}

	return &OrderStatus{
		RefID:         trx.RefID,
		ProviderRefID: trx.TransactionID,
		Status:        b.mapStatus(trx.Status),
		Message:       trx.Message,
		SN:            trx.SN,
		UpdatedAt:     time.Now(),
	}, nil
}

// GetBalance returns the current balance with BangJeff
func (b *BangJeffProvider) GetBalance(ctx context.Context) (*Balance, error) {
	timestamp := time.Now().Unix()
	signData := fmt.Sprintf("%s%d", b.memberID, timestamp)
	signature := b.generateSign(signData)

	reqBody := map[string]interface{}{
		"member_id": b.memberID,
		"signature": signature,
		"timestamp": timestamp,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", b.baseURL+"/balance", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var bjResp BangJeffResponse
	if err := json.Unmarshal(respBody, &bjResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !bjResp.Success {
		return nil, fmt.Errorf("API error: %s", bjResp.Message)
	}

	var balance BangJeffBalance
	if err := json.Unmarshal(bjResp.Data, &balance); err != nil {
		return nil, fmt.Errorf("failed to unmarshal balance: %w", err)
	}

	return &Balance{
		Balance:   balance.Balance,
		Currency:  "IDR",
		UpdatedAt: time.Now(),
	}, nil
}

// HealthCheck checks if BangJeff is accessible
func (b *BangJeffProvider) HealthCheck(ctx context.Context) error {
	_, err := b.GetBalance(ctx)
	return err
}

// mapStatus maps BangJeff status to common status
func (b *BangJeffProvider) mapStatus(status string) string {
	switch status {
	case "success", "completed", "sukses":
		return StatusSuccess
	case "failed", "error", "gagal":
		return StatusFailed
	case "pending":
		return StatusPending
	case "processing", "proses":
		return StatusProcessing
	default:
		return StatusProcessing
	}
}

// ValidateWebhook validates a webhook callback from BangJeff
func (b *BangJeffProvider) ValidateWebhook(token string, data []byte) (*BangJeffTransaction, error) {
	if b.webhookToken != "" && token != b.webhookToken {
		return nil, fmt.Errorf("invalid webhook token")
	}

	var callback struct {
		Data BangJeffTransaction `json:"data"`
	}

	if err := json.Unmarshal(data, &callback); err != nil {
		return nil, fmt.Errorf("failed to unmarshal webhook: %w", err)
	}

	return &callback.Data, nil
}

