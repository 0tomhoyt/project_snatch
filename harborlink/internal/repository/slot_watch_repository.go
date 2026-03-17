package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/yourname/harborlink/internal/model"
)

// Common slot watch errors
var (
	ErrSlotWatchNotFound      = errors.New("slot watch not found")
	ErrSlotWatchAlreadyExists = errors.New("slot watch already exists")
	ErrSlotWatchExpired       = errors.New("slot watch has expired")
	ErrSlotWatchNotActive     = errors.New("slot watch is not active")
)

// SlotWatchFilter defines filters for listing slot watches
type SlotWatchFilter struct {
	TenantID    string
	Status      model.WatchStatus
	CarrierCode string
	POL         string
	POD         string
	StartDate   *time.Time
	EndDate     *time.Time
	Page        int
	PageSize    int
}

// SlotWatchRepository defines the interface for slot watch data access
type SlotWatchRepository interface {
	Create(ctx context.Context, watch *model.SlotWatch) error
	GetByID(ctx context.Context, id uint) (*model.SlotWatch, error)
	GetByReference(ctx context.Context, reference string) (*model.SlotWatch, error)
	GetByTenantReference(ctx context.Context, tenantID, reference string) (*model.SlotWatch, error)
	Update(ctx context.Context, watch *model.SlotWatch) error
	UpdateStatus(ctx context.Context, id uint, status model.WatchStatus) error
	Delete(ctx context.Context, reference string) error
	List(ctx context.Context, filter *SlotWatchFilter) ([]model.SlotWatch, int64, error)
	ListActiveByCarrier(ctx context.Context, carrierCode string) ([]model.SlotWatch, error)
	ListActive(ctx context.Context) ([]model.SlotWatch, error)
	MarkTriggered(ctx context.Context, id uint, carrierCode string, bookingRef string) error
	IncrementRetry(ctx context.Context, id uint) error
	CleanupExpired(ctx context.Context) (int64, error)
}

// slotWatchRepository implements SlotWatchRepository
type slotWatchRepository struct {
	db *Database
}

// NewSlotWatchRepository creates a new slot watch repository
func NewSlotWatchRepository(db *Database) SlotWatchRepository {
	return &slotWatchRepository{db: db}
}

// Create creates a new slot watch record
func (r *slotWatchRepository) Create(ctx context.Context, watch *model.SlotWatch) error {
	// Marshal CarrierCodes to JSON
	if len(watch.CarrierCodes) > 0 {
		data, err := json.Marshal(watch.CarrierCodes)
		if err != nil {
			return err
		}
		watch.CarrierCodes = nil // Clear to avoid double marshaling
		result := r.db.WithContext(ctx).Exec(
			"INSERT INTO slot_watches (tenant_id, reference, carrier_codes, pol, pod, etd_from_date, etd_to_date, equipment_type, extend_info, lock_strategy, prebuilt_booking, status, priority, max_retries, expires_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			watch.TenantID, watch.Reference, data, watch.POL, watch.POD, watch.ETDFromDate, watch.ETDToDate, watch.EquipmentType, watch.ExtendInfo, watch.LockStrategy, watch.PrebuiltBooking, watch.Status, watch.Priority, watch.MaxRetries, watch.ExpiresAt, time.Now(), time.Now(),
		)
		if result.Error != nil {
			return result.Error
		}
		return nil
	}

	result := r.db.WithContext(ctx).Create(watch)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// GetByID gets a slot watch by ID
func (r *slotWatchRepository) GetByID(ctx context.Context, id uint) (*model.SlotWatch, error) {
	var watch model.SlotWatch
	result := r.db.WithContext(ctx).First(&watch, id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrSlotWatchNotFound
		}
		return nil, result.Error
	}

	return &watch, nil
}

// GetByReference gets a slot watch by reference
func (r *slotWatchRepository) GetByReference(ctx context.Context, reference string) (*model.SlotWatch, error) {
	var watch model.SlotWatch
	result := r.db.WithContext(ctx).
		Where("reference = ?", reference).
		First(&watch)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrSlotWatchNotFound
		}
		return nil, result.Error
	}

	return &watch, nil
}

// GetByTenantReference gets a slot watch by tenant ID and reference
func (r *slotWatchRepository) GetByTenantReference(ctx context.Context, tenantID, reference string) (*model.SlotWatch, error) {
	var watch model.SlotWatch
	result := r.db.WithContext(ctx).
		Where("tenant_id = ? AND reference = ?", tenantID, reference).
		First(&watch)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrSlotWatchNotFound
		}
		return nil, result.Error
	}

	return &watch, nil
}

