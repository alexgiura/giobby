package models

// Storage matches swagger StorageApi.
type Storage struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Managed     int32  `json:"managed"`
}

type StorageListQuery struct {
	ListQuery
	ID      string
	Managed *bool
}
