package booking

import (
	"container/heap"
	"context"
	"sync"
	"time"
)

// LockRequest represents a lock request in the queue
type LockRequest struct {
	WatchID    uint              `json:"watchId"`
	WatchRef   string            `json:"watchRef"`
	TenantID   string            `json:"tenantId"`
	Priority   int               `json:"priority"`
	Carrier    string            `json:"carrier"`
	Slot       interface{}       `json:"slot,omitempty"`
	BookingReq []byte            `json:"bookingReq,omitempty"`
	CreatedAt  time.Time         `json:"createdAt"`
	index      int               // index in the heap
}

// LockQueue is a priority queue for lock requests
type LockQueue struct {
	items    []*LockRequest
	mu       sync.Mutex
	notEmpty *sync.Cond
	closed   bool
}

// NewLockQueue creates a new lock queue
func NewLockQueue() *LockQueue {
	q := &LockQueue{
		items: make([]*LockRequest, 0),
	}
	q.notEmpty = sync.NewCond(&q.mu)
	return q
}

// Enqueue adds a request to the queue (sorted by priority)
func (q *LockQueue) Enqueue(req *LockRequest) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return
	}

	heap.Push(&lockQueueHeap{items: q.items}, req)
	q.notEmpty.Signal()
}

// Dequeue removes and returns the highest priority request
func (q *LockQueue) Dequeue(ctx context.Context) (*LockRequest, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for len(q.items) == 0 && !q.closed {
		// Wait for items or context cancellation
		done := make(chan struct{})
		go func() {
			q.notEmpty.Wait()
			close(done)
		}()

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-done:
		}

		if q.closed {
			return nil, nil
		}
	}

	if len(q.items) == 0 {
		return nil, nil
	}

	return heap.Pop(&lockQueueHeap{items: q.items}).(*LockRequest), nil
}

// Process continuously processes items from the queue
func (q *LockQueue) Process(ctx context.Context, handler func(*LockRequest) error) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			req, err := q.Dequeue(ctx)
			if err != nil || req == nil {
				if ctx.Err() != nil {
					return
				}
				continue
			}

			// Process the request
			if err := handler(req); err != nil {
				// Log error but continue processing
			}
		}
	}
}

// Close closes the queue
func (q *LockQueue) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.closed = true
	q.notEmpty.Broadcast()
}

// Length returns the current queue length
func (q *LockQueue) Length() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items)
}

// Peek returns the highest priority request without removing it
func (q *LockQueue) Peek() *LockRequest {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.items) == 0 {
		return nil
	}
	return q.items[0]
}

// lockQueueHeap implements heap.Interface for priority queue
type lockQueueHeap struct {
	items []*LockRequest
}

func (h lockQueueHeap) Len() int { return len(h.items) }

func (h lockQueueHeap) Less(i, j int) bool {
	// Higher priority first, then earlier created time
	if h.items[i].Priority != h.items[j].Priority {
		return h.items[i].Priority > h.items[j].Priority
	}
	return h.items[i].CreatedAt.Before(h.items[j].CreatedAt)
}

func (h lockQueueHeap) Swap(i, j int) {
	h.items[i], h.items[j] = h.items[j], h.items[i]
	h.items[i].index = i
	h.items[j].index = j
}

func (h *lockQueueHeap) Push(x interface{}) {
	n := len(h.items)
	item := x.(*LockRequest)
	item.index = n
	h.items = append(h.items, item)
}

func (h *lockQueueHeap) Pop() interface{} {
	old := h.items
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	h.items = old[0 : n-1]
	return item
}

// PriorityQueueStats contains queue statistics
type PriorityQueueStats struct {
	Length      int       `json:"length"`
	TopPriority int       `json:"topPriority,omitempty"`
	OldestItem  time.Time `json:"oldestItem,omitempty"`
}

// GetStats returns queue statistics
func (q *LockQueue) GetStats() PriorityQueueStats {
	q.mu.Lock()
	defer q.mu.Unlock()

	stats := PriorityQueueStats{
		Length: len(q.items),
	}

	if len(q.items) > 0 {
		stats.TopPriority = q.items[0].Priority
		stats.OldestItem = q.items[0].CreatedAt
	}

	return stats
}
