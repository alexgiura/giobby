package models

// Attribute matches swagger AttributeApi.
type Attribute struct {
	DocType                *int32 `json:"docType,omitempty"`
	IDDoc                  string `json:"idDoc,omitempty"`
	IDAttribute            *int32 `json:"idAttribute,omitempty"`
	SubID                  *int32 `json:"subId,omitempty"`
	Pos                    *int32 `json:"pos,omitempty"`
	AttributeValue         string `json:"attributeValue,omitempty"`
	AttributeName          string `json:"attributeName,omitempty"`
	DocTypeAttributeValue  *int32 `json:"docTypeAttributeValue,omitempty"`
	IDLang                 string `json:"idLang,omitempty"`
}

// ProductAttribute matches swagger ProductAttributeApi.
type ProductAttribute struct {
	IDAttribute      int32  `json:"idAttribute,omitempty"`
	AttributeNameIT  string `json:"attributeName_IT,omitempty"`
	AttributeNameEN  string `json:"attributeName_EN,omitempty"`
	AttributeNameES  string `json:"attributeName_ES,omitempty"`
	AttributeNameBG  string `json:"attributeName_BG,omitempty"`
}

// ProductAttributesPayload matches swagger ProductAttributesApi.
type ProductAttributesPayload struct {
	Attributes []Attribute `json:"attributes"`
}

type AttributeListQuery struct {
	ListQuery
	AttributeType  string
	IDDoc          string
	IDDocumentType *int32
	IDLanguage     string
}
