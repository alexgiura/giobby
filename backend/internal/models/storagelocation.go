package models

// StorageLocation matches swagger StorageLocationApi.
type StorageLocation struct {
	IDLocation  string `json:"idLocation,omitempty"`
	IDStorage   string `json:"idStorage,omitempty"`
	Description string `json:"description,omitempty"`
}

type StorageLocationListQuery struct {
	ListQuery
	IDStorage      string
	ManagedStorage *bool
}
