package adapter

import (
	"context"

	"github.com/yourname/harborlink/internal/model"
)

// CarrierAdapter defines the interface for shipping carrier integrations
// Each carrier (Maersk, MSC, CMA CGM, etc.) must implement this interface
type CarrierAdapter interface {
	// CreateBooking submits a new booking request to the carrier
	// Returns CreateBookingResponse with carrier's reference on success
	CreateBooking(ctx context.Context, req *model.CreateBooking) (*model.CreateBookingResponse, error)

	// GetBooking retrieves booking details from the carrier
	// reference can be carrierBookingRequestReference or carrierBookingReference
	GetBooking(ctx context.Context, reference string) (*model.Booking, error)

	// UpdateBooking updates an existing booking or creates an amendment
	UpdateBooking(ctx context.Context, reference string, req *model.UpdateBooking) (*model.Booking, error)

	// CancelBooking cancels a booking request
	CancelBooking(ctx context.Context, reference string, req *model.CancelBookingRequest) error

	// GetCarrierCode returns the carrier's SMDG or NMFTA code (e.g., "MAEU", "MSCU")
	GetCarrierCode() string

	// GetCarrierName returns the human-readable carrier name
	GetCarrierName() string

	// IsEnabled returns whether the adapter is currently enabled
	IsEnabled() bool

	// HealthCheck verifies the carrier API is accessible
	HealthCheck(ctx context.Context) error
}

// AdapterConfig contains configuration for a carrier adapter
type AdapterConfig struct {
	Code      string // SMDG code (e.g., "MAEU")
	Name      string // Human readable name (e.g., "Maersk")
	BaseURL   string // API base URL
	APIKey    string // API key for authentication
	RateLimit int    // Requests per minute limit
	Enabled   bool   // Whether adapter is enabled
}

// AdapterError represents an error from a carrier adapter
type AdapterError struct {
	CarrierCode string            // Carrier that returned the error
	StatusCode  int               // HTTP status code from carrier
	Code        string            // Carrier-specific error code
	Message     string            // Error message
	Details     *model.ErrorResponse // Full error response if available
}

// Error implements the error interface
func (e *AdapterError) Error() string {
	return e.Message
}

// IsAdapterError checks if an error is an AdapterError
func IsAdapterError(err error) bool {
	_, ok := err.(*AdapterError)
	return ok
}

// NewAdapterError creates a new AdapterError
func NewAdapterError(carrierCode string, statusCode int, code, message string) *AdapterError {
	return &AdapterError{
		CarrierCode: carrierCode,
		StatusCode:  statusCode,
		Code:        code,
		Message:     message,
	}
}

// WithDetails adds error response details to the AdapterError
func (e *AdapterError) WithDetails(details *model.ErrorResponse) *AdapterError {
	e.Details = details
	return e
}
