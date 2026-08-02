package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/services"

	"github.com/gorilla/mux"
)

type PricelistHandler struct {
	svc services.PricelistService
}

func NewPricelistHandler(svc services.PricelistService) *PricelistHandler {
	return &PricelistHandler{svc: svc}
}

func (h *PricelistHandler) ListPricelists(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.ListPricelists(r.Context(), models.PricelistListQuery{
		ListQuery:          models.ListQuery{Limit: limit, Offset: offset},
		Type:               r.URL.Query().Get("type"),
		ValidOnDateMillins: r.URL.Query().Get("validOnDateMillins"),
	})
	if err != nil {
		h.respondEntityError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *PricelistHandler) CreatePricelist(w http.ResponseWriter, r *http.Request) {
	var in models.Pricelist
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.CreatePricelist(r.Context(), in)
	if err != nil {
		h.respondEntityError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *PricelistHandler) GetPricelist(w http.ResponseWriter, r *http.Request) {
	id, ok := parseInt32Path(mux.Vars(r), "id")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid pricelist id", "")
		return
	}
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.GetPricelistRows(r.Context(), id, models.PricelistDetailQuery{
		ListQuery:          models.ListQuery{Limit: limit, Offset: offset},
		Type:               r.URL.Query().Get("type"),
		ValidOnDateMillins: r.URL.Query().Get("validOnDateMillins"),
		ProdDescription:    r.URL.Query().Get("prodDescription"),
		MaterialID:         r.URL.Query().Get("materialId"),
	})
	if err != nil {
		h.respondEntityError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *PricelistHandler) UpdatePricelist(w http.ResponseWriter, r *http.Request) {
	id, ok := parseInt32Path(mux.Vars(r), "id")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid pricelist id", "")
		return
	}
	var in models.Pricelist
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.UpdatePricelist(r.Context(), id, in)
	if err != nil {
		h.respondEntityError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *PricelistHandler) DeletePricelist(w http.ResponseWriter, r *http.Request) {
	id, ok := parseInt32Path(mux.Vars(r), "id")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid pricelist id", "")
		return
	}
	if err := h.svc.DeletePricelist(r.Context(), id); err != nil {
		h.respondEntityError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *PricelistHandler) ListPricelistSchemes(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.ListPricelistSchemes(r.Context(), models.ListQuery{Limit: limit, Offset: offset})
	if err != nil {
		h.respondEntityError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *PricelistHandler) GetPricelistScheme(w http.ResponseWriter, r *http.Request) {
	id, ok := parseInt32Path(mux.Vars(r), "id")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid scheme id", "")
		return
	}
	item, err := h.svc.GetPricelistScheme(r.Context(), id)
	if err != nil {
		h.respondEntityError(w, err)
		return
	}
	if item == nil {
		respondWithError(w, http.StatusNotFound, ErrCodeNotFound, "Pricelist scheme not found", "")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *PricelistHandler) respondEntityError(w http.ResponseWriter, err error) {
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

func (h *PricelistHandler) respondDBError(w http.ResponseWriter, err error) {
	status, code, msg := parseDatabaseError(err)
	respondWithError(w, status, code, msg, err.Error())
}
