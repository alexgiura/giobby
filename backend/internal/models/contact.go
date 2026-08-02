package models

// Contact matches swagger ContactApi (+ id, flags).
type Contact struct {
	ID             int32   `json:"id,omitempty"`
	Type           string  `json:"type,omitempty"`
	Name           string  `json:"name,omitempty"`
	LastName       *string `json:"lastName,omitempty"`
	Address        *string `json:"address,omitempty"`
	City           *string `json:"city,omitempty"`
	Country        *string `json:"country,omitempty"`
	State          *string `json:"state,omitempty"`
	PostalCode     *string `json:"postalCode,omitempty"`
	Pr             *string `json:"pr,omitempty"`
	FiscalCode     *string `json:"fiscalCode,omitempty"`
	VatCode        *string `json:"vatCode,omitempty"`
	Phone1         *string `json:"phone1,omitempty"`
	Phone2         *string `json:"phone2,omitempty"`
	Mobile         *string `json:"mobile,omitempty"`
	Fax            *string `json:"fax,omitempty"`
	Email          *string `json:"email,omitempty"`
	Email2         *string `json:"email2,omitempty"`
	VisibilityType *string `json:"visibilityType,omitempty"`
	Status         *string `json:"status,omitempty"`
	IDUserOwner    *int32  `json:"idUserOwner,omitempty"`
	Source         *string `json:"source,omitempty"`
	Note           *string `json:"note,omitempty"`
	PhotoBase64    *string `json:"photo,omitempty"`
}

// ContactOffice matches ContactOfficeApi (+ id).
type ContactOffice struct {
	ID           int32   `json:"id,omitempty"`
	IDOfficeType *int32  `json:"idOfficeType,omitempty"`
	Name         *string `json:"name,omitempty"`
	Address      *string `json:"address,omitempty"`
	City         *string `json:"city,omitempty"`
	Pr           *string `json:"pr,omitempty"`
	PostalCode   *string `json:"postalCode,omitempty"`
	Country      *string `json:"country,omitempty"`
	State        *string `json:"state,omitempty"`
	Email        *string `json:"email,omitempty"`
	Phone1       *string `json:"phone1,omitempty"`
	Phone2       *string `json:"phone2,omitempty"`
	Fax          *string `json:"fax,omitempty"`
	DefaultDest  bool    `json:"defaultDest,omitempty"`
}

// SubContactAssoc matches SubContactAssocApi (+ id).
type SubContactAssoc struct {
	ID            int32  `json:"id,omitempty"`
	IDSubContact  int32  `json:"idSubContact"`
	IDContactRole *int32 `json:"idContactRole,omitempty"`
}

// ContactSource lookup row.
type ContactSource struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

// Customer matches CustomerApi.
type Customer struct {
	Contact          Contact  `json:"contact"`
	ID               string   `json:"id,omitempty"`
	IDPaymentTerm    *int32   `json:"idPaymentTerm,omitempty"`
	IDUserAgent1     *int32   `json:"idUserAgent1,omitempty"`
	IDUserAgent2     *int32   `json:"idUserAgent2,omitempty"`
	Currency         *string  `json:"currency,omitempty"`
	Credit           *float64 `json:"credit,omitempty"`
	Dsc1             *float64 `json:"dsc1,omitempty"`
	Dsc2             *float64 `json:"dsc2,omitempty"`
	Dsc3             *float64 `json:"dsc3,omitempty"`
	Dsc4             *float64 `json:"dsc4,omitempty"`
	PriceListID      *string  `json:"priceListID,omitempty"`
	PriceListType    *string  `json:"priceListType,omitempty"`
	PriceListEnabled *bool    `json:"priceListEnabled,omitempty"`
	Bank             *string  `json:"bank,omitempty"`
	Iban             *string  `json:"iban,omitempty"`
	Swift            *string  `json:"swift,omitempty"`
	Locked           *bool    `json:"locked,omitempty"`
	RecipientCode    *string  `json:"recipientCode,omitempty"`
}

// Vendor matches VendorApi.
type Vendor struct {
	Contact          Contact  `json:"contact"`
	ID               string   `json:"id,omitempty"`
	IDPaymentTerm    *int32   `json:"idPaymentTerm,omitempty"`
	IDUserAgent1     *int32   `json:"idUserAgent1,omitempty"`
	IDUserAgent2     *int32   `json:"idUserAgent2,omitempty"`
	Currency         *string  `json:"currency,omitempty"`
	Credit           *float64 `json:"credit,omitempty"`
	Dsc1             *float64 `json:"dsc1,omitempty"`
	Dsc2             *float64 `json:"dsc2,omitempty"`
	Dsc3             *float64 `json:"dsc3,omitempty"`
	Dsc4             *float64 `json:"dsc4,omitempty"`
	PriceListID      *string  `json:"priceListID,omitempty"`
	PriceListType    *string  `json:"priceListType,omitempty"`
	PriceListEnabled *bool    `json:"priceListEnabled,omitempty"`
	Bank             *string  `json:"bank,omitempty"`
	Iban             *string  `json:"iban,omitempty"`
	Swift            *string  `json:"swift,omitempty"`
}

// PartnerReport stub for customer/vendor reports.
type PartnerReport struct {
	ID            string  `json:"id"`
	TotalInvoiced float64 `json:"totalInvoiced"`
	TotalPaid     float64 `json:"totalPaid"`
	TotalDue      float64 `json:"totalDue"`
}

type ContactListQuery struct {
	ListQuery
	Type          string
	VisibilityType string
	FullName      string
	FreeText      string
	Deleted       *bool
	Email         string
	FiscalCode    string
	VatCode       string
	OnlyCustomers bool
	OnlyVendors   bool
	OnlyLeads     bool
	RetrieveImage bool
}

type ContactSourceListQuery struct {
	ListQuery
	FreeText string
}

type OfficeListQuery struct {
	ListQuery
	DefaultDest *bool
}

type SubContactListQuery struct {
	ListQuery
	Type    string
	Deleted *bool
}

type CustomerListQuery struct {
	ListQuery
	Name      string
	FreeText  string
	Deleted   *bool
	IDCompany *int64
}

type VendorListQuery struct {
	ListQuery
	Name      string
	FreeText  string
	Deleted   *bool
	IDCompany *int64
}
