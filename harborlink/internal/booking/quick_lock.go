package booking

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/yourname/harborlink/internal/eventbus"
	"github.com/yourname/harborlink/internal/model"
	"github.com/yourname/harborlink/internal/repository"
	"github.com/yourname/harborlink/internal/service"
)

// Common errors
var (
	ErrLockTimeout       = errors.New("lock operation timeout")
	ErrWatchNotActive    = errors.New("watch is not active")
	ErrBookingFailed     = errors.New("booking creation failed")
	ErrConfirmTimeout    = errors.New("confirmation timeout")
)

// QuickLockHandler handles slot opened events and performs quick locking
type QuickLockHandler struct {
	bookingSvc   *service.BookingService
	watchRepo    repository.SlotWatchRepository
	eventBus     *eventbus.EventBus
	lockQueue    *LockQueue
	notifier     Notifier

	// Pending confirmations for NOTIFY_CONFIRM strategy
	pendingConfirms map[string]*PendingConfirm
	mu              sync.RWMutex

	// Configuration
	notifyTimeout time.Duration
	lockTimeout   time.Duration
}

// PendingConfirm represents a pending confirmation request
type PendingConfirm struct {
	Watch    *model.SlotWatch
	Slot     *model.SlotStatus
	Carrier  string
	ExpiresAt time.Time
}

// Notifier interface for sending notifications
type Notifier interface {
	NotifySlotOpened(ctx context.Context, tenantID string, watch *model.SlotWatch, slot *model.SlotStatus) error
	NotifyLockResult(ctx context.Context, tenantID string, watch *model.SlotWatch, success bool, bookingRef string, errMsg string) error
}

// NewQuickLockHandler creates a new quick lock handler
func NewQuickLockHandler(
	bookingSvc *service.BookingService,
	watchRepo repository.SlotWatchRepository,
	eventBus *eventbus.EventBus,
	notifier Notifier,
	notifyTimeout, lockTimeout time.Duration,
) *QuickLockHandler {
	h := &QuickLockHandler{
		bookingSvc:      bookingSvc,
		watchRepo:       watchRepo,
		eventBus:        eventBus,
		lockQueue:       NewLockQueue(),
		notifier:        notifier,
		pendingConfirms: make(map[string]*PendingConfirm),
		notifyTimeout:   notifyTimeout,
		lockTimeout:     lockTimeout,
	}

	// Subscribe to slot opened events
	eventBus.RegisterHandler(eventbus.EventSlotOpened, h.HandleSlotOpened)

	return h
}

// Start starts the lock queue processor
func (h *QuickLockHandler) Start(ctx context.Context) {
	h.lockQueue.Process(ctx, h.processLockRequest)
}

// Stop stops the handler
func (h *QuickLockHandler) Stop() {
	h.lockQueue.Close()
}

// HandleSlotOpened handles slot opened events
func (h *QuickLockHandler) HandleSlotOpened(event eventbus.Event) {
	ctx := context.Background()

	// Get watch details
	watch, err := h.watchRepo.GetByID(ctx, event.WatchID)
	if err != nil {
		log.Printf("[ERROR] QuickLock: failed to get watch %d: %v", event.WatchID, err)
		return
	}

	// Check if watch is still active
	if !watch.IsActive() {
		log.Printf("[DEBUG] QuickLock: watch %s is not active, skipping", watch.Reference)
		return
	}

	// Handle based on lock strategy
	switch watch.LockStrategy {
	case model.LockStrategyAutoLock:
		h.handleAutoLock(ctx, watch, event)

	case model.LockStrategyNotifyConfirm:
		h.handleNotifyConfirm(ctx, watch, event)

	default:
		log.Printf("[WARN] QuickLock: unknown lock strategy %s for watch %s", watch.LockStrategy, watch.Reference)
	}
}

