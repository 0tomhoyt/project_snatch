package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Response represents a standard API response
type Response struct {
	Data       interface{} `json:"data,omitempty"`
	Error      *ErrorBody  `json:"error,omitempty"`
	Meta       *Meta       `json:"meta,omitempty"`
}

// ErrorBody represents an error in the response
type ErrorBody struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	Description string `json:"description,omitempty"`
}

// Meta contains metadata for paginated responses
type Meta struct {
	Page       int   `json:"page,omitempty"`
	PageSize   int   `json:"pageSize,omitempty"`
	TotalCount int64 `json:"totalCount,omitempty"`
	TotalPages int   `json:"totalPages,omitempty"`
}

// Success sends a successful response with data
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Data: data,
	})
}

// SuccessWithMeta sends a successful response with data and pagination metadata
func SuccessWithMeta(c *gin.Context, data interface{}, meta *Meta) {
	c.JSON(http.StatusOK, Response{
		Data: data,
		Meta: meta,
	})
}

// Created sends a 201 Created response
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Data: data,
	})
}

// Accepted sends a 202 Accepted response (for async operations)
func Accepted(c *gin.Context, data interface{}) {
	c.JSON(http.StatusAccepted, Response{
		Data: data,
	})
}

// NoContent sends a 204 No Content response
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Error sends an error response
func Error(c *gin.Context, status int, code, message string) {
	c.JSON(status, Response{
		Error: &ErrorBody{
			Code:    code,
			Message: message,
		},
	})
}

// ErrorWithDescription sends an error response with additional description
func ErrorWithDescription(c *gin.Context, status int, code, message, description string) {
	c.JSON(status, Response{
		Error: &ErrorBody{
			Code:        code,
			Message:     message,
			Description: description,
		},
	})
}

// BadRequest sends a 400 Bad Request error
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, "BAD_REQUEST", message)
}

// ValidationError sends a 400 Bad Request error for validation failures
func ValidationError(c *gin.Context, message string, details string) {
	ErrorWithDescription(c, http.StatusBadRequest, "VALIDATION_ERROR", message, details)
}

// Unauthorized sends a 401 Unauthorized error
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// Forbidden sends a 403 Forbidden error
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, "FORBIDDEN", message)
}

// NotFound sends a 404 Not Found error
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, "NOT_FOUND", message)
}

// Conflict sends a 409 Conflict error
func Conflict(c *gin.Context, message string) {
	Error(c, http.StatusConflict, "CONFLICT", message)
}

// TooManyRequests sends a 429 Too Many Requests error
func TooManyRequests(c *gin.Context, message string) {
	Error(c, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", message)
}

// InternalError sends a 500 Internal Server Error
func InternalError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", message)
}

// ServiceUnavailable sends a 503 Service Unavailable error
func ServiceUnavailable(c *gin.Context, message string) {
	Error(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", message)
}

// CarrierError sends an error response from carrier adapter errors
func CarrierError(c *gin.Context, statusCode int, carrierCode, code, message string) {
	c.JSON(statusCode, Response{
		Error: &ErrorBody{
			Code:    code,
			Message: message,
		},
	})
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

// Health sends a health check response
func Health(c *gin.Context, version string) {
	c.JSON(http.StatusOK, HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC(),
		Version:   version,
	})
}

// CalculateTotalPages calculates the total number of pages
func CalculateTotalPages(totalCount int64, pageSize int) int {
	if pageSize <= 0 {
		return 0
	}
	pages := int(totalCount) / pageSize
	if int(totalCount)%pageSize > 0 {
		pages++
	}
	return pages
}
