package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/services"

	"github.com/gorilla/mux"
)

type CountryHandler struct {
	svc services.CountryService
}

func NewCountryHandler(svc services.CountryService) *CountryHandler {
	return &CountryHandler{svc: svc}
}

func (h *CountryHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.ListCountries(r.Context(), models.CountryListQuery{
		ListQuery:   models.ListQuery{Limit: limit, Offset: offset},
		Description: r.URL.Query().Get("description"),
		ShowDeleted: parseBoolQuery(r, "showDeleted", false),
	})
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *CountryHandler) Get(w http.ResponseWriter, r *http.Request) {
	item, err := h.svc.GetCountry(r.Context(), mux.Vars(r)["id"])
	if err != nil {
		h.respondError(w, err)
		return
	}
	if item == nil {
		respondWithError(w, http.StatusNotFound, ErrCodeNotFound, "Country not found", "")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *CountryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in models.Country
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.CreateCountry(r.Context(), in)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *CountryHandler) Update(w http.ResponseWriter, r *http.Request) {
	var in models.Country
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.UpdateCountry(r.Context(), mux.Vars(r)["id"], in)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *CountryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteCountry(r.Context(), mux.Vars(r)["id"]); err != nil {
		h.respondError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CountryHandler) Recover(w http.ResponseWriter, r *http.Request) {
	item, err := h.svc.RecoverCountry(r.Context(), mux.Vars(r)["id"])
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *CountryHandler) respondError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "required") || strings.Contains(msg, "does not match") {
		respondWithError(w, http.StatusBadRequest, ErrCodeValidationFailed, err.Error(), "")
		return
	}
	if strings.Contains(msg, "not found") {
		respondWithError(w, http.StatusNotFound, ErrCodeNotFound, "Country not found", "")
		return
	}
	status, code, message := parseDatabaseError(err)
	respondWithError(w, status, code, message, err.Error())
}
