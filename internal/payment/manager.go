package payment

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Manager manages multiple payment gateways
type Manager struct {
	gateways map[string]Gateway
	mu       sync.RWMutex

	// Channel to gateway mapping
	channelGateway map[string]string
	channelMu      sync.RWMutex

	// Gateway health status
	healthStatus map[string]HealthStatus
	healthMu     sync.RWMutex
}

// HealthStatus represents the health status of a gateway
type HealthStatus struct {
	Status    string    `json:"status"`
	LastCheck time.Time `json:"last_check"`
	Message   string    `json:"message,omitempty"`
}

// NewManager creates a new payment gateway manager
func NewManager() *Manager {
	return &Manager{
		gateways:       make(map[string]Gateway),
		channelGateway: make(map[string]string),
		healthStatus:   make(map[string]HealthStatus),
	}
}

// Register registers a gateway with the manager
func (m *Manager) Register(gateway Gateway) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.gateways[gateway.GetName()] = gateway

	// Register supported channels
	m.channelMu.Lock()
	for _, channel := range gateway.GetSupportedChannels() {
		m.channelGateway[channel] = gateway.GetName()
	}
	m.channelMu.Unlock()
}

// Get returns a gateway by name
func (m *Manager) Get(name string) (Gateway, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	gateway, ok := m.gateways[name]
	if !ok {
		return nil, fmt.Errorf("gateway not found: %s", name)
	}
	return gateway, nil
}

// GetByChannel returns a gateway for a specific payment channel
func (m *Manager) GetByChannel(channel string) (Gateway, error) {
	m.channelMu.RLock()
	gatewayName, ok := m.channelGateway[channel]
	m.channelMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("no gateway configured for channel: %s", channel)
	}

	return m.Get(gatewayName)
}

// GetAll returns all registered gateways
func (m *Manager) GetAll() []Gateway {
	m.mu.RLock()
	defer m.mu.RUnlock()

	gateways := make([]Gateway, 0, len(m.gateways))
	for _, g := range m.gateways {
		gateways = append(gateways, g)
	}
	return gateways
}

// GetSupportedChannels returns all supported payment channels
func (m *Manager) GetSupportedChannels() []string {
	m.channelMu.RLock()
	defer m.channelMu.RUnlock()

	channels := make([]string, 0, len(m.channelGateway))
	for channel := range m.channelGateway {
		channels = append(channels, channel)
	}
	return channels
}

// SetChannelGateway sets the gateway for a specific channel
func (m *Manager) SetChannelGateway(channel, gatewayName string) error {
	// Verify gateway exists
	if _, err := m.Get(gatewayName); err != nil {
		return err
	}

	m.channelMu.Lock()
	m.channelGateway[channel] = gatewayName
	m.channelMu.Unlock()

	return nil
}

// GetHealthStatus returns the health status of a gateway
func (m *Manager) GetHealthStatus(name string) (HealthStatus, error) {
	m.healthMu.RLock()
	defer m.healthMu.RUnlock()

	status, ok := m.healthStatus[name]
	if !ok {
		return HealthStatus{}, fmt.Errorf("gateway not found: %s", name)
	}
	return status, nil
}

// GetAllHealthStatus returns health status for all gateways
func (m *Manager) GetAllHealthStatus() map[string]HealthStatus {
	m.healthMu.RLock()
	defer m.healthMu.RUnlock()

	result := make(map[string]HealthStatus, len(m.healthStatus))
	for k, v := range m.healthStatus {
		result[k] = v
	}
	return result
}

