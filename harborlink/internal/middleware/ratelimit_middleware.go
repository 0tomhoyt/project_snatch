package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/yourname/harborlink/pkg/cache"
)

// RateLimiterConfig holds configuration for the rate limiter
type RateLimiterConfig struct {
	// Requests per minute allowed
	RequestsPerMinute int
	// Window duration for rate limiting
	Window time.Duration
	// Key function to generate rate limit key
	KeyFunc func(c *gin.Context) string
}

// DefaultRateLimiterConfig returns default rate limiter configuration
func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		RequestsPerMinute: 100,
		Window:           time.Minute,
		KeyFunc:          defaultKeyFunc,
	}
}

// defaultKeyFunc generates a rate limit key based on client IP
func defaultKeyFunc(c *gin.Context) string {
	// Use tenant ID if available, otherwise use IP
	tenantID := GetTenantID(c)
	if tenantID != "" {
		return fmt.Sprintf("ratelimit:tenant:%s", tenantID)
	}
	return fmt.Sprintf("ratelimit:ip:%s", c.ClientIP())
}

// RateLimiter creates a rate limiting middleware using Redis
func RateLimiter(bookingCache *cache.BookingCache, cfg RateLimiterConfig) gin.HandlerFunc {
	if cfg.RequestsPerMinute <= 0 {
		cfg.RequestsPerMinute = 100
	}
	if cfg.Window <= 0 {
		cfg.Window = time.Minute
	}
	if cfg.KeyFunc == nil {
		cfg.KeyFunc = defaultKeyFunc
	}

	return func(c *gin.Context) {
		// Skip rate limiting if cache is not available
		if bookingCache == nil {
			c.Next()
			return
		}

		key := cfg.KeyFunc(c)
		cacheKey := fmt.Sprintf("rl:%s", key)

		// Increment counter
		count, err := bookingCache.IncrementRateLimit(c.Request.Context(), cacheKey, cfg.Window)
		if err != nil {
			// If Redis fails, allow the request
			c.Next()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", cfg.RequestsPerMinute))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, int64(cfg.RequestsPerMinute)-count)))

		// Check if limit exceeded
		if count > int64(cfg.RequestsPerMinute) {
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", cfg.Window/time.Second))

			c.JSON(429, gin.H{
				"error": gin.H{
					"code":    "RATE_LIMIT_EXCEEDED",
					"message": "Rate limit exceeded. Please try again later.",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// TenantRateLimiter creates a rate limiter that uses per-tenant limits
func TenantRateLimiter(bookingCache *cache.BookingCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip if cache not available
		if bookingCache == nil {
			c.Next()
			return
		}

		// Get tenant-specific rate limit
		rateLimit := GetRateLimit(c)
		if rateLimit <= 0 {
			rateLimit = 1000
		}

		cfg := RateLimiterConfig{
			RequestsPerMinute: rateLimit,
			Window:           time.Hour, // Per hour for tenant limits
			KeyFunc: func(c *gin.Context) string {
				tenantID := GetTenantID(c)
				if tenantID != "" {
					return fmt.Sprintf("tenant_rl:%s", tenantID)
				}
				return fmt.Sprintf("ip_rl:%s", c.ClientIP())
			},
		}

		key := cfg.KeyFunc(c)
		cacheKey := fmt.Sprintf("trl:%s", key)

		count, err := bookingCache.IncrementRateLimit(c.Request.Context(), cacheKey, cfg.Window)
		if err != nil {
			c.Next()
			return
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", cfg.RequestsPerMinute))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, int64(cfg.RequestsPerMinute)-count)))

		if count > int64(cfg.RequestsPerMinute) {
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", cfg.Window/time.Second))

			c.JSON(429, gin.H{
				"error": gin.H{
					"code":    "RATE_LIMIT_EXCEEDED",
					"message": "Rate limit exceeded. Please try again later.",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CarrierRateLimiter creates a rate limiter specific to carrier API calls
func CarrierRateLimiter(bookingCache *cache.BookingCache, carrierCode string, requestsPerMinute int) gin.HandlerFunc {
	return func(c *gin.Context) {
		if bookingCache == nil {
			c.Next()
			return
		}

		key := fmt.Sprintf("carrier_rl:%s", carrierCode)

		count, err := bookingCache.IncrementRateLimit(c.Request.Context(), key, time.Minute)
		if err != nil {
			c.Next()
			return
		}

		if count > int64(requestsPerMinute) {
			c.JSON(429, gin.H{
				"error": gin.H{
					"code":    "CARRIER_RATE_LIMIT",
					"message": fmt.Sprintf("Carrier %s rate limit exceeded", carrierCode),
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
