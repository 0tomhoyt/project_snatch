package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/yourname/harborlink/pkg/config"
)

// Client wraps the Redis client
type Client struct {
	*redis.Client
}

// NewClient creates a new Redis client
func NewClient(cfg *config.RedisConfig) (*Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Client{client}, nil
}

// Close closes the Redis connection
func (c *Client) Close() error {
	return c.Client.Close()
}

// BookingCache provides caching operations for bookings
type BookingCache struct {
	client *Client
	prefix string
}

// NewBookingCache creates a new booking cache
func NewBookingCache(client *Client) *BookingCache {
	return &BookingCache{
		client: client,
		prefix: "booking",
	}
}

// bookingCacheKey generates the cache key for a booking
func (c *BookingCache) bookingCacheKey(reference string) string {
	return fmt.Sprintf("%s:%s", c.prefix, reference)
}

// bookingStatusKey generates the cache key for booking status
func (c *BookingCache) bookingStatusKey(reference string) string {
	return fmt.Sprintf("%s:status:%s", c.prefix, reference)
}

// carrierLockKey generates the cache key for carrier rate limiting
func (c *BookingCache) carrierLockKey(carrierCode string) string {
	return fmt.Sprintf("carrier:lock:%s", carrierCode)
}

// Get retrieves a cached booking
func (c *BookingCache) Get(ctx context.Context, reference string, dest interface{}) error {
	key := c.bookingCacheKey(reference)
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// Set stores a booking in cache
func (c *BookingCache) Set(ctx context.Context, reference string, value interface{}, ttl time.Duration) error {
	key := c.bookingCacheKey(reference)
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, ttl).Err()
}

// Delete removes a booking from cache
func (c *BookingCache) Delete(ctx context.Context, reference string) error {
	key := c.bookingCacheKey(reference)
	return c.client.Del(ctx, key).Err()
}

// GetStatus retrieves cached booking status
func (c *BookingCache) GetStatus(ctx context.Context, reference string) (string, error) {
	key := c.bookingStatusKey(reference)
	return c.client.Get(ctx, key).Result()
}

// SetStatus caches booking status
func (c *BookingCache) SetStatus(ctx context.Context, reference string, status string, ttl time.Duration) error {
	key := c.bookingStatusKey(reference)
	return c.client.Set(ctx, key, status, ttl).Err()
}

// Exists checks if a booking exists in cache
func (c *BookingCache) Exists(ctx context.Context, reference string) (bool, error) {
	key := c.bookingCacheKey(reference)
	result, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// SetNX sets a booking only if the key does not exist (for distributed locking)
func (c *BookingCache) SetNX(ctx context.Context, reference string, value interface{}, ttl time.Duration) (bool, error) {
	key := c.bookingCacheKey(reference)
	data, err := json.Marshal(value)
	if err != nil {
		return false, err
	}
	return c.client.SetNX(ctx, key, data, ttl).Result()
}

// AcquireCarrierLock acquires a distributed lock for carrier API calls
func (c *BookingCache) AcquireCarrierLock(ctx context.Context, carrierCode string, ttl time.Duration) (bool, error) {
	key := c.carrierLockKey(carrierCode)
	return c.client.SetNX(ctx, key, "1", ttl).Result()
}

// ReleaseCarrierLock releases the carrier lock
func (c *BookingCache) ReleaseCarrierLock(ctx context.Context, carrierCode string) error {
	key := c.carrierLockKey(carrierCode)
	return c.client.Del(ctx, key).Err()
}

// IncrementRateLimit increments and returns the rate limit counter
func (c *BookingCache) IncrementRateLimit(ctx context.Context, key string, window time.Duration) (int64, error) {
	val, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	// Set expiry on first increment
	if val == 1 {
		c.client.Expire(ctx, key, window)
	}

	return val, nil
}

// Default cache TTLs
const (
	DefaultBookingTTL    = 10 * time.Minute
	DefaultStatusTTL     = 5 * time.Minute
	DefaultLockTTL       = 30 * time.Second
	DefaultRateLimitWindow = 1 * time.Minute
)
