package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/yourname/harborlink/internal/model"
	"github.com/yourname/harborlink/internal/notification"
	"github.com/yourname/harborlink/internal/repository"
	"github.com/yourname/harborlink/internal/service"
)

// SlotWatchHandler handles HTTP requests for slot watch operations
type SlotWatchHandler struct {
	service *service.SlotWatchService
	wsHub   *notification.WebSocketHub
}

// NewSlotWatchHandler creates a new slot watch handler
func NewSlotWatchHandler(svc *service.SlotWatchService, wsHub *notification.WebSocketHub) *SlotWatchHandler {
	return &SlotWatchHandler{
		service: svc,
		wsHub:   wsHub,
	}
}

// CreateSlotWatch handles POST /v2/slot-watches
func (h *SlotWatchHandler) CreateSlotWatch(c *gin.Context) {
	var req model.CreateSlotWatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Validate required fields
	if len(req.CarrierCodes) == 0 {
		BadRequest(c, "carrierCodes is required and must have at least one carrier")
		return
	}

	if len(req.POL) != 5 {
		BadRequest(c, "pol must be a 5-character UN location code")
		return
	}

	if len(req.POD) != 5 {
		BadRequest(c, "pod must be a 5-character UN location code")
		return
	}

	// Get tenant ID from context (set by auth middleware)
	tenantID := getTenantID(c)
	if tenantID == "" {
		Unauthorized(c, "tenant ID not found")
		return
	}

	// Create watch
	response, err := h.service.CreateWatch(c.Request.Context(), tenantID, &req)
	if err != nil {
		if err == service.ErrInvalidLockStrategy {
			BadRequest(c, "Invalid lock strategy. Must be AUTO_LOCK or NOTIFY_CONFIRM")
			return
		}
		InternalError(c, err.Error())
		return
	}

	Created(c, response)
}

// GetSlotWatch handles GET /v2/slot-watches/:reference
func (h *SlotWatchHandler) GetSlotWatch(c *gin.Context) {
	reference := c.Param("reference")
	if reference == "" {
		BadRequest(c, "reference is required")
		return
	}

	tenantID := getTenantID(c)
	if tenantID == "" {
		Unauthorized(c, "tenant ID not found")
		return
	}

	response, err := h.service.GetWatch(c.Request.Context(), tenantID, reference)
	if err != nil {
		if err == service.ErrSlotWatchNotFound {
			NotFound(c, "Slot watch not found")
			return
		}
		InternalError(c, err.Error())
		return
	}

	Success(c, response)
}

// ListSlotWatches handles GET /v2/slot-watches
func (h *SlotWatchHandler) ListSlotWatches(c *gin.Context) {
	tenantID := getTenantID(c)
	if tenantID == "" {
		Unauthorized(c, "tenant ID not found")
		return
	}

	filter := h.parseListFilter(c)

	response, err := h.service.ListWatches(c.Request.Context(), tenantID, filter)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	SuccessWithMeta(c, response.Data, &Meta{
		Page:       response.Meta.Page,
		PageSize:   response.Meta.PageSize,
		TotalCount: response.Meta.TotalCount,
		TotalPages: response.Meta.TotalPages,
	})
}

// CancelSlotWatch handles DELETE /v2/slot-watches/:reference
func (h *SlotWatchHandler) CancelSlotWatch(c *gin.Context) {
	reference := c.Param("reference")
	if reference == "" {
		BadRequest(c, "reference is required")
		return
	}

	tenantID := getTenantID(c)
	if tenantID == "" {
		Unauthorized(c, "tenant ID not found")
		return
	}

	err := h.service.CancelWatch(c.Request.Context(), tenantID, reference)
	if err != nil {
		if err == service.ErrSlotWatchNotFound {
			NotFound(c, "Slot watch not found")
			return
		}
		if err == service.ErrSlotWatchNotActive {
			BadRequest(c, "Slot watch is not active and cannot be cancelled")
			return
		}
		InternalError(c, err.Error())
		return
	}

	NoContent(c)
}

// ConfirmLock handles POST /v2/slot-watches/:reference/confirm
func (h *SlotWatchHandler) ConfirmLock(c *gin.Context) {
	reference := c.Param("reference")
	if reference == "" {
		BadRequest(c, "reference is required")
		return
	}

	var req model.ConfirmLockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, "Invalid request body", err.Error())
		return
	}

	tenantID := getTenantID(c)
	if tenantID == "" {
		Unauthorized(c, "tenant ID not found")
		return
	}

	err := h.service.ConfirmLock(c.Request.Context(), tenantID, reference, req.Confirmed)
	if err != nil {
		if err == service.ErrSlotWatchNotFound {
			NotFound(c, "Slot watch not found")
			return
		}
		if err == service.ErrNoPendingConfirm {
			BadRequest(c, "No pending confirmation found for this watch")
			return
		}
		InternalError(c, err.Error())
		return
	}

	Success(c, gin.H{
		"reference": reference,
		"confirmed": req.Confirmed,
	})
}

// WebSocketHandler handles GET /v2/ws
func (h *SlotWatchHandler) WebSocketHandler(c *gin.Context) {
	tenantID := getTenantID(c)
	if tenantID == "" {
		Unauthorized(c, "tenant ID not found")
		return
	}

	notification.ServeWebSocket(h.wsHub, tenantID)(c)
}

// parseListFilter parses query parameters into a slot watch filter
func (h *SlotWatchHandler) parseListFilter(c *gin.Context) *repository.SlotWatchFilter {
	filter := &repository.SlotWatchFilter{
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
		filter.Status = model.WatchStatus(status)
	}
	if carrierCode := c.Query("carrierCode"); carrierCode != "" {
		filter.CarrierCode = carrierCode
	}
	if pol := c.Query("pol"); pol != "" {
		filter.POL = pol
	}
	if pod := c.Query("pod"); pod != "" {
		filter.POD = pod
	}

	return filter
}

// RegisterRoutes registers slot watch routes on the given router group
func (h *SlotWatchHandler) RegisterRoutes(rg *gin.RouterGroup) {
	slotWatches := rg.Group("/slot-watches")
	{
		slotWatches.GET("", h.ListSlotWatches)
		slotWatches.POST("", h.CreateSlotWatch)
		slotWatches.GET("/:reference", h.GetSlotWatch)
		slotWatches.DELETE("/:reference", h.CancelSlotWatch)
		slotWatches.POST("/:reference/confirm", h.ConfirmLock)
	}

	// WebSocket endpoint
	rg.GET("/ws", h.WebSocketHandler)
}

// getTenantID extracts tenant ID from the context
func getTenantID(c *gin.Context) string {
	// Try to get from context (set by auth middleware)
	if tenantID, exists := c.Get("tenantID"); exists {
		if id, ok := tenantID.(string); ok {
			return id
		}
	}

	// Fallback to API key header for development
	// In production, this should be validated properly
	apiKey := c.GetHeader("X-API-Key")
	if apiKey != "" {
		// For development, use API key as tenant ID
		return apiKey
	}

	// Default tenant for development
	return "default"
}
