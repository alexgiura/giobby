package models

// Stock matches swagger StockApi.
type Stock struct {
	ID                       int32   `json:"id,omitempty"`
	IDStorage                string  `json:"idStorage,omitempty"`
	IDLocation               string  `json:"idLocation,omitempty"`
	IDLot                    string  `json:"idLot,omitempty"`
	IDMaterial               string  `json:"idMaterial,omitempty"`
	Um                       string  `json:"um,omitempty"`
	Quantity                 float64 `json:"quantity,omitempty"`
	Value                    float64 `json:"value,omitempty"`
	NumberOfContainers       int32   `json:"numberOfContainers,omitempty"`
	StockState               int32   `json:"stockState,omitempty"`
	IDDocumentType           int32   `json:"idDocumentType,omitempty"`
	IDAttributeCombination   int32   `json:"idAttributeCombination,omitempty"`
	AttributeCombinationDesc string  `json:"attributeCombinationDesc,omitempty"`
	StorageDesc              string  `json:"storageDesc,omitempty"`
}

// StockAvailability is an aggregated availability row.
type StockAvailability struct {
	IDMaterial   string  `json:"idMaterial,omitempty"`
	MaterialDesc string  `json:"materialDesc,omitempty"`
	IDStorage    string  `json:"idStorage,omitempty"`
	StorageDesc  string  `json:"storageDesc,omitempty"`
	IDLocation   string  `json:"idLocation,omitempty"`
	LocationDesc string  `json:"locationDesc,omitempty"`
	IDLot        string  `json:"idLot,omitempty"`
	Quantity     float64 `json:"quantity,omitempty"`
	CommittedQty float64 `json:"committedQty,omitempty"`
	AvailableQty float64 `json:"availableQty,omitempty"`
	Um           string  `json:"um,omitempty"`
}

type StockAvailabilityQuery struct {
	ListQuery
	IDMaterial  string
	DetailLevel *int32
	IDStorage   string
	ZeroQtyRows *bool
	IDLocation  string
}

type StockAvailabilityReportQuery struct {
	ListQuery
	IDMaterial       string
	ZeroQtyRows      *bool
	AttributesDetail *bool
	StorageDetail    *bool
	LotsDetail       *bool
	IDStorage        string
	IDLocation       string
}
