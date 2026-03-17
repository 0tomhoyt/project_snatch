package eventbus

import (
	"sync"
	"testing"
	"time"

	"github.com/yourname/harborlink/internal/model"
)

func TestEventBus_SubscribePublish(t *testing.T) {
	bus := NewEventBus(10)
	defer bus.Close()

	// Subscribe to SLOT_OPENED events
	ch := bus.Subscribe(EventSlotOpened)

	// Publish an event
	slot := &model.SlotStatus{
		CarrierCode:   "MAEU",
		VesselName:    "Test Vessel",
		VoyageNumber:  "2401E",
		POL:           "CNSHA",
		POD:           "USLAX",
		Available:     true,
		AvailableQty:  50,
		FetchedAt:     time.Now(),
	}

	go func() {
		bus.PublishSlotOpened(1, "WATCH-001", "tenant-1", "MAEU", slot, model.LockStrategyAutoLock)
	}()

	select {
	case event := <-ch:
		if event.Type != EventSlotOpened {
				t.Errorf("expected event type %s, got %s", EventSlotOpened, event.Type)
			}
		if event.Carrier != "MAEU" {
			t.Errorf("expected carrier MAEU, got %s", event.Carrier)
		}
		if event.Slot.VesselName != "Test Vessel" {
			t.Errorf("expected vessel name Test Vessel, got %s", event.Slot.VesselName)
		}
	case <-time.After(1 * time.Second):
		t.Error("timeout waiting for event")
	}
}

func TestEventBus_MultipleSubscribers(t *testing.T) {
	bus := NewEventBus(10)
	defer bus.Close()

	// Subscribe multiple times
	ch1 := bus.Subscribe(EventSlotOpened)
	ch2 := bus.Subscribe(EventSlotOpened)

	// Publish an event
	bus.Publish(Event{
		Type:      EventSlotOpened,
		Carrier:   "MAEU",
		WatchID:   1,
		Timestamp: time.Now(),
	})

	// Both should receive
	select {
	case <-ch1:
		// Good
	case <-time.After(100 * time.Millisecond):
		t.Error("ch1 did not receive event")
	}

	select {
	case <-ch2:
		// Good
	case <-time.After(100 * time.Millisecond):
		t.Error("ch2 did not receive event")
	}
}

func TestEventBus_AllEventTypes(t *testing.T) {
	bus := NewEventBus(10)
	defer bus.Close()

	tests := []struct {
		eventType EventType
		publish   func()
	}{
		{EventSlotOpened, func() {
			bus.PublishSlotOpened(1, "WATCH-001", "tenant-1", "MAEU", &model.SlotStatus{}, model.LockStrategyAutoLock)
		}},
		{EventSlotClosed, func() {
			bus.PublishSlotClosed(1, "WATCH-001", "tenant-1", "MAEU", &model.SlotStatus{})
		}},
		{EventSlotChanged, func() {
			bus.PublishSlotChanged(1, "WATCH-001", "tenant-1", "MAEU", &model.SlotStatus{}, &model.SlotStatus{})
		}},
		{EventLockSuccess, func() {
			bus.PublishLockSuccess(1, "WATCH-001", "tenant-1", "MAEU", "BK123")
		}},
		{EventLockFailed, func() {
			bus.PublishLockFailed(1, "WATCH-001", "tenant-1", "MAEU", "test error")
		}},
		{EventLockPending, func() {
			bus.PublishLockPending(1, "WATCH-001", "tenant-1", "MAEU", &model.SlotStatus{})
		}},
	}

	for _, tt := range tests {
		t.Run(string(tt.eventType), func(t *testing.T) {
			ch := bus.Subscribe(tt.eventType)

			tt.publish()

			select {
			case event := <-ch:
				if event.Type != tt.eventType {
					t.Errorf("expected event type %s, got %s", tt.eventType, event.Type)
				}
			case <-time.After(100 * time.Millisecond):
				t.Error("timeout waiting for event")
			}
		})
	}
}

func TestEventBus_SubscribeAll(t *testing.T) {
	bus := NewEventBus(10)
	defer bus.Close()

	ch := bus.SubscribeAll()

	// Publish different event types
	bus.Publish(Event{Type: EventSlotOpened, Timestamp: time.Now()})
	bus.Publish(Event{Type: EventLockSuccess, Timestamp: time.Now()})

	// Should receive both
	count := 0
	timeout := time.After(200 * time.Millisecond)

	for count < 2 {
		select {
		case <-ch:
			count++
		case <-timeout:
			t.Errorf("only received %d events, expected 2", count)
			return
		}
	}
}

