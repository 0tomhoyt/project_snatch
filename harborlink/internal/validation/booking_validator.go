package validation

import (
	"errors"
	"fmt"
	"time"

	"github.com/yourname/harborlink/internal/model"
)

// Validation errors
var (
	ErrMissingShipmentLocations = errors.New("shipmentLocations is required")
	ErrMissingDocumentParties   = errors.New("documentParties is required")
	ErrMissingEquipment         = errors.New("requestedEquipments is required")
	ErrInvalidLocationType      = errors.New("invalid location type code")
	ErrInvalidDateRange         = errors.New("invalid date range")
	ErrInvalidEquipmentUnits    = errors.New("equipment units must be positive")
)

// ValidationError represents a validation error with details
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []*ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "validation failed"
	}
	return fmt.Sprintf("validation failed: %s", e[0].Message)
}

// BookingValidator validates booking requests
type BookingValidator struct{}

// NewBookingValidator creates a new booking validator
func NewBookingValidator() *BookingValidator {
	return &BookingValidator{}
}

// ValidateCreateBooking validates a booking creation request
func (v *BookingValidator) ValidateCreateBooking(req *model.CreateBooking) ValidationErrors {
	var errs ValidationErrors

	// Validate shipment locations
	if len(req.ShipmentLocations) == 0 {
		errs = append(errs, &ValidationError{
			Field:   "shipmentLocations",
			Message: ErrMissingShipmentLocations.Error(),
		})
	} else {
		if locErrs := v.validateShipmentLocations(req.ShipmentLocations); len(locErrs) > 0 {
			errs = append(errs, locErrs...)
		}
	}

	// Validate document parties
	if req.DocumentParties == nil {
		errs = append(errs, &ValidationError{
			Field:   "documentParties",
			Message: ErrMissingDocumentParties.Error(),
		})
	}

	// Validate requested equipments
	if len(req.RequestedEquipments) == 0 {
		errs = append(errs, &ValidationError{
			Field:   "requestedEquipments",
			Message: ErrMissingEquipment.Error(),
		})
	} else {
		if eqErrs := v.validateRequestedEquipments(req.RequestedEquipments); len(eqErrs) > 0 {
			errs = append(errs, eqErrs...)
		}
	}

	// Validate dates
	if req.ExpectedDepartureDate != nil {
		if req.ExpectedDepartureDate.Before(time.Now().Add(-24 * time.Hour)) {
			errs = append(errs, &ValidationError{
				Field:   "expectedDepartureDate",
				Message: "departure date cannot be in the past",
			})
		}
	}

	return errs
}

// ValidateUpdateBooking validates a booking update request
func (v *BookingValidator) ValidateUpdateBooking(req *model.UpdateBooking) ValidationErrors {
	var errs ValidationErrors

	if req.Booking == nil {
		errs = append(errs, &ValidationError{
			Field:   "booking",
			Message: "booking is required for update",
		})
		return errs
	}

	// Validate shipment locations if provided
	if len(req.Booking.ShipmentLocations) > 0 {
		if locErrs := v.validateShipmentLocations(req.Booking.ShipmentLocations); len(locErrs) > 0 {
			errs = append(errs, locErrs...)
		}
	}

	// Validate requested equipments if provided
	if len(req.Booking.RequestedEquipments) > 0 {
		if eqErrs := v.validateRequestedEquipmentsFull(req.Booking.RequestedEquipments); len(eqErrs) > 0 {
			errs = append(errs, eqErrs...)
		}
	}

	// Validate dates
	if req.Booking.ExpectedDepartureDate != nil {
		if req.Booking.ExpectedDepartureDate.Before(time.Now().Add(-24 * time.Hour)) {
			errs = append(errs, &ValidationError{
				Field:   "expectedDepartureDate",
				Message: "departure date cannot be in the past",
			})
		}
	}

	return errs
}

// ValidateCancelBooking validates a booking cancellation request
func (v *BookingValidator) ValidateCancelBooking(req *model.CancelBookingRequest) ValidationErrors {
	var errs ValidationErrors

	// Reason is optional but if provided should not be empty
	if req != nil && req.Reason == "" && req.BookingCancellationStatus != "" {
		// If status is provided without reason, that's fine
	}

	return errs
}

