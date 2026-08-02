package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/services"

	"github.com/gorilla/mux"
)

type SaleDocumentHandler struct {
	svc services.SaleDocumentService
}

func NewSaleDocumentHandler(svc services.SaleDocumentService) *SaleDocumentHandler {
	return &SaleDocumentHandler{svc: svc}
}

type docTransformIn struct {
	IDNumerator int32  `json:"idNumerator,omitempty"`
	DocDate     *int64 `json:"docDate,omitempty"`
	DocNumber   string `json:"docNumber,omitempty"`
}

// --- Offers ---

func (h *SaleDocumentHandler) ListOffers(w http.ResponseWriter, r *http.Request) {
	h.listDocuments(w, r, models.SaleDocOffer)
}

func (h *SaleDocumentHandler) CreateOffer(w http.ResponseWriter, r *http.Request) {
	h.createDocument(w, r, models.SaleDocOffer)
}

func (h *SaleDocumentHandler) GetOfferPDF(w http.ResponseWriter, r *http.Request) {
	h.servePDF(w, r)
}

func (h *SaleDocumentHandler) GetOffer(w http.ResponseWriter, r *http.Request) {
	h.getDocument(w, r, models.SaleDocOffer, false)
}

func (h *SaleDocumentHandler) UpdateOffer(w http.ResponseWriter, r *http.Request) {
	h.updateDocument(w, r, models.SaleDocOffer)
}

func (h *SaleDocumentHandler) DeleteOffer(w http.ResponseWriter, r *http.Request) {
	h.deleteDocument(w, r, models.SaleDocOffer)
}

func (h *SaleDocumentHandler) OfferToGoodsIssue(w http.ResponseWriter, r *http.Request) {
	h.transform(w, r, models.SaleDocGoodsIssue)
}

func (h *SaleDocumentHandler) OfferToInvoice(w http.ResponseWriter, r *http.Request) {
	h.transform(w, r, models.SaleDocInvoice)
}

func (h *SaleDocumentHandler) OfferToOrder(w http.ResponseWriter, r *http.Request) {
	h.transform(w, r, models.SaleDocOrder)
}

// --- Orders ---

func (h *SaleDocumentHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	h.listDocuments(w, r, models.SaleDocOrder)
}

func (h *SaleDocumentHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	h.createDocument(w, r, models.SaleDocOrder)
}

func (h *SaleDocumentHandler) GetOrderPDF(w http.ResponseWriter, r *http.Request) {
	h.servePDF(w, r)
}

func (h *SaleDocumentHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	h.getDocument(w, r, models.SaleDocOrder, false)
}

func (h *SaleDocumentHandler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	h.updateDocument(w, r, models.SaleDocOrder)
}

func (h *SaleDocumentHandler) DeleteOrder(w http.ResponseWriter, r *http.Request) {
	h.deleteDocument(w, r, models.SaleDocOrder)
}

func (h *SaleDocumentHandler) UpdateOrderPayments(w http.ResponseWriter, r *http.Request) {
	id, ok := saleDocPathID(mux.Vars(r))
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid document id", "")
		return
	}
	var rows []models.DocumentPaymentRowBase
	if err := json.NewDecoder(r.Body).Decode(&rows); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	if err := h.svc.UpdateOrderPayments(r.Context(), id, rows); err != nil {
		h.respondError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *SaleDocumentHandler) OrderToGoodsIssue(w http.ResponseWriter, r *http.Request) {
	h.transform(w, r, models.SaleDocGoodsIssue)
}

func (h *SaleDocumentHandler) OrderToInvoice(w http.ResponseWriter, r *http.Request) {
	h.transform(w, r, models.SaleDocInvoice)
}

func (h *SaleDocumentHandler) OrderToTicket(w http.ResponseWriter, r *http.Request) {
	h.transform(w, r, models.SaleDocTicket)
}

// --- Goods issue ---

func (h *SaleDocumentHandler) ListGoodsIssue(w http.ResponseWriter, r *http.Request) {
	h.listDocuments(w, r, models.SaleDocGoodsIssue)
}

func (h *SaleDocumentHandler) CreateGoodsIssue(w http.ResponseWriter, r *http.Request) {
	h.createDocument(w, r, models.SaleDocGoodsIssue)
}

