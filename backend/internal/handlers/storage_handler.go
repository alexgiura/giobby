package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/services"

	"github.com/gorilla/mux"
)

type StorageHandler struct {
	svc    services.StorageService
	locSvc services.StorageLocationService
}

func NewStorageHandler(svc services.StorageService, locSvc services.StorageLocationService) *StorageHandler {
	return &StorageHandler{svc: svc, locSvc: locSvc}
}

func (h *StorageHandler) ListStorages(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.ListStorages(r.Context(), models.StorageListQuery{
		ListQuery: models.ListQuery{Limit: limit, Offset: offset},
		ID:        r.URL.Query().Get("id"),
		Managed:   parseOptionalBoolQuery(r, "managed"),
	})
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *StorageHandler) GetStorage(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	item, err := h.svc.GetStorage(r.Context(), id)
	if err != nil {
		h.respondError(w, err)
		return
	}
	if item == nil {
		respondWithError(w, http.StatusNotFound, ErrCodeNotFound, "Storage not found", "")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *StorageHandler) CreateStorage(w http.ResponseWriter, r *http.Request) {
	var in models.Storage
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.CreateStorage(r.Context(), in)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *StorageHandler) UpdateStorage(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var in models.Storage
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.UpdateStorage(r.Context(), id, in)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *StorageHandler) DeleteStorage(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteStorage(r.Context(), mux.Vars(r)["id"]); err != nil {
		h.respondError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *StorageHandler) CreateNestedLocation(w http.ResponseWriter, r *http.Request) {
	idStorage := mux.Vars(r)["idStorage"]
	var in models.StorageLocation
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	in.IDStorage = idStorage
	item, err := h.locSvc.CreateStorageLocation(r.Context(), in)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *StorageHandler) GetNestedLocation(w http.ResponseWriter, r *http.Request) {
	idLocation := mux.Vars(r)["idLocation"]
	item, err := h.locSvc.GetStorageLocation(r.Context(), idLocation)
	if err != nil {
		h.respondError(w, err)
		return
	}
	if item == nil || item.IDStorage != mux.Vars(r)["idStorage"] {
		respondWithError(w, http.StatusNotFound, ErrCodeNotFound, "Storage location not found", "")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *StorageHandler) UpdateNestedLocation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var in models.StorageLocation
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	in.IDStorage = vars["idStorage"]
	item, err := h.locSvc.UpdateStorageLocation(r.Context(), vars["idLocation"], in)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *StorageHandler) DeleteNestedLocation(w http.ResponseWriter, r *http.Request) {
	if err := h.locSvc.DeleteStorageLocation(r.Context(), mux.Vars(r)["idLocation"]); err != nil {
		h.respondError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *StorageHandler) respondError(w http.ResponseWriter, err error) {
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
