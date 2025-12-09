package provider

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Manager manages multiple product providers
type Manager struct {
	providers map[string]Provider
	mu        sync.RWMutex
	
	// Provider health status
	healthStatus map[string]HealthStatus
	healthMu     sync.RWMutex
}

// HealthStatus represents the health status of a provider
type HealthStatus struct {
	Status    string    `json:"status"`
	LastCheck time.Time `json:"last_check"`
	Message   string    `json:"message,omitempty"`
}

// NewManager creates a new provider manager
func NewManager() *Manager {
	return &Manager{
		providers:    make(map[string]Provider),
		healthStatus: make(map[string]HealthStatus),
	}
}

// Register registers a provider with the manager
func (m *Manager) Register(provider Provider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.providers[provider.GetName()] = provider
}

// Get returns a provider by name
func (m *Manager) Get(name string) (Provider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	provider, ok := m.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", name)
	}
	return provider, nil
}

// GetAll returns all registered providers
func (m *Manager) GetAll() []Provider {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	providers := make([]Provider, 0, len(m.providers))
	for _, p := range m.providers {
		providers = append(providers, p)
	}
	return providers
}

// GetHealthStatus returns the health status of a provider
func (m *Manager) GetHealthStatus(name string) (HealthStatus, error) {
	m.healthMu.RLock()
	defer m.healthMu.RUnlock()
	
	status, ok := m.healthStatus[name]
	if !ok {
		return HealthStatus{}, fmt.Errorf("provider not found: %s", name)
	}
	return status, nil
}

// GetAllHealthStatus returns health status for all providers
func (m *Manager) GetAllHealthStatus() map[string]HealthStatus {
	m.healthMu.RLock()
	defer m.healthMu.RUnlock()
	
	result := make(map[string]HealthStatus, len(m.healthStatus))
	for k, v := range m.healthStatus {
		result[k] = v
	}
	return result
}

// CheckHealth checks the health of all providers
func (m *Manager) CheckHealth(ctx context.Context) {
	m.mu.RLock()
	providers := make([]Provider, 0, len(m.providers))
	for _, p := range m.providers {
		providers = append(providers, p)
	}
	m.mu.RUnlock()
	
	var wg sync.WaitGroup
	for _, provider := range providers {
		wg.Add(1)
		go func(p Provider) {
			defer wg.Done()
			
			status := HealthStatus{
				LastCheck: time.Now(),
			}
			
			err := p.HealthCheck(ctx)
			if err != nil {
				status.Status = "UNHEALTHY"
				status.Message = err.Error()
			} else {
				status.Status = "HEALTHY"
			}
			
			m.healthMu.Lock()
			m.healthStatus[p.GetName()] = status
			m.healthMu.Unlock()
		}(provider)
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

// GetProductsFromAll fetches products from all providers
func (m *Manager) GetProductsFromAll(ctx context.Context) (map[string][]Product, error) {
	m.mu.RLock()
	providers := make([]Provider, 0, len(m.providers))
	for _, p := range m.providers {
		providers = append(providers, p)
	}
	m.mu.RUnlock()
	
	result := make(map[string][]Product)
	var mu sync.Mutex
	var wg sync.WaitGroup
	var firstErr error
	var errMu sync.Mutex
	
	for _, provider := range providers {
		wg.Add(1)
		go func(p Provider) {
			defer wg.Done()
			
			products, err := p.GetProducts(ctx)
			if err != nil {
				errMu.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("provider %s: %w", p.GetName(), err)
				}
				errMu.Unlock()
				return
			}
			
			mu.Lock()
			result[p.GetName()] = products
			mu.Unlock()
		}(provider)
	}
	wg.Wait()
	
	// Return partial results even if some providers failed
	return result, firstErr
}

// GetBalanceFromAll fetches balance from all providers
func (m *Manager) GetBalanceFromAll(ctx context.Context) (map[string]*Balance, error) {
	m.mu.RLock()
	providers := make([]Provider, 0, len(m.providers))
	for _, p := range m.providers {
		providers = append(providers, p)
	}
	m.mu.RUnlock()
	
	result := make(map[string]*Balance)
	var mu sync.Mutex
	var wg sync.WaitGroup
	
	for _, provider := range providers {
		wg.Add(1)
		go func(p Provider) {
			defer wg.Done()
			
			balance, err := p.GetBalance(ctx)
			if err != nil {
				return
			}
			
			mu.Lock()
			result[p.GetName()] = balance
			mu.Unlock()
		}(provider)
	}
	wg.Wait()
	
	return result, nil
}

// SelectProvider selects the best provider for a product based on availability and price
func (m *Manager) SelectProvider(ctx context.Context, sku string, preferredProvider string) (Provider, *PriceInfo, error) {
	// If preferred provider is specified and healthy, try it first
	if preferredProvider != "" {
		if provider, err := m.Get(preferredProvider); err == nil {
			if status, _ := m.GetHealthStatus(preferredProvider); status.Status == "HEALTHY" {
				price, err := provider.CheckPrice(ctx, sku)
				if err == nil && price.IsAvailable {
					return provider, price, nil
				}
			}
		}
	}
	
	// Otherwise, find the best available provider
	m.mu.RLock()
	providers := make([]Provider, 0, len(m.providers))
	for _, p := range m.providers {
		providers = append(providers, p)
	}
	m.mu.RUnlock()
	
	var bestProvider Provider
	var bestPrice *PriceInfo
	
	for _, provider := range providers {
		// Check health status
		if status, _ := m.GetHealthStatus(provider.GetName()); status.Status != "HEALTHY" {
			continue
		}
		
		price, err := provider.CheckPrice(ctx, sku)
		if err != nil || !price.IsAvailable {
			continue
		}
		
		if bestPrice == nil || price.Price < bestPrice.Price {
			bestProvider = provider
			bestPrice = price
		}
	}
	
	if bestProvider == nil {
		return nil, nil, fmt.Errorf("no available provider for SKU: %s", sku)
	}
	
	return bestProvider, bestPrice, nil
}

