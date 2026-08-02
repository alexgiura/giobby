package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/services"

	"github.com/gorilla/mux"
)

type VendorHandler struct {
	svc services.VendorService
}

func NewVendorHandler(svc services.VendorService) *VendorHandler {
	return &VendorHandler{svc: svc}
}

func (h *VendorHandler) ListVendors(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.ListVendors(r.Context(), models.VendorListQuery{
		ListQuery: models.ListQuery{Limit: limit, Offset: offset},
		Name:      r.URL.Query().Get("name"),
		FreeText:  r.URL.Query().Get("freeText"),
		Deleted:   parseOptionalBoolQuery(r, "deleted"),
	})
	if err != nil {
		h.respondDBError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *VendorHandler) CreateVendor(w http.ResponseWriter, r *http.Request) {
	var in models.Vendor
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.CreateVendor(r.Context(), in)
	if err != nil {
		h.respondValidationOrDB(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *VendorHandler) GetVendor(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["idvendor"]
	item, err := h.svc.GetVendor(r.Context(), id, parseBoolQuery(r, "onlyVendorData", false))
	if err != nil {
		h.respondValidationOrDB(w, err)
		return
	}
	if item == nil {
		respondWithError(w, http.StatusNotFound, ErrCodeNotFound, "Vendor not found", "")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *VendorHandler) UpdateVendor(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["idvendor"]
	var in models.Vendor
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.UpdateVendor(r.Context(), id, in)
	if err != nil {
		h.respondValidationOrDB(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *VendorHandler) PatchVendor(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["idvendor"]
	var patch map[string]any
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.PatchVendor(r.Context(), id, patch)
	if err != nil {
		h.respondValidationOrDB(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *VendorHandler) DeleteVendor(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["idvendor"]
	if err := h.svc.DeleteVendor(r.Context(), id); err != nil {
		h.respondValidationOrDB(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *VendorHandler) GetVendorReport(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["idvendor"]
	item, err := h.svc.GetVendorReport(r.Context(), id)
	if err != nil {
		h.respondValidationOrDB(w, err)
		return
	}
	if item == nil {
		respondWithError(w, http.StatusNotFound, ErrCodeNotFound, "Vendor not found", "")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *VendorHandler) respondValidationOrDB(w http.ResponseWriter, err error) {
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

func (h *VendorHandler) respondDBError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	status, code, msg := parseDatabaseError(err)
	respondWithError(w, status, code, msg, err.Error())
}
