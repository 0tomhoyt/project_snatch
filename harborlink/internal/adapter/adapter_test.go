package adapter

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/yourname/harborlink/internal/model"
	"github.com/yourname/harborlink/pkg/config"
)

func TestMockAdapter_CreateBooking(t *testing.T) {
	adapter := newMockAdapterWithConfig()

	req := &model.CreateBooking{
		ReceiptTypeAtOrigin:            model.ReceiptDeliveryTypeCY,
		DeliveryTypeAtDestination:      model.ReceiptDeliveryTypeCY,
		CargoMovementTypeAtOrigin:      model.CargoMovementTypeFCL,
		CargoMovementTypeAtDestination: model.CargoMovementTypeFCL,
		DocumentParties: &model.DocumentPartiesReq{
			BookingAgent: &model.BookingAgent{
				PartyName: "Test Agent",
			},
		},
		ShipmentLocations: []model.ShipmentLocation{
			{
				Location:         model.Location{UNLocationCode: "DEBRV"},
				LocationTypeCode: model.LocationTypePol,
			},
		},
		RequestedEquipments: []model.RequestedEquipmentShipper{
			{
				ISOEquipmentCode: "42G1",
				Units:            1,
				IsShipperOwned:   false,
			},
		},
	}

	resp, err := adapter.CreateBooking(context.Background(), req)
	if err != nil {
		t.Errorf("failed to create booking: %v", err)
	}

	if resp.CarrierBookingRequestReference == "" {
		t.Error("expected non-empty carrier booking request reference")
	}
}

func TestMockAdapter_GetBooking(t *testing.T) {
	adapter := newMockAdapterWithConfig()

	// First create a booking
	req := &model.CreateBooking{
		ReceiptTypeAtOrigin:            model.ReceiptDeliveryTypeCY,
		DeliveryTypeAtDestination:      model.ReceiptDeliveryTypeCY,
		CargoMovementTypeAtOrigin:      model.CargoMovementTypeFCL,
		CargoMovementTypeAtDestination: model.CargoMovementTypeFCL,
		DocumentParties: &model.DocumentPartiesReq{
			BookingAgent: &model.BookingAgent{
				PartyName: "Test Agent",
			},
		},
		ShipmentLocations: []model.ShipmentLocation{
			{
				Location:         model.Location{UNLocationCode: "DEBRV"},
				LocationTypeCode: model.LocationTypePol,
			},
		},
		RequestedEquipments: []model.RequestedEquipmentShipper{
			{
				ISOEquipmentCode: "42G1",
				Units:            1,
				IsShipperOwned:   false,
			},
		},
	}

	createResp, _ := adapter.CreateBooking(context.Background(), req)

	// Now get the booking
	booking, err := adapter.GetBooking(context.Background(), createResp.CarrierBookingRequestReference)
	if err != nil {
		t.Errorf("failed to get booking: %v", err)
	}

	if booking.BookingStatus != model.BookingStatusReceived {
		t.Errorf("expected status RECEIVED, got %s", booking.BookingStatus)
	}
}

func TestMockAdapter_GetBooking_NotFound(t *testing.T) {
	adapter := newMockAdapterWithConfig()

	_, err := adapter.GetBooking(context.Background(), "non-existent-ref")
	if err == nil {
		t.Error("expected error for non-existent booking")
		return
	}

	var adapterErr *AdapterError
	if !errors.As(err, &adapterErr) {
		t.Errorf("expected AdapterError, got %T", err)
		return
	}

	if adapterErr.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", adapterErr.StatusCode)
	}
}

