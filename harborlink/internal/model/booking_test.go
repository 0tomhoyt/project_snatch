package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestBookingStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   BookingStatus
		expected string
	}{
		{"Received", BookingStatusReceived, "RECEIVED"},
		{"Confirmed", BookingStatusConfirmed, "CONFIRMED"},
		{"Cancelled", BookingStatusCancelled, "CANCELLED"},
		{"Rejected", BookingStatusRejected, "REJECTED"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.status)
			}
		})
	}
}

func TestCreateBookingRequest(t *testing.T) {
	now := time.Now()
	req := CreateBooking{
		ReceiptTypeAtOrigin:            ReceiptDeliveryTypeCY,
		DeliveryTypeAtDestination:      ReceiptDeliveryTypeCY,
		CargoMovementTypeAtOrigin:      CargoMovementTypeFCL,
		CargoMovementTypeAtDestination: CargoMovementTypeFCL,
		ServiceContractReference:       "HHL51800000",
		FreightPaymentTermCode:         FreightPaymentTermPre,
		IsEquipmentSubstitutionAllowed: false,
		ExpectedDepartureDate:          &now,
		DocumentParties: &DocumentPartiesReq{
			BookingAgent: &BookingAgent{
				PartyName: "Test Agent",
			},
		},
		ShipmentLocations: []ShipmentLocation{
			{
				Location: Location{
					UNLocationCode: "DEBRV",
					LocationName:   "Bremerhaven",
				},
				LocationTypeCode: LocationTypePol,
			},
		},
		RequestedEquipments: []RequestedEquipmentShipper{
			{
				ISOEquipmentCode: "42G1",
				Units:            3,
				IsShipperOwned:   false,
			},
		},
	}

	// Test JSON serialization
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal CreateBooking: %v", err)
	}

	var unmarshaled CreateBooking
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal CreateBooking: %v", err)
	}

	if unmarshaled.ReceiptTypeAtOrigin != req.ReceiptTypeAtOrigin {
		t.Errorf("expected ReceiptTypeAtOrigin %s, got %s", req.ReceiptTypeAtOrigin, unmarshaled.ReceiptTypeAtOrigin)
	}

	if unmarshaled.ServiceContractReference != req.ServiceContractReference {
		t.Errorf("expected ServiceContractReference %s, got %s", req.ServiceContractReference, unmarshaled.ServiceContractReference)
	}

	if len(unmarshaled.ShipmentLocations) != 1 {
		t.Errorf("expected 1 ShipmentLocation, got %d", len(unmarshaled.ShipmentLocations))
	}

	if len(unmarshaled.RequestedEquipments) != 1 {
		t.Errorf("expected 1 RequestedEquipment, got %d", len(unmarshaled.RequestedEquipments))
	}
}

func TestCreateBookingResponse(t *testing.T) {
	resp := CreateBookingResponse{
		CarrierBookingRequestReference: "cbrr-123",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal CreateBookingResponse: %v", err)
	}

	var unmarshaled CreateBookingResponse
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal CreateBookingResponse: %v", err)
	}

	if unmarshaled.CarrierBookingRequestReference != resp.CarrierBookingRequestReference {
		t.Errorf("expected %s, got %s", resp.CarrierBookingRequestReference, unmarshaled.CarrierBookingRequestReference)
	}
}

func TestErrorResponse(t *testing.T) {
	errResp := NewErrorResponse(
		HTTPMethodPost,
		"/v2/bookings",
		400,
		"Bad Request",
		"receiptTypeAtOrigin not found",
	)
	errResp.ProviderCorrelationReference = "4426d965-0dd8-4005-8c63-dc68b01c4962"
	errResp.AddError(7003, "receiptTypeAtOrigin", "", "mandatory property missing", "receiptTypeAtOrigin must be provided")

	data, err := json.Marshal(errResp)
	if err != nil {
		t.Fatalf("failed to marshal ErrorResponse: %v", err)
	}

	var unmarshaled ErrorResponse
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal ErrorResponse: %v", err)
	}

	if unmarshaled.StatusCode != 400 {
		t.Errorf("expected StatusCode 400, got %d", unmarshaled.StatusCode)
	}

	if len(unmarshaled.Errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(unmarshaled.Errors))
	}

	if unmarshaled.Errors[0].ErrorCode != 7003 {
		t.Errorf("expected ErrorCode 7003, got %d", unmarshaled.Errors[0].ErrorCode)
	}
}

func TestVessel(t *testing.T) {
	vessel := Vessel{
		Name:            "MAERSK IOWA",
		VesselIMONumber: "9298686",
	}

	data, err := json.Marshal(vessel)
	if err != nil {
		t.Fatalf("failed to marshal Vessel: %v", err)
	}

	var unmarshaled Vessel
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal Vessel: %v", err)
	}

	if unmarshaled.Name != vessel.Name {
		t.Errorf("expected Name %s, got %s", vessel.Name, unmarshaled.Name)
	}
}

func TestLocation(t *testing.T) {
	loc := Location{
		LocationName:   "Bremerhaven",
		UNLocationCode: "DEBRV",
	}

	data, err := json.Marshal(loc)
	if err != nil {
		t.Fatalf("failed to marshal Location: %v", err)
	}

	var unmarshaled Location
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal Location: %v", err)
	}

	if unmarshaled.UNLocationCode != loc.UNLocationCode {
		t.Errorf("expected UNLocationCode %s, got %s", loc.UNLocationCode, unmarshaled.UNLocationCode)
	}
}

func TestRequestedEquipment(t *testing.T) {
	eq := RequestedEquipment{
		ISOEquipmentCode: "42G1",
		Units:            3,
		IsShipperOwned:   false,
		CargoGrossWeight: &CargoGrossWeightReq{
			Value: 36000,
			Unit:  WeightUnitKGM,
		},
	}

	data, err := json.Marshal(eq)
	if err != nil {
		t.Fatalf("failed to marshal RequestedEquipment: %v", err)
	}

	var unmarshaled RequestedEquipment
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal RequestedEquipment: %v", err)
	}

	if unmarshaled.Units != eq.Units {
		t.Errorf("expected Units %d, got %d", eq.Units, unmarshaled.Units)
	}

	if unmarshaled.CargoGrossWeight.Value != eq.CargoGrossWeight.Value {
		t.Errorf("expected CargoGrossWeight.Value %f, got %f", eq.CargoGrossWeight.Value, unmarshaled.CargoGrossWeight.Value)
	}
}

func TestBookingNotification(t *testing.T) {
	notification := BookingNotification{
		SpecVersion:          "1.0",
		ID:                   "3cecb101-7a1a-43a4-9d62-e88a131651e2",
		Source:               "https://carrier.com/",
		Type:                 "org.dcsa.booking.v2",
		Time:                 time.Now(),
		DataContentType:      "application/json",
		SubscriptionReference: "30675492-50ff-4e17-a7df-7a487a8ad343",
	}

	data, err := json.Marshal(notification)
	if err != nil {
		t.Fatalf("failed to marshal BookingNotification: %v", err)
	}

	var unmarshaled BookingNotification
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal BookingNotification: %v", err)
	}

	if unmarshaled.SpecVersion != notification.SpecVersion {
		t.Errorf("expected SpecVersion %s, got %s", notification.SpecVersion, unmarshaled.SpecVersion)
	}
}
