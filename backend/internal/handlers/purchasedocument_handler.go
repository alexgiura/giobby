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

type PurchaseDocumentHandler struct {
	svc services.PurchaseDocumentService
}

func NewPurchaseDocumentHandler(svc services.PurchaseDocumentService) *PurchaseDocumentHandler {
	return &PurchaseDocumentHandler{svc: svc}
}

// --- Orders ---

func (h *PurchaseDocumentHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	h.listDocuments(w, r, models.PurchaseDocOrder)
}

func (h *PurchaseDocumentHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	h.createDocument(w, r, models.PurchaseDocOrder)
}

func (h *PurchaseDocumentHandler) GetOrderPDF(w http.ResponseWriter, r *http.Request) {
	h.servePDF(w, r)
}

func (h *PurchaseDocumentHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	onlyHeader := parseBoolQuery(r, "onlyHeaderData", false)
	h.getDocument(w, r, models.PurchaseDocOrder, onlyHeader)
}

func (h *PurchaseDocumentHandler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	h.updateDocument(w, r, models.PurchaseDocOrder)
}

// --- Goods receipt ---

func (h *PurchaseDocumentHandler) ListGoodsReceipt(w http.ResponseWriter, r *http.Request) {
	h.listDocuments(w, r, models.PurchaseDocGoodsReceipt)
}

func (h *PurchaseDocumentHandler) CreateGoodsReceipt(w http.ResponseWriter, r *http.Request) {
	h.createDocument(w, r, models.PurchaseDocGoodsReceipt)
}

func (h *PurchaseDocumentHandler) GetGoodsReceiptPDF(w http.ResponseWriter, r *http.Request) {
	h.servePDF(w, r)
}

func (h *PurchaseDocumentHandler) GetGoodsReceipt(w http.ResponseWriter, r *http.Request) {
	onlyHeader := parseBoolQuery(r, "onlyHeaderData", false)
	h.getDocument(w, r, models.PurchaseDocGoodsReceipt, onlyHeader)
}

func (h *PurchaseDocumentHandler) UpdateGoodsReceipt(w http.ResponseWriter, r *http.Request) {
	h.updateDocument(w, r, models.PurchaseDocGoodsReceipt)
}

func (h *PurchaseDocumentHandler) DeleteGoodsReceipt(w http.ResponseWriter, r *http.Request) {
	h.deleteDocument(w, r, models.PurchaseDocGoodsReceipt)
}

func (h *PurchaseDocumentHandler) GoodsReceiptsToInvoice(w http.ResponseWriter, r *http.Request) {
	h.transformList(w, r, models.PurchaseDocInvoice)
}

// --- Invoices ---

func (h *PurchaseDocumentHandler) ListInvoices(w http.ResponseWriter, r *http.Request) {
	h.listDocuments(w, r, models.PurchaseDocInvoice)
}

func (h *PurchaseDocumentHandler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
	h.createDocument(w, r, models.PurchaseDocInvoice)
}

func (h *PurchaseDocumentHandler) OpenBalance(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.OpenBalance(r.Context(), r.URL.Query().Get("idVendor"), limit, offset)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *PurchaseDocumentHandler) GetInvoicePDF(w http.ResponseWriter, r *http.Request) {
	h.servePDF(w, r)
}

func (h *PurchaseDocumentHandler) GetInvoice(w http.ResponseWriter, r *http.Request) {
	onlyHeader := parseBoolQuery(r, "onlyHeaderData", false)
	h.getDocument(w, r, models.PurchaseDocInvoice, onlyHeader)
}

func (h *PurchaseDocumentHandler) UpdateInvoice(w http.ResponseWriter, r *http.Request) {
	h.updateDocument(w, r, models.PurchaseDocInvoice)
}

func (h *PurchaseDocumentHandler) DeleteInvoice(w http.ResponseWriter, r *http.Request) {
	h.deleteDocument(w, r, models.PurchaseDocInvoice)
}

// --- Vendor payments ---

func (h *PurchaseDocumentHandler) ListVendorPayments(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.ListVendorPayments(r.Context(), models.ListQuery{Limit: limit, Offset: offset})
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *PurchaseDocumentHandler) CreateVendorPayment(w http.ResponseWriter, r *http.Request) {
	var in models.VendorPayment
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.CreateVendorPayment(r.Context(), in)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *PurchaseDocumentHandler) GetVendorPaymentByID(w http.ResponseWriter, r *http.Request) {
	id, ok := parseInt32Path(mux.Vars(r), "id")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid payment id", "")
		return
	}
	item, err := h.svc.GetVendorPayment(r.Context(), id)
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