func TestMockAdapter_CancelBooking(t *testing.T) {
	adapter := newMockAdapterWithConfig()

	// First create a booking
	req := &model.CreateBooking{
		ReceiptTypeAtOrigin:            model.ReceiptDeliveryTypeCY,
		DeliveryTypeAtDestination:      model.ReceiptDeliveryTypeCY,
		CargoMovementTypeAtOrigin:      model.CargoMovementTypeFCL,
		CargoMovementTypeAtDestination: model.CargoMovementTypeFCL,
		DocumentParties: &model.DocumentPartiesReq{
			BookingAgent: &model.BookingAgent{
				PartyName: "Test Agent",
			},
		},
		ShipmentLocations: []model.ShipmentLocation{
			{
				Location:         model.Location{UNLocationCode: "DEBRV"},
				LocationTypeCode: model.LocationTypePol,
			},
		},
		RequestedEquipments: []model.RequestedEquipmentShipper{
			{
				ISOEquipmentCode: "42G1",
				Units:            1,
				IsShipperOwned:   false,
			},
		},
	}

	createResp, _ := adapter.CreateBooking(context.Background(), req)

	// Cancel the booking
	err := adapter.CancelBooking(context.Background(), createResp.CarrierBookingRequestReference, &model.CancelBookingRequest{
		BookingStatus: model.BookingStatusCancelled,
		Reason:        "Test cancellation",
	})
	if err != nil {
		t.Errorf("failed to cancel booking: %v", err)
	}
}

func TestMockAdapter_ShouldFail(t *testing.T) {
	adapter := newMockAdapterWithConfig()
	adapter.SetShouldFail(true)

	req := &model.CreateBooking{
		ReceiptTypeAtOrigin:            model.ReceiptDeliveryTypeCY,
		DeliveryTypeAtDestination:      model.ReceiptDeliveryTypeCY,
		CargoMovementTypeAtOrigin:      model.CargoMovementTypeFCL,
		CargoMovementTypeAtDestination: model.CargoMovementTypeFCL,
		DocumentParties: &model.DocumentPartiesReq{
			BookingAgent: &model.BookingAgent{
				PartyName: "Test Agent",
			},
		},
		ShipmentLocations: []model.ShipmentLocation{
			{
				Location:         model.Location{UNLocationCode: "DEBRV"},
				LocationTypeCode: model.LocationTypePol,
			},
		},
		RequestedEquipments: []model.RequestedEquipmentShipper{
			{
				ISOEquipmentCode: "42G1",
				Units:            1,
				IsShipperOwned:   false,
			},
		},
	}

	_, err := adapter.CreateBooking(context.Background(), req)
	if err == nil {
		t.Error("expected error when shouldFail is true")
	}
}

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry(nil)

	// Create adapter manually
	adapter := newMockAdapterWithConfig()

	err := registry.Register(adapter)
	if err != nil {
		t.Errorf("failed to register adapter: %v", err)
	}

	// Verify registration
	a, err := registry.Get("MOCK")
	if err != nil {
		t.Errorf("failed to get adapter: %v", err)
		return
	}

	if a.GetCarrierCode() != "MOCK" {
		t.Errorf("expected code MOCK, got %s", a.GetCarrierCode())
	}
}

func TestRegistry_GetAll(t *testing.T) {
	registry := NewRegistry(nil)

	// Register multiple adapters
	for i := 0; i < 3; i++ {
		cfg := &config.CarrierConfig{
			Name:    fmt.Sprintf("Mock Carrier %d", i),
			Code:    fmt.Sprintf("MOCK%d", i),
			Adapter: "mock",
			Enabled: i < 2, // Only first two are enabled
		}
		adapter := NewMockAdapter(NewBaseAdapter(cfg, nil))
		registry.Register(adapter)
	}

	// Get all enabled adapters
	enabled := registry.GetEnabled()
	if len(enabled) != 2 {
		t.Errorf("expected 2 enabled adapters, got %d", len(enabled))
	}
}

func newMockAdapterWithConfig() *MockAdapter {
	cfg := &config.CarrierConfig{
		Name:      "Mock Carrier",
		Code:      "MOCK",
		Adapter:   "mock",
		Enabled:   true,
		BaseURL:   "https://mock.carrier.com",
		RateLimit: 100,
	}
	base := NewBaseAdapter(cfg, nil)
	return NewMockAdapter(base)
}
