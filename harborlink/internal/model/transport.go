package model

import (
	"time"
)

// Transport represents a transport leg in the transport plan
type Transport struct {
	TransportPlanStage             TransportPlanStage `json:"transportPlanStage"`
	TransportPlanStageSequenceNumber int              `json:"transportPlanStageSequenceNumber"`
	LoadLocation                   LoadLocation       `json:"loadLocation"`
	DischargeLocation              DischargeLocation  `json:"dischargeLocation"`
	PlannedDepartureDate           *time.Time         `json:"plannedDepartureDate"`
	PlannedArrivalDate             *time.Time         `json:"plannedArrivalDate"`
	ModeOfTransport                ModeOfTransport    `json:"modeOfTransport,omitempty"`
	VesselName                     string             `json:"vesselName,omitempty"`
	VesselIMONumber                string             `json:"vesselIMONumber,omitempty"`
	CarrierServiceCode             string             `json:"carrierServiceCode,omitempty"`
	UniversalServiceReference      string             `json:"universalServiceReference,omitempty"`
	CarrierExportVoyageNumber      string             `json:"carrierExportVoyageNumber,omitempty"`
	UniversalExportVoyageReference string             `json:"universalExportVoyageReference,omitempty"`
}

// ShipmentCutOffTime represents cut-off times for a shipment
type ShipmentCutOffTime struct {
	CutOffDateTime    *time.Time      `json:"cutOffDateTime,omitempty"`
	CutOffTimeTypeCode CutOffTimeTypeCode `json:"cutOffTimeTypeCode"`
}

// CutOffTimeTypeCode represents the type of cut-off time
type CutOffTimeTypeCode string

const (
	CutOffTimeTypeCOD CutOffTimeTypeCode = "COD" // Cargo on board deadline
	CutOffTimeTypeCSD CutOffTimeTypeCode = "CSD" // Cargo stacking deadline
	CutOffTimeTypeVGM CutOffTimeTypeCode = "VGM" // VGM submission deadline
	CutOffTimeTypeSIC CutOffTimeTypeCode = "SIC" // Shipping instruction deadline
	CutOffTimeTypeDOC CutOffTimeTypeCode = "DOC" // Documentation deadline
	CutOffTimeTypeCUS CutOffTimeTypeCode = "CUS" // Customs clearance deadline
)

// AdvanceManifestFiling represents advance manifest filing info
type AdvanceManifestFiling struct {
	CountryCode                 string    `json:"countryCode"`
	AdvanceManifestFilingType   string    `json:"advanceManifestFilingType,omitempty"`
	IsRequired                  bool      `json:"isRequired,omitempty"`
	FilingReference             string    `json:"filingReference,omitempty"`
	ExpectedFilingDateTime      *time.Time `json:"expectedFilingDateTime,omitempty"`
	ExpectedApprovalDateTime    *time.Time `json:"expectedApprovalDateTime,omitempty"`
	ActualFilingDateTime        *time.Time `json:"actualFilingDateTime,omitempty"`
	ActualApprovalDateTime      *time.Time `json:"actualApprovalDateTime,omitempty"`
	StatusCode                  string    `json:"statusCode,omitempty"`
	StatusDescription           string    `json:"statusDescription,omitempty"`
}

// Charge represents a charge for the booking
type Charge struct {
	ChargeType           string          `json:"chargeType,omitempty"`
	CurrencyAmount       *CurrencyAmount `json:"currencyAmount,omitempty"`
	PaymentTermCode      FreightPaymentTermCode `json:"paymentTermCode,omitempty"`
	ChargeParty          *ChargeParty    `json:"chargeParty,omitempty"`
	ChargeReference      string          `json:"chargeReference,omitempty"`
	UnitOfMeasure        string          `json:"unitOfMeasure,omitempty"`
	UnitPrice            float64         `json:"unitPrice,omitempty"`
	Quantity             float64         `json:"quantity,omitempty"`
	TaxAmount            float64         `json:"taxAmount,omitempty"`
	TotalAmount          float64         `json:"totalAmount,omitempty"`
}

// CurrencyAmount represents a monetary value with currency
type CurrencyAmount struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// ChargeParty represents the party responsible for a charge
type ChargeParty struct {
	PartyFunction PartyFunction `json:"partyFunction,omitempty"`
	Party         *Party        `json:"party,omitempty"`
}
