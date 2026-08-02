package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/services"

	"github.com/gorilla/mux"
)

type StorageLocationHandler struct {
	svc services.StorageLocationService
}

func NewStorageLocationHandler(svc services.StorageLocationService) *StorageLocationHandler {
	return &StorageLocationHandler{svc: svc}
}

func (h *StorageLocationHandler) ListStorageLocations(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.ListStorageLocations(r.Context(), models.StorageLocationListQuery{
		ListQuery:      models.ListQuery{Limit: limit, Offset: offset},
		IDStorage:      r.URL.Query().Get("idStorage"),
		ManagedStorage: parseOptionalBoolQuery(r, "managedStorage"),
	})
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *StorageLocationHandler) GetStorageLocation(w http.ResponseWriter, r *http.Request) {
	item, err := h.svc.GetStorageLocation(r.Context(), mux.Vars(r)["idLocation"])
	if err != nil {
		h.respondError(w, err)
		return
	}
	if item == nil {
		respondWithError(w, http.StatusNotFound, ErrCodeNotFound, "Storage location not found", "")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *StorageLocationHandler) CreateStorageLocation(w http.ResponseWriter, r *http.Request) {
	var in models.StorageLocation
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.CreateStorageLocation(r.Context(), in)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *StorageLocationHandler) UpdateStorageLocation(w http.ResponseWriter, r *http.Request) {
	var in models.StorageLocation
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.UpdateStorageLocation(r.Context(), mux.Vars(r)["idLocation"], in)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *StorageLocationHandler) DeleteStorageLocation(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteStorageLocation(r.Context(), mux.Vars(r)["idLocation"]); err != nil {
		h.respondError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *StorageLocationHandler) respondError(w http.ResponseWriter, err error) {
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
