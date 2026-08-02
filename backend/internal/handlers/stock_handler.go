package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/services"
)

type StockHandler struct {
	svc services.StockService
}

func NewStockHandler(svc services.StockService) *StockHandler {
	return &StockHandler{svc: svc}
}

func (h *StockHandler) CreateStock(w http.ResponseWriter, r *http.Request) {
	var in models.Stock
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.CreateStock(r.Context(), in)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *StockHandler) UpdateStock(w http.ResponseWriter, r *http.Request) {
	var in models.Stock
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.UpdateStock(r.Context(), in)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *StockHandler) ListAvailability(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.ListAvailability(r.Context(), models.StockAvailabilityQuery{
		ListQuery:   models.ListQuery{Limit: limit, Offset: offset},
		IDMaterial:  r.URL.Query().Get("idMaterial"),
		DetailLevel: parseOptionalInt32Query(r, "detailLevel"),
		IDStorage:   r.URL.Query().Get("idStorage"),
		ZeroQtyRows: parseOptionalBoolQuery(r, "zeroQtyRows"),
		IDLocation:  r.URL.Query().Get("idLocation"),
	})
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *StockHandler) ListAvailabilityReport(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.ListAvailabilityReport(r.Context(), models.StockAvailabilityReportQuery{
		ListQuery:        models.ListQuery{Limit: limit, Offset: offset},
		IDMaterial:       r.URL.Query().Get("idMaterial"),
		ZeroQtyRows:      parseOptionalBoolQuery(r, "zeroQtyRows"),
		AttributesDetail: parseOptionalBoolQuery(r, "attributesDetail"),
		StorageDetail:    parseOptionalBoolQuery(r, "storageDetail"),
		LotsDetail:       parseOptionalBoolQuery(r, "lotsDetail"),
		IDStorage:        r.URL.Query().Get("idStorage"),
		IDLocation:       r.URL.Query().Get("idLocation"),
	})
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *StockHandler) respondError(w http.ResponseWriter, err error) {
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
