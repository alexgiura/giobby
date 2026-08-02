package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/services"

	"github.com/gorilla/mux"
)

type LotHandler struct {
	svc services.LotService
}

func NewLotHandler(svc services.LotService) *LotHandler {
	return &LotHandler{svc: svc}
}

func (h *LotHandler) ListLots(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.ListLots(r.Context(), models.LotListQuery{
		ListQuery:  models.ListQuery{Limit: limit, Offset: offset},
		IDMaterial: r.URL.Query().Get("idMaterial"),
		IDLot:      r.URL.Query().Get("idLot"),
	})
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *LotHandler) GetLot(w http.ResponseWriter, r *http.Request) {
	idLot := lotPathID(mux.Vars(r))
	item, err := h.svc.GetLot(r.Context(), idLot)
	if err != nil {
		h.respondError(w, err)
		return
	}
	if item == nil {
		respondWithError(w, http.StatusNotFound, ErrCodeNotFound, "Lot not found", "")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *LotHandler) CreateLot(w http.ResponseWriter, r *http.Request) {
	var in models.Lot
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.CreateLot(r.Context(), in)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *LotHandler) UpdateLot(w http.ResponseWriter, r *http.Request) {
	idLot := lotPathID(mux.Vars(r))
	var in models.Lot
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.UpdateLot(r.Context(), idLot, in)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *LotHandler) DeleteLot(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteLot(r.Context(), lotPathID(mux.Vars(r))); err != nil {
		h.respondError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func lotPathID(vars map[string]string) string {
	if v := vars["idLot"]; v != "" {
		return v
	}
	return vars["idlot"]
}

func (h *LotHandler) respondError(w http.ResponseWriter, err error) {
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
