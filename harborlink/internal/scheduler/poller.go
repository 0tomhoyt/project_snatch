package scheduler

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/yourname/harborlink/internal/adapter"
	"github.com/yourname/harborlink/internal/eventbus"
	"github.com/yourname/harborlink/internal/model"
	"github.com/yourname/harborlink/internal/repository"
	"github.com/yourname/harborlink/pkg/config"
)

// Default configuration values
const (
	defaultPollInterval = 10 * time.Second
	defaultPollTimeout  = 5 * time.Second
)

// Errors
var (
	ErrPollerNotFound = errors.New("poller not found")
)

// PollerStats contains poller statistics
type PollerStats struct {
	CarrierCode   string        `json:"carrierCode"`
	WatchCount    int           `json:"watchCount"`
	PollCount     int64         `json:"pollCount"`
	LastError     string        `json:"lastError,omitempty"`
	LastPollTime  time.Time     `json:"lastPollTime"`
	PollInterval  time.Duration `json:"pollInterval"`
	IsRunning     bool          `json:"isRunning"`
}

// SlotChangeEvent represents a change in slot status
type SlotChangeEvent struct {
	WatchID  uint
	Old      *model.SlotStatus
	New      *model.SlotStatus
	ChangeType eventbus.EventType
}

// CarrierPoller handles polling for a specific carrier
type CarrierPoller struct {
	carrier   adapter.CarrierAdapter
	config    *config.CarrierConfig
	eventBus  *eventbus.EventBus
	watchRepo repository.SlotWatchRepository

	watches    map[uint]bool // watchID -> active
	lastResult map[uint]*model.SlotStatus // watchID -> last known status
	mu         sync.RWMutex

	ticker    *time.Ticker
	triggerCh chan struct{}
	stopCh    chan struct{}

	stats PollerStats
}

// NewCarrierPoller creates a new carrier poller
func NewCarrierPoller(
	carrier adapter.CarrierAdapter,
	cfg *config.CarrierConfig,
	eventBus *eventbus.EventBus,
	watchRepo repository.SlotWatchRepository,
) *CarrierPoller {
	return &CarrierPoller{
		carrier:    carrier,
		config:     cfg,
		eventBus:   eventBus,
		watchRepo:  watchRepo,
		watches:    make(map[uint]bool),
		lastResult: make(map[uint]*model.SlotStatus),
		triggerCh:  make(chan struct{}, 1),
		stopCh:     make(chan struct{}),
		stats: PollerStats{
			CarrierCode:  cfg.Code,
			PollInterval: cfg.PollInterval,
			IsRunning:    false,
		},
	}
}

// Start starts the poller
func (p *CarrierPoller) Start(ctx context.Context) {
	p.mu.Lock()
	p.stats.IsRunning = true
	p.mu.Unlock()

	defer func() {
		p.mu.Lock()
		p.stats.IsRunning = false
		p.mu.Unlock()
	}()

	// Use configured interval or default
	interval := p.config.PollInterval
	if interval <= 0 {
		interval = defaultPollInterval
	}

	p.ticker = time.NewTicker(interval)
	defer p.ticker.Stop()

	log.Printf("[INFO] Poller[%s]: started with interval %s", p.config.Code, interval)

	for {
		select {
		case <-ctx.Done():
			log.Printf("[INFO] Poller[%s]: context cancelled", p.config.Code)
			return
		case <-p.stopCh:
			log.Printf("[INFO] Poller[%s]: stopped", p.config.Code)
			return
		case <-p.ticker.C:
			p.poll(ctx)
		case <-p.triggerCh:
			p.poll(ctx)
		}
	}
}

// Stop stops the poller
func (p *CarrierPoller) Stop() {
	close(p.stopCh)
}

// AddWatch adds a watch to be monitored
func (p *CarrierPoller) AddWatch(watchID uint) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.watches[watchID] = true
	p.stats.WatchCount = len(p.watches)
}

// RemoveWatch removes a watch from monitoring
func (p *CarrierPoller) RemoveWatch(watchID uint) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.watches, watchID)
	delete(p.lastResult, watchID)
	p.stats.WatchCount = len(p.watches)
}