// handleAutoLock handles AUTO_LOCK strategy
func (h *QuickLockHandler) handleAutoLock(ctx context.Context, watch *model.SlotWatch, event eventbus.Event) {
	// Create lock request
	req := &LockRequest{
		WatchID:    watch.ID,
		WatchRef:   watch.Reference,
		TenantID:   watch.TenantID,
		Priority:   watch.Priority,
		Carrier:    event.Carrier,
		Slot:       event.Slot,
		BookingReq: watch.PrebuiltBooking,
		CreatedAt:  time.Now(),
	}

	// Enqueue to priority queue
	h.lockQueue.Enqueue(req)

	log.Printf("[INFO] QuickLock: enqueued AUTO_LOCK request for watch %s (priority: %d)", watch.Reference, watch.Priority)
}

// handleNotifyConfirm handles NOTIFY_CONFIRM strategy
func (h *QuickLockHandler) notifyAndWait(ctx context.Context, watch *model.SlotWatch, event eventbus.Event) error {
	// Store pending confirmation
	h.mu.Lock()
	h.pendingConfirms[watch.Reference] = &PendingConfirm{
		Watch:     watch,
		Slot:      event.Slot,
		Carrier:   event.Carrier,
		ExpiresAt: time.Now().Add(h.notifyTimeout),
	}
	h.mu.Unlock()

	// Update watch status to pending
	if err := h.watchRepo.UpdateStatus(ctx, watch.ID, model.WatchStatusPending); err != nil {
		return err
	}

	// Publish lock pending event
	h.eventBus.PublishLockPending(watch.ID, watch.Reference, watch.TenantID, event.Carrier, event.Slot)

	// Notify client
	if h.notifier != nil {
		if err := h.notifier.NotifySlotOpened(ctx, watch.TenantID, watch, event.Slot); err != nil {
			log.Printf("[ERROR] QuickLock: failed to notify client: %v", err)
		}
	}

	log.Printf("[INFO] QuickLock: sent NOTIFY_CONFIRM for watch %s, waiting for confirmation", watch.Reference)

	return nil
}

// handleNotifyConfirm handles NOTIFY_CONFIRM strategy
func (h *QuickLockHandler) handleNotifyConfirm(ctx context.Context, watch *model.SlotWatch, event eventbus.Event) {
	_ = h.notifyAndWait(ctx, watch, event)
}

// ConfirmLock confirms a pending lock request
func (h *QuickLockHandler) ConfirmLock(ctx context.Context, reference string, confirmed bool) error {
	h.mu.Lock()
	pending, exists := h.pendingConfirms[reference]
	if !exists {
		h.mu.Unlock()
		return errors.New("no pending confirmation found")
	}

	// Check if expired
	if time.Now().After(pending.ExpiresAt) {
		delete(h.pendingConfirms, reference)
		h.mu.Unlock()
		return ErrConfirmTimeout
	}

	delete(h.pendingConfirms, reference)
	h.mu.Unlock()

	if !confirmed {
		// User declined, cancel the watch
		_ = h.watchRepo.UpdateStatus(ctx, pending.Watch.ID, model.WatchStatusCancelled)
		h.eventBus.PublishLockFailed(pending.Watch.ID, reference, pending.Watch.TenantID, pending.Carrier, "user declined")
		return nil
	}

	// User confirmed, proceed with auto lock
	event := eventbus.Event{
		Type:     eventbus.EventSlotOpened,
		Carrier:  pending.Carrier,
		WatchID:  pending.Watch.ID,
		WatchRef: reference,
		TenantID: pending.Watch.TenantID,
		Slot:     pending.Slot,
	}
	h.handleAutoLock(ctx, pending.Watch, event)

	return nil
}