// Update updates an existing slot watch
func (r *slotWatchRepository) Update(ctx context.Context, watch *model.SlotWatch) error {
	result := r.db.WithContext(ctx).Save(watch)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrSlotWatchNotFound
	}
	return nil
}

// UpdateStatus updates only the status of a slot watch
func (r *slotWatchRepository) UpdateStatus(ctx context.Context, id uint, status model.WatchStatus) error {
	result := r.db.WithContext(ctx).
		Model(&model.SlotWatch{}).
		Where("id = ?", id).
		Update("status", status)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrSlotWatchNotFound
	}
	return nil
}

// Delete soft deletes a slot watch
func (r *slotWatchRepository) Delete(ctx context.Context, reference string) error {
	result := r.db.WithContext(ctx).
		Where("reference = ?", reference).
		Delete(&model.SlotWatch{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrSlotWatchNotFound
	}
	return nil
}

// List returns a paginated list of slot watches with filters
func (r *slotWatchRepository) List(ctx context.Context, filter *SlotWatchFilter) ([]model.SlotWatch, int64, error) {
	var watches []model.SlotWatch
	var total int64

	query := r.db.WithContext(ctx).Model(&model.SlotWatch{})

	// Apply filters
	if filter.TenantID != "" {
		query = query.Where("tenant_id = ?", filter.TenantID)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.CarrierCode != "" {
		query = query.Where("carrier_codes::jsonb @> ?::jsonb", `["`+filter.CarrierCode+`"]`)
	}
	if filter.POL != "" {
		query = query.Where("pol = ?", filter.POL)
	}
	if filter.POD != "" {
		query = query.Where("pod = ?", filter.POD)
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
	result := query.Offset(offset).Limit(filter.PageSize).Order("priority DESC, created_at DESC").Find(&watches)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return watches, total, nil
}

// ListActiveByCarrier returns all active watches for a specific carrier
func (r *slotWatchRepository) ListActiveByCarrier(ctx context.Context, carrierCode string) ([]model.SlotWatch, error) {
	var watches []model.SlotWatch

	query := r.db.WithContext(ctx).
		Where("status = ?", model.WatchStatusActive).
		Where("carrier_codes::jsonb @> ?::jsonb", `["`+carrierCode+`"]`).
		Where("(expires_at IS NULL OR expires_at > ?)", time.Now())

	result := query.Order("priority DESC, created_at ASC").Find(&watches)
	if result.Error != nil {
		return nil, result.Error
	}

	return watches, nil
}

// ListActive returns all active watches
func (r *slotWatchRepository) ListActive(ctx context.Context) ([]model.SlotWatch, error) {
	var watches []model.SlotWatch

	query := r.db.WithContext(ctx).
		Where("status = ?", model.WatchStatusActive).
		Where("(expires_at IS NULL OR expires_at > ?)", time.Now())

	result := query.Order("priority DESC, created_at ASC").Find(&watches)
	if result.Error != nil {
		return nil, result.Error
	}

	return watches, nil
}

// MarkTriggered marks a watch as triggered with carrier and booking reference
func (r *slotWatchRepository) MarkTriggered(ctx context.Context, id uint, carrierCode string, bookingRef string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":               model.WatchStatusTriggered,
		"triggered_at":         &now,
		"triggered_by_carrier": carrierCode,
		"booking_ref":          bookingRef,
	}

	result := r.db.WithContext(ctx).
		Model(&model.SlotWatch{}).
		Where("id = ? AND status = ?", id, model.WatchStatusActive).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrSlotWatchNotActive
	}
	return nil
}

// IncrementRetry increments the retry count for a watch
func (r *slotWatchRepository) IncrementRetry(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).
		Model(&model.SlotWatch{}).
		Where("id = ?", id).
		UpdateColumn("retry_count", gorm.Expr("retry_count + 1"))

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrSlotWatchNotFound
	}
	return nil
}

// CleanupExpired marks all expired watches
func (r *slotWatchRepository) CleanupExpired(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).
		Model(&model.SlotWatch{}).
		Where("status = ? AND expires_at IS NOT NULL AND expires_at <= ?", model.WatchStatusActive, time.Now()).
		Update("status", model.WatchStatusExpired)

	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}
