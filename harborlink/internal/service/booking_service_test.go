package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yourname/harborlink/internal/model"
	"github.com/yourname/harborlink/internal/repository"
)

// mockBookingRepository implements repository.BookingRepository for testing
type mockBookingRepository struct {
	bookings map[string]*model.BookingRecord
}

func newMockBookingRepository() *mockBookingRepository {
	return &mockBookingRepository{
		bookings: make(map[string]*model.BookingRecord),
	}
}

func (m *mockBookingRepository) Create(ctx context.Context, booking *model.BookingRecord) error {
	m.bookings[booking.CarrierBookingRequestReference] = booking
	return nil
}

func (m *mockBookingRepository) GetByReference(ctx context.Context, reference string) (*model.BookingRecord, error) {
	if booking, ok := m.bookings[reference]; ok {
		return booking, nil
	}
	return nil, repository.ErrBookingNotFound
}

func (m *mockBookingRepository) GetByCarrierReference(ctx context.Context, carrierRef string) (*model.BookingRecord, error) {
	for _, booking := range m.bookings {
		if booking.CarrierBookingReference == carrierRef {
			return booking, nil
		}
	}
	return nil, repository.ErrBookingNotFound
}

func (m *mockBookingRepository) Update(ctx context.Context, booking *model.BookingRecord) error {
	if _, ok := m.bookings[booking.CarrierBookingRequestReference]; ok {
		m.bookings[booking.CarrierBookingRequestReference] = booking
		return nil
	}
	return repository.ErrBookingNotFound
}

func (m *mockBookingRepository) UpdateStatus(ctx context.Context, reference string, status model.BookingStatus) error {
	if booking, ok := m.bookings[reference]; ok {
		booking.BookingStatus = status
		return nil
	}
	return repository.ErrBookingNotFound
}

func (m *mockBookingRepository) Delete(ctx context.Context, reference string) error {
	delete(m.bookings, reference)
	return nil
}

func (m *mockBookingRepository) List(ctx context.Context, filter *repository.BookingFilter) ([]model.BookingRecord, int64, error) {
	var results []model.BookingRecord
	for _, b := range m.bookings {
		results = append(results, *b)
	}
	return results, int64(len(results)), nil
}

func TestBookingService_GetBooking(t *testing.T) {
	ctx := context.Background()
	repo := newMockBookingRepository()

	// Pre-populate a booking
	repo.bookings["test-ref-123"] = &model.BookingRecord{
		CarrierBookingRequestReference: "test-ref-123",
		BookingStatus:                  model.BookingStatusReceived,
		CarrierCode:                    "MAEU",
	}

	svc := &BookingService{
		repo:  repo,
		cache: nil,
	}

	// Test getting existing booking
	booking, err := svc.GetBooking(ctx, "test-ref-123")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if booking == nil {
		t.Error("expected booking to be returned")
	}

	// Test getting non-existent booking
	_, err = svc.GetBooking(ctx, "non-existent")
	if !errors.Is(err, ErrBookingNotFound) {
		t.Errorf("expected ErrBookingNotFound, got: %v", err)
	}
}

func TestBookingService_UpdateBookingStatus(t *testing.T) {
	ctx := context.Background()
	repo := newMockBookingRepository()

	// Pre-populate a booking
	repo.bookings["test-ref-123"] = &model.BookingRecord{
		CarrierBookingRequestReference: "test-ref-123",
		BookingStatus:                  model.BookingStatusReceived,
		CarrierCode:                    "MAEU",
	}

	svc := &BookingService{
		repo:  repo,
		cache: nil,
	}

	// Test updating status
	err := svc.UpdateBookingStatus(ctx, "test-ref-123", model.BookingStatusConfirmed)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	// Verify status was updated
	booking, _ := repo.GetByReference(ctx, "test-ref-123")
	if booking.BookingStatus != model.BookingStatusConfirmed {
		t.Errorf("expected status CONFIRMED, got: %s", booking.BookingStatus)
	}
}

