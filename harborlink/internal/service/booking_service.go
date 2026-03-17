package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/yourname/harborlink/internal/adapter"
	"github.com/yourname/harborlink/internal/model"
	"github.com/yourname/harborlink/internal/repository"
	"github.com/yourname/harborlink/pkg/cache"
)

// Common service errors
var (
	ErrBookingNotFound      = errors.New("booking not found")
	ErrInvalidCarrier       = errors.New("invalid carrier")
	ErrBookingAlreadyExists = errors.New("booking already exists")
	ErrInvalidStatus        = errors.New("invalid booking status for this operation")
)

// BookingService handles business logic for booking operations
type BookingService struct {
	repo   repository.BookingRepository
	cache  *cache.BookingCache
	router *CarrierRouter
}

// NewBookingService creates a new booking service
func NewBookingService(
	repo repository.BookingRepository,
	cache *cache.BookingCache,
	router *CarrierRouter,
) *BookingService {
	return &BookingService{
		repo:   repo,
		cache:  cache,
		router: router,
	}
}

// CreateBooking creates a new booking request
func (s *BookingService) CreateBooking(ctx context.Context, req *model.CreateBooking) (*model.Booking, error) {
	// Determine carrier code from request (from document parties or default)
	carrierCode := s.extractCarrierCode(req)
	if carrierCode == "" {
		return nil, ErrInvalidCarrier
	}

	// Get carrier adapter
	carrierAdapter, err := s.router.Route(carrierCode)
	if err != nil {
		return nil, fmt.Errorf("failed to route to carrier: %w", err)
	}

	// Submit booking to carrier
	response, err := carrierAdapter.CreateBooking(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("carrier rejected booking: %w", err)
	}

	// Generate reference if not provided
	reference := response.CarrierBookingRequestReference
	if reference == "" {
		reference = generateReference()
	}

	// Create booking record for database
	record := s.createBookingRecord(reference, carrierCode, req)

	// Store in database
	if err := s.repo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to save booking: %w", err)
	}

	// Cache the booking
	if s.cache != nil {
		booking := s.recordToBooking(record)
		_ = s.cache.Set(ctx, reference, booking, cache.DefaultBookingTTL)
		_ = s.cache.SetStatus(ctx, reference, string(record.BookingStatus), cache.DefaultStatusTTL)
	}

	// Return the booking
	return s.recordToBooking(record), nil
}

// GetBooking retrieves a booking by reference
func (s *BookingService) GetBooking(ctx context.Context, reference string) (*model.Booking, error) {
	// Try cache first
	if s.cache != nil {
		var booking model.Booking
		if err := s.cache.Get(ctx, reference, &booking); err == nil {
			return &booking, nil
		}
	}

	// Get from database
	record, err := s.repo.GetByReference(ctx, reference)
	if err != nil {
		if errors.Is(err, repository.ErrBookingNotFound) {
			return nil, ErrBookingNotFound
		}
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}

	booking := s.recordToBooking(record)

	// Cache for future requests
	if s.cache != nil {
		_ = s.cache.Set(ctx, reference, booking, cache.DefaultBookingTTL)
	}

	return booking, nil
}

// UpdateBooking updates an existing booking
func (s *BookingService) UpdateBooking(ctx context.Context, reference string, req *model.UpdateBooking) (*model.Booking, error) {
	// Get existing booking
	record, err := s.repo.GetByReference(ctx, reference)
	if err != nil {
		if errors.Is(err, repository.ErrBookingNotFound) {
			return nil, ErrBookingNotFound
		}
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}

	// Check if booking can be updated
	if !s.canUpdate(record.BookingStatus) {
		return nil, ErrInvalidStatus
	}

	// Get carrier adapter
	carrierAdapter, err := s.router.Route(record.CarrierCode)
	if err != nil {
		return nil, fmt.Errorf("failed to route to carrier: %w", err)
	}

	// Update via carrier adapter
	updatedBooking, err := carrierAdapter.UpdateBooking(ctx, reference, req)
	if err != nil {
		return nil, fmt.Errorf("carrier rejected update: %w", err)
	}

	// Update status to pending
	record.BookingStatus = model.BookingStatusPendingUpdate
	if err := s.repo.UpdateStatus(ctx, reference, record.BookingStatus); err != nil {
		return nil, fmt.Errorf("failed to update status: %w", err)
	}

	// Update cache
	if s.cache != nil {
		_ = s.cache.Delete(ctx, reference)
	}

	// Return updated booking (merge with carrier response if available)
	if updatedBooking != nil {
		return updatedBooking, nil
	}

	return s.recordToBooking(record), nil
}

