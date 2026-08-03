package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"dnsc_microservice/internal/middleware"
	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/services"

	"github.com/gorilla/mux"
)

type LoggedUserHandler struct {
	svc  services.LoggedUserService
	auth services.AuthService
}

func NewLoggedUserHandler(svc services.LoggedUserService, auth services.AuthService) *LoggedUserHandler {
	return &LoggedUserHandler{svc: svc, auth: auth}
}

func (h *LoggedUserHandler) Get(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok || user == nil {
		respondWithError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "Unauthorized", "")
		return
	}
	info, err := h.svc.GetInfo(r.Context(), user)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, info)
}

func (h *LoggedUserHandler) ChangeLanguage(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok || user == nil {
		respondWithError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "Unauthorized", "")
		return
	}
	var in models.ChangeLanguageRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	lang := in.Language
	if lang == "" {
		lang = in.Type
	}
	if err := h.svc.ChangeLanguage(r.Context(), user.ID, lang); err != nil {
		h.respondError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *LoggedUserHandler) AddDeviceToken(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok || user == nil {
		respondWithError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "Unauthorized", "")
		return
	}
	var in models.DeviceToken
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.AddDeviceToken(r.Context(), user.ID, in)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *LoggedUserHandler) DeleteDeviceToken(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok || user == nil {
		respondWithError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "Unauthorized", "")
		return
	}
	id, ok := parseInt64Path(mux.Vars(r), "id")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid token id", "")
		return
	}
	if err := h.svc.DeleteDeviceToken(r.Context(), user.ID, id); err != nil {
		h.respondError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *LoggedUserHandler) SetImage(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok || user == nil {
		respondWithError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "Unauthorized", "")
		return
	}
	var in map[string]any
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	url, _ := in["imageUrl"].(string)
	if url == "" {
		url, _ = in["url"].(string)
	}
	if err := h.svc.SetImage(r.Context(), user.ID, url); err != nil {
		h.respondError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *LoggedUserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
	if token == "" {
		respondWithError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "Bearer access token required", "")
		return
	}
	if err := h.auth.LogoutByAccessToken(r.Context(), token); err != nil {
		h.respondError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *LoggedUserHandler) respondError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "required") || strings.Contains(msg, "invalid") {
		respondWithError(w, http.StatusBadRequest, ErrCodeValidationFailed, err.Error(), "")
		return
	}
	if strings.Contains(msg, "not found") {
		respondWithError(w, http.StatusNotFound, ErrCodeNotFound, err.Error(), "")
		return
	}
	status, code, m := parseDatabaseError(err)
	respondWithError(w, status, code, m, err.Error())
}