func (h *SaleDocumentHandler) GetGoodsIssuePDF(w http.ResponseWriter, r *http.Request) {
	h.servePDF(w, r)
}

func (h *SaleDocumentHandler) GetGoodsIssue(w http.ResponseWriter, r *http.Request) {
	h.getDocument(w, r, models.SaleDocGoodsIssue, false)
}

func (h *SaleDocumentHandler) UpdateGoodsIssue(w http.ResponseWriter, r *http.Request) {
	h.updateDocument(w, r, models.SaleDocGoodsIssue)
}

func (h *SaleDocumentHandler) DeleteGoodsIssue(w http.ResponseWriter, r *http.Request) {
	h.deleteDocument(w, r, models.SaleDocGoodsIssue)
}

func (h *SaleDocumentHandler) GoodsIssuesToInvoice(w http.ResponseWriter, r *http.Request) {
	h.transformList(w, r, models.SaleDocInvoice)
}

func (h *SaleDocumentHandler) GoodsIssueToInvoice(w http.ResponseWriter, r *http.Request) {
	h.transform(w, r, models.SaleDocInvoice)
}

// --- Invoices ---

func (h *SaleDocumentHandler) ListInvoices(w http.ResponseWriter, r *http.Request) {
	h.listDocuments(w, r, models.SaleDocInvoice)
}

func (h *SaleDocumentHandler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
	h.createDocument(w, r, models.SaleDocInvoice)
}

func (h *SaleDocumentHandler) OpenBalance(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.OpenBalance(r.Context(), r.URL.Query().Get("idCustomer"), limit, offset)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *SaleDocumentHandler) GetInvoicePDF(w http.ResponseWriter, r *http.Request) {
	h.servePDF(w, r)
}

func (h *SaleDocumentHandler) EInvoice(w http.ResponseWriter, r *http.Request) {
	id, ok := saleDocPathID(mux.Vars(r))
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid document id", "")
		return
	}
	out, err := h.svc.EInvoice(r.Context(), id)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *SaleDocumentHandler) AddInvoicePayment(w http.ResponseWriter, r *http.Request) {
	id, ok := saleDocPathID(mux.Vars(r))
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid document id", "")
		return
	}
	var payload models.DocumentPaymentPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	if err := h.svc.AddInvoicePayment(r.Context(), id, payload); err != nil {
		h.respondError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *SaleDocumentHandler) GetInvoice(w http.ResponseWriter, r *http.Request) {
	h.getDocument(w, r, models.SaleDocInvoice, false)
}

func (h *SaleDocumentHandler) UpdateInvoice(w http.ResponseWriter, r *http.Request) {
	h.updateDocument(w, r, models.SaleDocInvoice)
}

func (h *SaleDocumentHandler) DeleteInvoice(w http.ResponseWriter, r *http.Request) {
	h.deleteDocument(w, r, models.SaleDocInvoice)
}

func (h *SaleDocumentHandler) InvoiceToCreditNote(w http.ResponseWriter, r *http.Request) {
	h.transform(w, r, models.SaleDocWriteoff)
}

// --- Tickets ---

func (h *SaleDocumentHandler) ListTickets(w http.ResponseWriter, r *http.Request) {
	h.listDocuments(w, r, models.SaleDocTicket)
}

func (h *SaleDocumentHandler) CreateTicket(w http.ResponseWriter, r *http.Request) {
	h.createDocument(w, r, models.SaleDocTicket)
}

func (h *SaleDocumentHandler) GetTicket(w http.ResponseWriter, r *http.Request) {
	h.getDocument(w, r, models.SaleDocTicket, false)
}

func (h *SaleDocumentHandler) UpdateTicket(w http.ResponseWriter, r *http.Request) {
	h.updateDocument(w, r, models.SaleDocTicket)
}

func (h *SaleDocumentHandler) DeleteTicket(w http.ResponseWriter, r *http.Request) {
	h.deleteDocument(w, r, models.SaleDocTicket)
}

// --- Writeoff ---

func (h *SaleDocumentHandler) CreateWriteoff(w http.ResponseWriter, r *http.Request) {
	h.createDocument(w, r, models.SaleDocWriteoff)
}

