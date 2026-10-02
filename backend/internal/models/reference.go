package models

// Country matches swagger country resource (id = ISO code).
type Country struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

// City matches swagger CityApi (+ id from DB).
type City struct {
	ID      int32  `json:"id,omitempty"`
	Name    string `json:"name"`
	Pr      string `json:"pr,omitempty"`
	Zip     string `json:"zip,omitempty"`
	Country string `json:"country,omitempty"`
	State   string `json:"state,omitempty"`
}

// Currency matches swagger currency resource.
type Currency struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

// Um matches swagger UmApi.
type Um struct {
	DescriptionIT      string `json:"description_IT,omitempty"`
	DescriptionEN      string `json:"description_EN,omitempty"`
	DescriptionES      string `json:"description_ES,omitempty"`
	DescriptionBG      string `json:"description_BG,omitempty"`
	ShortDescriptionIT string `json:"shortDescription_IT,omitempty"`
	ShortDescriptionEN string `json:"shortDescription_EN,omitempty"`
	ShortDescriptionES string `json:"shortDescription_ES,omitempty"`
	ShortDescriptionBG string `json:"shortDescription_BG,omitempty"`
	Um                 string `json:"um,omitempty"`
}

// OfficeType lookup for contact offices.
type OfficeType struct {
	ID          int32  `json:"id"`
	Description string `json:"description"`
}

// ContactRole lookup for sub-contacts.
type ContactRole struct {
	ID          int32  `json:"id"`
	Description string `json:"description"`
}

// PaymentTerm matches swagger PaymentTermApi.
type PaymentTerm struct {
	ID            int32            `json:"id,omitempty"`
	IDPaymentType *int32           `json:"idPaymentType,omitempty"`
	Description   string           `json:"description,omitempty"`
	EndOfMonth    bool             `json:"endOfMonth,omitempty"`
	ExtraDays     int32            `json:"extraDays,omitempty"`
	Pos           []PaymentTermPos `json:"pos,omitempty"`
}

// PaymentTermPos matches swagger PaymentTermPosApi.
type PaymentTermPos struct {
	ID         int32   `json:"id,omitempty"`
	Days       int32   `json:"days,omitempty"`
	Percentage float64 `json:"percentage,omitempty"`
}

// ListQuery common pagination/filter params.
type ListQuery struct {
	Limit  int
	Offset int
}

type CountryListQuery struct {
	ListQuery
	Description string
}

type CityListQuery struct {
	ListQuery
	Name string
}

type UmListQuery struct {
	ListQuery
	Description      string
	ShortDescription string
	ShowDeleted      bool
}

type PaymentTermListQuery struct {
	ListQuery
	Description string
}

type OfficeTypeListQuery struct {
	ListQuery
	Description string
}

type ContactRoleListQuery struct {
	ListQuery
	Description string
}
