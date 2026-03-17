package adapter

import (
	"context"
	"fmt"
	"sync"

	"github.com/yourname/harborlink/pkg/cache"
	"github.com/yourname/harborlink/pkg/config"
)

// Registry manages all carrier adapters
type Registry struct {
	mu       sync.RWMutex
	adapters map[string]CarrierAdapter
	cache    *cache.BookingCache
}

// NewRegistry creates a new adapter registry
func NewRegistry(cache *cache.BookingCache) *Registry {
	return &Registry{
		adapters: make(map[string]CarrierAdapter),
		cache:    cache,
	}
}

// Register adds a carrier adapter to the registry
func (r *Registry) Register(adapter CarrierAdapter) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	code := adapter.GetCarrierCode()
	if _, exists := r.adapters[code]; exists {
		return fmt.Errorf("adapter for carrier %s already registered", code)
	}

	r.adapters[code] = adapter
	return nil
}

// Get retrieves an adapter by carrier code
func (r *Registry) Get(carrierCode string) (CarrierAdapter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	adapter, exists := r.adapters[carrierCode]
	if !exists {
		return nil, fmt.Errorf("no adapter found for carrier %s", carrierCode)
	}

	return adapter, nil
}

// GetAll returns all registered adapters
func (r *Registry) GetAll() map[string]CarrierAdapter {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]CarrierAdapter, len(r.adapters))
	for k, v := range r.adapters {
		result[k] = v
	}
	return result
}

// GetEnabled returns all enabled adapters
func (r *Registry) GetEnabled() []CarrierAdapter {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var enabled []CarrierAdapter
	for _, adapter := range r.adapters {
		if adapter.IsEnabled() {
			enabled = append(enabled, adapter)
		}
	}
	return enabled
}

// HealthCheckAll performs health check on all adapters
func (r *Registry) HealthCheckAll(ctx context.Context) map[string]error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	results := make(map[string]error)
	for code, adapter := range r.adapters {
		results[code] = adapter.HealthCheck(ctx)
	}
	return results
}

// InitializeFromConfig initializes adapters from configuration
func (r *Registry) InitializeFromConfig(carriers []config.CarrierConfig) error {
	for _, cfg := range carriers {
		if !cfg.Enabled {
			continue
		}

		// Create adapter based on adapter type
		adapter, err := r.createAdapter(&cfg)
		if err != nil {
			return fmt.Errorf("failed to create adapter for %s: %w", cfg.Code, err)
		}

		if err := r.Register(adapter); err != nil {
			return err
		}
	}
	return nil
}

// createAdapter creates an adapter based on configuration
func (r *Registry) createAdapter(cfg *config.CarrierConfig) (CarrierAdapter, error) {
	base := NewBaseAdapter(cfg, r.cache)

	switch cfg.Adapter {
	case "mock":
		return NewMockAdapter(base), nil
	// Add cases for real carriers as they are implemented
	// case "maersk":
	//     return NewMaerskAdapter(base), nil
	// case "msc":
	//     return NewMSCAdapter(base), nil
	default:
		return nil, fmt.Errorf("unknown adapter type: %s", cfg.Adapter)
	}
}
