package model

// Commodity represents type of goods
type Commodity struct {
	CommoditySubReference   string                  `json:"commoditySubReference,omitempty"`
	CommodityType           string                  `json:"commodityType"`
	HSCodes                 []string                `json:"HSCodes,omitempty"`
	NationalCommodityCodes  []NationalCommodityCode `json:"nationalCommodityCodes,omitempty"`
	CargoGrossWeight        *CargoGrossWeight       `json:"cargoGrossWeight,omitempty"`
	CargoGrossVolume        *CargoGrossVolume       `json:"cargoGrossVolume,omitempty"`
	CargoNetWeight          *CargoNetWeight         `json:"cargoNetWeight,omitempty"`
	CargoNetVolume          *CargoNetVolume         `json:"cargoNetVolume,omitempty"`
	ExportLicense           *ExportLicense          `json:"exportLicense,omitempty"`
	ImportLicense           *ImportLicense          `json:"importLicense,omitempty"`
	OuterPackaging          *OuterPackaging         `json:"outerPackaging,omitempty"`
	References              []Reference             `json:"references,omitempty"`
	CustomsReferences       []CustomsReference      `json:"customsReferences,omitempty"`
}

// CommodityShipper represents commodity info from shipper
type CommodityShipper struct {
	CommodityType          string                  `json:"commodityType"`
	HSCodes                []string                `json:"HSCodes,omitempty"`
	NationalCommodityCodes []NationalCommodityCode `json:"nationalCommodityCodes,omitempty"`
	CargoGrossWeight       *CargoGrossWeight       `json:"cargoGrossWeight,omitempty"`
	CargoGrossVolume       *CargoGrossVolume       `json:"cargoGrossVolume,omitempty"`
	CargoNetWeight         *CargoNetWeight         `json:"cargoNetWeight,omitempty"`
	CargoNetVolume         *CargoNetVolume         `json:"cargoNetVolume,omitempty"`
	ExportLicense          *ExportLicense          `json:"exportLicense,omitempty"`
	ImportLicense          *ImportLicense          `json:"importLicense,omitempty"`
	OuterPackaging         *OuterPackaging         `json:"outerPackaging,omitempty"`
	References             []ReferenceShipper      `json:"references,omitempty"`
	CustomsReferences      []CustomsReference      `json:"customsReferences,omitempty"`
}

// NationalCommodityCode represents national commodity classification
type NationalCommodityCode struct {
	Type        string   `json:"type"`
	CountryCode string   `json:"countryCode,omitempty"`
	Values      []string `json:"values"`
}

// ExportLicense represents export license info
type ExportLicense struct {
	IsRequired bool   `json:"isRequired,omitempty"`
	Reference  string `json:"reference,omitempty"`
	IssueDate  string `json:"issueDate,omitempty"`
	ExpiryDate string `json:"expiryDate,omitempty"`
}

// ImportLicense represents import license info
type ImportLicense struct {
	IsRequired bool   `json:"isRequired,omitempty"`
	Reference  string `json:"reference,omitempty"`
	IssueDate  string `json:"issueDate,omitempty"`
	ExpiryDate string `json:"expiryDate,omitempty"`
}

// OuterPackaging represents outer packaging/overpack specification
type OuterPackaging struct {
	PackageCode      string           `json:"packageCode,omitempty"`
	IMOPackagingCode string           `json:"imoPackagingCode,omitempty"`
	NumberOfPackages int              `json:"numberOfPackages,omitempty"`
	Description      string           `json:"description,omitempty"`
	DangerousGoods   []DangerousGoods `json:"dangerousGoods,omitempty"`
}

