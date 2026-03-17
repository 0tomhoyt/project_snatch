package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/yourname/harborlink/internal/model"
	"github.com/yourname/harborlink/internal/repository"
	"github.com/yourname/harborlink/pkg/config"
)

// Common service errors
var (
	ErrSlotWatchNotFound      = errors.New("slot watch not found")
	ErrSlotWatchAlreadyExists = errors.New("slot watch already exists")
	ErrSlotWatchExpired       = errors.New("slot watch has expired")
	ErrSlotWatchNotActive     = errors.New("slot watch is not active")
	ErrInvalidLockStrategy    = errors.New("invalid lock strategy")
	ErrNoPendingConfirm       = errors.New("no pending confirmation found")
)

// SlotWatchService handles business logic for slot watch operations
type SlotWatchService struct {
	repo      repository.SlotWatchRepository
	scheduler Scheduler
	config    *config.SlotWatchConfig
}

// Scheduler interface for registering/unregistering watches
type Scheduler interface {
	RegisterWatch(watchID uint, carrierCodes []string) error
	UnregisterWatch(watchID uint, carrierCodes []string) error
}

// NewSlotWatchService creates a new slot watch service
func NewSlotWatchService(
	repo repository.SlotWatchRepository,
	scheduler Scheduler,
	cfg *config.SlotWatchConfig,
) *SlotWatchService {
	return &SlotWatchService{
		repo:      repo,
		scheduler: scheduler,
		config:    cfg,
	}
}

// CreateWatch creates a new slot watch request
func (s *SlotWatchService) CreateWatch(ctx context.Context, tenantID string, req *model.CreateSlotWatchRequest) (*model.SlotWatchResponse, error) {
	// Validate lock strategy
	if req.LockStrategy != model.LockStrategyAutoLock && req.LockStrategy != model.LockStrategyNotifyConfirm {
		return nil, ErrInvalidLockStrategy
	}

	// Generate unique reference
	reference := generateWatchReference()

	// Set default priority
	priority := req.Priority
	if priority <= 0 {
		priority = s.config.DefaultPriority
		if priority <= 0 {
			priority = 5
		}
	}

	// Create watch entity
	watch := &model.SlotWatch{
		TenantID:        tenantID,
		Reference:       reference,
		CarrierCodes:    req.CarrierCodes,
		POL:             req.POL,
		POD:             req.POD,
		ETDFromDate:     req.ETDFromDate,
		ETDToDate:       req.ETDToDate,
		EquipmentType:   req.EquipmentType,
		ExtendInfo:      req.ExtendInfo,
		LockStrategy:    req.LockStrategy,
		PrebuiltBooking: req.PrebuiltBooking,
		Status:          model.WatchStatusActive,
		Priority:        priority,
		MaxRetries:      req.MaxRetries,
		ExpiresAt:       req.ExpiresAt,
	}

	// Set default max retries
	if watch.MaxRetries <= 0 {
		watch.MaxRetries = 3
	}

	// Save to database
	if err := s.repo.Create(ctx, watch); err != nil {
		return nil, fmt.Errorf("failed to create watch: %w", err)
	}

	// Register with scheduler
	if err := s.scheduler.RegisterWatch(watch.ID, watch.CarrierCodes); err != nil {
		// Log but don't fail - scheduler will pick it up on next poll
		fmt.Printf("[WARN] Failed to register watch with scheduler: %v\n", err)
	}

	return watch.ToResponse(), nil
}

// GetWatch retrieves a watch by reference
func (s *SlotWatchService) GetWatch(ctx context.Context, tenantID, reference string) (*model.SlotWatchResponse, error) {
	watch, err := s.repo.GetByTenantReference(ctx, tenantID, reference)
	if err != nil {
		if errors.Is(err, repository.ErrSlotWatchNotFound) {
			return nil, ErrSlotWatchNotFound
		}
		return nil, fmt.Errorf("failed to get watch: %w", err)
	}

	return watch.ToResponse(), nil
}

// ListWatches lists watches with filters
func (s *SlotWatchService) ListWatches(ctx context.Context, tenantID string, filter *repository.SlotWatchFilter) (*model.SlotWatchListResponse, error) {
	// Set tenant filter
	filter.TenantID = tenantID

	watches, total, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list watches: %w", err)
	}

	// Convert to response
	data := make([]model.SlotWatchResponse, len(watches))
	for i, watch := range watches {
		data[i] = *watch.ToResponse()
	}

	// Calculate total pages
	totalPages := int(total) / filter.PageSize
	if int(total)%filter.PageSize > 0 {
		totalPages++
	}

	return &model.SlotWatchListResponse{
		Data: data,
		Meta: &model.ListMeta{
			Page:       filter.Page,
			PageSize:   filter.PageSize,
			TotalCount: total,
			TotalPages: totalPages,
		},
	}, nil
}

