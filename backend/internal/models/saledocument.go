package models

import "time"

// Sale document kinds (swagger paths).
const (
	SaleDocOffer      = "offer"
	SaleDocOrder      = "order"
	SaleDocGoodsIssue = "goodsissue"
	SaleDocInvoice    = "invoice"
	SaleDocTicket     = "ticket"
	SaleDocWriteoff   = "writeoff"
)

// DocumentRow matches swagger DocumentRowApi (+ idPos from DB).
type DocumentRow struct {
	IDPos                  int32   `json:"idPos,omitempty"`
	IDMaterial             string  `json:"idMaterial,omitempty"`
	IDAttributeCombination int32   `json:"idAttributeCombination,omitempty"`
	IDPosType              int32   `json:"idPosType,omitempty"`
	Description            string  `json:"description,omitempty"`
	Price                  float64 `json:"price,omitempty"`
	Quantity               float64 `json:"quantity,omitempty"`
	Packages               int32   `json:"packages,omitempty"`
	Dsc1                   float64 `json:"dsc1,omitempty"`
	Dsc2                   float64 `json:"dsc2,omitempty"`
	Dsc3                   float64 `json:"dsc3,omitempty"`
	Dsc4                   float64 `json:"dsc4,omitempty"`
	IDVat                  string  `json:"idVat,omitempty"`
	Um                     string  `json:"um,omitempty"`
}

// DocumentPaymentRow matches swagger DocumentPaymentRowApi.
type DocumentPaymentRow struct {
	RateNumber  int32   `json:"rateNumber,omitempty"`
	PaidAmount  float64 `json:"paidAmount,omitempty"`
	Date        *int64  `json:"date,omitempty"`
	Description string  `json:"description,omitempty"`
	Amount      float64 `json:"amount,omitempty"`
	PaymentDate *int64  `json:"paymentDate,omitempty"`
}

// Document matches swagger DocumentApi (+ id, docNumber).
type Document struct {
	ID               int32                `json:"id,omitempty"`
	IDDocumentType   int32                `json:"idDocumentType,omitempty"`
	IDContact        int32                `json:"idContact,omitempty"`
	IDCustomer       string               `json:"idCustomer,omitempty"`
	IDVendor         string               `json:"idVendor,omitempty"`
	DocDate          *int64               `json:"docDate,omitempty"`
	DocStatus        string               `json:"docStatus,omitempty"`
	DocCurrency      string               `json:"docCurrency,omitempty"`
	Tdsc1            float64              `json:"tdsc1,omitempty"`
	Tdsc2            float64              `json:"tdsc2,omitempty"`
	Tdsc3            float64              `json:"tdsc3,omitempty"`
	Tdsc4            float64              `json:"tdsc4,omitempty"`
	IDBu             string               `json:"idBu,omitempty"`
	IDNumerator      int32                `json:"idNumerator,omitempty"`
	DocNumber        string               `json:"docNumber,omitempty"`
	SequentialNumber int32                `json:"sequentialNumber,omitempty"`
	Note             string               `json:"note,omitempty"`
	InternalNote     string               `json:"internalNote,omitempty"`
	IDPaymentTerm    *int32               `json:"idPaymentTerm,omitempty"`
	Reference        string               `json:"reference,omitempty"`
	DestIDStorage    string               `json:"destIdStorage,omitempty"`
	Rows             []DocumentRow        `json:"rows,omitempty"`
	Payments         []DocumentPaymentRow `json:"payments,omitempty"`
}

type SaleDocumentListQuery struct {
	ListQuery
	FreeText        string
	IDContact       *int32
	DocStatus       string
	NeDocStatus     string
	GtDocDateMillis string
	LtDocDateMillis string
}

type DocumentListTransformation struct {
	IDNumerator    int32  `json:"idNumerator,omitempty"`
	DocNumber      string `json:"docNumber,omitempty"`
	DocDate        *int64 `json:"docDate,omitempty"`
	IDDocumentList string `json:"idDocumentList,omitempty"`
}

type DocumentPaymentPayload struct {
	DocumentPaymentRows []DocumentPaymentRowBase `json:"documentPaymentRows,omitempty"`
	RegDate             *int64                   `json:"regDate,omitempty"`
	PaymentDate         *int64                   `json:"paymentDate,omitempty"`
	IDPaymentType       *int32                   `json:"idPaymentType,omitempty"`
	Description         string                   `json:"description,omitempty"`
	Amount              float64                  `json:"amount,omitempty"`
}

type DocumentPaymentRowBase struct {
	RateNumber int32   `json:"rateNumber,omitempty"`
	PaidAmount float64 `json:"paidAmount,omitempty"`
}

// CustomerPayment matches swagger PaymentApi (subset).
type CustomerPayment struct {
	ID               int32   `json:"id,omitempty"`
	IDDocumentType   int32   `json:"idDocumentType,omitempty"`
	Reference        string  `json:"reference,omitempty"`
	IDAgent          *int32  `json:"idAgent,omitempty"`
	TotalPaidIc      float64 `json:"totalPaidIc,omitempty"`
	TotalInvoiceIc   float64 `json:"totalInvoiceIc,omitempty"`
	TotalRemainingIc float64 `json:"totalRemainingIc,omitempty"`
	DocDate          *int64  `json:"docDate,omitempty"`
	Note             string  `json:"note,omitempty"`
	DocNumber        string  `json:"docNumber,omitempty"`
	Deleted          bool    `json:"deleted,omitempty"`
}

type OpenBalanceRow struct {
	IDCustomer    string  `json:"idCustomer,omitempty"`
	IDDocument    int32   `json:"idDocument,omitempty"`
	DocNumber     string  `json:"docNumber,omitempty"`
	Amount        float64 `json:"amount,omitempty"`
	PaidAmount    float64 `json:"paidAmount,omitempty"`
	Remaining     float64 `json:"remaining,omitempty"`
	PaymentStatus string  `json:"paymentStatus,omitempty"`
}

type SaleDocumentAttachment struct {
	ID          int32  `json:"id,omitempty"`
	Filename    string `json:"filename,omitempty"`
	ContentType string `json:"contentType,omitempty"`
}

func TimeToMillisPtr(t time.Time) *int64 {
	ms := t.UnixMilli()
	return &ms
}

func MillisToTime(ms *int64) time.Time {
	if ms == nil {
		return time.Now()
	}
	return time.UnixMilli(*ms)
}
