package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger creates a request logging middleware
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log after request is processed
		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()

		// Get tenant info if available
		tenantID := GetTenantID(c)
		tenantInfo := ""
		if tenantID != "" {
			tenantInfo = " tenant=" + tenantID
		}

		if query != "" {
			path = path + "?" + query
		}

		// Log based on status code
		if status >= 500 {
			log.Printf("[ERROR] %s %s %s%s %d %v",
				method, path, clientIP, tenantInfo, status, latency)
		} else if status >= 400 {
			log.Printf("[WARN] %s %s %s%s %d %v",
				method, path, clientIP, tenantInfo, status, latency)
		} else {
			log.Printf("[INFO] %s %s %s%s %d %v",
				method, path, clientIP, tenantInfo, status, latency)
		}
	}
}

// RequestLoggerWithConfig creates a request logging middleware with custom configuration
type LoggingConfig struct {
	// Skip paths from logging
	SkipPaths []string
	// Log request body
	LogBody bool
	// Log request headers
	LogHeaders bool
}

// RequestLoggerWithConfig creates a configured request logging middleware
func RequestLoggerWithConfig(cfg LoggingConfig) gin.HandlerFunc {
	skipPaths := make(map[string]bool)
	for _, path := range cfg.SkipPaths {
		skipPaths[path] = true
	}

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Skip logging for certain paths
		if skipPaths[path] {
			c.Next()
			return
		}

		// Process request
		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()

		log.Printf("[INFO] %s %s %s %d %v",
			method, path, clientIP, status, latency)
	}
}

// CorrelationID adds a correlation ID to requests for tracing
func CorrelationID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if correlation ID is already set
		correlationID := c.GetHeader("X-Correlation-ID")
		if correlationID == "" {
			// Generate a simple correlation ID using timestamp
			correlationID = time.Now().UTC().Format("20060102150405") + "-" + randomString(8)
		}

		// Set in context and response header
		c.Set("correlation_id", correlationID)
		c.Header("X-Correlation-ID", correlationID)

		c.Next()
	}
}

// GetCorrelationID extracts correlation ID from context
func GetCorrelationID(c *gin.Context) string {
	if correlationID, exists := c.Get("correlation_id"); exists {
		if id, ok := correlationID.(string); ok {
			return id
		}
	}
	return ""
}

// randomString generates a random string of given length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}

// RecoveryLogger creates a panic recovery middleware with logging
func RecoveryLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				correlationID := GetCorrelationID(c)
				tenantID := GetTenantID(c)

				log.Printf("[PANIC] Recovered from panic: %v, correlation_id=%s, tenant_id=%s, path=%s",
					err, correlationID, tenantID, c.Request.URL.Path)

				c.JSON(500, gin.H{
					"error": gin.H{
						"code":    "INTERNAL_ERROR",
						"message": "An internal error occurred",
					},
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