// processLockRequest processes a lock request from the queue
func (h *QuickLockHandler) processLockRequest(req *LockRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), h.lockTimeout)
	defer cancel()

	// Get latest watch state
	watch, err := h.watchRepo.GetByID(ctx, req.WatchID)
	if err != nil {
		return err
	}

	// Check if still active
	if !watch.CanTrigger() {
		return ErrWatchNotActive
	}

	// Convert slot to proper type
	var slot *model.SlotStatus
	if req.Slot != nil {
		if s, ok := req.Slot.(*model.SlotStatus); ok {
			slot = s
		}
	}

	// Attempt to create booking
	bookingRef, err := h.createBooking(ctx, watch, req.Carrier, slot)
	if err != nil {
		// Handle failure
		log.Printf("[ERROR] QuickLock: failed to create booking for watch %s: %v", watch.Reference, err)

		// Increment retry count
		_ = h.watchRepo.IncrementRetry(ctx, watch.ID)

		// Check if can retry
		updatedWatch, _ := h.watchRepo.GetByID(ctx, req.WatchID)
		if updatedWatch != nil && updatedWatch.CanRetry() {
			// Re-enqueue for retry
			h.lockQueue.Enqueue(req)
			return err
		}

		// Mark as failed
		_ = h.watchRepo.UpdateStatus(ctx, watch.ID, model.WatchStatusFailed)
		h.eventBus.PublishLockFailed(watch.ID, watch.Reference, watch.TenantID, req.Carrier, err.Error())

		// Notify client
		if h.notifier != nil {
			_ = h.notifier.NotifyLockResult(ctx, watch.TenantID, watch, false, "", err.Error())
		}

		return err
	}

	// Mark as triggered
	_ = h.watchRepo.MarkTriggered(ctx, watch.ID, req.Carrier, bookingRef)

	// Publish success event
	h.eventBus.PublishLockSuccess(watch.ID, watch.Reference, watch.TenantID, req.Carrier, bookingRef)

	// Notify client
	if h.notifier != nil {
		_ = h.notifier.NotifyLockResult(ctx, watch.TenantID, watch, true, bookingRef, "")
	}

	log.Printf("[INFO] QuickLock: successfully locked slot for watch %s, booking: %s", watch.Reference, bookingRef)

	return nil
}

// createBooking creates a booking from the prebuilt request
func (h *QuickLockHandler) createBooking(ctx context.Context, watch *model.SlotWatch, carrier string, slot *model.SlotStatus) (string, error) {
	// Parse prebuilt booking
	var bookingReq model.CreateBooking
	if len(watch.PrebuiltBooking) > 0 {
		if err := json.Unmarshal(watch.PrebuiltBooking, &bookingReq); err != nil {
			return "", err
		}
	}

	// Fill in slot details
	if slot != nil {
		bookingReq.CarrierServiceCode = slot.VesselName
		bookingReq.CarrierExportVoyageNumber = slot.VoyageNumber
		if slot.ETD != (time.Time{}) {
			bookingReq.ExpectedDepartureDate = &slot.ETD
		}

		// Add shipment locations if not present
		if len(bookingReq.ShipmentLocations) == 0 {
			bookingReq.ShipmentLocations = []model.ShipmentLocation{
				{
					LocationTypeCode: model.LocationTypePol,
					Location: model.Location{
						UNLocationCode: slot.POL,
					},
				},
				{
					LocationTypeCode: model.LocationTypePod,
					Location: model.Location{
						UNLocationCode: slot.POD,
					},
				},
			}
		}
	}

	// Create booking via service
	booking, err := h.bookingSvc.CreateBooking(ctx, &bookingReq)
	if err != nil {
		return "", err
	}

	return booking.CarrierBookingRequestReference, nil
}

// GetPendingConfirm returns a pending confirmation by reference
func (h *QuickLockHandler) GetPendingConfirm(reference string) (*PendingConfirm, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	pending, exists := h.pendingConfirms[reference]
	return pending, exists
}

// CleanupExpiredConfirms removes expired pending confirmations
func (h *QuickLockHandler) CleanupExpiredConfirms(ctx context.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()

	now := time.Now()
	for ref, pending := range h.pendingConfirms {
		if now.After(pending.ExpiresAt) {
			delete(h.pendingConfirms, ref)
			_ = h.watchRepo.UpdateStatus(ctx, pending.Watch.ID, model.WatchStatusExpired)
			h.eventBus.PublishLockFailed(pending.Watch.ID, ref, pending.Watch.TenantID, pending.Carrier, "confirmation timeout")
			log.Printf("[INFO] QuickLock: expired pending confirmation for watch %s", ref)
		}
	}
}

// GetQueueLength returns the current queue length
func (h *QuickLockHandler) GetQueueLength() int {
	return h.lockQueue.Length()
}
