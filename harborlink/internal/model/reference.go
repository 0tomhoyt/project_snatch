package model

// Reference represents a reference provided by shipper or freight forwarder
type Reference struct {
	Type  ReferenceTypeCode `json:"type"`
	Value string            `json:"value"`
}

// ReferenceShipper represents references specific to shippers
type ReferenceShipper struct {
	Type  ReferenceTypeCode `json:"type"`
	Value string            `json:"value"`
}

// CustomsReference represents customs-related references
type CustomsReference struct {
	Type        string   `json:"type"`
	CountryCode string   `json:"countryCode"`
	Values      []string `json:"values"`
}
