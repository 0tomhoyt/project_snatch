package eventbus

import (
	"log"
	"sync"
	"time"

	"github.com/yourname/harborlink/internal/model"
)

// EventType defines the type of slot event
type EventType string

const (
	EventSlotOpened   EventType = "SLOT_OPENED"
	EventSlotClosed   EventType = "SLOT_CLOSED"
	EventSlotChanged  EventType = "SLOT_CHANGED"
	EventLockSuccess  EventType = "LOCK_SUCCESS"
	EventLockFailed   EventType = "LOCK_FAILED"
	EventLockPending  EventType = "LOCK_PENDING"
	EventWatchExpired EventType = "WATCH_EXPIRED"
)

// Event represents a slot event
type Event struct {
	Type          EventType            `json:"type"`
	Carrier       string               `json:"carrier"`
	WatchID       uint                 `json:"watchId"`
	WatchRef      string               `json:"watchRef"`
	TenantID      string               `json:"tenantId"`
	Slot          *model.SlotStatus    `json:"slot,omitempty"`
	OldSlot       *model.SlotStatus    `json:"oldSlot,omitempty"`
	BookingRef    string               `json:"bookingRef,omitempty"`
	Error         string               `json:"error,omitempty"`
	LockStrategy  model.LockStrategy   `json:"lockStrategy,omitempty"`
	Timestamp     time.Time            `json:"timestamp"`
}

// Handler is a function that handles events
type Handler func(event Event)

// EventBus manages event publishing and subscription
type EventBus struct {
	mu          sync.RWMutex
	subscribers map[EventType][]chan Event
	handlers    map[EventType][]Handler
	bufferSize  int
}

// NewEventBus creates a new event bus
func NewEventBus(bufferSize int) *EventBus {
	if bufferSize <= 0 {
		bufferSize = 100
	}
	return &EventBus{
		subscribers: make(map[EventType][]chan Event),
		handlers:    make(map[EventType][]Handler),
		bufferSize:  bufferSize,
	}
}

// Subscribe creates a subscription channel for a specific event type
func (b *EventBus) Subscribe(eventType EventType) <-chan Event {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan Event, b.bufferSize)
	b.subscribers[eventType] = append(b.subscribers[eventType], ch)
	return ch
}

// SubscribeAll creates a subscription channel for all events
func (b *EventBus) SubscribeAll() <-chan Event {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan Event, b.bufferSize)
	for _, eventType := range []EventType{EventSlotOpened, EventSlotClosed, EventSlotChanged, EventLockSuccess, EventLockFailed, EventLockPending, EventWatchExpired} {
		b.subscribers[eventType] = append(b.subscribers[eventType], ch)
	}
	return ch
}

// RegisterHandler registers a handler function for a specific event type
func (b *EventBus) RegisterHandler(eventType EventType, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[eventType] = append(b.handlers[eventType], handler)
}

// Unsubscribe removes a subscription channel
func (b *EventBus) Unsubscribe(eventType EventType, ch <-chan Event) {
	b.mu.Lock()
	defer b.mu.Unlock()

	subscribers := b.subscribers[eventType]
	for i, sub := range subscribers {
		if sub == ch {
			// Close the channel and remove from slice
			close(sub)
			b.subscribers[eventType] = append(subscribers[:i], subscribers[i+1:]...)
			break
		}
	}
}

// Publish publishes an event to all subscribers
func (b *EventBus) Publish(event Event) {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	b.mu.RLock()
	subscribers := b.subscribers[event.Type]
	handlers := b.handlers[event.Type]
	b.mu.RUnlock()

	// Send to channel subscribers
	for _, ch := range subscribers {
		select {
		case ch <- event:
		default:
			// Channel full, log warning but don't block
			log.Printf("[WARN] EventBus: channel full for event type %s, dropping event", event.Type)
		}
	}

	// Call registered handlers
	for _, handler := range handlers {
		go func(h Handler) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[ERROR] EventBus: handler panic: %v", r)
				}
			}()
			h(event)
		}(handler)
	}
}

// PublishSlotOpened publishes a slot opened event
func (b *EventBus) PublishSlotOpened(watchID uint, watchRef, tenantID, carrier string, slot *model.SlotStatus, lockStrategy model.LockStrategy) {
	b.Publish(Event{
		Type:         EventSlotOpened,
		Carrier:      carrier,
		WatchID:      watchID,
		WatchRef:     watchRef,
		TenantID:     tenantID,
		Slot:         slot,
		LockStrategy: lockStrategy,
		Timestamp:    time.Now(),
	})
}

// PublishSlotClosed publishes a slot closed event
func (b *EventBus) PublishSlotClosed(watchID uint, watchRef, tenantID, carrier string, slot *model.SlotStatus) {
	b.Publish(Event{
		Type:       EventSlotClosed,
		Carrier:    carrier,
		WatchID:    watchID,
		WatchRef:   watchRef,
		TenantID:   tenantID,
		Slot:       slot,
		Timestamp:  time.Now(),
	})
}

// PublishSlotChanged publishes a slot changed event
func (b *EventBus) PublishSlotChanged(watchID uint, watchRef, tenantID, carrier string, oldSlot, newSlot *model.SlotStatus) {
	b.Publish(Event{
		Type:       EventSlotChanged,
		Carrier:    carrier,
		WatchID:    watchID,
		WatchRef:   watchRef,
		TenantID:   tenantID,
		OldSlot:    oldSlot,
		Slot:       newSlot,
		Timestamp:  time.Now(),
	})
}

// PublishLockSuccess publishes a lock success event
func (b *EventBus) PublishLockSuccess(watchID uint, watchRef, tenantID, carrier, bookingRef string) {
	b.Publish(Event{
		Type:       EventLockSuccess,
		Carrier:    carrier,
		WatchID:    watchID,
		WatchRef:   watchRef,
		TenantID:   tenantID,
		BookingRef: bookingRef,
		Timestamp:  time.Now(),
	})
}

// PublishLockFailed publishes a lock failed event
func (b *EventBus) PublishLockFailed(watchID uint, watchRef, tenantID, carrier, errMsg string) {
	b.Publish(Event{
		Type:       EventLockFailed,
		Carrier:    carrier,
		WatchID:    watchID,
		WatchRef:   watchRef,
		TenantID:   tenantID,
		Error:      errMsg,
		Timestamp:  time.Now(),
	})
}

// PublishLockPending publishes a lock pending event (for NOTIFY_CONFIRM)
func (b *EventBus) PublishLockPending(watchID uint, watchRef, tenantID, carrier string, slot *model.SlotStatus) {
	b.Publish(Event{
		Type:       EventLockPending,
		Carrier:    carrier,
		WatchID:    watchID,
		WatchRef:   watchRef,
		TenantID:   tenantID,
		Slot:       slot,
		Timestamp:  time.Now(),
	})
}

// Close closes all subscriber channels
func (b *EventBus) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	for eventType, subscribers := range b.subscribers {
		for _, ch := range subscribers {
			close(ch)
		}
		delete(b.subscribers, eventType)
	}
}

// Stats returns statistics about the event bus
func (b *EventBus) Stats() map[string]interface{} {
	b.mu.RLock()
	defer b.mu.RUnlock()

	stats := make(map[string]interface{})
	for eventType, subscribers := range b.subscribers {
		stats[string(eventType)+"_subscribers"] = len(subscribers)
	}
	for eventType, handlers := range b.handlers {
		stats[string(eventType)+"_handlers"] = len(handlers)
	}
	return stats
}
