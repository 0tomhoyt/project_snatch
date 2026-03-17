package model

import (
	"encoding/json"
	"time"
)

// LockStrategy defines how slot locks are handled
type LockStrategy string

const (
	LockStrategyAutoLock      LockStrategy = "AUTO_LOCK"       // 发现即锁定
	LockStrategyNotifyConfirm LockStrategy = "NOTIFY_CONFIRM" // 通知后确认
)

// WatchStatus defines the status of a slot watch
type WatchStatus string

const (
	WatchStatusActive     WatchStatus = "ACTIVE"     // 监控中
	WatchStatusTriggered  WatchStatus = "TRIGGERED"  // 已触发
	WatchStatusCancelled  WatchStatus = "CANCELLED"  // 已取消
	WatchStatusExpired    WatchStatus = "EXPIRED"    // 已过期
	WatchStatusPending    WatchStatus = "PENDING"    // 等待确认
	WatchStatusConfirmed  WatchStatus = "CONFIRMED"  // 已确认锁定
	WatchStatusFailed     WatchStatus = "FAILED"     // 锁定失败
)

// SlotWatch represents a slot monitoring request
type SlotWatch struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TenantID  string    `gorm:"size:50;index" json:"tenantId"`
	Reference string    `gorm:"uniqueIndex;size:50;not null" json:"reference"`

	// 监控目标（可扩展）
	CarrierCodes      []string        `gorm:"type:json;serializer:json" json:"carrierCodes"` // 多载体竞争
	POL               string          `gorm:"size:5;index" json:"pol"`        // 起运港
	POD               string          `gorm:"size:5;index" json:"pod"`        // 目的港
	ETDFromDate       *time.Time      `json:"etdFromDate"`                    // ETD开始日期
	ETDToDate         *time.Time      `json:"etdToDate"`                      // ETD结束日期
	EquipmentType     string          `gorm:"size:4" json:"equipmentType"`    // 柜型
	ExtendInfo        json.RawMessage `gorm:"type:json" json:"extendInfo"`   // 扩展信息

	// 锁定策略
	LockStrategy      LockStrategy    `gorm:"size:20;not null" json:"lockStrategy"`
	PrebuiltBooking   json.RawMessage `gorm:"type:json" json:"prebuiltBooking"` // 预构建的booking请求

	// 状态
	Status            WatchStatus     `gorm:"size:20;not null;index" json:"status"`
	TriggeredAt       *time.Time      `json:"triggeredAt"`        // 触发时间
	TriggeredByCarrier string         `gorm:"size:10" json:"triggeredByCarrier"` // 触发的载体
	BookingRef        *string         `gorm:"size:100" json:"bookingRef"` // 锁定后的booking引用
	LockResult        json.RawMessage `gorm:"type:json" json:"lockResult"` // 锁定结果详情

	// 优先级
	Priority          int             `gorm:"default:5" json:"priority"` // 优先级 (1-10, 越高越优先)

	// 重试计数
	RetryCount        int             `gorm:"default:0" json:"retryCount"`
	MaxRetries        int             `gorm:"default:3" json:"maxRetries"`

	// 时间戳
	CreatedAt         time.Time       `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt         time.Time       `gorm:"autoUpdateTime" json:"updatedAt"`
	ExpiresAt         *time.Time      `json:"expiresAt"` // 过期时间
}

// TableName returns the table name for SlotWatch
func (SlotWatch) TableName() string {
	return "slot_watches"
}

// SlotStatus represents slot availability from carrier API
type SlotStatus struct {
	CarrierCode   string    `json:"carrierCode"`
	VesselName    string    `json:"vesselName"`
	VoyageNumber  string    `json:"voyageNumber"`
	POL           string    `json:"pol"`
	POD           string    `json:"pod"`
	ETD           time.Time `json:"etd"`
	ETA           time.Time `json:"eta,omitempty"`
	EquipmentType string    `json:"equipmentType"`
	Available     bool      `json:"available"`
	AvailableQty  int       `json:"availableQty"`
	FetchedAt     time.Time `json:"fetchedAt"`
}

// SlotInfo represents detailed slot information for API responses
type SlotInfo struct {
	CarrierCode   string    `json:"carrierCode"`
	VesselName    string    `json:"vesselName"`
	VoyageNumber  string    `json:"voyageNumber"`
	POL           string    `json:"pol"`
	POD           string    `json:"pod"`
	ETD           time.Time `json:"etd"`
	ETA           time.Time `json:"eta,omitempty"`
	EquipmentType string    `json:"equipmentType"`
	Available     bool      `json:"available"`
	AvailableQty  int       `json:"availableQty"`
}

// QuerySlotsRequest represents a request to query slot availability
type QuerySlotsRequest struct {
	POL           string     `json:"pol"`
	POD           string     `json:"pod"`
	ETDFromDate   *time.Time `json:"etdFromDate,omitempty"`
	ETDToDate     *time.Time `json:"etdToDate,omitempty"`
	EquipmentType string     `json:"equipmentType,omitempty"`
}

// QuerySlotsResponse represents the response from slot availability query
type QuerySlotsResponse struct {
	Slots []SlotInfo `json:"slots"`
}

// CreateSlotWatchRequest represents the request to create a slot watch
type CreateSlotWatchRequest struct {
	CarrierCodes    []string        `json:"carrierCodes" binding:"required,min=1"`
	POL             string          `json:"pol" binding:"required,len=5"`
	POD             string          `json:"pod" binding:"required,len=5"`
	ETDFromDate     *time.Time      `json:"etdFromDate,omitempty"`
	ETDToDate       *time.Time      `json:"etdToDate,omitempty"`
	EquipmentType   string          `json:"equipmentType,omitempty"`
	ExtendInfo      json.RawMessage `json:"extendInfo,omitempty"`
	LockStrategy    LockStrategy    `json:"lockStrategy" binding:"required"`
	PrebuiltBooking json.RawMessage `json:"prebuiltBooking,omitempty"`
	Priority        int             `json:"priority,omitempty"`
	ExpiresAt       *time.Time      `json:"expiresAt,omitempty"`
	MaxRetries      int             `json:"maxRetries,omitempty"`
}

// SlotWatchResponse represents the response for a slot watch
type SlotWatchResponse struct {
	Reference          string       `json:"reference"`
	TenantID           string       `json:"tenantId"`
	CarrierCodes       []string     `json:"carrierCodes"`
	POL                string       `json:"pol"`
	POD                string       `json:"pod"`
	ETDFromDate        *time.Time   `json:"etdFromDate,omitempty"`
	ETDToDate          *time.Time   `json:"etdToDate,omitempty"`
	EquipmentType      string       `json:"equipmentType,omitempty"`
	LockStrategy       LockStrategy `json:"lockStrategy"`
	Status             WatchStatus  `json:"status"`
	Priority           int          `json:"priority"`
	TriggeredAt        *time.Time   `json:"triggeredAt,omitempty"`
	TriggeredByCarrier string       `json:"triggeredByCarrier,omitempty"`
	BookingRef         *string      `json:"bookingRef,omitempty"`
	CreatedAt          time.Time    `json:"createdAt"`
	ExpiresAt          *time.Time   `json:"expiresAt,omitempty"`
}

// SlotWatchListResponse represents a list of slot watches with pagination
type SlotWatchListResponse struct {
	Data       []SlotWatchResponse `json:"data"`
	Meta       *ListMeta           `json:"meta,omitempty"`
}

// ListMeta contains pagination metadata
type ListMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	TotalCount int64 `json:"totalCount"`
	TotalPages int   `json:"totalPages"`
}

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	Type          string      `json:"type"` // SLOT_OPENED, SLOT_CLOSED, LOCK_SUCCESS, LOCK_FAILED, etc.
	WatchReference string     `json:"watchReference,omitempty"`
	Carrier       string      `json:"carrier,omitempty"`
	Slot          *SlotInfo   `json:"slot,omitempty"`
	BookingRef    string      `json:"bookingRef,omitempty"`
	Error         string      `json:"error,omitempty"`
	Timestamp     time.Time   `json:"timestamp"`
	Data          interface{} `json:"data,omitempty"`
}

// ConfirmLockRequest represents a request to confirm a lock
type ConfirmLockRequest struct {
	Confirmed bool `json:"confirmed"`
}

// IsExpired checks if the watch has expired
func (w *SlotWatch) IsExpired() bool {
	if w.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*w.ExpiresAt)
}

// IsActive checks if the watch is still active
func (w *SlotWatch) IsActive() bool {
	return w.Status == WatchStatusActive && !w.IsExpired()
}

// CanTrigger checks if the watch can be triggered
func (w *SlotWatch) CanTrigger() bool {
	return w.IsActive()
}

// CanRetry checks if the watch can be retried
func (w *SlotWatch) CanRetry() bool {
	return w.RetryCount < w.MaxRetries
}

// ToResponse converts SlotWatch to SlotWatchResponse
func (w *SlotWatch) ToResponse() *SlotWatchResponse {
	return &SlotWatchResponse{
		Reference:          w.Reference,
		TenantID:           w.TenantID,
		CarrierCodes:       w.CarrierCodes,
		POL:                w.POL,
		POD:                w.POD,
		ETDFromDate:        w.ETDFromDate,
		ETDToDate:          w.ETDToDate,
		EquipmentType:      w.EquipmentType,
		LockStrategy:       w.LockStrategy,
		Status:             w.Status,
		Priority:           w.Priority,
		TriggeredAt:        w.TriggeredAt,
		TriggeredByCarrier: w.TriggeredByCarrier,
		BookingRef:         w.BookingRef,
		CreatedAt:          w.CreatedAt,
		ExpiresAt:          w.ExpiresAt,
	}
}