// CancelBooking cancels a booking
func (s *BookingService) CancelBooking(ctx context.Context, reference string, req *model.CancelBookingRequest) error {
	// Get existing booking
	record, err := s.repo.GetByReference(ctx, reference)
	if err != nil {
		if errors.Is(err, repository.ErrBookingNotFound) {
			return ErrBookingNotFound
		}
		return fmt.Errorf("failed to get booking: %w", err)
	}

	// Check if booking can be cancelled
	if !s.canCancel(record.BookingStatus) {
		return ErrInvalidStatus
	}

	// Get carrier adapter
	carrierAdapter, err := s.router.Route(record.CarrierCode)
	if err != nil {
		return fmt.Errorf("failed to route to carrier: %w", err)
	}

	// Set default cancellation status if not provided
	if req == nil {
		req = &model.CancelBookingRequest{}
	}
	if req.BookingCancellationStatus == "" {
		req.BookingCancellationStatus = model.CancellationStatusReceived
	}

	// Cancel via carrier adapter
	if err := carrierAdapter.CancelBooking(ctx, reference, req); err != nil {
		return fmt.Errorf("carrier rejected cancellation: %w", err)
	}

	// Update status
	record.BookingStatus = model.BookingStatusCancelled
	if err := s.repo.UpdateStatus(ctx, reference, record.BookingStatus); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	// Invalidate cache
	if s.cache != nil {
		_ = s.cache.Delete(ctx, reference)
		_ = s.cache.SetStatus(ctx, reference, string(record.BookingStatus), cache.DefaultStatusTTL)
	}

	return nil
}

// ListBookings lists bookings with filters
func (s *BookingService) ListBookings(ctx context.Context, filter *repository.BookingFilter) ([]model.Booking, int64, error) {
	records, total, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list bookings: %w", err)
	}

	bookings := make([]model.Booking, len(records))
	for i, record := range records {
		bookings[i] = *s.recordToBooking(&record)
	}

	return bookings, total, nil
}

// extractCarrierCode extracts carrier code from the request
func (s *BookingService) extractCarrierCode(req *model.CreateBooking) string {
	// Default to first enabled carrier
	carriers := s.router.GetEnabledCarriers()
	if len(carriers) > 0 {
		return carriers[0].GetCarrierCode()
	}

	return ""
}

// createBookingRecord creates a database record from a create request
func (s *BookingService) createBookingRecord(reference, carrierCode string, req *model.CreateBooking) *model.BookingRecord {
	record := &model.BookingRecord{
		CarrierBookingRequestReference: reference,
		CarrierCode:                    carrierCode,
		BookingStatus:                  model.BookingStatusReceived,
		ReceiptTypeAtOrigin:            req.ReceiptTypeAtOrigin,
		DeliveryTypeAtDestination:      req.DeliveryTypeAtDestination,
		CargoMovementTypeAtOrigin:      req.CargoMovementTypeAtOrigin,
		CargoMovementTypeAtDestination: req.CargoMovementTypeAtDestination,
		ServiceContractReference:       req.ServiceContractReference,
		ContractQuotationReference:     req.ContractQuotationReference,
		CarrierServiceCode:             req.CarrierServiceCode,
		CarrierExportVoyageNumber:      req.CarrierExportVoyageNumber,
		ExpectedDepartureDate:          req.ExpectedDepartureDate,
	}

	// Extract vessel info
	if req.Vessel != nil {
		record.VesselName = req.Vessel.Name
		record.VesselIMONumber = req.Vessel.VesselIMONumber
	}

	// Extract location info
	for _, loc := range req.ShipmentLocations {
		switch loc.LocationTypeCode {
		case model.LocationTypePol:
			record.POLUNLocationCode = loc.Location.UNLocationCode
		case model.LocationTypePod:
			record.PODUNLocationCode = loc.Location.UNLocationCode
		case model.LocationTypePre:
			record.PlaceOfReceiptUNLocationCode = loc.Location.UNLocationCode
		case model.LocationTypePde:
			record.PlaceOfDeliveryUNLocationCode = loc.Location.UNLocationCode
		}
	}

	// Store complex nested data as JSON
	if len(req.RequestedEquipments) > 0 {
		data, _ := json.Marshal(req.RequestedEquipments)
		record.RequestedEquipmentJSON = data
	}

	if req.DocumentParties != nil {
		data, _ := json.Marshal(req.DocumentParties)
		record.DocumentPartiesJSON = data
	}

	if len(req.ShipmentLocations) > 0 {
		data, _ := json.Marshal(req.ShipmentLocations)
		record.ShipmentLocationsJSON = data
	}

	return record
}