// TriggerPoll triggers an immediate poll
func (p *CarrierPoller) TriggerPoll() error {
	select {
	case p.triggerCh <- struct{}{}:
		return nil
	default:
		return errors.New("poll already pending")
	}
}

// GetStats returns poller statistics
func (p *CarrierPoller) GetStats() PollerStats {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.stats
}

// poll performs a single poll cycle
func (p *CarrierPoller) poll(ctx context.Context) {
	p.mu.Lock()
	watchIDs := make([]uint, 0, len(p.watches))
	for id := range p.watches {
		watchIDs = append(watchIDs, id)
	}
	p.mu.Unlock()

	if len(watchIDs) == 0 {
		return
	}

	// Create poll context with timeout
	timeout := p.config.PollTimeout
	if timeout <= 0 {
		timeout = defaultPollTimeout
	}
	pollCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Poll for each watch
	for _, watchID := range watchIDs {
		if err := p.pollWatch(pollCtx, watchID); err != nil {
			log.Printf("[ERROR] Poller[%s]: failed to poll watch %d: %v", p.config.Code, watchID, err)
		}
	}

	p.mu.Lock()
	p.stats.PollCount++
	p.stats.LastPollTime = time.Now()
	p.mu.Unlock()
}

// pollWatch polls for a specific watch
func (p *CarrierPoller) pollWatch(ctx context.Context, watchID uint) error {
	// Get watch details
	watch, err := p.watchRepo.GetByID(ctx, watchID)
	if err != nil {
		if errors.Is(err, repository.ErrSlotWatchNotFound) {
			// Watch was deleted, remove from tracking
			p.RemoveWatch(watchID)
			return nil
		}
		return err
	}

	// Check if watch is still active
	if !watch.IsActive() {
		p.RemoveWatch(watchID)
		return nil
	}

	// Query slots from carrier
	req := &model.QuerySlotsRequest{
		POL:           watch.POL,
		POD:           watch.POD,
		ETDFromDate:   watch.ETDFromDate,
		ETDToDate:     watch.ETDToDate,
		EquipmentType: watch.EquipmentType,
	}

	slots, err := p.querySlots(ctx, req)
	if err != nil {
		p.mu.Lock()
		p.stats.LastError = err.Error()
		p.mu.Unlock()
		return err
	}

	// Process results
	for _, slot := range slots {
		p.processSlot(ctx, watch, slot)
	}

	return nil
}

// querySlots queries slot availability from the carrier adapter
func (p *CarrierPoller) querySlots(ctx context.Context, req *model.QuerySlotsRequest) ([]model.SlotStatus, error) {
	// Check if adapter supports QuerySlots
	slotsAdapter, ok := p.carrier.(interface {
		QuerySlots(context.Context, *model.QuerySlotsRequest) (*model.QuerySlotsResponse, error)
	})

	if !ok {
		// Adapter doesn't support QuerySlots, return mock data for development
		return p.mockQuerySlots(req), nil
	}

	resp, err := slotsAdapter.QuerySlots(ctx, req)
	if err != nil {
		return nil, err
	}

	// Convert SlotInfo to SlotStatus
	result := make([]model.SlotStatus, len(resp.Slots))
	for i, slot := range resp.Slots {
		result[i] = model.SlotStatus{
			CarrierCode:   slot.CarrierCode,
			VesselName:    slot.VesselName,
			VoyageNumber:  slot.VoyageNumber,
			POL:           slot.POL,
			POD:           slot.POD,
			ETD:           slot.ETD,
			ETA:           slot.ETA,
			EquipmentType: slot.EquipmentType,
			Available:     slot.Available,
			AvailableQty:  slot.AvailableQty,
			FetchedAt:     time.Now(),
		}
	}

	return result, nil
}