func (h *SaleDocumentHandler) TicketToWriteoff(w http.ResponseWriter, r *http.Request) {
	h.transform(w, r, models.SaleDocWriteoff)
}

func (h *SaleDocumentHandler) UpdateWriteoff(w http.ResponseWriter, r *http.Request) {
	h.updateDocument(w, r, models.SaleDocWriteoff)
}

// --- Customer payments ---

func (h *SaleDocumentHandler) ListCustomerPayments(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.ListCustomerPayments(r.Context(), models.ListQuery{Limit: limit, Offset: offset})
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *SaleDocumentHandler) CreateCustomerPayment(w http.ResponseWriter, r *http.Request) {
	var in models.CustomerPayment
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.CreateCustomerPayment(r.Context(), in)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *SaleDocumentHandler) GetCustomerPaymentByID(w http.ResponseWriter, r *http.Request) {
	id, ok := parseInt32Path(mux.Vars(r), "id")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid payment id", "")
		return
	}
	item, err := h.svc.GetCustomerPayment(r.Context(), id)
	if err != nil {
		h.respondError(w, err)
		return
	}
	if item == nil {
		respondWithError(w, http.StatusNotFound, ErrCodeNotFound, "Payment not found", "")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *SaleDocumentHandler) GetCustomerPaymentByDoc(w http.ResponseWriter, r *http.Request) {
	id, ok := saleDocPathID(mux.Vars(r))
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid document id", "")
		return
	}
	item, err := h.svc.GetCustomerPaymentByDoc(r.Context(), id)
	if err != nil {
		h.respondError(w, err)
		return
	}
	if item == nil {
		respondWithError(w, http.StatusNotFound, ErrCodeNotFound, "Payment not found", "")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *SaleDocumentHandler) UpdateCustomerPayment(w http.ResponseWriter, r *http.Request) {
	id, ok := saleDocPathID(mux.Vars(r))
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid payment id", "")
		return
	}
	var in models.CustomerPayment
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.UpdateCustomerPayment(r.Context(), id, in)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *SaleDocumentHandler) DeleteCustomerPayment(w http.ResponseWriter, r *http.Request) {
	id, ok := saleDocPathID(mux.Vars(r))
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid payment id", "")
		return
	}
	if err := h.svc.DeleteCustomerPayment(r.Context(), id); err != nil {
		h.respondError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Reports, PDF, attachments ---

func (h *SaleDocumentHandler) Reports(w http.ResponseWriter, r *http.Request) {
	out, err := h.svc.Reports(r.Context(),
		r.URL.Query().Get("reportType"),
		r.URL.Query().Get("startDate"),
		r.URL.Query().Get("endDate"),
	)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *SaleDocumentHandler) GetDocumentPDF(w http.ResponseWriter, r *http.Request) {
	h.servePDF(w, r)
}

func (h *SaleDocumentHandler) GetAttachment(w http.ResponseWriter, r *http.Request) {
	id, ok := saleDocPathID(mux.Vars(r))
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid document id", "")
		return
	}
	item, err := h.svc.GetAttachment(r.Context(), id)
	if err != nil {
		h.respondError(w, err)
		return
	}
	if item == nil {
		respondWithError(w, http.StatusNotFound, ErrCodeNotFound, "Attachment not found", "")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *SaleDocumentHandler) UploadAttachment(w http.ResponseWriter, r *http.Request) {
	id, ok := saleDocPathID(mux.Vars(r))
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid document id", "")
		return
	}
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid multipart form", err.Error())
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		file, header, err = r.FormFile("attachment")
	}
	if err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Attachment file required", err.Error())
		return
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Failed to read attachment", err.Error())
		return
	}
	mime := header.Header.Get("Content-Type")
	if mime == "" {
		mime = "application/octet-stream"
	}
	if err := h.svc.SaveAttachment(r.Context(), id, header.Filename, mime, data); err != nil {
		h.respondError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *SaleDocumentHandler) DeleteAttachment(w http.ResponseWriter, r *http.Request) {
	id, ok := saleDocPathID(mux.Vars(r))
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid document id", "")
		return
	}
	if err := h.svc.DeleteAttachment(r.Context(), id); err != nil {
		h.respondError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- helpers ---

func (h *SaleDocumentHandler) listDocuments(w http.ResponseWriter, r *http.Request, kind string) {
	limit, offset := parseLimitOffset(r)
	q := saleDocListQuery(r, limit, offset)
	items, err := h.svc.ListDocuments(r.Context(), kind, q)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *SaleDocumentHandler) createDocument(w http.ResponseWriter, r *http.Request, kind string) {
	var in models.Document
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.CreateDocument(r.Context(), kind, in)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *SaleDocumentHandler) getDocument(w http.ResponseWriter, r *http.Request, kind string, onlyHeader bool) {
	id, ok := saleDocPathID(mux.Vars(r))
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid document id", "")
		return
	}
	item, err := h.svc.GetDocument(r.Context(), id, kind, onlyHeader)
	if err != nil {
		h.respondError(w, err)
		return
	}
	if item == nil {
		respondWithError(w, http.StatusNotFound, ErrCodeNotFound, "Document not found", "")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *SaleDocumentHandler) updateDocument(w http.ResponseWriter, r *http.Request, kind string) {
	id, ok := saleDocPathID(mux.Vars(r))
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid document id", "")
		return
	}
	var in models.Document
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.UpdateDocument(r.Context(), id, kind, in)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *SaleDocumentHandler) deleteDocument(w http.ResponseWriter, r *http.Request, kind string) {
	id, ok := saleDocPathID(mux.Vars(r))
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid document id", "")
		return
	}
	if err := h.svc.DeleteDocument(r.Context(), id, kind); err != nil {
		h.respondError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *SaleDocumentHandler) transform(w http.ResponseWriter, r *http.Request, targetKind string) {
	id, ok := saleDocPathID(mux.Vars(r))
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid document id", "")
		return
	}
	var body docTransformIn
	_ = json.NewDecoder(r.Body).Decode(&body)
	item, err := h.svc.Transform(r.Context(), id, targetKind, body.DocDate, body.DocNumber)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *SaleDocumentHandler) transformList(w http.ResponseWriter, r *http.Request, targetKind string) {
	var payload models.DocumentListTransformation
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	items, err := h.svc.TransformList(r.Context(), targetKind, payload)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSONArray(w, http.StatusCreated, items)
}