// recordToBooking converts a database record to a Booking model
func (s *BookingService) recordToBooking(record *model.BookingRecord) *model.Booking {
	booking := &model.Booking{
		CarrierBookingRequestReference: record.CarrierBookingRequestReference,
		CarrierBookingReference:        record.CarrierBookingReference,
		BookingStatus:                  record.BookingStatus,
		CarrierCode:                    record.CarrierCode,
		ReceiptTypeAtOrigin:            record.ReceiptTypeAtOrigin,
		DeliveryTypeAtDestination:      record.DeliveryTypeAtDestination,
		CargoMovementTypeAtOrigin:      record.CargoMovementTypeAtOrigin,
		CargoMovementTypeAtDestination: record.CargoMovementTypeAtDestination,
		ServiceContractReference:       record.ServiceContractReference,
		ContractQuotationReference:     record.ContractQuotationReference,
		CarrierServiceCode:             record.CarrierServiceCode,
		CarrierExportVoyageNumber:      record.CarrierExportVoyageNumber,
		ExpectedDepartureDate:          record.ExpectedDepartureDate,
		IsEquipmentSubstitutionAllowed: true,
	}

	// Add vessel info
	if record.VesselName != "" || record.VesselIMONumber != "" {
		booking.Vessel = &model.Vessel{
			Name:            record.VesselName,
			VesselIMONumber: record.VesselIMONumber,
		}
	}

	// Unmarshal nested JSON data
	if len(record.RequestedEquipmentJSON) > 0 {
		var equipments []model.RequestedEquipmentShipper
		if err := json.Unmarshal(record.RequestedEquipmentJSON, &equipments); err == nil {
			booking.RequestedEquipments = make([]model.RequestedEquipment, len(equipments))
			for i, eq := range equipments {
				booking.RequestedEquipments[i] = model.RequestedEquipment{
					ISOEquipmentCode: eq.ISOEquipmentCode,
					Units:            eq.Units,
				}
			}
		}
	}

	if len(record.DocumentPartiesJSON) > 0 {
		var parties model.DocumentPartiesReq
		if err := json.Unmarshal(record.DocumentPartiesJSON, &parties); err == nil {
			booking.DocumentParties = &model.DocumentParties{
				Shipper:   parties.Shipper,
				Consignee: parties.Consignee,
			}
		}
	}

	if len(record.ShipmentLocationsJSON) > 0 {
		_ = json.Unmarshal(record.ShipmentLocationsJSON, &booking.ShipmentLocations)
	}

	return booking
}

// canUpdate checks if a booking can be updated based on status
func (s *BookingService) canUpdate(status model.BookingStatus) bool {
	switch status {
	case model.BookingStatusReceived,
		model.BookingStatusPendingUpdate,
		model.BookingStatusConfirmed:
		return true
	default:
		return false
	}
}

// canCancel checks if a booking can be cancelled based on status
func (s *BookingService) canCancel(status model.BookingStatus) bool {
	switch status {
	case model.BookingStatusReceived,
		model.BookingStatusPendingUpdate,
		model.BookingStatusConfirmed,
		model.BookingStatusPendingAmendment:
		return true
	default:
		return false
	}
}

// UpdateBookingStatus updates the booking status (called by webhook or polling)
func (s *BookingService) UpdateBookingStatus(ctx context.Context, reference string, status model.BookingStatus) error {
	if err := s.repo.UpdateStatus(ctx, reference, status); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	// Invalidate cache
	if s.cache != nil {
		_ = s.cache.Delete(ctx, reference)
		_ = s.cache.SetStatus(ctx, reference, string(status), cache.DefaultStatusTTL)
	}

	return nil
}

// SyncBookingFromCarrier fetches latest booking data from carrier
func (s *BookingService) SyncBookingFromCarrier(ctx context.Context, reference string) (*model.Booking, error) {
	// Get local record first
	record, err := s.repo.GetByReference(ctx, reference)
	if err != nil {
		if errors.Is(err, repository.ErrBookingNotFound) {
			return nil, ErrBookingNotFound
		}
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}

	// Get carrier adapter
	carrierAdapter, err := s.router.Route(record.CarrierCode)
	if err != nil {
		return nil, fmt.Errorf("failed to route to carrier: %w", err)
	}

	// Fetch from carrier
	booking, err := carrierAdapter.GetBooking(ctx, reference)
	if err != nil {
		// If it's an adapter error, wrap it appropriately
		if adapter.IsAdapterError(err) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to fetch from carrier: %w", err)
	}

	// Update cache
	if s.cache != nil {
		_ = s.cache.Set(ctx, reference, booking, cache.DefaultBookingTTL)
		if booking.BookingStatus != "" {
			_ = s.cache.SetStatus(ctx, reference, string(booking.BookingStatus), cache.DefaultStatusTTL)
		}
	}

	return booking, nil
}

// GetBookingStatus gets just the status of a booking (fast cache lookup)
func (s *BookingService) GetBookingStatus(ctx context.Context, reference string) (model.BookingStatus, error) {
	// Try cache first
	if s.cache != nil {
		status, err := s.cache.GetStatus(ctx, reference)
		if err == nil && status != "" {
			return model.BookingStatus(status), nil
		}
	}

	// Fall back to full booking lookup
	booking, err := s.GetBooking(ctx, reference)
	if err != nil {
		return "", err
	}

	return booking.BookingStatus, nil
}

// HealthCheck checks if the service is healthy
func (s *BookingService) HealthCheck(ctx context.Context) error {
	// Could add database ping here
	return nil
}

// GetCarrierAdapter returns the carrier adapter for a given carrier code
func (s *BookingService) GetCarrierAdapter(carrierCode string) (adapter.CarrierAdapter, error) {
	return s.router.Route(carrierCode)
}

// Now returns current time (useful for testing)
var now = func() time.Time {
	return time.Now().UTC()
}

// generateReference generates a unique booking reference
func generateReference() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return "BK" + hex.EncodeToString(bytes)
}
