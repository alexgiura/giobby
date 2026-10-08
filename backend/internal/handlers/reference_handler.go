package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/services"

	"github.com/gorilla/mux"
)

type ReferenceHandler struct {
	svc services.ReferenceService
}

func NewReferenceHandler(svc services.ReferenceService) *ReferenceHandler {
	return &ReferenceHandler{svc: svc}
}

func (h *ReferenceHandler) ListCities(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.ListCities(r.Context(), models.CityListQuery{
		ListQuery: models.ListQuery{Limit: limit, Offset: offset},
		Name:      r.URL.Query().Get("name"),
	})
	if err != nil {
		h.respondDBError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *ReferenceHandler) CreateCity(w http.ResponseWriter, r *http.Request) {
	var in models.City
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.CreateCity(r.Context(), in)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "required") {
			respondWithError(w, http.StatusBadRequest, ErrCodeValidationFailed, err.Error(), "")
			return
		}
		h.respondDBError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *ReferenceHandler) ListCurrencies(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.ListCurrencies(r.Context(), models.ListQuery{Limit: limit, Offset: offset})
	if err != nil {
		h.respondDBError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *ReferenceHandler) GetCurrency(w http.ResponseWriter, r *http.Request) {
	code := mux.Vars(r)["code"]
	item, err := h.svc.GetCurrency(r.Context(), code)
	if err != nil {
		h.respondDBError(w, err)
		return
	}
	if item == nil {
		respondWithError(w, http.StatusNotFound, ErrCodeNotFound, "Currency not found", "")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *ReferenceHandler) ListUms(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.ListUms(r.Context(), models.UmListQuery{
		ListQuery:        models.ListQuery{Limit: limit, Offset: offset},
		Description:      r.URL.Query().Get("description"),
		ShortDescription: r.URL.Query().Get("shortdescription"),
		ShowDeleted:      parseBoolQuery(r, "showdeleted", false),
	})
	if err != nil {
		h.respondDBError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *ReferenceHandler) CreateUm(w http.ResponseWriter, r *http.Request) {
	var in models.Um
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	if err := h.svc.CreateUm(r.Context(), in); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "required") {
			respondWithError(w, http.StatusBadRequest, ErrCodeValidationFailed, err.Error(), "")
			return
		}
		h.respondDBError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, in)
}

func (h *ReferenceHandler) GetUm(w http.ResponseWriter, r *http.Request) {
	idUm := mux.Vars(r)["idUm"]
	item, err := h.svc.GetUm(r.Context(), idUm, parseBoolQuery(r, "showdeleted", false))
	if err != nil {
		h.respondDBError(w, err)
		return
	}
	if item == nil {
		respondWithError(w, http.StatusNotFound, ErrCodeNotFound, "Unit of measure not found", "")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *ReferenceHandler) UpdateUm(w http.ResponseWriter, r *http.Request) {
	idUm := mux.Vars(r)["idUm"]
	var in models.Um
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	if err := h.svc.UpdateUm(r.Context(), idUm, in); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			respondWithError(w, http.StatusNotFound, ErrCodeNotFound, "Unit of measure not found", "")
			return
		}
		h.respondDBError(w, err)
		return
	}
	in.Um = idUm
	writeJSON(w, http.StatusOK, in)
}

func (h *ReferenceHandler) DeleteUm(w http.ResponseWriter, r *http.Request) {
	idUm := mux.Vars(r)["idUm"]
	if err := h.svc.DeleteUm(r.Context(), idUm); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			respondWithError(w, http.StatusNotFound, ErrCodeNotFound, "Unit of measure not found", "")
			return
		}
		h.respondDBError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ReferenceHandler) ListOfficeTypes(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.ListOfficeTypes(r.Context(), models.OfficeTypeListQuery{
		ListQuery:   models.ListQuery{Limit: limit, Offset: offset},
		Description: r.URL.Query().Get("description"),
	})
	if err != nil {
		h.respondDBError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *ReferenceHandler) ListContactRoles(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.ListContactRoles(r.Context(), models.ContactRoleListQuery{
		ListQuery:   models.ListQuery{Limit: limit, Offset: offset},
		Description: r.URL.Query().Get("description"),
	})
	if err != nil {
		h.respondDBError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *ReferenceHandler) ListPaymentTerms(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.ListPaymentTerms(r.Context(), models.PaymentTermListQuery{
		ListQuery:   models.ListQuery{Limit: limit, Offset: offset},
		Description: r.URL.Query().Get("description"),
	})
	if err != nil {
		h.respondDBError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *ReferenceHandler) GetPaymentTerm(w http.ResponseWriter, r *http.Request) {
	id, ok := parseInt32Path(mux.Vars(r), "id")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid payment term id", "")
		return
	}
	item, err := h.svc.GetPaymentTerm(r.Context(), id)
	if err != nil {
		h.respondDBError(w, err)
		return
	}
	if item == nil {
		respondWithError(w, http.StatusNotFound, ErrCodeNotFound, "Payment term not found", "")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *ReferenceHandler) CreatePaymentTerm(w http.ResponseWriter, r *http.Request) {
	var in models.PaymentTerm
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.CreatePaymentTerm(r.Context(), in)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "required") {
			respondWithError(w, http.StatusBadRequest, ErrCodeValidationFailed, err.Error(), "")
			return
		}
		h.respondDBError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *ReferenceHandler) respondDBError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	status, code, msg := parseDatabaseError(err)
	respondWithError(w, status, code, msg, err.Error())
}
