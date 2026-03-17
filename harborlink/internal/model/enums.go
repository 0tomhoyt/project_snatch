package model

// BookingStatus represents the status of a Booking
type BookingStatus string

const (
	BookingStatusReceived        BookingStatus = "RECEIVED"
	BookingStatusPendingUpdate   BookingStatus = "PENDING_UPDATE"
	BookingStatusUpdateReceived  BookingStatus = "UPDATE_RECEIVED"
	BookingStatusConfirmed       BookingStatus = "CONFIRMED"
	BookingStatusPendingAmendment BookingStatus = "PENDING_AMENDMENT"
	BookingStatusRejected        BookingStatus = "REJECTED"
	BookingStatusDeclined        BookingStatus = "DECLINED"
	BookingStatusCancelled       BookingStatus = "CANCELLED"
	BookingStatusCompleted       BookingStatus = "COMPLETED"
)

// AmendedBookingStatus represents the status of a booking amendment
type AmendedBookingStatus string

const (
	AmendedBookingStatusReceived  AmendedBookingStatus = "AMENDMENT_RECEIVED"
	AmendedBookingStatusConfirmed AmendedBookingStatus = "AMENDMENT_CONFIRMED"
	AmendedBookingStatusDeclined  AmendedBookingStatus = "AMENDMENT_DECLINED"
	AmendedBookingStatusCancelled AmendedBookingStatus = "AMENDMENT_CANCELLED"
)

// BookingCancellationStatus represents the status of a booking cancellation
type BookingCancellationStatus string

const (
	CancellationStatusReceived  BookingCancellationStatus = "CANCELLATION_RECEIVED"
	CancellationStatusDeclined  BookingCancellationStatus = "CANCELLATION_DECLINED"
	CancellationStatusConfirmed BookingCancellationStatus = "CANCELLATION_CONFIRMED"
)

// ReceiptDeliveryType indicates the type of service at origin/destination
type ReceiptDeliveryType string

const (
	ReceiptDeliveryTypeCY  ReceiptDeliveryType = "CY"  // Container Yard
	ReceiptDeliveryTypeSD  ReceiptDeliveryType = "SD"  // Store Door
	ReceiptDeliveryTypeCFS ReceiptDeliveryType = "CFS" // Container Freight Station
)

// CargoMovementType refers to the shipment term at loading/unloading
type CargoMovementType string

const (
	CargoMovementTypeFCL CargoMovementType = "FCL" // Full Container Load
	CargoMovementTypeLCL CargoMovementType = "LCL" // Less than Container Load
)

// FreightPaymentTermCode indicates whether freight is prepaid or collect
type FreightPaymentTermCode string

const (
	FreightPaymentTermPre FreightPaymentTermCode = "PRE" // Prepaid
	FreightPaymentTermCol FreightPaymentTermCode = "COL" // Collect
)

// CarrierCodeListProvider indicates the source of carrier codes
type CarrierCodeListProvider string

const (
	CarrierCodeListProviderSMDG  CarrierCodeListProvider = "SMDG"  // Ship Message Design Group
	CarrierCodeListProviderNMFTA CarrierCodeListProvider = "NMFTA" // National Motor Freight Traffic Association
)

// TransportDocumentTypeCode specifies the type of transport document
type TransportDocumentTypeCode string

const (
	TransportDocumentTypeBOL TransportDocumentTypeCode = "BOL" // Bill of Lading
	TransportDocumentTypeSWB TransportDocumentTypeCode = "SWB" // Sea Waybill
)

// TemperatureUnit represents temperature measurement unit
type TemperatureUnit string

const (
	TemperatureUnitCel TemperatureUnit = "CEL" // Celsius
	TemperatureUnitFah TemperatureUnit = "FAH" // Fahrenheit
)

// WeightUnit represents weight measurement unit
type WeightUnit string

const (
	WeightUnitKGM WeightUnit = "KGM" // Kilograms
	WeightUnitLBR WeightUnit = "LBR" // Pounds
)

// VolumeUnit represents volume measurement unit
type VolumeUnit string

const (
	VolumeUnitMTQ VolumeUnit = "MTQ" // Cubic meter
	VolumeUnitFTQ VolumeUnit = "FTQ" // Cubic foot
)