func TestEventBus_Unsubscribe(t *testing.T) {
	bus := NewEventBus(10)
	defer bus.Close()

	ch := bus.Subscribe(EventSlotOpened)

	// Publish should work
	bus.Publish(Event{Type: EventSlotOpened, Timestamp: time.Now()})
	select {
	case <-ch:
		// Good
	case <-time.After(100 * time.Millisecond):
		t.Error("did not receive event")
	}

	// Unsubscribe
	bus.Unsubscribe(EventSlotOpened, ch)

	// Publish again - should not receive (channel closed)
	bus.Publish(Event{Type: EventSlotOpened, Timestamp: time.Now()})
	select {
	case _, ok := <-ch:
		if ok {
			t.Error("should not receive event after unsubscribe")
		}
	default:
		// Good - no event received
	}
}

func TestEventBus_RegisterHandler(t *testing.T) {
	bus := NewEventBus(10)
	defer bus.Close()

	var received Event
	var mu sync.Mutex

	bus.RegisterHandler(EventSlotOpened, func(e Event) {
		mu.Lock()
		received = e
		mu.Unlock()
	})

	bus.Publish(Event{
		Type:      EventSlotOpened,
		Carrier:   "MAEU",
		WatchID:   1,
		Timestamp: time.Now(),
	})

	// Wait for handler to be called
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	if received.Type != EventSlotOpened {
		t.Errorf("expected handler to receive event, got type %s", received.Type)
	}
	if received.Carrier != "MAEU" {
		t.Errorf("expected carrier MAEU, got %s", received.Carrier)
	}
	mu.Unlock()
}

func TestEventBus_ConcurrentAccess(t *testing.T) {
	bus := NewEventBus(100)
	defer bus.Close()

	ch := bus.Subscribe(EventSlotOpened)

	var wg sync.WaitGroup

	// Concurrent publishers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				bus.Publish(Event{
					Type:      EventSlotOpened,
					Carrier:   "MAEU",
					WatchID:   uint(id*10 + j),
					Timestamp: time.Now(),
				})
			}
		}(i)
	}

	// Wait for all publishers
	go func() {
		wg.Wait()
	}()

	// Count received events
	count := 0
	timeout := time.After(2 * time.Second)

	for count < 100 {
		select {
		case <-ch:
			count++
		case <-timeout:
			t.Errorf("timeout: only received %d events, expected 100", count)
			return
		}
	}
}

func TestEventBus_BufferOverflow(t *testing.T) {
	// Small buffer
	bus := NewEventBus(2)
	defer bus.Close()

	ch := bus.Subscribe(EventSlotOpened)

	// Publish more than buffer size
	for i := 0; i < 10; i++ {
		bus.Publish(Event{
			Type:      EventSlotOpened,
			WatchID:   uint(i),
			Timestamp: time.Now(),
		})
	}

	// Should receive at least buffer size events
	count := 0
	timeout := time.After(100 * time.Millisecond)

	for {
		select {
		case <-ch:
			count++
		case <-timeout:
			if count < 2 {
				t.Errorf("expected at least 2 events, got %d", count)
			}
			return
		}
	}
}

func TestEventBus_Stats(t *testing.T) {
	bus := NewEventBus(10)
	defer bus.Close()

	// Subscribe
	_ = bus.Subscribe(EventSlotOpened)
	_ = bus.Subscribe(EventSlotClosed)
	_ = bus.Subscribe(EventSlotOpened)

	// Register handler
	bus.RegisterHandler(EventSlotOpened, func(e Event) {})

	stats := bus.Stats()

	if stats["SLOT_OPENED_subscribers"] != 2 {
		t.Errorf("expected 2 SLOT_OPENED subscribers, got %v", stats["SLOT_OPENED_subscribers"])
	}
	if stats["SLOT_CLOSED_subscribers"] != 1 {
		t.Errorf("expected 1 SLOT_CLOSED subscriber, got %v", stats["SLOT_CLOSED_subscribers"])
	}
	if stats["SLOT_OPENED_handlers"] != 1 {
		t.Errorf("expected 1 SLOT_OPENED handler, got %v", stats["SLOT_OPENED_handlers"])
	}
}
