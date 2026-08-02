package models

// AccountMovement matches swagger AccountMovementApi.
type AccountMovement struct {
	ID              int32   `json:"id,omitempty"`
	Description     string  `json:"description,omitempty"`
	FromAccountCode string  `json:"fromAccountCode,omitempty"`
	ToAccountCode   string  `json:"toAccountCode,omitempty"`
	AccountCode     string  `json:"accountCode,omitempty"`
	AccountSign     string  `json:"accountSign,omitempty"`
	Amount          float64 `json:"amount,omitempty"`
}

// AccountMovementRegistration matches swagger AccountMovementRegistrationApi.
type AccountMovementRegistration struct {
	IDDoc                   int32             `json:"idDoc,omitempty"`
	Bu                      string            `json:"bu,omitempty"`
	Bup                     string            `json:"bup,omitempty"`
	IDNumerator             int32             `json:"idNumerator,omitempty"`
	IDAccountmovementType   int32             `json:"idAccountmovementType,omitempty"`
	DocDate                 *int64            `json:"docDate,omitempty"`
	RegDate                 *int64            `json:"regDate,omitempty"`
	IDAccountCenter         *int32            `json:"idAccountCenter,omitempty"`
	Description             string            `json:"description,omitempty"`
	Rows                    []AccountMovement `json:"rows,omitempty"`
}

type AccountMovementListQuery struct {
	ListQuery
	EntryType      string
	IDBu           string
	IDBup          string
	IDDoc          *int32
	AccountCenter  *int32
	FreeText       string
	ShowZero       bool
	FromDateMillis string
	ToDateMillis   string
}
