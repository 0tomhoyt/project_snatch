package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/yourname/harborlink/internal/repository"
)

// APIKeyAuth creates an API key authentication middleware
func APIKeyAuth(apiKeyRepo repository.APIKeyRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get API key from header
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			// Also try Authorization header with Bearer prefix
			authHeader := c.GetHeader("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				apiKey = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if apiKey == "" {
			c.JSON(401, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "API key is required",
				},
			})
			c.Abort()
			return
		}

		// Validate API key
		keyRecord, err := apiKeyRepo.GetByKey(c.Request.Context(), apiKey)
		if err != nil {
			c.JSON(401, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Invalid or expired API key",
				},
			})
			c.Abort()
			return
		}

		// Update last used timestamp (async, don't block request)
		go func() {
			_ = apiKeyRepo.UpdateLastUsed(c.Request.Context(), apiKey)
		}()

		// Store API key info in context
		c.Set("api_key_id", keyRecord.ID)
		c.Set("tenant_id", keyRecord.TenantID)
		c.Set("api_key_name", keyRecord.Name)
		c.Set("rate_limit", keyRecord.RateLimit)

		c.Next()
	}
}

// OptionalAPIKey creates an optional API key authentication middleware
// It will validate the key if provided but won't reject requests without one
func OptionalAPIKey(apiKeyRepo repository.APIKeyRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get API key from header
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			authHeader := c.GetHeader("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				apiKey = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if apiKey == "" {
			c.Next()
			return
		}

		// Validate API key
		keyRecord, err := apiKeyRepo.GetByKey(c.Request.Context(), apiKey)
		if err == nil {
			// Update last used timestamp (async)
			go func() {
				_ = apiKeyRepo.UpdateLastUsed(c.Request.Context(), apiKey)
			}()

			// Store API key info in context
			c.Set("api_key_id", keyRecord.ID)
			c.Set("tenant_id", keyRecord.TenantID)
			c.Set("api_key_name", keyRecord.Name)
			c.Set("rate_limit", keyRecord.RateLimit)
		}

		c.Next()
	}
}

// GetTenantID extracts tenant ID from context
func GetTenantID(c *gin.Context) string {
	if tenantID, exists := c.Get("tenant_id"); exists {
		if id, ok := tenantID.(string); ok {
			return id
		}
	}
	return ""
}

// GetAPIKeyID extracts API key ID from context
func GetAPIKeyID(c *gin.Context) uint {
	if apiKeyID, exists := c.Get("api_key_id"); exists {
		if id, ok := apiKeyID.(uint); ok {
			return id
		}
	}
	return 0
}

// GetRateLimit extracts rate limit from context
func GetRateLimit(c *gin.Context) int {
	if rateLimit, exists := c.Get("rate_limit"); exists {
		if limit, ok := rateLimit.(int); ok {
			return limit
		}
	}
	return 1000 // Default rate limit
}
