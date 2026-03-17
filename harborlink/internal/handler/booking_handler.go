package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/yourname/harborlink/internal/model"
	"github.com/yourname/harborlink/internal/repository"
	"github.com/yourname/harborlink/internal/service"
)

// BookingHandler handles HTTP requests for booking operations
type BookingHandler struct {
	service *service.BookingService
}

// NewBookingHandler creates a new booking handler
func NewBookingHandler(svc *service.BookingService) *BookingHandler {
	return &BookingHandler{
		service: svc,
	}
}

// CreateBooking handles POST /v2/bookings
func (h *BookingHandler) CreateBooking(c *gin.Context) {
	var req model.CreateBooking
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Validate required fields
	if len(req.ShipmentLocations) == 0 {
		BadRequest(c, "shipmentLocations is required")
		return
	}

	booking, err := h.service.CreateBooking(c.Request.Context(), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	Accepted(c, booking)
}

// GetBooking handles GET /v2/bookings/:reference
func (h *BookingHandler) GetBooking(c *gin.Context) {
	reference := c.Param("reference")
	if reference == "" {
		BadRequest(c, "reference is required")
		return
	}

	// Check if we should sync from carrier
	sync := c.Query("sync") == "true"

	var booking *model.Booking
	var err error

	if sync {
		booking, err = h.service.SyncBookingFromCarrier(c.Request.Context(), reference)
	} else {
		booking, err = h.service.GetBooking(c.Request.Context(), reference)
	}

	if err != nil {
		h.handleError(c, err)
		return
	}

	Success(c, booking)
}

// UpdateBooking handles PUT /v2/bookings/:reference
func (h *BookingHandler) UpdateBooking(c *gin.Context) {
	reference := c.Param("reference")
	if reference == "" {
		BadRequest(c, "reference is required")
		return
	}

	var req model.UpdateBooking
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, "Invalid request body", err.Error())
		return
	}

	booking, err := h.service.UpdateBooking(c.Request.Context(), reference, &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	Accepted(c, booking)
}

// CancelBooking handles DELETE /v2/bookings/:reference
func (h *BookingHandler) CancelBooking(c *gin.Context) {
	reference := c.Param("reference")
	if reference == "" {
		BadRequest(c, "reference is required")
		return
	}

	var req model.CancelBookingRequest
	// Body is optional for cancel
	_ = c.ShouldBindJSON(&req)

	if err := h.service.CancelBooking(c.Request.Context(), reference, &req); err != nil {
		h.handleError(c, err)
		return
	}

	NoContent(c)
}

// ListBookings handles GET /v2/bookings
func (h *BookingHandler) ListBookings(c *gin.Context) {
	filter := h.parseListFilter(c)

	bookings, total, err := h.service.ListBookings(c.Request.Context(), filter)
	if err != nil {
		h.handleError(c, err)
		return
	}

	meta := &Meta{
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalCount: total,
		TotalPages: CalculateTotalPages(total, filter.PageSize),
	}

	SuccessWithMeta(c, bookings, meta)
}

// GetBookingStatus handles GET /v2/bookings/:reference/status
func (h *BookingHandler) GetBookingStatus(c *gin.Context) {
	reference := c.Param("reference")
	if reference == "" {
		BadRequest(c, "reference is required")
		return
	}

	status, err := h.service.GetBookingStatus(c.Request.Context(), reference)
	if err != nil {
		h.handleError(c, err)
		return
	}

	Success(c, gin.H{
		"carrierBookingRequestReference": reference,
		"bookingStatus":                  status,
	})
}

// parseListFilter parses query parameters into a booking filter
func (h *BookingHandler) parseListFilter(c *gin.Context) *repository.BookingFilter {
	filter := &repository.BookingFilter{
		Page:     1,
		PageSize: 20,
	}

	// Parse pagination
	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			filter.Page = p
		}
	}
	if pageSize := c.Query("pageSize"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 && ps <= 100 {
			filter.PageSize = ps
		}
	}

	// Parse filters
	if status := c.Query("status"); status != "" {
		filter.Status = model.BookingStatus(status)
	}
	if carrierCode := c.Query("carrierCode"); carrierCode != "" {
		filter.CarrierCode = carrierCode
	}

	return filter
}

// handleError maps service errors to HTTP responses
func (h *BookingHandler) handleError(c *gin.Context, err error) {
	switch err {
	case service.ErrBookingNotFound:
		NotFound(c, "Booking not found")
	case service.ErrInvalidCarrier:
		BadRequest(c, "Invalid or unavailable carrier")
	case service.ErrBookingAlreadyExists:
		Conflict(c, "Booking already exists")
	case service.ErrInvalidStatus:
		BadRequest(c, "Booking status does not allow this operation")
	default:
		// Check for adapter errors
		InternalError(c, err.Error())
	}
}

// RegisterRoutes registers booking routes on the given router group
func (h *BookingHandler) RegisterRoutes(rg *gin.RouterGroup) {
	bookings := rg.Group("/bookings")
	{
		bookings.GET("", h.ListBookings)
		bookings.POST("", h.CreateBooking)
		bookings.GET("/:reference", h.GetBooking)
		bookings.PUT("/:reference", h.UpdateBooking)
		bookings.DELETE("/:reference", h.CancelBooking)
		bookings.GET("/:reference/status", h.GetBookingStatus)
	}
}
