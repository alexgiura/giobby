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

type ContactHandler struct {
	svc services.ContactService
}

func NewContactHandler(svc services.ContactService) *ContactHandler {
	return &ContactHandler{svc: svc}
}

func (h *ContactHandler) ListContacts(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.ListContacts(r.Context(), models.ContactListQuery{
		ListQuery:      models.ListQuery{Limit: limit, Offset: offset},
		Type:           r.URL.Query().Get("type"),
		VisibilityType: r.URL.Query().Get("visibilityType"),
		FullName:       r.URL.Query().Get("fullName"),
		FreeText:       r.URL.Query().Get("freeText"),
		Deleted:        parseOptionalBoolQuery(r, "deleted"),
		Email:          r.URL.Query().Get("email"),
		FiscalCode:     r.URL.Query().Get("fiscalCode"),
		VatCode:        r.URL.Query().Get("vatCode"),
		OnlyCustomers:  parseBoolQuery(r, "onlyCustomers", false),
		OnlyVendors:    parseBoolQuery(r, "onlyVendors", false),
		OnlyLeads:      parseBoolQuery(r, "onlyLeads", false),
		RetrieveImage:  parseBoolQuery(r, "retrieveImage", false),
	})
	if err != nil {
		h.respondDBError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *ContactHandler) CreateContact(w http.ResponseWriter, r *http.Request) {
	var in models.Contact
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.CreateContact(r.Context(), in)
	if err != nil {
		h.respondValidationOrDB(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *ContactHandler) ListContactSources(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.ListContactSources(r.Context(), models.ContactSourceListQuery{
		ListQuery: models.ListQuery{Limit: limit, Offset: offset},
		FreeText:  r.URL.Query().Get("freeText"),
	})
	if err != nil {
		h.respondDBError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *ContactHandler) GetContact(w http.ResponseWriter, r *http.Request) {
	id, ok := parseContactID(mux.Vars(r))
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid contact id", "")
		return
	}
	item, err := h.svc.GetContact(r.Context(), id, false)
	if err != nil {
		h.respondDBError(w, err)
		return
	}
	if item == nil {
		respondWithError(w, http.StatusNotFound, ErrCodeNotFound, "Contact not found", "")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *ContactHandler) UpdateContact(w http.ResponseWriter, r *http.Request) {
	id, ok := parseContactID(mux.Vars(r))
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid contact id", "")
		return
	}
	var in models.Contact
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.UpdateContact(r.Context(), id, in)
	if err != nil {
		h.respondValidationOrDB(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *ContactHandler) PatchContact(w http.ResponseWriter, r *http.Request) {
	id, ok := parseContactIDFromKey(mux.Vars(r), "idcontact")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid contact id", "")
		return
	}
	var patch map[string]any
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.PatchContact(r.Context(), id, patch)
	if err != nil {
		h.respondValidationOrDB(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *ContactHandler) DeleteContact(w http.ResponseWriter, r *http.Request) {
	id, ok := parseContactIDFromKey(mux.Vars(r), "idcontact")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid contact id", "")
		return
	}
	if err := h.svc.DeleteContact(r.Context(), id); err != nil {
		h.respondValidationOrDB(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ContactHandler) UploadPhoto(w http.ResponseWriter, r *http.Request) {
	id, ok := parseContactID(mux.Vars(r))
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid contact id", "")
		return
	}
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid multipart form", err.Error())
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		file, header, err = r.FormFile("photo")
	}
	if err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Photo file required", err.Error())
		return
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Failed to read photo", err.Error())
		return
	}
	mime := header.Header.Get("Content-Type")
	if mime == "" {
		mime = "application/octet-stream"
	}
	if err := h.svc.UpdateContactPhoto(r.Context(), id, data, mime); err != nil {
		h.respondValidationOrDB(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ContactHandler) ListOffices(w http.ResponseWriter, r *http.Request) {
	contactID, ok := parseContactIDFromKey(mux.Vars(r), "idcontact")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid contact id", "")
		return
	}
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.ListOffices(r.Context(), contactID, models.OfficeListQuery{
		ListQuery:   models.ListQuery{Limit: limit, Offset: offset},
		DefaultDest: parseOptionalBoolQuery(r, "defaultDest"),
	})
	if err != nil {
		h.respondDBError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *ContactHandler) CreateOffice(w http.ResponseWriter, r *http.Request) {
	contactID, ok := parseContactIDFromKey(mux.Vars(r), "idcontact")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid contact id", "")
		return
	}
	var in models.ContactOffice
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.CreateOffice(r.Context(), contactID, in)
	if err != nil {
		h.respondValidationOrDB(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *ContactHandler) UpdateOffice(w http.ResponseWriter, r *http.Request) {
	contactID, ok := parseContactIDFromKey(mux.Vars(r), "idcontact")
	officeID, ok2 := parseInt32Path(mux.Vars(r), "idoffice")
	if !ok || !ok2 {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid ids", "")
		return
	}
	var in models.ContactOffice
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.UpdateOffice(r.Context(), contactID, officeID, in)
	if err != nil {
		h.respondValidationOrDB(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *ContactHandler) DeleteOffice(w http.ResponseWriter, r *http.Request) {
	contactID, ok := parseContactIDFromKey(mux.Vars(r), "idcontact")
	officeID, ok2 := parseInt32Path(mux.Vars(r), "idoffice")
	if !ok || !ok2 {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid ids", "")
		return
	}
	if err := h.svc.DeleteOffice(r.Context(), contactID, officeID); err != nil {
		h.respondValidationOrDB(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ContactHandler) ListSubContacts(w http.ResponseWriter, r *http.Request) {
	contactID, ok := parseContactIDFromKey(mux.Vars(r), "idcontact")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid contact id", "")
		return
	}
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.ListSubContacts(r.Context(), contactID, models.SubContactListQuery{
		ListQuery: models.ListQuery{Limit: limit, Offset: offset},
		Type:      r.URL.Query().Get("type"),
		Deleted:   parseOptionalBoolQuery(r, "deleted"),
	})
	if err != nil {
		h.respondDBError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *ContactHandler) CreateSubContact(w http.ResponseWriter, r *http.Request) {
	contactID, ok := parseContactIDFromKey(mux.Vars(r), "idcontact")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid contact id", "")
		return
	}
	var in models.SubContactAssoc
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.CreateSubContact(r.Context(), contactID, in)
	if err != nil {
		h.respondValidationOrDB(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *ContactHandler) DeleteSubContact(w http.ResponseWriter, r *http.Request) {
	contactID, ok := parseContactIDFromKey(mux.Vars(r), "idcontact")
	subID, ok2 := parseInt32Path(mux.Vars(r), "idsubcontact")
	if !ok || !ok2 {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid ids", "")
		return
	}
	if err := h.svc.DeleteSubContact(r.Context(), contactID, subID); err != nil {
		h.respondValidationOrDB(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseContactID(vars map[string]string) (int32, bool) {
	if id, ok := parseInt32Path(vars, "id"); ok {
		return id, true
	}
	return parseContactIDFromKey(vars, "idcontact")
}

func parseContactIDFromKey(vars map[string]string, key string) (int32, bool) {
	return parseInt32Path(vars, key)
}

func (h *ContactHandler) respondValidationOrDB(w http.ResponseWriter, err error) {
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
	h.respondDBError(w, err)
}

func (h *ContactHandler) respondDBError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	status, code, msg := parseDatabaseError(err)
	respondWithError(w, status, code, msg, err.Error())
}
