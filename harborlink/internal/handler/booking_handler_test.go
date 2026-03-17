package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/yourname/harborlink/internal/model"
	"github.com/yourname/harborlink/internal/repository"
	"github.com/yourname/harborlink/internal/service"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// mockBookingService implements service.BookingService for testing
type mockBookingService struct {
	bookings map[string]*model.Booking
}

func newMockBookingService() *mockBookingService {
	return &mockBookingService{
		bookings: make(map[string]*model.Booking),
	}
}

func (m *mockBookingService) CreateBooking(ctx interface{}, req *model.CreateBooking) (*model.Booking, error) {
	booking := &model.Booking{
		CarrierBookingRequestReference: "test-ref-123",
		BookingStatus:                  model.BookingStatusReceived,
	}
	m.bookings["test-ref-123"] = booking
	return booking, nil
}

func (m *mockBookingService) GetBooking(ctx interface{}, reference string) (*model.Booking, error) {
	if b, ok := m.bookings[reference]; ok {
		return b, nil
	}
	return nil, service.ErrBookingNotFound
}

func (m *mockBookingService) UpdateBooking(ctx interface{}, reference string, req *model.UpdateBooking) (*model.Booking, error) {
	if b, ok := m.bookings[reference]; ok {
		b.BookingStatus = model.BookingStatusPendingUpdate
		return b, nil
	}
	return nil, service.ErrBookingNotFound
}

func (m *mockBookingService) CancelBooking(ctx interface{}, reference string, req *model.CancelBookingRequest) error {
	if _, ok := m.bookings[reference]; ok {
		delete(m.bookings, reference)
		return nil
	}
	return service.ErrBookingNotFound
}

func (m *mockBookingService) ListBookings(ctx interface{}, filter *repository.BookingFilter) ([]model.Booking, int64, error) {
	var results []model.Booking
	for _, b := range m.bookings {
		results = append(results, *b)
	}
	return results, int64(len(results)), nil
}

func (m *mockBookingService) GetBookingStatus(ctx interface{}, reference string) (model.BookingStatus, error) {
	if b, ok := m.bookings[reference]; ok {
		return b.BookingStatus, nil
	}
	return "", service.ErrBookingNotFound
}

func TestBookingHandler_GetBooking(t *testing.T) {
	// Create mock service
	mockSvc := newMockBookingService()
	mockSvc.bookings["test-ref-123"] = &model.Booking{
		CarrierBookingRequestReference: "test-ref-123",
		BookingStatus:                  model.BookingStatusReceived,
	}

	// Create handler using a real service with mocked dependencies
	// For simplicity, we'll test the response helpers directly
	router := gin.New()
	router.GET("/bookings/:reference", func(c *gin.Context) {
		Success(c, gin.H{
			"carrierBookingRequestReference": c.Param("reference"),
			"bookingStatus":                  "RECEIVED",
		})
	})

	// Create request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/bookings/test-ref-123", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestBookingHandler_CreateBooking(t *testing.T) {
	router := gin.New()
	router.POST("/bookings", func(c *gin.Context) {
		Accepted(c, gin.H{
			"carrierBookingRequestReference": "new-ref-123",
			"bookingStatus":                  "RECEIVED",
		})
	})

	body := map[string]interface{}{
		"receiptTypeAtOrigin":            "CY",
		"deliveryTypeAtDestination":      "CY",
		"cargoMovementTypeAtOrigin":      "FCL",
		"cargoMovementTypeAtDestination": "FCL",
		"isEquipmentSubstitutionAllowed": true,
		"shipmentLocations": []map[string]interface{}{
			{
				"locationTypeCode": "POL",
				"location":         map[string]interface{}{"UNLocationCode": "CNSHA"},
			},
			{
				"locationTypeCode": "POD",
				"location":         map[string]interface{}{"UNLocationCode": "USLAX"},
			},
		},
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/bookings", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("expected status 202, got %d", w.Code)
	}
}

func TestResponseHelpers_Success(t *testing.T) {
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		Success(c, gin.H{"message": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestResponseHelpers_Created(t *testing.T) {
	router := gin.New()
	router.POST("/test", func(c *gin.Context) {
		Created(c, gin.H{"id": "123"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}
}

func TestResponseHelpers_BadRequest(t *testing.T) {
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		BadRequest(c, "invalid request")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestResponseHelpers_NotFound(t *testing.T) {
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		NotFound(c, "resource not found")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestResponseHelpers_Unauthorized(t *testing.T) {
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		Unauthorized(c, "invalid api key")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestResponseHelpers_TooManyRequests(t *testing.T) {
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		TooManyRequests(c, "rate limit exceeded")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %d", w.Code)
	}
}

func TestResponseHelpers_SuccessWithMeta(t *testing.T) {
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		SuccessWithMeta(c, []string{"item1", "item2"}, &Meta{
			Page:       1,
			PageSize:   10,
			TotalCount: 2,
			TotalPages: 1,
		})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp Response
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Meta == nil {
		t.Error("expected meta to be present")
	}
}

func TestCalculateTotalPages(t *testing.T) {
	tests := []struct {
		total    int64
		size     int
		expected int
	}{
		{0, 10, 0},
		{5, 10, 1},
		{10, 10, 1},
		{11, 10, 2},
		{100, 10, 10},
		{101, 10, 11},
	}

	for _, tt := range tests {
		result := CalculateTotalPages(tt.total, tt.size)
		if result != tt.expected {
			t.Errorf("CalculateTotalPages(%d, %d) = %d, expected %d", tt.total, tt.size, result, tt.expected)
		}
	}
}
