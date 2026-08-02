package models

// MachineDataTracking matches swagger MachineDataTrackingApi (+ id from DB).
type MachineDataTracking struct {
	ID              int32   `json:"id,omitempty"`
	IDMachine       string  `json:"idMachine,omitempty"`
	IDOperator      string  `json:"idOperator,omitempty"`
	Materials       string  `json:"materials,omitempty"`
	IDLot           string  `json:"idLot,omitempty"`
	BoxQty          float64 `json:"boxQty,omitempty"`
	BoxNumber       float64 `json:"boxNumber,omitempty"`
	ProductionWaste float64 `json:"productionWaste,omitempty"`
	Barcode         string  `json:"barcode,omitempty"`
	StartTime       string  `json:"startTime,omitempty"`
	EndTime         string  `json:"endTime,omitempty"`
	Duration        string  `json:"duration,omitempty"`
	MachineStatus   string  `json:"machineStatus,omitempty"`
	AlarmCode       string  `json:"alarmCode,omitempty"`
	AlarmTime       string  `json:"alarmTime,omitempty"`
	Date            string  `json:"date,omitempty"`
	IDAttachment    string  `json:"idAttachment,omitempty"`
}

type MachineDataTrackingListQuery struct {
	ListQuery
	ID              *int32
	IDMachine       string
	IDOperator      string
	Materials       string
	IDLot           string
	BoxQty          *float64
	BoxNumber       *float64
	ProductionWaste *float64
	Barcode         string
	StartTime       string
	EndTime         string
	Duration        string
	MachineStatus   string
	AlarmCode       string
	AlarmTime       string
	Date            string
	IDAttachment    string
	Deleted         *bool
}
