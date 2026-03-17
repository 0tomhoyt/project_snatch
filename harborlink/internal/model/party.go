package model

// DocumentParties represents all parties with associated roles
type DocumentParties struct {
	BookingAgent         *BookingAgent         `json:"bookingAgent,omitempty"`
	Shipper              *Shipper              `json:"shipper,omitempty"`
	Consignee            *Consignee            `json:"consignee,omitempty"`
	ServiceContractOwner *ServiceContractOwner `json:"serviceContractOwner,omitempty"`
	CarrierBookingOffice *CarrierBookingOffice `json:"carrierBookingOffice,omitempty"`
	IssueTo              *IssueToParty         `json:"issueTo,omitempty"`
	Other                []OtherDocumentParty  `json:"other,omitempty"`
}

// DocumentPartiesReq represents document parties for booking request
type DocumentPartiesReq struct {
	BookingAgent         *BookingAgent         `json:"bookingAgent"`
	Shipper              *Shipper              `json:"shipper,omitempty"`
	Consignee            *Consignee            `json:"consignee,omitempty"`
	ServiceContractOwner *ServiceContractOwner `json:"serviceContractOwner,omitempty"`
	CarrierBookingOffice *CarrierBookingOffice `json:"carrierBookingOffice,omitempty"`
	IssueTo              *IssueToParty         `json:"issueTo,omitempty"`
	Other                []OtherDocumentParty  `json:"other,omitempty"`
}

// Party represents a company or legal entity
type Party struct {
	PartyName           string               `json:"partyName"`
	Address             *PartyAddress        `json:"address,omitempty"`
	IdentifyingCodes    []IdentifyingCode    `json:"identifyingCodes,omitempty"`
	TaxLegalReferences  []TaxLegalReference  `json:"taxLegalReferences,omitempty"`
	PartyContactDetails []PartyContactDetail `json:"partyContactDetails,omitempty"`
	Reference           string               `json:"reference,omitempty"`
}

// BookingAgent represents the party placing the booking request
type BookingAgent struct {
	PartyName           string               `json:"partyName"`
	Address             *PartyAddress        `json:"address,omitempty"`
	PartyContactDetails []PartyContactDetail `json:"partyContactDetails,omitempty"`
	IdentifyingCodes    []IdentifyingCode    `json:"identifyingCodes,omitempty"`
	TaxLegalReferences  []TaxLegalReference  `json:"taxLegalReferences,omitempty"`
	Reference           string               `json:"reference,omitempty"`
}

// Shipper represents the party by whom goods are delivered to carrier
type Shipper struct {
	PartyName               string               `json:"partyName"`
	Address                 *PartyAddress        `json:"address,omitempty"`
	PartyContactDetails     []PartyContactDetail `json:"partyContactDetails,omitempty"`
	IdentifyingCodes        []IdentifyingCode    `json:"identifyingCodes,omitempty"`
	TaxLegalReferences      []TaxLegalReference  `json:"taxLegalReferences,omitempty"`
	Reference               string               `json:"reference,omitempty"`
	PurchaseOrderReferences []string             `json:"purchaseOrderReferences,omitempty"`
}

// Consignee represents the party to which goods are consigned
type Consignee struct {
	PartyName               string               `json:"partyName"`
	Address                 *PartyAddress        `json:"address,omitempty"`
	PartyContactDetails     []PartyContactDetail `json:"partyContactDetails,omitempty"`
	IdentifyingCodes        []IdentifyingCode    `json:"identifyingCodes,omitempty"`
	TaxLegalReferences      []TaxLegalReference  `json:"taxLegalReferences,omitempty"`
	Reference               string               `json:"reference,omitempty"`
	PurchaseOrderReferences []string             `json:"purchaseOrderReferences,omitempty"`
}

// ServiceContractOwner represents the party signing the service contract
type ServiceContractOwner struct {
	PartyName           string               `json:"partyName"`
	Address             *PartyAddress        `json:"address,omitempty"`
	PartyContactDetails []PartyContactDetail `json:"partyContactDetails,omitempty"`
	IdentifyingCodes    []IdentifyingCode    `json:"identifyingCodes,omitempty"`
	TaxLegalReferences  []TaxLegalReference  `json:"taxLegalReferences,omitempty"`
	Reference           string               `json:"reference,omitempty"`
}

// CarrierBookingOffice represents the carrier office processing the booking
type CarrierBookingOffice struct {
	PartyName           string               `json:"partyName"`
	UNLocationCode      string               `json:"UNLocationCode"`
	Address             *Address             `json:"address,omitempty"`
	PartyContactDetails []PartyContactDetail `json:"partyContactDetails,omitempty"`
}

// IssueToParty represents the party that receives the original Bill of Lading
type IssueToParty struct {
	PartyName          string              `json:"partyName"`
	SendToPlatform     string              `json:"sendToPlatform,omitempty"`
	IdentifyingCodes   []IdentifyingCode   `json:"identifyingCodes,omitempty"`
	TaxLegalReferences []TaxLegalReference `json:"taxLegalReferences,omitempty"`
}

// OtherDocumentParty represents additional document parties
type OtherDocumentParty struct {
	Party         Party         `json:"party"`
	PartyFunction PartyFunction `json:"partyFunction"`
}

// PartyAddress represents an address for a party
type PartyAddress struct {
	Street         string `json:"street"`
	StreetNumber   string `json:"streetNumber,omitempty"`
	Floor          string `json:"floor,omitempty"`
	PostCode       string `json:"postCode,omitempty"`
	POBox          string `json:"POBox,omitempty"`
	City           string `json:"city"`
	UNLocationCode string `json:"UNLocationCode,omitempty"`
	StateRegion    string `json:"stateRegion,omitempty"`
	CountryCode    string `json:"countryCode"`
}

// PartyContactDetail represents contact details for a party
type PartyContactDetail struct {
	Name  string `json:"name"`
	Phone string `json:"phone,omitempty"`
	Email string `json:"email,omitempty"`
}

// IdentifyingCode represents a code that uniquely identifies a party
type IdentifyingCode struct {
	CodeListProvider string `json:"codeListProvider"`
	PartyCode        string `json:"partyCode"`
	CodeListName     string `json:"codeListName,omitempty"`
}

// TaxLegalReference represents tax and legal references for a party
type TaxLegalReference struct {
	Type        string `json:"type"`
	CountryCode string `json:"countryCode"`
	Value       string `json:"value"`
}
