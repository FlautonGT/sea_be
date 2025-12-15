package provider

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// VIPResellerProvider implements the Provider interface for VIP Reseller
type VIPResellerProvider struct {
	apiID   string
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewVIPResellerProvider creates a new VIP Reseller provider instance
func NewVIPResellerProvider(apiID, apiKey string) *VIPResellerProvider {
	return &VIPResellerProvider{
		apiID:   apiID,
		apiKey:  apiKey,
		baseURL: "https://vip-reseller.co.id/api",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetName returns the provider name
func (v *VIPResellerProvider) GetName() string {
	return "vipreseller"
}

// generateSign generates MD5 signature for VIP Reseller API
func (v *VIPResellerProvider) generateSign(apiID string) string {
	data := v.apiID + v.apiKey
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

// VIPRequest represents a generic VIP Reseller API request
type VIPRequest struct {
	Key     string `json:"key"`
	Sign    string `json:"sign"`
	Type    string `json:"type,omitempty"`
	Service string `json:"service,omitempty"`
	Data    string `json:"data,omitempty"`
	RefID   string `json:"ref_id,omitempty"`
}

// VIPResponse represents a generic VIP Reseller API response
type VIPResponse struct {
	Result  bool            `json:"result"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// VIPProduct represents a product from VIP Reseller
type VIPProduct struct {
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Status      string  `json:"status"`
	Category    string  `json:"category"`
	Brand       string  `json:"brand"`
	Type        string  `json:"type"`
	Description string  `json:"description"`
}

// VIPTransaction represents a transaction from VIP Reseller
type VIPTransaction struct {
	TrxID     string  `json:"trx_id"`
	RefID     string  `json:"ref_id"`
	Service   string  `json:"service"`
	Data      string  `json:"data"`
	Price     float64 `json:"price"`
	Status    string  `json:"status"`
	Message   string  `json:"message"`
	SN        string  `json:"sn"`
	Balance   float64 `json:"balance"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// VIPBalance represents balance info from VIP Reseller
type VIPBalance struct {
	Balance float64 `json:"balance"`
}

// GetProducts fetches all available products from VIP Reseller
func (v *VIPResellerProvider) GetProducts(ctx context.Context) ([]Product, error) {
	sign := v.generateSign(v.apiID)

	reqBody := map[string]string{
		"key":  v.apiID,
		"sign": sign,
		"type": "services",
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", v.baseURL+"/game-feature", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := v.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var vipResp VIPResponse
	if err := json.Unmarshal(respBody, &vipResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !vipResp.Result {
		return nil, fmt.Errorf("API error: %s", vipResp.Message)
	}

	var vipProducts []VIPProduct
	if err := json.Unmarshal(vipResp.Data, &vipProducts); err != nil {
		return nil, fmt.Errorf("failed to unmarshal products: %w", err)
	}

	products := make([]Product, 0, len(vipProducts))
	for _, vp := range vipProducts {
		isActive := vp.Status == "available" || vp.Status == "active"
		products = append(products, Product{
			SKU:          vp.Code,
			Name:         vp.Name,
			Description:  vp.Description,
			Category:     vp.Category,
			Brand:        vp.Brand,
			Type:         vp.Type,
			SellerPrice:  vp.Price,
			Price:        vp.Price,
			BuyerSKUCode: vp.Code,
			IsActive:     isActive,
			IsAvailable:  isActive,
			Stock:        -1, // VIP doesn't provide stock info
			Unlimited:    true,
		})
	}

	return products, nil
}

// CheckPrice checks the current price for a product
func (v *VIPResellerProvider) CheckPrice(ctx context.Context, sku string) (*PriceInfo, error) {
	products, err := v.GetProducts(ctx)
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

// CreateOrder creates an order with VIP Reseller
func (v *VIPResellerProvider) CreateOrder(ctx context.Context, req *OrderRequest) (*OrderResponse, error) {
	sign := v.generateSign(v.apiID)

	vipReq := map[string]string{
		"key":     v.apiID,
		"sign":    sign,
		"type":    "order",
		"service": req.SKU,
		"data":    req.CustomerNo,
	}

	body, err := json.Marshal(vipReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", v.baseURL+"/game-feature", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := v.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var vipResp VIPResponse
	if err := json.Unmarshal(respBody, &vipResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !vipResp.Result {
		return nil, fmt.Errorf("API error: %s", vipResp.Message)
	}

	var trx VIPTransaction
	if err := json.Unmarshal(vipResp.Data, &trx); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction: %w", err)
	}

	status := v.mapStatus(trx.Status)

	return &OrderResponse{
		RefID:         req.RefID,
		ProviderRefID: trx.TrxID,
		SKU:           trx.Service,
		CustomerNo:    trx.Data,
		Price:         trx.Price,
		SellingPrice:  trx.Price,
		Status:        status,
		Message:       trx.Message,
		SN:            trx.SN,
		CreatedAt:     time.Now(),
		RawRequest:    body,
		RawResponse:   respBody,
	}, nil
}

// CheckStatus checks the status of an order
func (v *VIPResellerProvider) CheckStatus(ctx context.Context, refID string) (*OrderStatus, error) {
	sign := v.generateSign(v.apiID)

	vipReq := map[string]string{
		"key":    v.apiID,
		"sign":   sign,
		"type":   "status",
		"trx_id": refID,
	}

	body, err := json.Marshal(vipReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", v.baseURL+"/game-feature", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := v.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var vipResp VIPResponse
	if err := json.Unmarshal(respBody, &vipResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !vipResp.Result {
		return nil, fmt.Errorf("API error: %s", vipResp.Message)
	}

	var trx VIPTransaction
	if err := json.Unmarshal(vipResp.Data, &trx); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction: %w", err)
	}

	return &OrderStatus{
		RefID:         trx.RefID,
		ProviderRefID: trx.TrxID,
		Status:        v.mapStatus(trx.Status),
		Message:       trx.Message,
		SN:            trx.SN,
		UpdatedAt:     time.Now(),
	}, nil
}

// GetBalance returns the current balance with VIP Reseller
func (v *VIPResellerProvider) GetBalance(ctx context.Context) (*Balance, error) {
	sign := v.generateSign(v.apiID)

	reqBody := map[string]string{
		"key":  v.apiID,
		"sign": sign,
		"type": "profile",
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", v.baseURL+"/profile", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := v.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var vipResp VIPResponse
	if err := json.Unmarshal(respBody, &vipResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !vipResp.Result {
		return nil, fmt.Errorf("API error: %s", vipResp.Message)
	}

	var profile struct {
		Balance float64 `json:"balance"`
	}
	if err := json.Unmarshal(vipResp.Data, &profile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal balance: %w", err)
	}

	return &Balance{
		Balance:   profile.Balance,
		Currency:  "IDR",
		UpdatedAt: time.Now(),
	}, nil
}

// HealthCheck checks if VIP Reseller is accessible
func (v *VIPResellerProvider) HealthCheck(ctx context.Context) error {
	_, err := v.GetBalance(ctx)
	return err
}

// mapStatus maps VIP Reseller status to common status
func (v *VIPResellerProvider) mapStatus(status string) string {
	switch status {
	case "success", "completed":
		return StatusSuccess
	case "failed", "error":
		return StatusFailed
	case "pending":
		return StatusPending
	case "processing":
		return StatusProcessing
	default:
		return StatusProcessing
	}
}