func TestBookingService_ListBookings(t *testing.T) {
	ctx := context.Background()
	repo := newMockBookingRepository()

	// Pre-populate bookings
	repo.bookings["ref-1"] = &model.BookingRecord{
		CarrierBookingRequestReference: "ref-1",
		BookingStatus:                  model.BookingStatusReceived,
	}
	repo.bookings["ref-2"] = &model.BookingRecord{
		CarrierBookingRequestReference: "ref-2",
		BookingStatus:                  model.BookingStatusConfirmed,
	}

	svc := &BookingService{
		repo:  repo,
		cache: nil,
	}

	bookings, total, err := svc.ListBookings(ctx, &repository.BookingFilter{Page: 1, PageSize: 10})
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if total != 2 {
		t.Errorf("expected total 2, got: %d", total)
	}
	if len(bookings) != 2 {
		t.Errorf("expected 2 bookings, got: %d", len(bookings))
	}
}

func TestBookingService_CanUpdate(t *testing.T) {
	svc := &BookingService{}

	tests := []struct {
		status    model.BookingStatus
		canUpdate bool
	}{
		{model.BookingStatusReceived, true},
		{model.BookingStatusPendingUpdate, true},
		{model.BookingStatusConfirmed, true},
		{model.BookingStatusCancelled, false},
		{model.BookingStatusRejected, false},
		{model.BookingStatusCompleted, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			result := svc.canUpdate(tt.status)
			if result != tt.canUpdate {
				t.Errorf("expected canUpdate(%s) = %v, got %v", tt.status, tt.canUpdate, result)
			}
		})
	}
}

func TestBookingService_CanCancel(t *testing.T) {
	svc := &BookingService{}

	tests := []struct {
		status    model.BookingStatus
		canCancel bool
	}{
		{model.BookingStatusReceived, true},
		{model.BookingStatusPendingUpdate, true},
		{model.BookingStatusConfirmed, true},
		{model.BookingStatusPendingAmendment, true},
		{model.BookingStatusCancelled, false},
		{model.BookingStatusRejected, false},
		{model.BookingStatusCompleted, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			result := svc.canCancel(tt.status)
			if result != tt.canCancel {
				t.Errorf("expected canCancel(%s) = %v, got %v", tt.status, tt.canCancel, result)
			}
		})
	}
}

func TestGenerateReference(t *testing.T) {
	ref1 := generateReference()
	ref2 := generateReference()

	// Should generate different references
	if ref1 == ref2 {
		t.Error("expected different references to be generated")
	}

	// Should start with "BK"
	if len(ref1) < 2 || ref1[:2] != "BK" {
		t.Errorf("expected reference to start with 'BK', got: %s", ref1)
	}
}

func TestBookingService_HealthCheck(t *testing.T) {
	svc := &BookingService{}
	ctx := context.Background()

	err := svc.HealthCheck(ctx)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestBookingService_RecordToBooking(t *testing.T) {
	svc := &BookingService{}

	now := time.Now()
	record := &model.BookingRecord{
		CarrierBookingRequestReference: "test-ref",
		CarrierBookingReference:        "carrier-ref",
		BookingStatus:                  model.BookingStatusConfirmed,
		CarrierCode:                    "MAEU",
		ReceiptTypeAtOrigin:            model.ReceiptDeliveryTypeCY,
		DeliveryTypeAtDestination:      model.ReceiptDeliveryTypeCY,
		VesselName:                     "Test Vessel",
		VesselIMONumber:                "1234567",
		ExpectedDepartureDate:          &now,
	}

	booking := svc.recordToBooking(record)

	if booking.CarrierBookingRequestReference != "test-ref" {
		t.Errorf("expected reference 'test-ref', got: %s", booking.CarrierBookingRequestReference)
	}
	if booking.BookingStatus != model.BookingStatusConfirmed {
		t.Errorf("expected status CONFIRMED, got: %s", booking.BookingStatus)
	}
	if booking.Vessel == nil {
		t.Error("expected vessel to be populated")
	}
	if booking.Vessel.Name != "Test Vessel" {
		t.Errorf("expected vessel name 'Test Vessel', got: %s", booking.Vessel.Name)
	}
}
