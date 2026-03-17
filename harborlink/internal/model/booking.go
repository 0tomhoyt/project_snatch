package model

import (
	"time"
)

// Booking represents the core booking entity in DCSA
type Booking struct {
	CarrierBookingRequestReference string                   `json:"carrierBookingRequestReference,omitempty"`
	CarrierBookingReference        string                   `json:"carrierBookingReference,omitempty"`
	BookingStatus                  BookingStatus            `json:"bookingStatus"`
	AmendedBookingStatus           AmendedBookingStatus     `json:"amendedBookingStatus,omitempty"`
	BookingCancellationStatus      BookingCancellationStatus `json:"bookingCancellationStatus,omitempty"`

	// Service details
	ReceiptTypeAtOrigin            ReceiptDeliveryType      `json:"receiptTypeAtOrigin"`
	DeliveryTypeAtDestination      ReceiptDeliveryType      `json:"deliveryTypeAtDestination"`
	CargoMovementTypeAtOrigin      CargoMovementType        `json:"cargoMovementTypeAtOrigin"`
	CargoMovementTypeAtDestination CargoMovementType        `json:"cargoMovementTypeAtDestination"`
	ServiceContractReference       string                   `json:"serviceContractReference,omitempty"`
	FreightPaymentTermCode         FreightPaymentTermCode   `json:"freightPaymentTermCode,omitempty"`
	OriginChargesPaymentTerm       *OriginChargesPaymentTerm  `json:"originChargesPaymentTerm,omitempty"`
	DestinationChargesPaymentTerm  *DestinationChargesPaymentTerm `json:"destinationChargesPaymentTerm,omitempty"`
	ContractQuotationReference     string                   `json:"contractQuotationReference,omitempty"`

	// Vessel and voyage details
	Vessel                         *Vessel                  `json:"vessel,omitempty"`
	CarrierServiceName             string                   `json:"carrierServiceName,omitempty"`
	CarrierServiceCode             string                   `json:"carrierServiceCode,omitempty"`
	UniversalServiceReference      string                   `json:"universalServiceReference,omitempty"`
	CarrierExportVoyageNumber      string                   `json:"carrierExportVoyageNumber,omitempty"`
	UniversalExportVoyageReference string                   `json:"universalExportVoyageReference,omitempty"`
	RoutingReference               string                   `json:"routingReference,omitempty"`

	// Carrier details
	CarrierCode             string                `json:"carrierCode,omitempty"`
	CarrierCodeListProvider CarrierCodeListProvider `json:"carrierCodeListProvider,omitempty"`

	// Value and currency
	DeclaredValue           float64 `json:"declaredValue,omitempty"`
	DeclaredValueCurrency   string  `json:"declaredValueCurrency,omitempty"`

	// Load details
	IsPartialLoadAllowed         bool       `json:"isPartialLoadAllowed,omitempty"`
	IsExportDeclarationRequired  bool       `json:"isExportDeclarationRequired,omitempty"`
	ExportDeclarationReference   string     `json:"exportDeclarationReference,omitempty"`

	// Dates
	ExpectedDepartureDate                     *time.Time `json:"expectedDepartureDate,omitempty"`
	ExpectedDepartureFromPlaceOfReceiptDate   *time.Time `json:"expectedDepartureFromPlaceOfReceiptDate,omitempty"`
	ExpectedArrivalAtPlaceOfDeliveryStartDate *time.Time `json:"expectedArrivalAtPlaceOfDeliveryStartDate,omitempty"`
	ExpectedArrivalAtPlaceOfDeliveryEndDate   *time.Time `json:"expectedArrivalAtPlaceOfDeliveryEndDate,omitempty"`

	// Transport document
	TransportDocumentTypeCode           TransportDocumentTypeCode `json:"transportDocumentTypeCode,omitempty"`
	TransportDocumentReference          string                    `json:"transportDocumentReference,omitempty"` // Deprecated
	TransportDocumentReferences         []string                  `json:"transportDocumentReferences,omitempty"`
	RequestedNumberOfTransportDocuments int                       `json:"requestedNumberOfTransportDocuments,omitempty"`

	// Booking channel
	BookingChannelReference string `json:"bookingChannelReference,omitempty"`

	// Incoterms
	IncoTerms string `json:"incoTerms,omitempty"`

	// Equipment
	IsEquipmentSubstitutionAllowed bool `json:"isEquipmentSubstitutionAllowed"`

	// Terms and conditions
	TermsAndConditions string `json:"termsAndConditions,omitempty"`

	// Invoice and B/L issue location
	InvoicePayableAt *InvoicePayableAt `json:"invoicePayableAt,omitempty"`
	PlaceOfBLIssue   *PlaceOfBLIssue   `json:"placeOfBLIssue,omitempty"`

	// References
	References        []Reference        `json:"references,omitempty"`
	CustomsReferences []CustomsReference `json:"customsReferences,omitempty"`

	// Parties
	DocumentParties    *DocumentParties        `json:"documentParties,omitempty"`
	PartyContactDetails []PartyContactDetail   `json:"partyContactDetails,omitempty"`

	// Locations
	ShipmentLocations []ShipmentLocation `json:"shipmentLocations"`

	// Transport modes
	RequestedPreCarriageModeOfTransport  ModeOfTransport `json:"requestedPreCarriageModeOfTransport,omitempty"`
	RequestedOnCarriageModeOfTransport   ModeOfTransport `json:"requestedOnCarriageModeOfTransport,omitempty"`

	// Equipment
	RequestedEquipments []RequestedEquipment `json:"requestedEquipments"`
	ConfirmedEquipments []ConfirmedEquipment `json:"confirmedEquipments,omitempty"`

	// Transport plan (for confirmed bookings)
	TransportPlan []Transport `json:"transportPlan,omitempty"`

	// Cut-off times (for confirmed bookings)
	ShipmentCutOffTimes []ShipmentCutOffTime `json:"shipmentCutOffTimes,omitempty"`

	// Advance manifest filings
	AdvanceManifestFilings []AdvanceManifestFiling `json:"advanceManifestFilings,omitempty"`

	// Charges
	Charges []Charge `json:"charges,omitempty"`

	// Carrier clauses
	CarrierClauses []string `json:"carrierClauses,omitempty"`

	// Feedbacks
	Feedbacks []Feedback `json:"feedbacks,omitempty"`
}