// DangerousGoods represents dangerous goods specification
type DangerousGoods struct {
	UNNumber                             string                  `json:"UNNumber,omitempty"`
	NANumber                             string                  `json:"naNumber,omitempty"`
	CodedVariantList                     string                  `json:"codedVariantList,omitempty"`
	ProperShippingName                   string                  `json:"properShippingName"`
	TechnicalName                        string                  `json:"technicalName,omitempty"`
	IMOClass                             string                  `json:"imoClass"`
	SubsidiaryRisk1                      string                  `json:"subsidiaryRisk1,omitempty"`
	SubsidiaryRisk2                      string                  `json:"subsidiaryRisk2,omitempty"`
	IsMarinePollutant                    bool                    `json:"isMarinePollutant"`
	PackingGroup                         int                     `json:"packingGroup,omitempty"`
	IsLimitedQuantity                    bool                    `json:"isLimitedQuantity"`
	IsExceptedQuantity                   bool                    `json:"isExceptedQuantity"`
	IsSalvagePackings                    bool                    `json:"isSalvagePackings"`
	IsEmptyUncleanedResidue              bool                    `json:"isEmptyUncleanedResidue"`
	IsWaste                              bool                    `json:"isWaste"`
	IsHot                                bool                    `json:"isHot"`
	IsCompetentAuthorityApprovalRequired bool                    `json:"isCompetentAuthorityApprovalRequired"`
	CompetentAuthorityApproval           string                  `json:"competentAuthorityApproval,omitempty"`
	SegregationGroups                    []string                `json:"segregationGroups,omitempty"`
	InnerPackagings                      []InnerPackaging        `json:"innerPackagings,omitempty"`
	EmergencyContactDetails              *EmergencyContactDetails `json:"emergencyContactDetails"`
	EMSNumber                            string                  `json:"EMSNumber,omitempty"`
	EndOfHoldingTime                     string                  `json:"endOfHoldingTime,omitempty"`
	FumigationDateTime                   string                  `json:"fumigationDateTime,omitempty"`
	IsReportableQuantity                 bool                    `json:"isReportableQuantity"`
	InhalationZone                       string                  `json:"inhalationZone,omitempty"`
	GrossWeight                          *GrossWeight            `json:"grossWeight"`
	NetWeight                            *NetWeight              `json:"netWeight,omitempty"`
	NetExplosiveContent                  *NetExplosiveContent    `json:"netExplosiveContent,omitempty"`
	NetVolume                            *NetVolume              `json:"netVolume,omitempty"`
	Limits                               *Limits                 `json:"limits,omitempty"`
	SpecialCertificateNumber             string                  `json:"specialCertificateNumber,omitempty"`
	AdditionalContainerCargoHandling     string                  `json:"additionalContainerCargoHandling,omitempty"`
}

// InnerPackaging represents inner packaging details
type InnerPackaging struct {
	PackageCode      string `json:"packageCode,omitempty"`
	NumberOfPackages int    `json:"numberOfPackages,omitempty"`
	Description      string `json:"description,omitempty"`
}

// EmergencyContactDetails represents emergency contact info
type EmergencyContactDetails struct {
	Contact string `json:"contact,omitempty"`
	Phone   string `json:"phone,omitempty"`
	Email   string `json:"email,omitempty"`
}

// GrossWeight represents total weight including packaging
type GrossWeight struct {
	Value float64    `json:"value"`
	Unit  WeightUnit `json:"unit"`
}

// NetWeight represents weight excluding packaging
type NetWeight struct {
	Value float64    `json:"value"`
	Unit  WeightUnit `json:"unit"`
}

// NetExplosiveContent represents net explosive content
type NetExplosiveContent struct {
	Value float64    `json:"value"`
	Unit  WeightUnit `json:"unit"`
}

// NetVolume represents net volume
type NetVolume struct {
	Value float64    `json:"value"`
	Unit  VolumeUnit `json:"unit"`
}

// Limits represents limits for dangerous goods
type Limits struct {
	FlashPoint         float64 `json:"flashPoint,omitempty"`
	FlashPointUnit     string  `json:"flashPointUnit,omitempty"`
	Viscosity          float64 `json:"viscosity,omitempty"`
	ViscosityUnit      string  `json:"viscosityUnit,omitempty"`
	SolidificationPoint float64 `json:"solidificationPoint,omitempty"`
	SolidificationPointUnit string `json:"solidificationPointUnit,omitempty"`
	BoilingPoint       float64 `json:"boilingPoint,omitempty"`
	BoilingPointUnit   string  `json:"boilingPointUnit,omitempty"`
}
