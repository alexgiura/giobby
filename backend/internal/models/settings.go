package models

// AppSetting matches swagger SingleSettingApi.
type AppSetting struct {
	Key1  string `json:"key1,omitempty"`
	Key2  string `json:"key2,omitempty"`
	Key3  string `json:"key3,omitempty"`
	Key4  string `json:"key4,omitempty"`
	Key5  string `json:"key5,omitempty"`
	Value string `json:"value,omitempty"`
}

type AppSettingListQuery struct {
	Key1 string
	Key2 string
	Key3 string
	Key4 string
	Key5 string
}
