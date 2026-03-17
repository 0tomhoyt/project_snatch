package service

import (
	"errors"

	"github.com/yourname/harborlink/internal/adapter"
)

// Common errors
var (
	ErrCarrierNotFound = errors.New("carrier not found")
	ErrNoCarriers      = errors.New("no carriers available")
)

// CarrierRouter routes requests to the appropriate carrier adapter
type CarrierRouter struct {
	registry *adapter.Registry
}

// NewCarrierRouter creates a new carrier router
func NewCarrierRouter(registry *adapter.Registry) *CarrierRouter {
	return &CarrierRouter{
		registry: registry,
	}
}

// Route returns the carrier adapter for the given carrier code
func (r *CarrierRouter) Route(carrierCode string) (adapter.CarrierAdapter, error) {
	adapter, err := r.registry.Get(carrierCode)
	if err != nil {
		return nil, ErrCarrierNotFound
	}

	if !adapter.IsEnabled() {
		return nil, ErrCarrierNotFound
	}

	return adapter, nil
}

// GetEnabledCarriers returns all enabled carrier adapters
func (r *CarrierRouter) GetEnabledCarriers() []adapter.CarrierAdapter {
	return r.registry.GetEnabled()
}

// GetAllCarriers returns all registered carrier adapters
func (r *CarrierRouter) GetAllCarriers() map[string]adapter.CarrierAdapter {
	return r.registry.GetAll()
}

// HasCarrier checks if a carrier is registered and enabled
func (r *CarrierRouter) HasCarrier(carrierCode string) bool {
	adapter, err := r.registry.Get(carrierCode)
	if err != nil {
		return false
	}
	return adapter.IsEnabled()
}