// CancelWatch cancels a watch
func (s *SlotWatchService) CancelWatch(ctx context.Context, tenantID, reference string) error {
	watch, err := s.repo.GetByTenantReference(ctx, tenantID, reference)
	if err != nil {
		if errors.Is(err, repository.ErrSlotWatchNotFound) {
			return ErrSlotWatchNotFound
		}
		return fmt.Errorf("failed to get watch: %w", err)
	}

	// Check if already cancelled or completed
	if watch.Status == model.WatchStatusCancelled {
		return nil // Already cancelled
	}

	if watch.Status == model.WatchStatusTriggered || watch.Status == model.WatchStatusConfirmed {
		return ErrSlotWatchNotActive
	}

	// Update status
	if err := s.repo.UpdateStatus(ctx, watch.ID, model.WatchStatusCancelled); err != nil {
		return fmt.Errorf("failed to cancel watch: %w", err)
	}

	// Unregister from scheduler
	if err := s.scheduler.UnregisterWatch(watch.ID, watch.CarrierCodes); err != nil {
		fmt.Printf("[WARN] Failed to unregister watch from scheduler: %v\n", err)
	}

	return nil
}

// ConfirmLock confirms a pending lock request (for NOTIFY_CONFIRM strategy)
func (s *SlotWatchService) ConfirmLock(ctx context.Context, tenantID, reference string, confirmed bool) error {
	watch, err := s.repo.GetByTenantReference(ctx, tenantID, reference)
	if err != nil {
		if errors.Is(err, repository.ErrSlotWatchNotFound) {
			return ErrSlotWatchNotFound
		}
		return fmt.Errorf("failed to get watch: %w", err)
	}

	// Check if watch is in pending state
	if watch.Status != model.WatchStatusPending {
		return ErrNoPendingConfirm
	}

	if !confirmed {
		// User declined, cancel the watch
		return s.CancelWatch(ctx, tenantID, reference)
	}

	// Update status to active (will be picked up for locking)
	if err := s.repo.UpdateStatus(ctx, watch.ID, model.WatchStatusActive); err != nil {
		return fmt.Errorf("failed to confirm watch: %w", err)
	}

	return nil
}

// UpdateWatchStatus updates the status of a watch
func (s *SlotWatchService) UpdateWatchStatus(ctx context.Context, id uint, status model.WatchStatus) error {
	return s.repo.UpdateStatus(ctx, id, status)
}

// CleanupExpired cleans up expired watches
func (s *SlotWatchService) CleanupExpired(ctx context.Context) (int64, error) {
	return s.repo.CleanupExpired(ctx)
}

// GetWatchByID retrieves a watch by ID (internal use)
func (s *SlotWatchService) GetWatchByID(ctx context.Context, id uint) (*model.SlotWatch, error) {
	return s.repo.GetByID(ctx, id)
}

// generateWatchReference generates a unique watch reference
func generateWatchReference() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return "WATCH-" + time.Now().Format("20060102") + hex.EncodeToString(bytes)
}

// SlotWatchServiceWithQuickLock extends SlotWatchService with quick lock functionality
type SlotWatchServiceWithQuickLock struct {
	*SlotWatchService
	quickLock QuickLockHandler
}

// QuickLockHandler interface for quick lock operations
type QuickLockHandler interface {
	ConfirmLock(ctx context.Context, reference string, confirmed bool) error
	GetPendingConfirm(reference string) (interface{}, bool)
}

// NewSlotWatchServiceWithQuickLock creates a new service with quick lock support
func NewSlotWatchServiceWithQuickLock(
	repo repository.SlotWatchRepository,
	scheduler Scheduler,
	cfg *config.SlotWatchConfig,
	quickLock QuickLockHandler,
) *SlotWatchServiceWithQuickLock {
	return &SlotWatchServiceWithQuickLock{
		SlotWatchService: NewSlotWatchService(repo, scheduler, cfg),
		quickLock:        quickLock,
	}
}

// ConfirmLockWithHandler confirms a lock using the quick lock handler
func (s *SlotWatchServiceWithQuickLock) ConfirmLockWithHandler(ctx context.Context, reference string, confirmed bool) error {
	return s.quickLock.ConfirmLock(ctx, reference, confirmed)
}

// ToSlotWatchResponse is a helper function to convert model to response
func ToSlotWatchResponse(watch *model.SlotWatch) *model.SlotWatchResponse {
	if watch == nil {
		return nil
	}

	response := &model.SlotWatchResponse{
		Reference:          watch.Reference,
		TenantID:           watch.TenantID,
		CarrierCodes:       watch.CarrierCodes,
		POL:                watch.POL,
		POD:                watch.POD,
		ETDFromDate:        watch.ETDFromDate,
		ETDToDate:          watch.ETDToDate,
		EquipmentType:      watch.EquipmentType,
		LockStrategy:       watch.LockStrategy,
		Status:             watch.Status,
		Priority:           watch.Priority,
		TriggeredAt:        watch.TriggeredAt,
		TriggeredByCarrier: watch.TriggeredByCarrier,
		BookingRef:         watch.BookingRef,
		CreatedAt:          watch.CreatedAt,
		ExpiresAt:          watch.ExpiresAt,
	}

	return response
}

// MarshalPrebuiltBooking marshals a CreateBooking request for storage
func MarshalPrebuiltBooking(booking *model.CreateBooking) (json.RawMessage, error) {
	if booking == nil {
		return nil, nil
	}
	return json.Marshal(booking)
}

// UnmarshalPrebuiltBooking unmarshals a CreateBooking request from storage
func UnmarshalPrebuiltBooking(data json.RawMessage) (*model.CreateBooking, error) {
	if len(data) == 0 {
		return nil, nil
	}
	var booking model.CreateBooking
	err := json.Unmarshal(data, &booking)
	return &booking, err
}