// CheckHealth checks the health of all gateways
func (m *Manager) CheckHealth(ctx context.Context) {
	m.mu.RLock()
	gateways := make([]Gateway, 0, len(m.gateways))
	for _, g := range m.gateways {
		gateways = append(gateways, g)
	}
	m.mu.RUnlock()

	var wg sync.WaitGroup
	for _, gateway := range gateways {
		wg.Add(1)
		go func(g Gateway) {
			defer wg.Done()

			status := HealthStatus{
				LastCheck: time.Now(),
			}

			err := g.HealthCheck(ctx)
			if err != nil {
				status.Status = "UNHEALTHY"
				status.Message = err.Error()
			} else {
				status.Status = "HEALTHY"
			}

			m.healthMu.Lock()
			m.healthStatus[g.GetName()] = status
			m.healthMu.Unlock()
		}(gateway)
	}
	wg.Wait()
}

// StartHealthCheck starts periodic health checking
func (m *Manager) StartHealthCheck(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		// Initial health check
		m.CheckHealth(ctx)

		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				m.CheckHealth(ctx)
			}
		}
	}()
}

// CreatePayment creates a payment using the appropriate gateway
// If GatewayName is specified in the request, use that gateway directly.
// Otherwise, fall back to channel-to-gateway mapping.
func (m *Manager) CreatePayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	var gateway Gateway
	var err error

	// Use explicit gateway name if provided (from database)
	if req.GatewayName != "" {
		gateway, err = m.Get(req.GatewayName)
		if err != nil {
			return nil, fmt.Errorf("gateway %s not found: %w", req.GatewayName, err)
		}
	} else {
		// Fall back to channel-based lookup
		gateway, err = m.GetByChannel(req.Channel)
		if err != nil {
			return nil, err
		}
	}

	// Log health status but don't block - health check may not be accurate for all gateways
	// (e.g., DANA doesn't have a root endpoint so health check returns 404)
	if status, err := m.GetHealthStatus(gateway.GetName()); err == nil && status.Status == "UNHEALTHY" {
		// Just log warning, don't block payment attempt
		// The actual payment call will fail if the gateway is truly unavailable
		_ = status // logged elsewhere during health check
	}

	return gateway.CreatePayment(ctx, req)
}

// CheckPaymentStatus checks the status of a payment
func (m *Manager) CheckPaymentStatus(ctx context.Context, gatewayName, paymentID string) (*PaymentStatus, error) {
	gateway, err := m.Get(gatewayName)
	if err != nil {
		return nil, err
	}

	return gateway.CheckStatus(ctx, paymentID)
}

// ChannelConfig represents configuration for a payment channel
type ChannelConfig struct {
	Code          string   `json:"code"`
	Name          string   `json:"name"`
	Type          string   `json:"type"`
	Gateway       string   `json:"gateway"`
	IsActive      bool     `json:"is_active"`
	MinAmount     float64  `json:"min_amount"`
	MaxAmount     float64  `json:"max_amount"`
	FeeFixed      float64  `json:"fee_fixed"`
	FeePercentage float64  `json:"fee_percentage"`
	Icon          string   `json:"icon"`
	Instructions  []string `json:"instructions"`
}