func (h *SaleDocumentHandler) servePDF(w http.ResponseWriter, r *http.Request) {
	id, ok := saleDocPathID(mux.Vars(r))
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid document id", "")
		return
	}
	data, err := h.svc.GetPDF(r.Context(), id)
	if err != nil {
		h.respondError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func saleDocListQuery(r *http.Request, limit, offset int) models.SaleDocumentListQuery {
	gt := r.URL.Query().Get("gtDocDateMillins")
	if gt == "" {
		gt = r.URL.Query().Get("gtDocDateMillis")
	}
	lt := r.URL.Query().Get("ltDocDateMillins")
	if lt == "" {
		lt = r.URL.Query().Get("ltDocDateMillis")
	}
	return models.SaleDocumentListQuery{
		ListQuery:       models.ListQuery{Limit: limit, Offset: offset},
		FreeText:        r.URL.Query().Get("freeText"),
		IDContact:       parseOptionalInt32Query(r, "idContact"),
		DocStatus:       r.URL.Query().Get("docStatus"),
		NeDocStatus:     r.URL.Query().Get("neDocStatus"),
		GtDocDateMillis: gt,
		LtDocDateMillis: lt,
	}
}

func saleDocPathID(vars map[string]string) (int32, bool) {
	if id, ok := parseInt32Path(vars, "idDoc"); ok {
		return id, true
	}
	return parseInt32Path(vars, "id")
}

func (h *SaleDocumentHandler) respondError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "required") || strings.Contains(msg, "invalid") {
		respondWithError(w, http.StatusBadRequest, ErrCodeValidationFailed, err.Error(), "")
		return
	}
	if strings.Contains(msg, "not found") {
		respondWithError(w, http.StatusNotFound, ErrCodeNotFound, "Resource not found", "")
		return
	}
	status, code, m := parseDatabaseError(err)
	respondWithError(w, status, code, m, err.Error())
}