// AirExchangeUnit represents air exchange measurement unit
type AirExchangeUnit string

const (
	AirExchangeUnitMQH AirExchangeUnit = "MQH" // Cubic metre per hour
	AirExchangeUnitFQH AirExchangeUnit = "FQH" // Cubic foot per hour
)

// LocationTypeCode represents the type of location in shipment
type LocationTypeCode string

const (
	LocationTypePre LocationTypeCode = "PRE" // Place of Receipt
	LocationTypePol LocationTypeCode = "POL" // Port of Loading
	LocationTypePod LocationTypeCode = "POD" // Port of Discharge
	LocationTypePde LocationTypeCode = "PDE" // Place of Delivery
	LocationTypePcf LocationTypeCode = "PCF" // Pre-carriage From
	LocationTypeOir LocationTypeCode = "OIR" // Onward In-land Routing
	LocationTypeOri LocationTypeCode = "ORI" // Origin of goods
	LocationTypeIel LocationTypeCode = "IEL" // Container intermediate export stop off location
	LocationTypePtp LocationTypeCode = "PTP" // Prohibited transshipment port
	LocationTypeRtp LocationTypeCode = "RTP" // Requested transshipment port
	LocationTypeFcd LocationTypeCode = "FCD" // Full container drop-off location
	LocationTypeRou LocationTypeCode = "ROU" // Routing Reference
)

// ModeOfTransport represents the mode of transport
type ModeOfTransport string

const (
	ModeOfTransportVessel     ModeOfTransport = "VESSEL"
	ModeOfTransportRail       ModeOfTransport = "RAIL"
	ModeOfTransportTruck      ModeOfTransport = "TRUCK"
	ModeOfTransportBarge      ModeOfTransport = "BARGE"
	ModeOfTransportRailTruck  ModeOfTransport = "RAIL_TRUCK"
	ModeOfTransportBargeTruck ModeOfTransport = "BARGE_TRUCK"
	ModeOfTransportBargeRail  ModeOfTransport = "BARGE_RAIL"
	ModeOfTransportMultimodal ModeOfTransport = "MULTIMODAL"
)

// TransportPlanStage represents a stage in the transport plan
type TransportPlanStage string

const (
	TransportPlanStagePRC TransportPlanStage = "PRC" // Pre-Carriage
	TransportPlanStageMNC TransportPlanStage = "MNC" // Main Carriage Transport
	TransportPlanStageONC TransportPlanStage = "ONC" // On-Carriage Transport
)

// ReferenceTypeCode represents the type of reference
type ReferenceTypeCode string

const (
	ReferenceTypeCR  ReferenceTypeCode = "CR"  // Customer's Reference
	ReferenceTypeECR ReferenceTypeCode = "ECR" // Empty container release reference
	ReferenceTypeAKG ReferenceTypeCode = "AKG" // Vehicle Identification Number
	ReferenceTypeAEF ReferenceTypeCode = "AEF" // Customer Load Reference
)

// PartyFunction represents the role of a party
type PartyFunction string

const (
	PartyFunctionDDR PartyFunction = "DDR" // Consignor's freight forwarder
	PartyFunctionDDS PartyFunction = "DDS" // Consignee's freight forwarder
	PartyFunctionCOW PartyFunction = "COW" // Invoice payer on behalf of the consignor
	PartyFunctionCOX PartyFunction = "COX" // Invoice payer on behalf of the consignee
	PartyFunctionN1  PartyFunction = "N1"  // First Notify Party
	PartyFunctionN2  PartyFunction = "N2"  // Second Notify Party
	PartyFunctionNI  PartyFunction = "NI"  // Other Notify Party
	PartyFunctionNAC PartyFunction = "NAC" // Named Account Customer
)

// HTTPMethod represents HTTP methods
type HTTPMethod string

const (
	HTTPMethodGet    HTTPMethod = "GET"
	HTTPMethodHead   HTTPMethod = "HEAD"
	HTTPMethodPost   HTTPMethod = "POST"
	HTTPMethodPut    HTTPMethod = "PUT"
	HTTPMethodDelete HTTPMethod = "DELETE"
	HTTPMethodOption HTTPMethod = "OPTION"
	HTTPMethodPatch  HTTPMethod = "PATCH"
)
