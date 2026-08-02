package models

// CrmAccount matches swagger CrmAccountApi (+ id on response).
type CrmAccount struct {
	ID        int32  `json:"id,omitempty"`
	IDContact string `json:"idContact,omitempty"`
	Name      string `json:"name,omitempty"`
}

// CrmTask matches swagger CrmTaskApi.
type CrmTask struct {
	ID                             int32          `json:"id,omitempty"`
	IDRoot                         int32          `json:"idRoot,omitempty"`
	IDParent                       int32          `json:"idParent,omitempty"`
	Type                           string         `json:"type,omitempty"`
	Name                           string         `json:"name,omitempty"`
	Description                    string         `json:"description,omitempty"`
	Place                          string         `json:"place,omitempty"`
	Status                         string         `json:"status,omitempty"`
	StartDate                      *int64         `json:"startDate,omitempty"`
	EndDate                        *int64         `json:"endDate,omitempty"`
	CreateDate                     *int64         `json:"createDate,omitempty"`
	IDDocType                      *int32         `json:"idDocType,omitempty"`
	IDDoc                          string         `json:"idDoc,omitempty"`
	DocDescription                 string         `json:"docDescription,omitempty"`
	Priority                       int32          `json:"priority,omitempty"`
	Privacy                        string         `json:"privacy,omitempty"`
	IDUserAssignee                 *int32         `json:"idUserAssignee,omitempty"`
	UserAssigneeDescription        string         `json:"userAssigneeDescription,omitempty"`
	IDContactRepresentative        *int32         `json:"idContactRepresentative,omitempty"`
	ContactRepresentativeDescription string       `json:"contactRepresentativeDescription,omitempty"`
	Quantity                       float64        `json:"quantity,omitempty"`
	Amount                         float64        `json:"amount,omitempty"`
	Probability                    float64        `json:"probability,omitempty"`
	ForecastRevenue                float64        `json:"forecastRevenue,omitempty"`
	Attendees                      []CrmActivity  `json:"attendees,omitempty"`
}

// CrmActivity matches swagger CrmTaskObjectApi.
type CrmActivity struct {
	ID                 int32  `json:"id,omitempty"`
	IDParent           int32  `json:"idParent,omitempty"`
	IDRoot             int32  `json:"idRoot,omitempty"`
	IDUser             *int32 `json:"idUser,omitempty"`
	IDAttachment       string `json:"idAttachment,omitempty"`
	AttachmentFileName string `json:"attachmentFileName,omitempty"`
	Name               string `json:"name,omitempty"`
	Description        string `json:"description,omitempty"`
	Type               string `json:"type,omitempty"`
	IDDocumentType     *int32 `json:"idDocumentType,omitempty"`
	IDDocumentTypeExt  string `json:"idDocumentTypeExt,omitempty"`
	IDDoc              string `json:"idDoc,omitempty"`
	DocDescription     string `json:"docDescription,omitempty"`
}

type CrmTaskListQuery struct {
	ListQuery
	Types           string
	Status          string
	CreateDateStart string
	CreateDateEnd   string
	IncludeActivities bool
	IDCrmAccount    *int32
}

type CrmActivityListQuery struct {
	ListQuery
	Type      string
	IDDoc     string
	IDDocType *int32
}

type CrmAttachmentResult struct {
	IDAttachment       string `json:"idAttachment,omitempty"`
	AttachmentFileName string `json:"attachmentFileName,omitempty"`
}