// validateShipmentLocations validates shipment locations
func (v *BookingValidator) validateShipmentLocations(locations []model.ShipmentLocation) ValidationErrors {
	var errs ValidationErrors

	validLocationTypes := map[model.LocationTypeCode]bool{
		model.LocationTypePre: true,
		model.LocationTypePol: true,
		model.LocationTypePod: true,
		model.LocationTypePde: true,
		model.LocationTypePcf: true,
		model.LocationTypeOir: true,
		model.LocationTypeOri: true,
	}

	hasPOL := false
	hasPOD := false

	for i, loc := range locations {
		// Validate location type
		if !validLocationTypes[loc.LocationTypeCode] {
			errs = append(errs, &ValidationError{
				Field:   fmt.Sprintf("shipmentLocations[%d].locationTypeCode", i),
				Message: fmt.Sprintf("invalid location type: %s", loc.LocationTypeCode),
			})
		}

		// Check for required POL and POD
		if loc.LocationTypeCode == model.LocationTypePol {
			hasPOL = true
		}
		if loc.LocationTypeCode == model.LocationTypePod {
			hasPOD = true
		}

		// Validate UNLocationCode if provided
		if loc.Location.UNLocationCode != "" && len(loc.Location.UNLocationCode) != 5 {
			errs = append(errs, &ValidationError{
				Field:   fmt.Sprintf("shipmentLocations[%d].location.UNLocationCode", i),
				Message: "UNLocationCode must be 5 characters",
			})
		}
	}

	if !hasPOL {
		errs = append(errs, &ValidationError{
			Field:   "shipmentLocations",
			Message: "POL (Port of Loading) location is required",
		})
	}
	if !hasPOD {
		errs = append(errs, &ValidationError{
			Field:   "shipmentLocations",
			Message: "POD (Port of Discharge) location is required",
		})
	}

	return errs
}

// validateRequestedEquipments validates requested equipments (shipper version)
func (v *BookingValidator) validateRequestedEquipments(equipments []model.RequestedEquipmentShipper) ValidationErrors {
	var errs ValidationErrors

	for i, eq := range equipments {
		if eq.ISOEquipmentCode == "" {
			errs = append(errs, &ValidationError{
				Field:   fmt.Sprintf("requestedEquipments[%d].ISOEquipmentCode", i),
				Message: "ISOEquipmentCode is required",
			})
		}
		if eq.Units <= 0 {
			errs = append(errs, &ValidationError{
				Field:   fmt.Sprintf("requestedEquipments[%d].units", i),
				Message: ErrInvalidEquipmentUnits.Error(),
			})
		}
	}

	return errs
}

// validateRequestedEquipmentsFull validates requested equipments (full version)
func (v *BookingValidator) validateRequestedEquipmentsFull(equipments []model.RequestedEquipment) ValidationErrors {
	var errs ValidationErrors

	for i, eq := range equipments {
		if eq.ISOEquipmentCode == "" {
			errs = append(errs, &ValidationError{
				Field:   fmt.Sprintf("requestedEquipments[%d].ISOEquipmentCode", i),
				Message: "ISOEquipmentCode is required",
			})
		}
		if eq.Units <= 0 {
			errs = append(errs, &ValidationError{
				Field:   fmt.Sprintf("requestedEquipments[%d].units", i),
				Message: ErrInvalidEquipmentUnits.Error(),
			})
		}
	}

	return errs
}

// ValidateStatusTransition validates if a status transition is valid
func (v *BookingValidator) ValidateStatusTransition(currentStatus, newStatus model.BookingStatus) error {
	validTransitions := map[model.BookingStatus][]model.BookingStatus{
		model.BookingStatusReceived: {
			model.BookingStatusPendingUpdate,
			model.BookingStatusConfirmed,
			model.BookingStatusRejected,
			model.BookingStatusCancelled,
		},
		model.BookingStatusPendingUpdate: {
			model.BookingStatusConfirmed,
			model.BookingStatusRejected,
			model.BookingStatusCancelled,
		},
		model.BookingStatusConfirmed: {
			model.BookingStatusPendingAmendment,
			model.BookingStatusCompleted,
			model.BookingStatusDeclined,
			model.BookingStatusCancelled,
		},
		model.BookingStatusPendingAmendment: {
			model.BookingStatusConfirmed,
			model.BookingStatusDeclined,
			model.BookingStatusCancelled,
		},
	}

	allowedStatuses, ok := validTransitions[currentStatus]
	if !ok {
		return fmt.Errorf("invalid current status: %s", currentStatus)
	}

	for _, allowed := range allowedStatuses {
		if newStatus == allowed {
			return nil
		}
	}

	return fmt.Errorf("invalid status transition from %s to %s", currentStatus, newStatus)
}
