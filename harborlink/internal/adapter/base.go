package adapter

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/yourname/harborlink/pkg/cache"
	"github.com/yourname/harborlink/pkg/config"
)

// BaseAdapter provides common functionality for carrier adapters
type BaseAdapter struct {
	config    *config.CarrierConfig
	cache    *cache.BookingCache
	enabled  bool
	mu      sync.RWMutex
}

// NewBaseAdapter creates a new base adapter
func NewBaseAdapter(cfg *config.CarrierConfig, cache *cache.BookingCache) *BaseAdapter {
	return &BaseAdapter{
		config:   cfg,
		cache:    cache,
		enabled:  cfg.Enabled,
	}
}

// GetCarrierCode returns the carrier's SMDG code
func (a *BaseAdapter) GetCarrierCode() string {
	return a.config.Code
}

// GetCarrierName returns the carrier's human-readable name
func (a *BaseAdapter) GetCarrierName() string {
	return a.config.Name
}

// IsEnabled returns whether the adapter is enabled
func (a *BaseAdapter) IsEnabled() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.enabled
}

// SetEnabled enables or disables the adapter
func (a *BaseAdapter) SetEnabled(enabled bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.enabled = enabled
}

// GetConfig returns the carrier configuration
func (a *BaseAdapter) GetConfig() *config.CarrierConfig {
	return a.config
}

// AcquireLock acquires a distributed lock for carrier API call
func (a *BaseAdapter) AcquireLock(ctx context.Context) (bool, error) {
	if a.cache == nil {
		return true, nil
	}
	return a.cache.AcquireCarrierLock(ctx, a.config.Code, 30*time.Second)
}

// ReleaseLock releases the carrier lock
func (a *BaseAdapter) ReleaseLock(ctx context.Context) error {
	if a.cache == nil {
		return nil
	}
	return a.cache.ReleaseCarrierLock(ctx, a.config.Code)
}

// CheckRateLimit checks if the rate limit has been exceeded
func (a *BaseAdapter) CheckRateLimit(ctx context.Context) (bool, error) {
	if a.cache == nil || a.config.RateLimit <= 0 {
		return false, nil
	}

	key := "ratelimit:" + a.config.Code
	count, err := a.cache.IncrementRateLimit(ctx, key, time.Minute)
	if err != nil {
		return false, err
	}

	return count > int64(a.config.RateLimit), nil
}

// HealthCheck performs a basic health check (override in specific adapters)
func (a *BaseAdapter) HealthCheck(ctx context.Context) error {
	// Base implementation just checks if enabled
	if !a.IsEnabled() {
		return NewAdapterError(a.config.Code, http.StatusServiceUnavailable, "ADAPTER_DISABLED", "Carrier adapter is disabled")
	}
	return nil
}
