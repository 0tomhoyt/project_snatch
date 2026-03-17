package adapter

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/yourname/harborlink/internal/model"
)

// MockAdapter is a mock implementation for testing
type MockAdapter struct {
	*BaseAdapter
	mu            sync.RWMutex
	shouldFail    bool
	bookingDelay time.Duration
	bookings     map[string]*model.Booking
}

// NewMockAdapter creates a new mock adapter
func NewMockAdapter(base *BaseAdapter) *MockAdapter {
	return &MockAdapter{
		BaseAdapter: base,
		bookings:     make(map[string]*model.Booking),
	}
}

// SetShouldFail configures the adapter to fail all requests
func (a *MockAdapter) SetShouldFail(shouldFail bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.shouldFail = shouldFail
}

// SetBookingDelay configures artificial delay for testing
func (a *MockAdapter) SetBookingDelay(delay time.Duration) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.bookingDelay = delay
}

// CreateBooking simulates creating a booking
func (a *MockAdapter) CreateBooking(ctx context.Context, req *model.CreateBooking) (*model.CreateBookingResponse, error) {
	if a.shouldFail {
		return nil, NewAdapterError(a.GetCarrierCode(), http.StatusInternalServerError, "MOCK_ERROR", "Mock adapter is configured to fail")
	}

	if a.bookingDelay > 0 {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(a.bookingDelay):
		}
	}

	// Generate a mock reference
	ref := fmt.Sprintf("MOCK-%d", time.Now().UnixNano())

	// Store the booking
	booking := &model.Booking{
		CarrierBookingRequestReference: ref,
		BookingStatus:                 model.BookingStatusReceived,
		ReceiptTypeAtOrigin:           req.ReceiptTypeAtOrigin,
		DeliveryTypeAtDestination:     req.DeliveryTypeAtDestination,
	}

	a.mu.Lock()
	a.bookings[ref] = booking
	a.mu.Unlock()

	return &model.CreateBookingResponse{
		CarrierBookingRequestReference: ref,
	}, nil
}

// GetBooking simulates retrieving a booking
func (a *MockAdapter) GetBooking(ctx context.Context, reference string) (*model.Booking, error) {
	if a.shouldFail {
		return nil, NewAdapterError(a.GetCarrierCode(), http.StatusInternalServerError, "MOCK_ERROR", "Mock adapter is configured to fail")
	}

	a.mu.RLock()
	booking, exists := a.bookings[reference]
	a.mu.RUnlock()

	if !exists {
		return nil, NewAdapterError(a.GetCarrierCode(), http.StatusNotFound, "NOT_FOUND", "Booking not found")
	}

	return booking, nil
}

// UpdateBooking simulates updating a booking
func (a *MockAdapter) UpdateBooking(ctx context.Context, reference string, req *model.UpdateBooking) (*model.Booking, error) {
	if a.shouldFail {
		return nil, NewAdapterError(a.GetCarrierCode(), http.StatusInternalServerError, "MOCK_ERROR", "Mock adapter is configured to fail")
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	booking, exists := a.bookings[reference]
	if !exists {
		return nil, NewAdapterError(a.GetCarrierCode(), http.StatusNotFound, "NOT_FOUND", "Booking not found")
	}

	// Update status
	booking.BookingStatus = model.BookingStatusUpdateReceived

	return booking, nil
}

// CancelBooking simulates cancelling a booking
func (a *MockAdapter) CancelBooking(ctx context.Context, reference string, req *model.CancelBookingRequest) error {
	if a.shouldFail {
		return NewAdapterError(a.GetCarrierCode(), http.StatusInternalServerError, "MOCK_ERROR", "Mock adapter is configured to fail")
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	booking, exists := a.bookings[reference]
	if !exists {
		return NewAdapterError(a.GetCarrierCode(), http.StatusNotFound, "NOT_FOUND", "Booking not found")
	}

	booking.BookingStatus = model.BookingStatusCancelled
	return nil
}

// HealthCheck returns the health status
func (a *MockAdapter) HealthCheck(ctx context.Context) error {
	if a.shouldFail {
		return errors.New("mock adapter unhealthy")
	}
	return a.BaseAdapter.HealthCheck(ctx)
}
