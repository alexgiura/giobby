package models

// PersonalActivity matches swagger PersonalActivityApi.
type PersonalActivity struct {
	ID          int32  `json:"id,omitempty"`
	Description string `json:"description,omitempty"`
	RootName    string `json:"rootName,omitempty"`
	Date        *int64 `json:"date,omitempty"`
	Status      string `json:"status,omitempty"`
}

type PersonalActivityListQuery struct {
	ListQuery
	Status string
	Name   string
}
