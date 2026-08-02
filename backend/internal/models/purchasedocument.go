package models

// Purchase document kinds (swagger paths under /purchases/*).
const (
	PurchaseDocOrder        = "order"
	PurchaseDocGoodsReceipt = "goodsreceipt"
	PurchaseDocInvoice      = "invoice"
)

type PurchaseDocumentListQuery struct {
	ListQuery
	FreeText        string
	IDContact       *int32
	DocStatus       string
	NeDocStatus     string
	GtDocDateMillis string
	LtDocDateMillis string
}

// VendorPayment matches swagger PaymentApi for vendor payments.
type VendorPayment struct {
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

type VendorOpenBalanceRow struct {
	IDVendor      string  `json:"idVendor,omitempty"`
	IDDocument    int32   `json:"idDocument,omitempty"`
	DocNumber     string  `json:"docNumber,omitempty"`
	Amount        float64 `json:"amount,omitempty"`
	PaidAmount    float64 `json:"paidAmount,omitempty"`
	Remaining     float64 `json:"remaining,omitempty"`
	PaymentStatus string  `json:"paymentStatus,omitempty"`
}

type PurchaseDocumentAttachment struct {
	ID          int32  `json:"id,omitempty"`
	Filename    string `json:"filename,omitempty"`
	ContentType string `json:"contentType,omitempty"`
}
