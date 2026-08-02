package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/services"
)

type SettingsHandler struct {
	svc services.SettingsService
}

func NewSettingsHandler(svc services.SettingsService) *SettingsHandler {
	return &SettingsHandler{svc: svc}
}

func (h *SettingsHandler) ListSettings(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListSettings(r.Context(), models.AppSettingListQuery{
		Key1: r.URL.Query().Get("key1"),
		Key2: r.URL.Query().Get("key2"),
		Key3: r.URL.Query().Get("key3"),
		Key4: r.URL.Query().Get("key4"),
		Key5: r.URL.Query().Get("key5"),
	})
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *SettingsHandler) UpsertSetting(w http.ResponseWriter, r *http.Request) {
	var in models.AppSetting
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.UpsertSetting(r.Context(), in)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *SettingsHandler) respondError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "required") || strings.Contains(msg, "invalid") {
		respondWithError(w, http.StatusBadRequest, ErrCodeValidationFailed, err.Error(), "")
		return
	}
	status, code, m := parseDatabaseError(err)
	respondWithError(w, status, code, m, err.Error())
}