func (h *PurchaseDocumentHandler) GetVendorPaymentByDoc(w http.ResponseWriter, r *http.Request) {
	id, ok := purchaseDocPathID(mux.Vars(r))
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid document id", "")
		return
	}
	item, err := h.svc.GetVendorPaymentByDoc(r.Context(), id)
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

func (h *PurchaseDocumentHandler) UpdateVendorPayment(w http.ResponseWriter, r *http.Request) {
	id, ok := purchaseDocPathID(mux.Vars(r))
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid payment id", "")
		return
	}
	var in models.VendorPayment
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.UpdateVendorPayment(r.Context(), id, in)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *PurchaseDocumentHandler) DeleteVendorPayment(w http.ResponseWriter, r *http.Request) {
	id, ok := purchaseDocPathID(mux.Vars(r))
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid payment id", "")
		return
	}
	if err := h.svc.DeleteVendorPayment(r.Context(), id); err != nil {
		h.respondError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- PDF & attachments ---

func (h *PurchaseDocumentHandler) GetDocumentPDF(w http.ResponseWriter, r *http.Request) {
	h.servePDF(w, r)
}

func (h *PurchaseDocumentHandler) GetAttachment(w http.ResponseWriter, r *http.Request) {
	id, ok := purchaseDocPathID(mux.Vars(r))
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

func (h *PurchaseDocumentHandler) UploadAttachment(w http.ResponseWriter, r *http.Request) {
	id, ok := purchaseDocPathID(mux.Vars(r))
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

func (h *PurchaseDocumentHandler) DeleteAttachment(w http.ResponseWriter, r *http.Request) {
	id, ok := purchaseDocPathID(mux.Vars(r))
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

func (h *PurchaseDocumentHandler) listDocuments(w http.ResponseWriter, r *http.Request, kind string) {
	limit, offset := parseLimitOffset(r)
	q := purchaseDocListQuery(r, limit, offset)
	items, err := h.svc.ListDocuments(r.Context(), kind, q)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *PurchaseDocumentHandler) createDocument(w http.ResponseWriter, r *http.Request, kind string) {
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

func (h *PurchaseDocumentHandler) getDocument(w http.ResponseWriter, r *http.Request, kind string, onlyHeader bool) {
	id, ok := purchaseDocPathID(mux.Vars(r))
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

func (h *PurchaseDocumentHandler) updateDocument(w http.ResponseWriter, r *http.Request, kind string) {
	id, ok := purchaseDocPathID(mux.Vars(r))
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

func (h *PurchaseDocumentHandler) deleteDocument(w http.ResponseWriter, r *http.Request, kind string) {
	id, ok := purchaseDocPathID(mux.Vars(r))
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

func (h *PurchaseDocumentHandler) transformList(w http.ResponseWriter, r *http.Request, targetKind string) {
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

func (h *PurchaseDocumentHandler) servePDF(w http.ResponseWriter, r *http.Request) {
	id, ok := purchaseDocPathID(mux.Vars(r))
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

func purchaseDocListQuery(r *http.Request, limit, offset int) models.PurchaseDocumentListQuery {
	gt := r.URL.Query().Get("gtDocDateMillins")
	if gt == "" {
		gt = r.URL.Query().Get("gtDocDateMillis")
	}
	lt := r.URL.Query().Get("ltDocDateMillins")
	if lt == "" {
		lt = r.URL.Query().Get("ltDocDateMillis")
	}
	return models.PurchaseDocumentListQuery{
		ListQuery:       models.ListQuery{Limit: limit, Offset: offset},
		FreeText:        r.URL.Query().Get("freeText"),
		IDContact:       parseOptionalInt32Query(r, "idContact"),
		DocStatus:       r.URL.Query().Get("docStatus"),
		NeDocStatus:     r.URL.Query().Get("neDocStatus"),
		GtDocDateMillis: gt,
		LtDocDateMillis: lt,
	}
}

func purchaseDocPathID(vars map[string]string) (int32, bool) {
	if id, ok := parseInt32Path(vars, "idDoc"); ok {
		return id, true
	}
	return parseInt32Path(vars, "id")
}

func (h *PurchaseDocumentHandler) respondError(w http.ResponseWriter, err error) {
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
