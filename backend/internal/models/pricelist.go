package models

// Pricelist matches swagger PricelistApi.
type Pricelist struct {
	ID                int32   `json:"id,omitempty"`
	IDPricelistScheme *int32  `json:"idPricelistScheme,omitempty"`
	Description       string  `json:"description,omitempty"`
	Type              string  `json:"type,omitempty"`
	Note              string  `json:"note,omitempty"`
	ValidSince        *int64  `json:"validSince,omitempty"`
	ValidUntil        *int64  `json:"validUntil,omitempty"`
	Priority          int32   `json:"priority,omitempty"`
	SourceID          string  `json:"sourceId,omitempty"`
}

// PricelistScheme is a pricelist schema header.
type PricelistScheme struct {
	ID          int32  `json:"id"`
	Description string `json:"description"`
}

// PricelistRow is a row inside a price list (GET /pricelists/{id}).
type PricelistRow struct {
	MaterialID      string  `json:"materialId,omitempty"`
	ProdDescription string  `json:"prodDescription,omitempty"`
	Price           float64 `json:"price,omitempty"`
}

type PricelistListQuery struct {
	ListQuery
	Type               string
	ValidOnDateMillins string
}

type PricelistDetailQuery struct {
	ListQuery
	Type               string
	ValidOnDateMillins string
	ProdDescription    string
	MaterialID         string
}