// mockQuerySlots returns mock slot data for development
func (p *CarrierPoller) mockQuerySlots(req *model.QuerySlotsRequest) []model.SlotStatus {
	// Generate mock data based on request
	now := time.Now()
	etd := now.Add(7 * 24 * time.Hour) // 7 days from now

	return []model.SlotStatus{
		{
			CarrierCode:   p.config.Code,
			VesselName:    "MOCK VESSEL",
			VoyageNumber:  "2401E",
			POL:           req.POL,
			POD:           req.POD,
			ETD:           etd,
			EquipmentType: req.EquipmentType,
			Available:     true,
			AvailableQty:  50,
			FetchedAt:     now,
		},
	}
}

// processSlot processes a slot status and detects changes
func (p *CarrierPoller) processSlot(ctx context.Context, watch *model.SlotWatch, slot model.SlotStatus) {
	// Create unique key for this slot
	slotKey := slot.VesselName + "-" + slot.VoyageNumber

	p.mu.Lock()
	lastStatus, exists := p.lastResult[watch.ID]
	p.mu.Unlock()

	// Build current slot status pointer
	currentStatus := &slot

	// Detect change
	var change *SlotChangeEvent
	if !exists {
		// First time seeing this watch
		if slot.Available {
			change = &SlotChangeEvent{
				WatchID:    watch.ID,
				Old:        nil,
				New:        currentStatus,
				ChangeType: eventbus.EventSlotOpened,
			}
		}
	} else {
		// Compare with last status
		change = p.detectChange(watch.ID, lastStatus, currentStatus, slotKey)
	}

	// Update last result
	p.mu.Lock()
	p.lastResult[watch.ID] = currentStatus
	p.mu.Unlock()

	// Publish event if there's a change
	if change != nil {
		p.publishEvent(watch, change)
	}
}

// detectChange detects changes between old and new slot status
func (p *CarrierPoller) detectChange(watchID uint, old, new *model.SlotStatus, slotKey string) *SlotChangeEvent {
	if old == nil && new == nil {
		return nil
	}

	// Check for slot opened (was unavailable, now available)
	if old != nil && new != nil {
		if !old.Available && new.Available {
			return &SlotChangeEvent{
				WatchID:    watchID,
				Old:        old,
				New:        new,
				ChangeType: eventbus.EventSlotOpened,
			}
		}

		// Check for slot closed (was available, now unavailable)
		if old.Available && !new.Available {
			return &SlotChangeEvent{
				WatchID:    watchID,
				Old:        old,
				New:        new,
				ChangeType: eventbus.EventSlotClosed,
			}
		}

		// Check for quantity change
		if old.AvailableQty != new.AvailableQty {
			return &SlotChangeEvent{
				WatchID:    watchID,
				Old:        old,
				New:        new,
				ChangeType: eventbus.EventSlotChanged,
			}
		}
	}

	// New slot available
	if old == nil && new != nil && new.Available {
		return &SlotChangeEvent{
			WatchID:    watchID,
			Old:        nil,
			New:        new,
			ChangeType: eventbus.EventSlotOpened,
		}
	}

	return nil
}

// publishEvent publishes a slot change event to the event bus
func (p *CarrierPoller) publishEvent(watch *model.SlotWatch, change *SlotChangeEvent) {
	switch change.ChangeType {
	case eventbus.EventSlotOpened:
		p.eventBus.PublishSlotOpened(
			watch.ID,
			watch.Reference,
			watch.TenantID,
			change.New.CarrierCode,
			change.New,
			watch.LockStrategy,
		)
		log.Printf("[INFO] Poller[%s]: slot opened for watch %s (vessel: %s, voyage: %s)",
			p.config.Code, watch.Reference, change.New.VesselName, change.New.VoyageNumber)

	case eventbus.EventSlotClosed:
		p.eventBus.PublishSlotClosed(
			watch.ID,
			watch.Reference,
			watch.TenantID,
			change.New.CarrierCode,
			change.New,
		)
		log.Printf("[INFO] Poller[%s]: slot closed for watch %s", p.config.Code, watch.Reference)

	case eventbus.EventSlotChanged:
		p.eventBus.PublishSlotChanged(
			watch.ID,
			watch.Reference,
			watch.TenantID,
			change.New.CarrierCode,
			change.Old,
			change.New,
		)
	}
}
