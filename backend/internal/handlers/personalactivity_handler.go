package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/services"

	"github.com/gorilla/mux"
)

type PersonalActivityHandler struct {
	svc services.PersonalActivityService
}

func NewPersonalActivityHandler(svc services.PersonalActivityService) *PersonalActivityHandler {
	return &PersonalActivityHandler{svc: svc}
}

func (h *PersonalActivityHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.List(r.Context(), models.PersonalActivityListQuery{
		ListQuery: models.ListQuery{Limit: limit, Offset: offset},
		Status:    r.URL.Query().Get("status"),
		Name:      r.URL.Query().Get("name"),
	})
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *PersonalActivityHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseInt32Path(mux.Vars(r), "id")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid activity id", "")
		return
	}
	item, err := h.svc.Get(r.Context(), id)
	if err != nil {
		h.respondError(w, err)
		return
	}
	if item == nil {
		respondWithError(w, http.StatusNotFound, ErrCodeNotFound, "Activity not found", "")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *PersonalActivityHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in models.PersonalActivity
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.Create(r.Context(), in)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *PersonalActivityHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseInt32Path(mux.Vars(r), "id")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid activity id", "")
		return
	}
	var in models.PersonalActivity
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.Update(r.Context(), id, in)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *PersonalActivityHandler) Patch(w http.ResponseWriter, r *http.Request) {
	id, ok := parseInt32Path(mux.Vars(r), "id")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid activity id", "")
		return
	}
	var fields map[string]any
	if err := json.NewDecoder(r.Body).Decode(&fields); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.Patch(r.Context(), id, fields)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *PersonalActivityHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseInt32Path(mux.Vars(r), "id")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid activity id", "")
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		h.respondError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *PersonalActivityHandler) respondError(w http.ResponseWriter, err error) {
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
