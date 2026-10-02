package models

// CompanyUser matches swagger UserApi (+ id).
type CompanyUser struct {
	ID          int32  `json:"id,omitempty"`
	Active      bool   `json:"active,omitempty"`
	Username    string `json:"username,omitempty"`
	Password    string `json:"password,omitempty"`
	Firstname   string `json:"firstname,omitempty"`
	Lastname    string `json:"lastname,omitempty"`
	Email       string `json:"email,omitempty"`
	Mobilephone string `json:"mobilephone,omitempty"`
	IsAgent1    bool   `json:"isAgent1,omitempty"`
	IsAgent2    bool   `json:"isAgent2,omitempty"`
	IsEmployee  bool   `json:"isEmployee,omitempty"`
	Image       string `json:"image,omitempty"`
}

type CompanyUserListQuery struct {
	ListQuery
	IsAgent1      string
	IsAgent2      string
	RetrieveImage bool
}

type AgentCommission struct {
	ID              int32   `json:"id,omitempty"`
	IDUser          int32   `json:"idUser,omitempty"`
	IDBu            string  `json:"idbu,omitempty"`
	IDCustomer      string  `json:"idcustomer,omitempty"`
	IDMaterialGroup *int32  `json:"idmaterialgroup,omitempty"`
	IDMaterial      string  `json:"idmaterial,omitempty"`
	CommissionRate  float64 `json:"commissionRate,omitempty"`
	Description     string  `json:"description,omitempty"`
}

type AgentCommissionListQuery struct {
	ListQuery
	IDUser          *int32
	IDBu            string
	IDCustomer      string
	IDMaterialGroup *int32
	IDMaterial      string
}
