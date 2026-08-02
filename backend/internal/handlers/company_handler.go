package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/services"

	"github.com/gorilla/mux"
)

type CompanyHandler struct {
	svc services.CompanyService
}

func NewCompanyHandler(svc services.CompanyService) *CompanyHandler {
	return &CompanyHandler{svc: svc}
}

func (h *CompanyHandler) GetCompany(w http.ResponseWriter, r *http.Request) {
	id, ok := parseInt64Path(mux.Vars(r), "id")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid company id", "")
		return
	}
	_ = parseBoolQuery(r, "retrieveImage", false) // image not stored yet
	item, err := h.svc.GetCompany(r.Context(), id)
	if err != nil {
		h.respondDBError(w, err)
		return
	}
	if item == nil {
		respondWithError(w, http.StatusNotFound, ErrCodeNotFound, "Company not found", "")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *CompanyHandler) ListAccountCenters(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.ListAccountCenters(r.Context(), models.AccountCenterListQuery{
		ListQuery: models.ListQuery{Limit: limit, Offset: offset},
		Type:      r.URL.Query().Get("type"),
	})
	if err != nil {
		h.respondDBError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *CompanyHandler) CreateAccountCenter(w http.ResponseWriter, r *http.Request) {
	var in models.AccountCenter
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.CreateAccountCenter(r.Context(), in)
	if err != nil {
		h.respondValidationOrDB(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *CompanyHandler) GetAccountCenter(w http.ResponseWriter, r *http.Request) {
	id, ok := parseAccountCenterID(mux.Vars(r), "idAccountCenter")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid account center id", "")
		return
	}
	item, err := h.svc.GetAccountCenter(r.Context(), id)
	if err != nil {
		h.respondDBError(w, err)
		return
	}
	if item == nil {
		respondWithError(w, http.StatusNotFound, ErrCodeNotFound, "Account center not found", "")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *CompanyHandler) UpdateAccountCenter(w http.ResponseWriter, r *http.Request) {
	id, ok := parseAccountCenterID(mux.Vars(r), "idAccountCenter")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid account center id", "")
		return
	}
	var in models.AccountCenter
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.UpdateAccountCenter(r.Context(), id, in)
	if err != nil {
		h.respondValidationOrDB(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *CompanyHandler) DeleteAccountCenter(w http.ResponseWriter, r *http.Request) {
	id, ok := parseAccountCenterID(mux.Vars(r), "idAccountCenter")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid account center id", "")
		return
	}
	if err := h.svc.DeleteAccountCenter(r.Context(), id); err != nil {
		h.respondValidationOrDB(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CompanyHandler) ListAccountCodes(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.ListAccountCodes(r.Context(), models.AccountCodeListQuery{
		ListQuery:    models.ListQuery{Limit: limit, Offset: offset},
		Entry:        r.URL.Query().Get("type"),
		EntryType:    r.URL.Query().Get("entryType"),
		ShowDisabled: parseBoolQuery(r, "showDisabled", false),
	})
	if err != nil {
		h.respondDBError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *CompanyHandler) ListBanks(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.ListBanks(r.Context(), models.BankListQuery{
		ListQuery:   models.ListQuery{Limit: limit, Offset: offset},
		IsCashdesk:  parseOptionalBoolQuery(r, "isCashdesk"),
		AccountCode: r.URL.Query().Get("accountCode"),
		Iban:        r.URL.Query().Get("iban"),
		Sia:         r.URL.Query().Get("sia"),
		Swift:       r.URL.Query().Get("swift"),
		Cuc:         r.URL.Query().Get("cuc"),
		ShowDeleted: parseBoolQuery(r, "showDeleted", false),
	})
	if err != nil {
		h.respondDBError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *CompanyHandler) CreateBank(w http.ResponseWriter, r *http.Request) {
	var in models.BankCashdesk
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.CreateBank(r.Context(), in)
	if err != nil {
		h.respondValidationOrDB(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *CompanyHandler) GetBank(w http.ResponseWriter, r *http.Request) {
	id, ok := parseInt32Path(mux.Vars(r), "id")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid bank id", "")
		return
	}
	item, err := h.svc.GetBank(r.Context(), id)
	if err != nil {
		h.respondDBError(w, err)
		return
	}
	if item == nil {
		respondWithError(w, http.StatusNotFound, ErrCodeNotFound, "Bank not found", "")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *CompanyHandler) UpdateBank(w http.ResponseWriter, r *http.Request) {
	id, ok := parseInt32Path(mux.Vars(r), "id")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid bank id", "")
		return
	}
	var in models.BankCashdesk
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.UpdateBank(r.Context(), id, in)
	if err != nil {
		h.respondValidationOrDB(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *CompanyHandler) DeleteBank(w http.ResponseWriter, r *http.Request) {
	id, ok := parseInt32Path(mux.Vars(r), "id")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid bank id", "")
		return
	}
	if err := h.svc.DeleteBank(r.Context(), id); err != nil {
		h.respondValidationOrDB(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CompanyHandler) ListBupNumerators(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListBupNumerators(r.Context(), models.BupNumeratorQuery{
		IDDocumentType: parseOptionalInt32Query(r, "idDocumentType"),
		IDPlugin:       r.URL.Query().Get("idPlugin"),
	})
	if err != nil {
		h.respondDBError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *CompanyHandler) GetCurrencyChange(w http.ResponseWriter, r *http.Request) {
	source := mux.Vars(r)["idSourceCurr"]
	var asOf *time.Time
	if ms := strings.TrimSpace(r.URL.Query().Get("currencyChangeDateMillis")); ms != "" {
		if v, err := strconv.ParseInt(ms, 10, 64); err == nil {
			t := time.UnixMilli(v).UTC()
			asOf = &t
		}
	}
	item, err := h.svc.GetCurrencyChange(r.Context(), source, asOf)
	if err != nil {
		h.respondValidationOrDB(w, err)
		return
	}
	if item == nil {
		respondWithError(w, http.StatusNotFound, ErrCodeNotFound, "Currency exchange not found", "")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *CompanyHandler) ListTaxs(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.ListVatRates(r.Context(), models.VatListQuery{
		ListQuery:    models.ListQuery{Limit: limit, Offset: offset},
		IDBu:         r.URL.Query().Get("idBu"),
		Country:      r.URL.Query().Get("country"),
		State:        r.URL.Query().Get("state"),
		ContactType:  r.URL.Query().Get("contactType"),
		TaxSuperType: parseOptionalInt32Query(r, "taxSuperType"),
		Rate:         parseOptionalFloatQuery(r, "rate"),
	})
	if err != nil {
		h.respondDBError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *CompanyHandler) GetTax(w http.ResponseWriter, r *http.Request) {
	idVat := mux.Vars(r)["idVat"]
	item, err := h.svc.GetVatRate(r.Context(), idVat, models.VatGetQuery{
		IDBu:         r.URL.Query().Get("idBu"),
		Country:      r.URL.Query().Get("country"),
		State:        r.URL.Query().Get("state"),
		ContactType:  r.URL.Query().Get("contactType"),
		TaxSuperType: parseOptionalInt32Query(r, "taxSuperType"),
	})
	if err != nil {
		h.respondValidationOrDB(w, err)
		return
	}
	if item == nil {
		respondWithError(w, http.StatusNotFound, ErrCodeNotFound, "VAT rate not found", "")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func parseAccountCenterID(vars map[string]string, key string) (int32, bool) {
	v := strings.TrimSpace(vars[key])
	if v == "" {
		return 0, false
	}
	n, err := strconv.ParseInt(v, 10, 32)
	if err != nil {
		return 0, false
	}
	return int32(n), true
}

func (h *CompanyHandler) respondValidationOrDB(w http.ResponseWriter, err error) {
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

func (h *CompanyHandler) respondDBError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	status, code, msg := parseDatabaseError(err)
	respondWithError(w, status, code, msg, err.Error())
}
