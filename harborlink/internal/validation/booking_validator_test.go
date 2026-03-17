package validation

import (
	"testing"
	"time"

	"github.com/yourname/harborlink/internal/model"
)

func TestBookingValidator_ValidateCreateBooking(t *testing.T) {
	validator := NewBookingValidator()

	tests := []struct {
		name    string
		req     *model.CreateBooking
		wantErr bool
	}{
		{
			name: "valid booking",
			req: &model.CreateBooking{
				ReceiptTypeAtOrigin:            model.ReceiptDeliveryTypeCY,
				DeliveryTypeAtDestination:      model.ReceiptDeliveryTypeCY,
				CargoMovementTypeAtOrigin:      model.CargoMovementTypeFCL,
				CargoMovementTypeAtDestination: model.CargoMovementTypeFCL,
				IsEquipmentSubstitutionAllowed: true,
				ShipmentLocations: []model.ShipmentLocation{
					{
						LocationTypeCode: model.LocationTypePol,
						Location:         model.Location{UNLocationCode: "CNSHA"},
					},
					{
						LocationTypeCode: model.LocationTypePod,
						Location:         model.Location{UNLocationCode: "USLAX"},
					},
				},
				DocumentParties: &model.DocumentPartiesReq{},
				RequestedEquipments: []model.RequestedEquipmentShipper{
					{
						ISOEquipmentCode: "22G1",
						Units:            1,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing shipment locations",
			req: &model.CreateBooking{
				ReceiptTypeAtOrigin:            model.ReceiptDeliveryTypeCY,
				DeliveryTypeAtDestination:      model.ReceiptDeliveryTypeCY,
				CargoMovementTypeAtOrigin:      model.CargoMovementTypeFCL,
				CargoMovementTypeAtDestination: model.CargoMovementTypeFCL,
				IsEquipmentSubstitutionAllowed: true,
				ShipmentLocations:              nil,
				DocumentParties:                &model.DocumentPartiesReq{},
				RequestedEquipments: []model.RequestedEquipmentShipper{
					{ISOEquipmentCode: "22G1", Units: 1},
				},
			},
			wantErr: true,
		},
		{
			name: "missing document parties",
			req: &model.CreateBooking{
				ReceiptTypeAtOrigin:            model.ReceiptDeliveryTypeCY,
				DeliveryTypeAtDestination:      model.ReceiptDeliveryTypeCY,
				CargoMovementTypeAtOrigin:      model.CargoMovementTypeFCL,
				CargoMovementTypeAtDestination: model.CargoMovementTypeFCL,
				IsEquipmentSubstitutionAllowed: true,
				ShipmentLocations: []model.ShipmentLocation{
					{
						LocationTypeCode: model.LocationTypePol,
						Location:         model.Location{UNLocationCode: "CNSHA"},
					},
					{
						LocationTypeCode: model.LocationTypePod,
						Location:         model.Location{UNLocationCode: "USLAX"},
					},
				},
				DocumentParties:    nil,
				RequestedEquipments: []model.RequestedEquipmentShipper{
					{ISOEquipmentCode: "22G1", Units: 1},
				},
			},
			wantErr: true,
		},
		{
			name: "missing equipment",
			req: &model.CreateBooking{
				ReceiptTypeAtOrigin:            model.ReceiptDeliveryTypeCY,
				DeliveryTypeAtDestination:      model.ReceiptDeliveryTypeCY,
				CargoMovementTypeAtOrigin:      model.CargoMovementTypeFCL,
				CargoMovementTypeAtDestination: model.CargoMovementTypeFCL,
				IsEquipmentSubstitutionAllowed: true,
				ShipmentLocations: []model.ShipmentLocation{
					{
						LocationTypeCode: model.LocationTypePol,
						Location:         model.Location{UNLocationCode: "CNSHA"},
					},
					{
						LocationTypeCode: model.LocationTypePod,
						Location:         model.Location{UNLocationCode: "USLAX"},
					},
				},
				DocumentParties:      &model.DocumentPartiesReq{},
				RequestedEquipments: nil,
			},
			wantErr: true,
		},
		{
			name: "missing POL",
			req: &model.CreateBooking{
				ReceiptTypeAtOrigin:            model.ReceiptDeliveryTypeCY,
				DeliveryTypeAtDestination:      model.ReceiptDeliveryTypeCY,
				CargoMovementTypeAtOrigin:      model.CargoMovementTypeFCL,
				CargoMovementTypeAtDestination: model.CargoMovementTypeFCL,
				IsEquipmentSubstitutionAllowed: true,
				ShipmentLocations: []model.ShipmentLocation{
					{
						LocationTypeCode: model.LocationTypePod,
						Location:         model.Location{UNLocationCode: "USLAX"},
					},
				},
				DocumentParties: &model.DocumentPartiesReq{},
				RequestedEquipments: []model.RequestedEquipmentShipper{
					{ISOEquipmentCode: "22G1", Units: 1},
				},
			},
			wantErr: true,
		},
		{
			name: "missing POD",
			req: &model.CreateBooking{
				ReceiptTypeAtOrigin:            model.ReceiptDeliveryTypeCY,
				DeliveryTypeAtDestination:      model.ReceiptDeliveryTypeCY,
				CargoMovementTypeAtOrigin:      model.CargoMovementTypeFCL,
				CargoMovementTypeAtDestination: model.CargoMovementTypeFCL,
				IsEquipmentSubstitutionAllowed: true,
				ShipmentLocations: []model.ShipmentLocation{
					{
						LocationTypeCode: model.LocationTypePol,
						Location:         model.Location{UNLocationCode: "CNSHA"},
					},
				},
				DocumentParties: &model.DocumentPartiesReq{},
				RequestedEquipments: []model.RequestedEquipmentShipper{
					{ISOEquipmentCode: "22G1", Units: 1},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid equipment units",
			req: &model.CreateBooking{
				ReceiptTypeAtOrigin:            model.ReceiptDeliveryTypeCY,
				DeliveryTypeAtDestination:      model.ReceiptDeliveryTypeCY,
				CargoMovementTypeAtOrigin:      model.CargoMovementTypeFCL,
				CargoMovementTypeAtDestination: model.CargoMovementTypeFCL,
				IsEquipmentSubstitutionAllowed: true,
				ShipmentLocations: []model.ShipmentLocation{
					{
						LocationTypeCode: model.LocationTypePol,
						Location:         model.Location{UNLocationCode: "CNSHA"},
					},
					{
						LocationTypeCode: model.LocationTypePod,
						Location:         model.Location{UNLocationCode: "USLAX"},
					},
				},
				DocumentParties: &model.DocumentPartiesReq{},
				RequestedEquipments: []model.RequestedEquipmentShipper{
					{ISOEquipmentCode: "22G1", Units: 0},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validator.ValidateCreateBooking(tt.req)
			if tt.wantErr && len(errs) == 0 {
				t.Error("expected validation errors, got none")
			}
			if !tt.wantErr && len(errs) > 0 {
				t.Errorf("expected no validation errors, got: %v", errs)
			}
		})
	}
}

func TestBookingValidator_ValidateUpdateBooking(t *testing.T) {
	validator := NewBookingValidator()

	tests := []struct {
		name    string
		req     *model.UpdateBooking
		wantErr bool
	}{
		{
			name: "nil booking",
			req: &model.UpdateBooking{
				Booking: nil,
			},
			wantErr: true,
		},
		{
			name: "valid update",
			req: &model.UpdateBooking{
				Booking: &model.Booking{
					ShipmentLocations: []model.ShipmentLocation{
						{
							LocationTypeCode: model.LocationTypePol,
							Location:         model.Location{UNLocationCode: "CNSHA"},
						},
						{
							LocationTypeCode: model.LocationTypePod,
							Location:         model.Location{UNLocationCode: "USLAX"},
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validator.ValidateUpdateBooking(tt.req)
			if tt.wantErr && len(errs) == 0 {
				t.Error("expected validation errors, got none")
			}
			if !tt.wantErr && len(errs) > 0 {
				t.Errorf("expected no validation errors, got: %v", errs)
			}
		})
	}
}

func TestBookingValidator_ValidateStatusTransition(t *testing.T) {
	validator := NewBookingValidator()

	tests := []struct {
		current   model.BookingStatus
		new       model.BookingStatus
		wantError bool
	}{
		{model.BookingStatusReceived, model.BookingStatusConfirmed, false},
		{model.BookingStatusReceived, model.BookingStatusCancelled, false},
		{model.BookingStatusReceived, model.BookingStatusCompleted, true},
		{model.BookingStatusConfirmed, model.BookingStatusCompleted, false},
		{model.BookingStatusConfirmed, model.BookingStatusCancelled, false},
		{model.BookingStatusCancelled, model.BookingStatusConfirmed, true},
		{model.BookingStatusCompleted, model.BookingStatusCancelled, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.current)+"->"+string(tt.new), func(t *testing.T) {
			err := validator.ValidateStatusTransition(tt.current, tt.new)
			if tt.wantError && err == nil {
				t.Error("expected error for invalid transition")
			}
			if !tt.wantError && err != nil {
				t.Errorf("expected no error for valid transition, got: %v", err)
			}
		})
	}
}

func TestBookingValidator_ValidateCancelBooking(t *testing.T) {
	validator := NewBookingValidator()

	// Nil request should be valid
	errs := validator.ValidateCancelBooking(nil)
	if len(errs) > 0 {
		t.Errorf("expected no errors for nil request, got: %v", errs)
	}

	// Empty request should be valid
	errs = validator.ValidateCancelBooking(&model.CancelBookingRequest{})
	if len(errs) > 0 {
		t.Errorf("expected no errors for empty request, got: %v", errs)
	}

	// Request with reason should be valid
	errs = validator.ValidateCancelBooking(&model.CancelBookingRequest{
		Reason: "test reason",
	})
	if len(errs) > 0 {
		t.Errorf("expected no errors for request with reason, got: %v", errs)
	}
}

func TestValidationError_Error(t *testing.T) {
	err := &ValidationError{
		Field:   "testField",
		Message: "test message",
	}

	if err.Error() != "testField: test message" {
		t.Errorf("expected 'testField: test message', got: %s", err.Error())
	}
}

func TestValidationErrors_Error(t *testing.T) {
	errs := ValidationErrors{
		{Field: "field1", Message: "error1"},
	}

	if errs.Error() != "validation failed: error1" {
		t.Errorf("expected 'validation failed: error1', got: %s", errs.Error())
	}

	// Empty errors
	emptyErrs := ValidationErrors{}
	if emptyErrs.Error() != "validation failed" {
		t.Errorf("expected 'validation failed', got: %s", emptyErrs.Error())
	}
}

func TestBookingValidator_DateValidation(t *testing.T) {
	validator := NewBookingValidator()

	// Past date should cause error
	pastDate := time.Now().Add(-48 * time.Hour)
	errs := validator.ValidateCreateBooking(&model.CreateBooking{
		ShipmentLocations: []model.ShipmentLocation{
			{LocationTypeCode: model.LocationTypePol, Location: model.Location{UNLocationCode: "CNSHA"}},
			{LocationTypeCode: model.LocationTypePod, Location: model.Location{UNLocationCode: "USLAX"}},
		},
		DocumentParties:      &model.DocumentPartiesReq{},
		RequestedEquipments: []model.RequestedEquipmentShipper{{ISOEquipmentCode: "22G1", Units: 1}},
		ExpectedDepartureDate: &pastDate,
	})

	if len(errs) == 0 {
		t.Error("expected validation error for past date")
	}

	// Future date should be valid
	futureDate := time.Now().Add(48 * time.Hour)
	errs = validator.ValidateCreateBooking(&model.CreateBooking{
		ShipmentLocations: []model.ShipmentLocation{
			{LocationTypeCode: model.LocationTypePol, Location: model.Location{UNLocationCode: "CNSHA"}},
			{LocationTypeCode: model.LocationTypePod, Location: model.Location{UNLocationCode: "USLAX"}},
		},
		DocumentParties:      &model.DocumentPartiesReq{},
		RequestedEquipments: []model.RequestedEquipmentShipper{{ISOEquipmentCode: "22G1", Units: 1}},
		ExpectedDepartureDate: &futureDate,
	})

	if len(errs) > 0 {
		t.Errorf("expected no validation errors for future date, got: %v", errs)
	}
}
