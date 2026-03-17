package model

import (
	"time"
)

// ErrorResponse represents an API error response
type ErrorResponse struct {
	HTTPMethod                  HTTPMethod    `json:"httpMethod"`
	RequestURI                  string        `json:"requestUri"`
	StatusCode                  int           `json:"statusCode"`
	StatusCodeText              string        `json:"statusCodeText"`
	StatusCodeMessage           string        `json:"statusCodeMessage,omitempty"`
	ProviderCorrelationReference string        `json:"providerCorrelationReference,omitempty"`
	ErrorDateTime               time.Time     `json:"errorDateTime"`
	Errors                      []DetailedError `json:"errors"`
}

// DetailedError represents a detailed error
type DetailedError struct {
	ErrorCode       int    `json:"errorCode,omitempty"`
	Property        string `json:"property,omitempty"`
	Value           string `json:"value,omitempty"`
	JsonPath        string `json:"jsonPath,omitempty"`
	ErrorCodeText   string `json:"errorCodeText"`
	ErrorCodeMessage string `json:"errorCodeMessage"`
}

// BookingNotification represents a CloudEvents notification for booking
type BookingNotification struct {
	SpecVersion          string    `json:"specversion"`
	ID                   string    `json:"id"`
	Source               string    `json:"source"`
	Type                 string    `json:"type"`
	Time                 time.Time `json:"time"`
	DataContentType      string    `json:"datacontenttype"`
	SubscriptionReference string    `json:"subscriptionReference"`
	Data                 *Booking  `json:"data,omitempty"`
}

// NewErrorResponse creates a new ErrorResponse
func NewErrorResponse(method HTTPMethod, uri string, statusCode int, statusText string, message string) *ErrorResponse {
	return &ErrorResponse{
		HTTPMethod:        method,
		RequestURI:        uri,
		StatusCode:        statusCode,
		StatusCodeText:    statusText,
		StatusCodeMessage: message,
		ErrorDateTime:     time.Now(),
		Errors:            []DetailedError{},
	}
}

// AddError adds a detailed error to the response
func (e *ErrorResponse) AddError(errorCode int, property, value, errorCodeText, errorCodeMessage string) {
	e.Errors = append(e.Errors, DetailedError{
		ErrorCode:        errorCode,
		Property:         property,
		Value:            value,
		ErrorCodeText:    errorCodeText,
		ErrorCodeMessage: errorCodeMessage,
	})
}
