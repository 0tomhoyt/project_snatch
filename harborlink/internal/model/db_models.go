package model

import (
	"time"

	"gorm.io/gorm"
)

// BookingRecord represents the database entity for a booking
type BookingRecord struct {
	ID                            uint           `gorm:"primaryKey" json:"-"`
	CarrierBookingRequestReference string         `gorm:"uniqueIndex;size:100;not null" json:"carrierBookingRequestReference"`
	CarrierBookingReference       string         `gorm:"index;size:35" json:"carrierBookingReference,omitempty"`
	BookingStatus                 BookingStatus  `gorm:"size:50;not null;index" json:"bookingStatus"`
	CarrierCode                   string         `gorm:"size:10;index" json:"carrierCode,omitempty"`

	// Service details
	ReceiptTypeAtOrigin           ReceiptDeliveryType `gorm:"size:3" json:"receiptTypeAtOrigin"`
	DeliveryTypeAtDestination     ReceiptDeliveryType `gorm:"size:3" json:"deliveryTypeAtDestination"`
	CargoMovementTypeAtOrigin     CargoMovementType   `gorm:"size:3" json:"cargoMovementTypeAtOrigin"`
	CargoMovementTypeAtDestination CargoMovementType `gorm:"size:3" json:"cargoMovementTypeAtDestination"`
	ServiceContractReference      string              `gorm:"size:30" json:"serviceContractReference,omitempty"`
	ContractQuotationReference    string              `gorm:"size:35" json:"contractQuotationReference,omitempty"`

	// Vessel and voyage
	VesselName                    string `gorm:"size:50" json:"vesselName,omitempty"`
	VesselIMONumber               string `gorm:"size:8" json:"vesselIMONumber,omitempty"`
	CarrierServiceCode            string `gorm:"size:11" json:"carrierServiceCode,omitempty"`
	CarrierExportVoyageNumber     string `gorm:"size:50" json:"carrierExportVoyageNumber,omitempty"`

	// Locations
	POLUNLocationCode             string `gorm:"size:5" json:"polUNLocationCode,omitempty"`
	PODUNLocationCode             string `gorm:"size:5" json:"podUNLocationCode,omitempty"`
	PlaceOfReceiptUNLocationCode  string `gorm:"size:5" json:"placeOfReceiptUNLocationCode,omitempty"`
	PlaceOfDeliveryUNLocationCode string `gorm:"size:5" json:"placeOfDeliveryUNLocationCode,omitempty"`

	// Dates
	ExpectedDepartureDate        *time.Time `json:"expectedDepartureDate,omitempty"`
	ExpectedArrivalDate          *time.Time `json:"expectedArrivalDate,omitempty"`

	// Equipment summary
	TotalEquipmentUnits          int `json:"totalEquipmentUnits,omitempty"`

	// Raw JSON storage for complex nested objects
	RequestedEquipmentsJSON      []byte `gorm:"type:jsonb" json:"-"`
	ConfirmedEquipmentsJSON      []byte `gorm:"type:jsonb" json:"-"`
	DocumentPartiesJSON          []byte `gorm:"type:jsonb" json:"-"`
	ShipmentLocationsJSON        []byte `gorm:"type:jsonb" json:"-"`
	RequestedEquipmentJSON       []byte `gorm:"type:jsonb" json:"-"`

	// Metadata
	CreatedAt                    time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt                    time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt                    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName returns the table name for BookingRecord
func (BookingRecord) TableName() string {
	return "bookings"
}

// CarrierConfig represents the database entity for carrier configuration
type CarrierConfigRecord struct {
	ID          uint   `gorm:"primaryKey" json:"-"`
	Name        string `gorm:"size:100;not null" json:"name"`
	Code        string `gorm:"size:10;uniqueIndex;not null" json:"code"`
	AdapterType string `gorm:"size:50;not null" json:"adapterType"`
	Enabled     bool   `gorm:"default:true" json:"enabled"`
	BaseURL     string `gorm:"size:255" json:"baseUrl,omitempty"`
	APIKey      string `gorm:"size:255" json:"-"`
	RateLimit   int    `gorm:"default:100" json:"rateLimit"`

	// Metadata
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName returns the table name for CarrierConfigRecord
func (CarrierConfigRecord) TableName() string {
	return "carrier_configs"
}

// APIKey represents an API key for multi-tenant access
type APIKeyRecord struct {
	ID          uint   `gorm:"primaryKey" json:"-"`
	Key         string `gorm:"size:64;uniqueIndex;not null" json:"key"`
	Name        string `gorm:"size:100;not null" json:"name"`
	TenantID    string `gorm:"size:50;index" json:"tenantId"`
	Active      bool   `gorm:"default:true" json:"active"`
	Permissions string `gorm:"size:255" json:"permissions,omitempty"` // JSON array of permissions

	// Rate limiting
	RateLimit   int    `gorm:"default:1000" json:"rateLimit"` // requests per hour

	// Timestamps
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
	LastUsedAt  *time.Time `json:"lastUsedAt,omitempty"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName returns the table name for APIKeyRecord
func (APIKeyRecord) TableName() string {
	return "api_keys"
}

// BookingAuditLog represents an audit trail for booking changes
type BookingAuditLog struct {
	ID               uint           `gorm:"primaryKey" json:"-"`
	BookingID        uint           `gorm:"index;not null" json:"bookingId"`
	BookingReference string         `gorm:"size:100;index" json:"bookingReference"`
	Action           string         `gorm:"size:50;not null" json:"action"` // CREATED, UPDATED, STATUS_CHANGED, CANCELLED
	OldStatus        BookingStatus  `gorm:"size:50" json:"oldStatus,omitempty"`
	NewStatus        BookingStatus  `gorm:"size:50" json:"newStatus,omitempty"`
	ChangedBy        string         `gorm:"size:100" json:"changedBy,omitempty"`
	Details          string         `gorm:"type:text" json:"details,omitempty"` // JSON details of changes
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"createdAt"`
}

// TableName returns the table name for BookingAuditLog
func (BookingAuditLog) TableName() string {
	return "booking_audit_logs"
}
