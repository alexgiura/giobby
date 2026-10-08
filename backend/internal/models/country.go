package models

// Country matches swagger CountryApi (id = ISO code).
type Country struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

// CountryListQuery filters GET /countries.
type CountryListQuery struct {
	ListQuery
	Description string
	ShowDeleted bool
}
