package models

// ProductGroup matches swagger ProductGroupApi.
type ProductGroup struct {
	ID            int32  `json:"id,omitempty"`
	IDParent      *int32 `json:"idParent,omitempty"`
	DescriptionIT string `json:"description_IT,omitempty"`
	DescriptionEN string `json:"description_EN,omitempty"`
	DescriptionES string `json:"description_ES,omitempty"`
	DescriptionBG string `json:"description_BG,omitempty"`
}

type ProductGroupListQuery struct {
	ListQuery
	Description string
}