// CreateBooking represents a booking creation request
type CreateBooking struct {
	ReceiptTypeAtOrigin            ReceiptDeliveryType      `json:"receiptTypeAtOrigin"`
	DeliveryTypeAtDestination      ReceiptDeliveryType      `json:"deliveryTypeAtDestination"`
	CargoMovementTypeAtOrigin      CargoMovementType        `json:"cargoMovementTypeAtOrigin"`
	CargoMovementTypeAtDestination CargoMovementType        `json:"cargoMovementTypeAtDestination"`
	ServiceContractReference       string                   `json:"serviceContractReference,omitempty"`
	FreightPaymentTermCode         FreightPaymentTermCode   `json:"freightPaymentTermCode,omitempty"`
	OriginChargesPaymentTerm       *OriginChargesPaymentTerm  `json:"originChargesPaymentTerm,omitempty"`
	DestinationChargesPaymentTerm  *DestinationChargesPaymentTerm `json:"destinationChargesPaymentTerm,omitempty"`
	ContractQuotationReference     string                   `json:"contractQuotationReference,omitempty"`

	Vessel                         *Vessel                  `json:"vessel,omitempty"`
	CarrierServiceCode             string                   `json:"carrierServiceCode,omitempty"`
	CarrierExportVoyageNumber      string                   `json:"carrierExportVoyageNumber,omitempty"`
	IsPartialLoadAllowed           bool                     `json:"isPartialLoadAllowed,omitempty"`
	IsExportDeclarationRequired    bool                     `json:"isExportDeclarationRequired,omitempty"`
	ExpectedDepartureDate          *time.Time               `json:"expectedDepartureDate,omitempty"`
	IncoTerms                      string                   `json:"incoTerms,omitempty"`
	IsEquipmentSubstitutionAllowed bool                     `json:"isEquipmentSubstitutionAllowed"`

	References         []Reference         `json:"references,omitempty"`
	DocumentParties    *DocumentPartiesReq `json:"documentParties"`
	PartyContactDetails []PartyContactDetail `json:"partyContactDetails,omitempty"`
	ShipmentLocations  []ShipmentLocation  `json:"shipmentLocations"`
	RequestedEquipments []RequestedEquipmentShipper `json:"requestedEquipments"`
}

// CreateBookingResponse represents the response after creating a booking
type CreateBookingResponse struct {
	CarrierBookingRequestReference string `json:"carrierBookingRequestReference"`
}

// UpdateBooking represents a booking update request
type UpdateBooking struct {
	Booking *Booking `json:"booking,omitempty"`
	Reason  string   `json:"reason,omitempty"`
}

// CancelBookingRequest represents a booking cancellation request
type CancelBookingRequest struct {
	BookingStatus             BookingStatus            `json:"bookingStatus,omitempty"`
	AmendedBookingStatus      AmendedBookingStatus     `json:"amendedBookingStatus,omitempty"`
	BookingCancellationStatus BookingCancellationStatus `json:"bookingCancellationStatus,omitempty"`
	Reason                    string                   `json:"reason,omitempty"`
}

// OriginChargesPaymentTerm represents payment terms for origin charges
type OriginChargesPaymentTerm struct {
	HaulageChargesPaymentTermCode FreightPaymentTermCode `json:"haulageChargesPaymentTermCode,omitempty"`
	PortChargesPaymentTermCode    FreightPaymentTermCode `json:"portChargesPaymentTermCode,omitempty"`
	OtherChargesPaymentTermCode   FreightPaymentTermCode `json:"otherChargesPaymentTermCode,omitempty"`
}

// DestinationChargesPaymentTerm represents payment terms for destination charges
type DestinationChargesPaymentTerm struct {
	HaulageChargesPaymentTermCode FreightPaymentTermCode `json:"haulageChargesPaymentTermCode,omitempty"`
	PortChargesPaymentTermCode    FreightPaymentTermCode `json:"portChargesPaymentTermCode,omitempty"`
	OtherChargesPaymentTermCode   FreightPaymentTermCode `json:"otherChargesPaymentTermCode,omitempty"`
}

// InvoicePayableAt represents where invoice payment takes place
type InvoicePayableAt struct {
	UNLocationCode string `json:"UNLocationCode"`
}

// PlaceOfBLIssue represents where the Bill of Lading will be issued
type PlaceOfBLIssue struct {
	LocationName   string `json:"locationName,omitempty"`
	UNLocationCode string `json:"UNLocationCode,omitempty"`
	CountryCode    string `json:"countryCode,omitempty"`
}

// Feedback represents feedback information
type Feedback struct {
	FeedbackType    string `json:"feedbackType,omitempty"`
	Property        string `json:"property,omitempty"`
	OriginalValue   string `json:"originalValue,omitempty"`
	ReplacedValue   string `json:"replacedValue,omitempty"`
	FeedbackMessage string `json:"feedbackMessage,omitempty"`
}
