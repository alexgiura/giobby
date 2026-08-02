package models

// Lot matches swagger LotApi.
type Lot struct {
	IDLot        string `json:"idLot,omitempty"`
	IDMaterial   string `json:"idMaterial,omitempty"`
	IncomingDate *int64 `json:"incomingDate,omitempty"`
	ExpireDate   *int64 `json:"expireDate,omitempty"`
	Reference    string `json:"reference,omitempty"`
}

type LotListQuery struct {
	ListQuery
	IDMaterial string
	IDLot      string
}
