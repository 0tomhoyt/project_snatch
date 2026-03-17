package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/yourname/harborlink/internal/model"
)

// Common errors
var (
	ErrBookingNotFound      = errors.New("booking not found")
	ErrBookingAlreadyExists = errors.New("booking already exists")
	ErrInvalidBookingStatus = errors.New("invalid booking status")
)

// BookingRepository defines the interface for booking data access
type BookingRepository interface {
	Create(ctx context.Context, booking *model.BookingRecord) error
	GetByReference(ctx context.Context, reference string) (*model.BookingRecord, error)
	GetByCarrierReference(ctx context.Context, carrierRef string) (*model.BookingRecord, error)
	Update(ctx context.Context, booking *model.BookingRecord) error
	UpdateStatus(ctx context.Context, reference string, status model.BookingStatus) error
	Delete(ctx context.Context, reference string) error
	List(ctx context.Context, filter *BookingFilter) ([]model.BookingRecord, int64, error)
}

// BookingFilter defines filters for listing bookings
type BookingFilter struct {
	Status       model.BookingStatus
	CarrierCode  string
	StartDate    *time.Time
	EndDate      *time.Time
	Page         int
	PageSize     int
}

// bookingRepository implements BookingRepository
type bookingRepository struct {
	db *Database
}

// NewBookingRepository creates a new booking repository
func NewBookingRepository(db *Database) BookingRepository {
	return &bookingRepository{db: db}
}

// Create creates a new booking record
func (r *bookingRepository) Create(ctx context.Context, booking *model.BookingRecord) error {
	result := r.db.WithContext(ctx).Create(booking)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// GetByReference gets a booking by carrier booking request reference
func (r *bookingRepository) GetByReference(ctx context.Context, reference string) (*model.BookingRecord, error) {
	var booking model.BookingRecord
	result := r.db.WithContext(ctx).
		Where("carrier_booking_request_reference = ? OR carrier_booking_reference = ?", reference, reference).
		First(&booking)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrBookingNotFound
		}
		return nil, result.Error
	}

	return &booking, nil
}

// GetByCarrierReference gets a booking by carrier booking reference
func (r *bookingRepository) GetByCarrierReference(ctx context.Context, carrierRef string) (*model.BookingRecord, error) {
	var booking model.BookingRecord
	result := r.db.WithContext(ctx).
		Where("carrier_booking_reference = ?", carrierRef).
		First(&booking)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrBookingNotFound
		}
		return nil, result.Error
	}

	return &booking, nil
}

// Update updates an existing booking
func (r *bookingRepository) Update(ctx context.Context, booking *model.BookingRecord) error {
	result := r.db.WithContext(ctx).Save(booking)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrBookingNotFound
	}
	return nil
}

// UpdateStatus updates only the status of a booking
func (r *bookingRepository) UpdateStatus(ctx context.Context, reference string, status model.BookingStatus) error {
	result := r.db.WithContext(ctx).
		Model(&model.BookingRecord{}).
		Where("carrier_booking_request_reference = ? OR carrier_booking_reference = ?", reference, reference).
		Update("booking_status", status)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrBookingNotFound
	}
	return nil
}

// Delete soft deletes a booking
func (r *bookingRepository) Delete(ctx context.Context, reference string) error {
	result := r.db.WithContext(ctx).
		Where("carrier_booking_request_reference = ? OR carrier_booking_reference = ?", reference, reference).
		Delete(&model.BookingRecord{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrBookingNotFound
	}
	return nil
}

// List returns a paginated list of bookings with filters
func (r *bookingRepository) List(ctx context.Context, filter *BookingFilter) ([]model.BookingRecord, int64, error) {
	var bookings []model.BookingRecord
	var total int64

	query := r.db.WithContext(ctx).Model(&model.BookingRecord{})

	// Apply filters
	if filter.Status != "" {
		query = query.Where("booking_status = ?", filter.Status)
	}
	if filter.CarrierCode != "" {
		query = query.Where("carrier_code = ?", filter.CarrierCode)
	}
	if filter.StartDate != nil {
		query = query.Where("created_at >= ?", filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("created_at <= ?", filter.EndDate)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	offset := (filter.Page - 1) * filter.PageSize
	result := query.Offset(offset).Limit(filter.PageSize).Order("created_at DESC").Find(&bookings)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return bookings, total, nil
}
