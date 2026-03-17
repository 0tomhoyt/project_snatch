package service

import (
	"context"
	"testing"

	"github.com/yourname/harborlink/internal/adapter"
	"github.com/yourname/harborlink/internal/model"
)

// mockAdapter implements adapter.CarrierAdapter for testing
type mockAdapter struct {
	carrierCode string
	carrierName string
	enabled     bool
}

func (m *mockAdapter) CreateBooking(ctx context.Context, req *model.CreateBooking) (*model.CreateBookingResponse, error) {
	return &model.CreateBookingResponse{CarrierBookingRequestReference: "test-ref"}, nil
}

func (m *mockAdapter) GetBooking(ctx context.Context, reference string) (*model.Booking, error) {
	return &model.Booking{CarrierBookingRequestReference: reference}, nil
}

func (m *mockAdapter) UpdateBooking(ctx context.Context, reference string, req *model.UpdateBooking) (*model.Booking, error) {
	return &model.Booking{CarrierBookingRequestReference: reference}, nil
}

func (m *mockAdapter) CancelBooking(ctx context.Context, reference string, req *model.CancelBookingRequest) error {
	return nil
}

func (m *mockAdapter) GetCarrierCode() string {
	return m.carrierCode
}

func (m *mockAdapter) GetCarrierName() string {
	return m.carrierName
}

func (m *mockAdapter) IsEnabled() bool {
	return m.enabled
}

func (m *mockAdapter) HealthCheck(ctx context.Context) error {
	return nil
}

// Test that CarrierRouter can be created
func TestNewCarrierRouter(t *testing.T) {
	router := NewCarrierRouter(nil)
	if router == nil {
		t.Error("expected router to be created")
	}
}

// Test HasCarrier method
func TestCarrierRouter_HasCarrier(t *testing.T) {
	registry := adapter.NewRegistry(nil)

	registry.Register(&mockAdapter{carrierCode: "MAEU", carrierName: "Maersk", enabled: true})

	router := NewCarrierRouter(registry)

	if !router.HasCarrier("MAEU") {
		t.Error("expected HasCarrier(MAEU) to return true")
	}

	if router.HasCarrier("UNKNOWN") {
		t.Error("expected HasCarrier(UNKNOWN) to return false")
	}
}

// Test GetEnabledCarriers method
func TestCarrierRouter_GetEnabledCarriers(t *testing.T) {
	registry := adapter.NewRegistry(nil)

	registry.Register(&mockAdapter{carrierCode: "MAEU", enabled: true})
	registry.Register(&mockAdapter{carrierCode: "MSCU", enabled: true})
	registry.Register(&mockAdapter{carrierCode: "CMAU", enabled: false})

	router := NewCarrierRouter(registry)

	enabled := router.GetEnabledCarriers()
	if len(enabled) != 2 {
		t.Errorf("expected 2 enabled carriers, got %d", len(enabled))
	}
}

// Test GetAllCarriers method
func TestCarrierRouter_GetAllCarriers(t *testing.T) {
	registry := adapter.NewRegistry(nil)

	registry.Register(&mockAdapter{carrierCode: "MAEU", enabled: true})
	registry.Register(&mockAdapter{carrierCode: "MSCU", enabled: false})

	router := NewCarrierRouter(registry)

	all := router.GetAllCarriers()
	if len(all) != 2 {
		t.Errorf("expected 2 total carriers, got %d", len(all))
	}
}

// Test Route method
func TestCarrierRouter_Route(t *testing.T) {
	registry := adapter.NewRegistry(nil)

	registry.Register(&mockAdapter{carrierCode: "MAEU", carrierName: "Maersk", enabled: true})
	registry.Register(&mockAdapter{carrierCode: "MSCU", carrierName: "MSC", enabled: false})

	router := NewCarrierRouter(registry)

	// Test routing to enabled carrier
	adapt, err := router.Route("MAEU")
	if err != nil {
		t.Errorf("expected no error for enabled carrier, got: %v", err)
	}
	if adapt == nil {
		t.Error("expected adapter for MAEU")
	}

	// Test routing to disabled carrier
	_, err = router.Route("MSCU")
	if err == nil {
		t.Error("expected error for disabled carrier")
	}

	// Test routing to non-existent carrier
	_, err = router.Route("UNKNOWN")
	if err == nil {
		t.Error("expected error for non-existent carrier")
	}
}
