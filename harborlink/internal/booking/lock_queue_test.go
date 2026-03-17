package booking

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestLockQueue_EnqueueDequeue(t *testing.T) {
	q := NewLockQueue()

	req := &LockRequest{
		WatchID:   1,
		WatchRef:  "WATCH-001",
		TenantID:  "tenant-1",
		Priority:  5,
		Carrier:   "MAEU",
		CreatedAt: time.Now(),
	}

	q.Enqueue(req)

	if q.Length() != 1 {
		t.Errorf("expected queue length 1, got %d", q.Length())
	}

	// Dequeue
	ctx := context.Background()
	dequeued, err := q.Dequeue(ctx)
	if err != nil {
		t.Errorf("failed to dequeue: %v", err)
	}
	if dequeued.WatchID != 1 {
		t.Errorf("expected watch ID 1, got %d", dequeued.WatchID)
	}
	if q.Length() != 0 {
		t.Errorf("expected queue length 0 after dequeue, got %d", q.Length())
	}
}

func TestLockQueue_PriorityOrder(t *testing.T) {
	q := NewLockQueue()

	// Add items with different priorities
	req1 := &LockRequest{
		WatchID:   1,
		Priority:  5,
		CreatedAt: time.Now(),
	}
	req2 := &LockRequest{
		WatchID:   2,
		Priority:  10, // Higher priority
		CreatedAt: time.Now(),
	}
	req3 := &LockRequest{
		WatchID:   3,
		Priority:  3,
		CreatedAt: time.Now(),
	}

	// Enqueue in order 1, 2, 3
	q.Enqueue(req1)
	q.Enqueue(req2)
	q.Enqueue(req3)

	// Should dequeue in priority order: 2, 1, 3
	ctx := context.Background()

	first, _ := q.Dequeue(ctx)
	if first.WatchID != 2 {
		t.Errorf("expected watch ID 2 (highest priority), got %d", first.WatchID)
	}

	second, _ := q.Dequeue(ctx)
	if second.WatchID != 1 {
		t.Errorf("expected watch ID 1 (second priority), got %d", second.WatchID)
	}

	third, _ := q.Dequeue(ctx)
	if third.WatchID != 3 {
		t.Errorf("expected watch ID 3 (lowest priority), got %d", third.WatchID)
	}
}

func TestLockQueue_SamePriorityFIFO(t *testing.T) {
	q := NewLockQueue()

	now := time.Now()

	// Add items with same priority but different times
	req1 := &LockRequest{
		WatchID:   1,
		Priority:  5,
		CreatedAt: now.Add(-2 * time.Second), // Earlier
	}
	req2 := &LockRequest{
		WatchID:   2,
		Priority:  5,
		CreatedAt: now.Add(-1 * time.Second), // Later
	}

	q.Enqueue(req2) // Add later one first
	q.Enqueue(req1) // Add earlier one second

	ctx := context.Background()

	// Earlier one should come out first (FIFO for same priority)
	first, _ := q.Dequeue(ctx)
	if first.WatchID != 1 {
		t.Errorf("expected watch ID 1 (earlier), got %d", first.WatchID)
	}

	second, _ := q.Dequeue(ctx)
	if second.WatchID != 2 {
		t.Errorf("expected watch ID 2 (later), got %d", second.WatchID)
	}
}

func TestLockQueue_DequeueEmpty(t *testing.T) {
	q := NewLockQueue()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := q.Dequeue(ctx)
	if err != context.DeadlineExceeded {
		t.Errorf("expected context.DeadlineExceeded for empty queue, got %v", err)
	}
}

