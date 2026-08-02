package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"dnsc_microservice/internal/middleware"
	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/services"
)

type AuthHandler struct {
	auth services.AuthService
}

func NewAuthHandler(auth services.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

// Register POST /api/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var in models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	tokens, err := h.auth.Register(r.Context(), in.Username, in.Password, in.Email)
	if err != nil {
		h.respondAuthError(w, err, true)
		return
	}
	writeJSON(w, http.StatusCreated, tokens)
}

// Login POST /api/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var in models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	tokens, err := h.auth.Login(r.Context(), in.Username, in.Password)
	if err != nil {
		h.respondAuthError(w, err, false)
		return
	}
	writeJSON(w, http.StatusOK, tokens)
}

// Refresh POST /api/auth/refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var in models.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	tokens, err := h.auth.Refresh(r.Context(), in.RefreshToken)
	if err != nil {
		h.respondAuthError(w, err, false)
		return
	}
	writeJSON(w, http.StatusOK, tokens)
}

// Logout POST /api/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	token, ok := bearerToken(r)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "Bearer access token required", "")
		return
	}
	if err := h.auth.LogoutByAccessToken(r.Context(), token); err != nil {
		respondWithError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "Authentication failed", "")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Me GET /api/auth/me
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.UserFromContext(r.Context())
	if !ok || u == nil {
		respondWithError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "Unauthorized", "")
		return
	}
	writeJSON(w, http.StatusOK, models.MeResponse{User: h.auth.ToPublic(u)})
}

func bearerToken(r *http.Request) (string, bool) {
	h := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(h, "Bearer ") {
		return "", false
	}
	t := strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
	return t, t != ""
}

func (h *AuthHandler) respondAuthError(w http.ResponseWriter, err error, register bool) {
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "already exists"):
		respondWithError(w, http.StatusConflict, ErrCodeConflict, "Registration failed", "")
	case strings.Contains(msg, "required"), strings.Contains(msg, "at least"):
		respondWithError(w, http.StatusBadRequest, ErrCodeValidationFailed, err.Error(), "")
	case strings.Contains(msg, "invalid"), strings.Contains(msg, "expired"), strings.Contains(msg, "revoked"), strings.Contains(msg, "not found"):
		if register {
			respondWithError(w, http.StatusBadRequest, ErrCodeValidationFailed, err.Error(), "")
		} else {
			respondWithError(w, http.StatusUnauthorized, ErrCodeUnauthorized, "Authentication failed", "")
		}
	default:
		respondWithError(w, http.StatusInternalServerError, ErrCodeInternalError, "Request failed", err.Error())
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
