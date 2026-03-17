package model

import (
	"time"
)

// RequestedEquipment represents equipment requested in a booking
type RequestedEquipment struct {
	ISOEquipmentCode           string                    `json:"ISOEquipmentCode"`
	Units                      int                       `json:"units"`
	ContainerPositionings      []ContainerPositioning    `json:"containerPositionings,omitempty"`
	EmptyContainerPickup       *EmptyContainerPickup     `json:"emptyContainerPickup,omitempty"`
	OriginEmptyContainerPickup *OriginEmptyContainerPickup `json:"originEmptyContainerPickup,omitempty"`
	FullContainerPickupDateTime *time.Time               `json:"fullContainerPickupDateTime,omitempty"`
	EquipmentReferences        []string                  `json:"equipmentReferences,omitempty"`
	TareWeight                 *TareWeight               `json:"tareWeight,omitempty"`
	CargoGrossWeight           *CargoGrossWeightReq      `json:"cargoGrossWeight,omitempty"`
	IsShipperOwned             bool                      `json:"isShipperOwned"`
	IsNonOperatingReefer       bool                      `json:"isNonOperatingReefer,omitempty"`
	ActiveReeferSettings       *ActiveReeferSettings     `json:"activeReeferSettings,omitempty"`
	References                 []Reference               `json:"references,omitempty"`
	CustomsReferences          []CustomsReference        `json:"customsReferences,omitempty"`
	Commodities                []Commodity               `json:"commodities,omitempty"`
}

// RequestedEquipmentShipper represents equipment requested by shipper
type RequestedEquipmentShipper struct {
	ISOEquipmentCode           string                    `json:"ISOEquipmentCode"`
	Units                      int                       `json:"units"`
	ContainerPositionings      []ContainerPositioning    `json:"containerPositionings,omitempty"`
	EmptyContainerPickup       *EmptyContainerPickup     `json:"emptyContainerPickup,omitempty"`
	OriginEmptyContainerPickup *OriginEmptyContainerPickup `json:"originEmptyContainerPickup,omitempty"`
	FullContainerPickupDateTime *time.Time               `json:"fullContainerPickupDateTime,omitempty"`
	EquipmentReferences        []string                  `json:"equipmentReferences,omitempty"`
	TareWeight                 *TareWeight               `json:"tareWeight,omitempty"`
	CargoGrossWeight           *CargoGrossWeightReq      `json:"cargoGrossWeight,omitempty"`
	IsShipperOwned             bool                      `json:"isShipperOwned"`
	IsNonOperatingReefer       bool                      `json:"isNonOperatingReefer,omitempty"`
	ActiveReeferSettings       *ActiveReeferSettings     `json:"activeReeferSettings,omitempty"`
	References                 []ReferenceShipper        `json:"references,omitempty"`
	CustomsReferences          []CustomsReference        `json:"customsReferences,omitempty"`
	Commodities                []CommodityShipper        `json:"commodities,omitempty"`
}

// ConfirmedEquipment represents confirmed equipment for a booking
type ConfirmedEquipment struct {
	ISOEquipmentCode           string                         `json:"ISOEquipmentCode"`
	Units                      int                            `json:"units"`
	ContainerPositionings      []ContainerPositioningEstimated `json:"containerPositionings,omitempty"`
	EmptyContainerPickup       *EmptyContainerPickup          `json:"emptyContainerPickup,omitempty"`
	OriginEmptyContainerPickup *OriginEmptyContainerPickup    `json:"originEmptyContainerPickup,omitempty"`
}

// ContainerPositioning represents location and time for container positioning
type ContainerPositioning struct {
	DateTime *time.Time                  `json:"dateTime,omitempty"`
	Location ContainerPositioningLocation `json:"location"`
}

// ContainerPositioningEstimated represents estimated container positioning
type ContainerPositioningEstimated struct {
	EstimatedDateTime *time.Time                  `json:"estimatedDateTime,omitempty"`
	Location          ContainerPositioningLocation `json:"location"`
}

// EmptyContainerPickup represents details for empty container pickup
type EmptyContainerPickup struct {
	DateTime            *time.Time                     `json:"dateTime,omitempty"`
	DepotReleaseLocation *EmptyContainerDepotReleaseLocation `json:"depotReleaseLocation,omitempty"`
}

// OriginEmptyContainerPickup represents origin empty container pickup details
type OriginEmptyContainerPickup struct {
	DateTime            *time.Time                     `json:"dateTime,omitempty"`
	DepotReleaseLocation *EmptyContainerDepotReleaseLocation `json:"depotReleaseLocation,omitempty"`
}

// TareWeight represents the weight of an empty container
type TareWeight struct {
	Value float64    `json:"value"`
	Unit  WeightUnit `json:"unit"`
}

// CargoGrossWeightReq represents cargo gross weight for request
type CargoGrossWeightReq struct {
	Value float64    `json:"value"`
	Unit  WeightUnit `json:"unit"`
}

// CargoGrossWeight represents cargo gross weight
type CargoGrossWeight struct {
	Value float64    `json:"value"`
	Unit  WeightUnit `json:"unit"`
}

// CargoGrossVolume represents cargo gross volume
type CargoGrossVolume struct {
	Value float64    `json:"value"`
	Unit  VolumeUnit `json:"unit"`
}

// CargoNetWeight represents cargo net weight
type CargoNetWeight struct {
	Value float64    `json:"value"`
	Unit  WeightUnit `json:"unit"`
}

// CargoNetVolume represents cargo net volume
type CargoNetVolume struct {
	Value float64    `json:"value"`
	Unit  VolumeUnit `json:"unit"`
}

// ActiveReeferSettings represents reefer container settings
type ActiveReeferSettings struct {
	TemperatureSetpoint             float64        `json:"temperatureSetpoint"`
	TemperatureUnit                 TemperatureUnit `json:"temperatureUnit"`
	O2Setpoint                      float64        `json:"o2Setpoint,omitempty"`
	CO2Setpoint                     float64        `json:"co2Setpoint,omitempty"`
	HumiditySetpoint                float64        `json:"humiditySetpoint,omitempty"`
	AirExchangeSetpoint             float64        `json:"airExchangeSetpoint,omitempty"`
	AirExchangeUnit                 AirExchangeUnit `json:"airExchangeUnit,omitempty"`
	IsVentilationOpen               bool           `json:"isVentilationOpen,omitempty"`
	IsDrainholesOpen                bool           `json:"isDrainholesOpen,omitempty"`
	IsBulbMode                      bool           `json:"isBulbMode,omitempty"`
	IsColdTreatmentRequired         bool           `json:"isColdTreatmentRequired,omitempty"`
	IsControlledAtmosphereRequired  bool           `json:"isControlledAtmosphereRequired,omitempty"`
	IsPreCoolingRequired            bool           `json:"isPreCoolingRequired,omitempty"`
	IsGeneratorSetRequired          bool           `json:"isGeneratorSetRequired,omitempty"`
}
