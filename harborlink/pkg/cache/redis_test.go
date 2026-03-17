package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// mockRedis creates an in-memory Redis for testing
func mockRedis(t *testing.T) (*Client, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return &Client{client}, mr
}

func TestBookingCache_Set_Get(t *testing.T) {
	client, mr := mockRedis(t)
	defer client.Close()
	defer mr.Close()

	cache := NewBookingCache(client)
	ctx := context.Background()
	ref := "test-ref-001"

	booking := map[string]interface{}{
		"carrierBookingRequestReference": ref,
		"bookingStatus":                 "RECEIVED",
	}

	// Set booking
	err := cache.Set(ctx, ref, booking, DefaultBookingTTL)
	if err != nil {
		t.Errorf("failed to set booking: %v", err)
	}

	// Get booking
	var result map[string]interface{}
	err = cache.Get(ctx, ref, &result)
	if err != nil {
		t.Errorf("failed to get booking: %v", err)
	}

	if result["carrierBookingRequestReference"] != ref {
		t.Errorf("expected ref %s, got %v", ref, result["carrierBookingRequestReference"])
	}
}

func TestBookingCache_Delete(t *testing.T) {
	client, mr := mockRedis(t)
	defer client.Close()
	defer mr.Close()

	cache := NewBookingCache(client)
	ctx := context.Background()
	ref := "test-ref-002"

	// Set booking
	cache.Set(ctx, ref, map[string]string{"status": "RECEIVED"}, DefaultBookingTTL)

	// Delete booking
	err := cache.Delete(ctx, ref)
	if err != nil {
		t.Errorf("failed to delete booking: %v", err)
	}

	// Verify deleted
	var result map[string]string
	err = cache.Get(ctx, ref, &result)
	if err == nil {
		t.Error("expected error when getting deleted booking")
	}
}

func TestBookingCache_Status(t *testing.T) {
	client, mr := mockRedis(t)
	defer client.Close()
	defer mr.Close()

	cache := NewBookingCache(client)
	ctx := context.Background()
	ref := "test-ref-003"

	// Set status
	err := cache.SetStatus(ctx, ref, "CONFIRMED", DefaultStatusTTL)
	if err != nil {
		t.Errorf("failed to set status: %v", err)
	}

	// Get status
	status, err := cache.GetStatus(ctx, ref)
	if err != nil {
		t.Errorf("failed to get status: %v", err)
	}

	if status != "CONFIRMED" {
		t.Errorf("expected status CONFIRMED, got %s", status)
	}
}

func TestBookingCache_Exists(t *testing.T) {
	client, mr := mockRedis(t)
	defer client.Close()
	defer mr.Close()

	cache := NewBookingCache(client)
	ctx := context.Background()
	ref := "test-ref-004"

	// Check non-existent
	exists, err := cache.Exists(ctx, ref)
	if err != nil {
		t.Errorf("failed to check exists: %v", err)
	}
	if exists {
		t.Error("expected booking to not exist")
	}

	// Set booking
	cache.Set(ctx, ref, map[string]string{"status": "RECEIVED"}, DefaultBookingTTL)

	// Check exists
	exists, err = cache.Exists(ctx, ref)
	if err != nil {
		t.Errorf("failed to check exists: %v", err)
	}
	if !exists {
		t.Error("expected booking to exist")
	}
}

func TestBookingCache_CarrierLock(t *testing.T) {
	client, mr := mockRedis(t)
	defer client.Close()
	defer mr.Close()

	cache := NewBookingCache(client)
	ctx := context.Background()
	carrier := "MAEU"

	// Acquire lock
	acquired, err := cache.AcquireCarrierLock(ctx, carrier, DefaultLockTTL)
	if err != nil {
		t.Errorf("failed to acquire lock: %v", err)
	}
	if !acquired {
		t.Error("expected to acquire lock")
	}

	// Try to acquire again (should fail)
	acquired, err = cache.AcquireCarrierLock(ctx, carrier, DefaultLockTTL)
	if err != nil {
		t.Errorf("failed on second acquire: %v", err)
	}
	if acquired {
		t.Error("expected lock to be already held")
	}

	// Release lock
	err = cache.ReleaseCarrierLock(ctx, carrier)
	if err != nil {
		t.Errorf("failed to release lock: %v", err)
	}

	// Acquire again (should succeed)
	acquired, err = cache.AcquireCarrierLock(ctx, carrier, DefaultLockTTL)
	if err != nil {
		t.Errorf("failed to re-acquire lock: %v", err)
	}
	if !acquired {
		t.Error("expected to re-acquire lock")
	}
}

func TestBookingCache_RateLimit(t *testing.T) {
	client, mr := mockRedis(t)
	defer client.Close()
	defer mr.Close()

	cache := NewBookingCache(client)
	ctx := context.Background()
	key := "ratelimit:tenant-001"

	// Increment counter
	for i := 1; i <= 5; i++ {
		count, err := cache.IncrementRateLimit(ctx, key, time.Minute)
		if err != nil {
			t.Errorf("failed to increment: %v", err)
		}
		if count != int64(i) {
			t.Errorf("expected count %d, got %d", i, count)
		}
	}
}

func TestBookingCache_SetNX(t *testing.T) {
	client, mr := mockRedis(t)
	defer client.Close()
	defer mr.Close()

	cache := NewBookingCache(client)
	ctx := context.Background()
	ref := "test-ref-setnx"

	// SetNX should succeed
	set, err := cache.SetNX(ctx, ref, map[string]string{"status": "RECEIVED"}, DefaultBookingTTL)
	if err != nil {
		t.Errorf("failed to SetNX: %v", err)
	}
	if !set {
		t.Error("expected SetNX to succeed")
	}

	// SetNX should fail (key already exists)
	set, err = cache.SetNX(ctx, ref, map[string]string{"status": "CONFIRMED"}, DefaultBookingTTL)
	if err != nil {
		t.Errorf("failed to SetNX: %v", err)
	}
	if set {
		t.Error("expected SetNX to fail (key already exists)")
	}
}