// DefaultChannelConfigs returns default channel configurations
func DefaultChannelConfigs() []ChannelConfig {
	return []ChannelConfig{
		{
			Code:          "QRIS",
			Name:          "QRIS",
			Type:          ChannelTypeQRIS,
			Gateway:       "linkqu",
			IsActive:      true,
			MinAmount:     1000,
			MaxAmount:     10000000,
			FeeFixed:      0,
			FeePercentage: 0.7,
			Icon:          "qris.png",
			Instructions: []string{
				"Buka aplikasi mobile banking atau e-wallet",
				"Pilih menu Scan QR atau QRIS",
				"Scan QR Code yang ditampilkan",
				"Konfirmasi pembayaran",
			},
		},
		{
			Code:          "VA_BCA",
			Name:          "BCA Virtual Account",
			Type:          ChannelTypeVirtualAccount,
			Gateway:       "bca",
			IsActive:      true,
			MinAmount:     10000,
			MaxAmount:     50000000,
			FeeFixed:      4000,
			FeePercentage: 0,
			Icon:          "bca.png",
			Instructions: []string{
				"Login ke BCA Mobile atau KlikBCA",
				"Pilih menu Transfer > Virtual Account",
				"Masukkan nomor Virtual Account",
				"Konfirmasi pembayaran",
			},
		},
		{
			Code:          "VA_BRI",
			Name:          "BRI Virtual Account",
			Type:          ChannelTypeVirtualAccount,
			Gateway:       "bri",
			IsActive:      true,
			MinAmount:     10000,
			MaxAmount:     50000000,
			FeeFixed:      4000,
			FeePercentage: 0,
			Icon:          "bri.png",
			Instructions: []string{
				"Login ke BRI Mobile atau Internet Banking BRI",
				"Pilih menu BRIVA",
				"Masukkan nomor BRIVA",
				"Konfirmasi pembayaran",
			},
		},
		{
			Code:          "VA_PERMATA",
			Name:          "Permata Virtual Account",
			Type:          ChannelTypeVirtualAccount,
			Gateway:       "xendit",
			IsActive:      true,
			MinAmount:     10000,
			MaxAmount:     50000000,
			FeeFixed:      4000,
			FeePercentage: 0,
			Icon:          "permata.png",
			Instructions: []string{
				"Login ke PermataNet atau PermataMobile",
				"Pilih menu Virtual Account",
				"Masukkan nomor Virtual Account",
				"Konfirmasi pembayaran",
			},
		},
		{
			Code:          "VA_MANDIRI",
			Name:          "Mandiri Virtual Account",
			Type:          ChannelTypeVirtualAccount,
			Gateway:       "xendit",
			IsActive:      true,
			MinAmount:     10000,
			MaxAmount:     50000000,
			FeeFixed:      4000,
			FeePercentage: 0,
			Icon:          "mandiri.png",
			Instructions: []string{
				"Login ke Livin by Mandiri",
				"Pilih menu Bayar > Multipayment",
				"Masukkan nomor Virtual Account",
				"Konfirmasi pembayaran",
			},
		},
		{
			Code:          "GOPAY",
			Name:          "GoPay",
			Type:          ChannelTypeEWallet,
			Gateway:       "midtrans",
			IsActive:      true,
			MinAmount:     1000,
			MaxAmount:     10000000,
			FeeFixed:      0,
			FeePercentage: 2,
			Icon:          "gopay.png",
			Instructions: []string{
				"Buka aplikasi Gojek atau GoPay",
				"Tap tombol Bayar",
				"Atau scan QR Code",
				"Konfirmasi pembayaran",
			},
		},
		{
			Code:          "SHOPEEPAY",
			Name:          "ShopeePay",
			Type:          ChannelTypeEWallet,
			Gateway:       "midtrans",
			IsActive:      true,
			MinAmount:     1000,
			MaxAmount:     10000000,
			FeeFixed:      0,
			FeePercentage: 2,
			Icon:          "shopeepay.png",
			Instructions: []string{
				"Buka aplikasi Shopee",
				"Pilih menu ShopeePay > Scan",
				"Scan QR Code yang ditampilkan",
				"Konfirmasi pembayaran",
			},
		},
		{
			Code:          "DANA",
			Name:          "DANA",
			Type:          ChannelTypeEWallet,
			Gateway:       "dana",
			IsActive:      true,
			MinAmount:     1000,
			MaxAmount:     10000000,
			FeeFixed:      0,
			FeePercentage: 1.5,
			Icon:          "dana.png",
			Instructions: []string{
				"Klik tombol Bayar dengan DANA",
				"Login ke akun DANA",
				"Periksa detail pembayaran",
				"Konfirmasi dengan PIN",
			},
		},
	}
}

// CalculateFee calculates the payment fee based on channel configuration
func CalculateFee(config ChannelConfig, amount float64) float64 {
	percentageFee := amount * (config.FeePercentage / 100)
	return config.FeeFixed + percentageFee
}