func TestLockQueue_DequeueClosed(t *testing.T) {
	q := NewLockQueue()

	ctx := context.Background()

	// Close the queue
	q.Close()

	// Dequeue should return nil
	req, err := q.Dequeue(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if req != nil {
		t.Errorf("expected nil request from closed queue, got %+v", req)
	}
}

func TestLockQueue_Peek(t *testing.T) {
	q := NewLockQueue()

	// Peek empty queue
	if q.Peek() != nil {
		t.Error("expected nil peek on empty queue")
	}

	// Add item
	req := &LockRequest{
		WatchID:   1,
		Priority:  5,
		CreatedAt: time.Now(),
	}
	q.Enqueue(req)

	// Peek should return item without removing
	peeked := q.Peek()
	if peeked == nil {
		t.Error("expected item from peek")
	}
	if peeked.WatchID != 1 {
		t.Errorf("expected watch ID 1, got %d", peeked.WatchID)
	}

	// Queue should still have the item
	if q.Length() != 1 {
		t.Errorf("expected queue length 1 after peek, got %d", q.Length())
	}
}

func TestLockQueue_ConcurrentEnqueue(t *testing.T) {
	q := NewLockQueue()

	var wg sync.WaitGroup

	// Concurrent enqueues
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			q.Enqueue(&LockRequest{
				WatchID:   uint(id),
				Priority:  id % 10,
				CreatedAt: time.Now(),
			})
		}(i)
	}

	wg.Wait()

	if q.Length() != 100 {
		t.Errorf("expected queue length 100, got %d", q.Length())
	}
}

func TestLockQueue_ConcurrentDequeue(t *testing.T) {
	q := NewLockQueue()

	// Add items first
	for i := 0; i < 100; i++ {
		q.Enqueue(&LockRequest{
			WatchID:   uint(i),
			Priority:  5,
			CreatedAt: time.Now(),
		})
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	dequeued := make(map[uint]bool)

	// Concurrent dequeues
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx := context.Background()
			for j := 0; j < 10; j++ {
				req, err := q.Dequeue(ctx)
				if err == nil && req != nil {
					mu.Lock()
					dequeued[req.WatchID] = true
					mu.Unlock()
				}
			}
		}()
	}

	wg.Wait()

	if len(dequeued) != 100 {
		t.Errorf("expected 100 unique items dequeued, got %d", len(dequeued))
	}
}

func TestLockQueue_Process(t *testing.T) {
	q := NewLockQueue()

	processed := make([]uint, 0)
	var mu sync.Mutex

	ctx, cancel := context.WithCancel(context.Background())

	// Start processor in goroutine
	go q.Process(ctx, func(req *LockRequest) error {
		mu.Lock()
		processed = append(processed, req.WatchID)
		mu.Unlock()
		return nil
	})

	// Enqueue items
	q.Enqueue(&LockRequest{WatchID: 1, Priority: 5, CreatedAt: time.Now()})
	q.Enqueue(&LockRequest{WatchID: 2, Priority: 5, CreatedAt: time.Now()})

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	cancel()

	mu.Lock()
	if len(processed) < 2 {
		t.Errorf("expected at least 2 items processed, got %d", len(processed))
	}
	mu.Unlock()
}

func TestLockQueue_GetStats(t *testing.T) {
	q := NewLockQueue()

	// Empty queue stats
	stats := q.GetStats()
	if stats.Length != 0 {
		t.Errorf("expected length 0, got %d", stats.Length)
	}

	// Add items
	now := time.Now()
	q.Enqueue(&LockRequest{WatchID: 1, Priority: 10, CreatedAt: now.Add(-1 * time.Second)})
	q.Enqueue(&LockRequest{WatchID: 2, Priority: 5, CreatedAt: now})

	stats = q.GetStats()
	if stats.Length != 2 {
		t.Errorf("expected length 2, got %d", stats.Length)
	}
	if stats.TopPriority != 10 {
		t.Errorf("expected top priority 10, got %d", stats.TopPriority)
	}
	if stats.OldestItem.IsZero() {
		t.Error("expected oldest item to be set")
	}
}

func TestLockQueue_Close(t *testing.T) {
	q := NewLockQueue()

	// Add item
	q.Enqueue(&LockRequest{WatchID: 1, Priority: 5, CreatedAt: time.Now()})

	// Close should not block
	done := make(chan struct{})
	go func() {
		q.Close()
		close(done)
	}()

	select {
	case <-done:
		// Good
	case <-time.After(100 * time.Millisecond):
		t.Error("Close blocked")
	}

	// Queue should be closed
	if q.Length() != 1 {
		t.Errorf("expected length 1 after close, got %d", q.Length())
	}
}

func TestLockQueue_EnqueueAfterClose(t *testing.T) {
	q := NewLockQueue()
	q.Close()

	// Should not add to closed queue
	q.Enqueue(&LockRequest{WatchID: 1, Priority: 5, CreatedAt: time.Now()})

	if q.Length() != 0 {
		t.Errorf("expected length 0 when enqueuing to closed queue, got %d", q.Length())
	}
}
