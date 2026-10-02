package models

// Company public profile (GET /companies/{id}).
type Company struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	VatNumber    *string `json:"vatNumber,omitempty"`
	Email        *string `json:"email,omitempty"`
	Phone        *string `json:"phone,omitempty"`
	Address      *string `json:"address,omitempty"`
	City         *string `json:"city,omitempty"`
	Zip          *string `json:"zip,omitempty"`
	Country      *string `json:"country,omitempty"`
	BaseCurrency *string `json:"baseCurrency,omitempty"`
}

// AccountCenter matches swagger AccountCenterApi (+ id).
type AccountCenter struct {
	ID          int32  `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
}

// AccountCode chart-of-accounts entry.
type AccountCode struct {
	ID          int32   `json:"id,omitempty"`
	Code        string  `json:"code"`
	Description *string `json:"description,omitempty"`
	Entry       *string `json:"entry,omitempty"`
	EntryType   *string `json:"entryType,omitempty"`
	Disabled    bool    `json:"disabled,omitempty"`
}

// BankCashdesk matches swagger BankCashdeskApi.
type BankCashdesk struct {
	ID          int32   `json:"id,omitempty"`
	Description string  `json:"description,omitempty"`
	IsCashdesk  bool    `json:"isCashdesk,omitempty"`
	AccountCode *string `json:"accountCode,omitempty"`
	Deleted     bool    `json:"deleted,omitempty"`
	Iban        *string `json:"iban,omitempty"`
	Sia         *string `json:"sia,omitempty"`
	Swift       *string `json:"swift,omitempty"`
	Cuc         *string `json:"cuc,omitempty"`
}

// CurrencyChange exchange rate response.
type CurrencyChange struct {
	SourceCurrency string  `json:"sourceCurrency"`
	TargetCurrency string  `json:"targetCurrency"`
	Rate           float64 `json:"rate"`
}

// VatRate VAT configuration per business unit.
type VatRate struct {
	IDVat        string  `json:"idVat"`
	IDBu         *string `json:"idBu,omitempty"`
	Country      *string `json:"country,omitempty"`
	State        *string `json:"state,omitempty"`
	ContactType  *string `json:"contactType,omitempty"`
	TaxSuperType *int32  `json:"taxSuperType,omitempty"`
	Rate         float64 `json:"rate"`
	Description  *string `json:"description,omitempty"`
}

// BupNumerator BU/BUP/numerator row.
type BupNumerator struct {
	IDBu           string `json:"idBu"`
	BuName         string `json:"buName"`
	IDBup          string `json:"idBup"`
	BupName        string `json:"bupName"`
	IDNumerator    int32  `json:"idNumerator"`
	IDDocumentType *int32 `json:"idDocumentType,omitempty"`
	Numerator      int32  `json:"numerator"`
}

type AccountCenterListQuery struct {
	ListQuery
	Type string
}

type AccountCodeListQuery struct {
	ListQuery
	EntryType    string
	Entry        string
	ShowDisabled bool
}

type BankListQuery struct {
	ListQuery
	IsCashdesk  *bool
	AccountCode string
	Iban        string
	Sia         string
	Swift       string
	Cuc         string
	ShowDeleted bool
}

type VatListQuery struct {
	ListQuery
	IDBu         string
	Country      string
	State        string
	ContactType  string
	TaxSuperType *int32
	Rate         *float64
}

type VatGetQuery struct {
	IDBu         string
	Country      string
	State        string
	ContactType  string
	TaxSuperType *int32
}

type BupNumeratorQuery struct {
	IDDocumentType *int32
	IDPlugin       string
}
