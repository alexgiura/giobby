package models

import "time"

// Product matches swagger ProductApi.
type Product struct {
	ID                    string   `json:"id,omitempty"`
	Barcode               string   `json:"barcode,omitempty"`
	BasicCode             string   `json:"basicCode,omitempty"`
	IDType                int32    `json:"idType,omitempty"`
	TypeDesc              string   `json:"typeDesc,omitempty"`
	Description           string   `json:"description,omitempty"`
	DescriptionIT         string   `json:"description_IT,omitempty"`
	DescriptionEN         string   `json:"description_EN,omitempty"`
	DescriptionES         string   `json:"description_ES,omitempty"`
	DescriptionBG         string   `json:"description_BG,omitempty"`
	Note                  string   `json:"note,omitempty"`
	IDMaterialGroup       *int32   `json:"idMaterialGroup,omitempty"`
	MaterialGroupDesc     string   `json:"materialGroupDesc,omitempty"`
	Um                    string   `json:"um,omitempty"`
	UmDescription         string   `json:"umDescription,omitempty"`
	SalesEnabled          bool     `json:"salesEnabled"`
	Locked                bool     `json:"locked"`
	StockEnabled          bool     `json:"stockEnabled"`
	PricelistEnabled      bool     `json:"pricelistEnabled"`
	SalesPrice            float64  `json:"salesPrice,omitempty"`
	SalesPriceIncludeVat  float64  `json:"salesPriceIncludeVat,omitempty"`
	IDVat                 string   `json:"idVat,omitempty"`
	PurchasePrice         float64  `json:"purchasePrice,omitempty"`
	ManageVariant         bool     `json:"manageVariant"`
	GrossWeight           float64  `json:"grossWeight,omitempty"`
	NetWeight             float64  `json:"netWeight,omitempty"`
	DimHeight             float64  `json:"dimHeight,omitempty"`
	DimLength             float64  `json:"dimLength,omitempty"`
	DimWidth              float64  `json:"dimWidth,omitempty"`
	BoxQty                float64  `json:"boxQty,omitempty"`
	Manufacturer          string   `json:"manufacturer,omitempty"`
	DefaultStorage        string   `json:"defaultStorage,omitempty"`
	ImagesList            []string `json:"imagesList,omitempty"`
	CreateDate            *int64   `json:"createDate,omitempty"`
	ModifyDate            *int64   `json:"modifyDate,omitempty"`
}

// AttributeCombination matches swagger variant row.
type AttributeCombination struct {
	IDMaterial             string  `json:"idMaterial,omitempty"`
	IDAttributeCombination int32   `json:"idAttributeCombination,omitempty"`
	AttributeIdsConcat     string  `json:"attributeIdsConcat,omitempty"`
	Barcode                string  `json:"barcode,omitempty"`
	DeltaPrice             float64 `json:"deltaPrice,omitempty"`
	Description            string  `json:"description,omitempty"`
}

// ProductVariantsPayload matches swagger ProductVariantsApi.
type ProductVariantsPayload struct {
	Variants []AttributeCombination `json:"variants"`
}

// RepomediaRef matches swagger RepomediaApi for product media link.
type RepomediaRef struct {
	ID string `json:"id"`
}

type ProductListQuery struct {
	ListQuery
	Description      string
	SalesEnabled     *bool
	Locked           *bool
	StockEnabled     *bool
	PricelistEnabled *bool
	IDPlugin         string
	CreateDateMillis string
	ModifyDateMillis string
}

func TimeToMillis(t time.Time) int64 {
	return t.UnixMilli()
}

func MillisPtr(t time.Time) *int64 {
	ms := t.UnixMilli()
	return &ms
}
