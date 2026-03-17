package model

// Location represents a geographical location
type Location struct {
	LocationName   string         `json:"locationName,omitempty"`
	Address        *Address       `json:"address,omitempty"`
	Facility       *Facility      `json:"facility,omitempty"`
	UNLocationCode string         `json:"UNLocationCode,omitempty"`
	GeoCoordinate  *GeoCoordinate `json:"geoCoordinate,omitempty"`
}

// Address represents a street address
type Address struct {
	Street         string `json:"street,omitempty"`
	StreetNumber   string `json:"streetNumber,omitempty"`
	Floor          string `json:"floor,omitempty"`
	PostCode       string `json:"postCode,omitempty"`
	POBox          string `json:"POBox,omitempty"`
	City           string `json:"city,omitempty"`
	UNLocationCode string `json:"UNLocationCode,omitempty"`
	StateRegion    string `json:"stateRegion,omitempty"`
	CountryCode    string `json:"countryCode,omitempty"`
}

// Facility represents a facility at a location
type Facility struct {
	FacilityCode             string `json:"facilityCode"`
	FacilityCodeListProvider string `json:"facilityCodeListProvider,omitempty"`
}

// GeoCoordinate represents geographic coordinates
type GeoCoordinate struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// ShipmentLocation maps the relationship between shipment and location
type ShipmentLocation struct {
	Location         Location         `json:"location"`
	LocationTypeCode LocationTypeCode `json:"locationTypeCode"`
}

// LoadLocation represents a loading location in transport
type LoadLocation struct {
	LocationName   string    `json:"locationName,omitempty"`
	Address        *Address  `json:"address,omitempty"`
	Facility       *Facility `json:"facility,omitempty"`
	UNLocationCode string    `json:"UNLocationCode,omitempty"`
}

// DischargeLocation represents a discharge location in transport
type DischargeLocation struct {
	LocationName   string    `json:"locationName,omitempty"`
	Address        *Address  `json:"address,omitempty"`
	Facility       *Facility `json:"facility,omitempty"`
	UNLocationCode string    `json:"UNLocationCode,omitempty"`
}

// ContainerPositioningLocation represents location for container positioning
type ContainerPositioningLocation struct {
	LocationName   string         `json:"locationName,omitempty"`
	Address        *Address       `json:"address,omitempty"`
	Facility       *Facility      `json:"facility,omitempty"`
	UNLocationCode string         `json:"UNLocationCode,omitempty"`
	GeoCoordinate  *GeoCoordinate `json:"geoCoordinate,omitempty"`
}

// EmptyContainerDepotReleaseLocation represents the depot release location
type EmptyContainerDepotReleaseLocation struct {
	LocationName   string         `json:"locationName,omitempty"`
	Address        *Address       `json:"address,omitempty"`
	Facility       *Facility      `json:"facility,omitempty"`
	UNLocationCode string         `json:"UNLocationCode,omitempty"`
	GeoCoordinate  *GeoCoordinate `json:"geoCoordinate,omitempty"`
}
