package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/services"

	"github.com/gorilla/mux"
)

type UserHandler struct {
	svc services.UserService
}

func NewUserHandler(svc services.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.ListUsers(r.Context(), models.CompanyUserListQuery{
		ListQuery:     models.ListQuery{Limit: limit, Offset: offset},
		IsAgent1:      r.URL.Query().Get("isAgent1"),
		IsAgent2:      r.URL.Query().Get("isAgent2"),
		RetrieveImage: parseBoolQuery(r, "retrieveImage", false),
	})
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var in models.CompanyUser
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.CreateUser(r.Context(), in)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, ok := userPathID(mux.Vars(r))
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid user id", "")
		return
	}
	item, err := h.svc.GetUser(r.Context(), id, parseBoolQuery(r, "retrieveImage", false))
	if err != nil {
		h.respondError(w, err)
		return
	}
	if item == nil {
		respondWithError(w, http.StatusNotFound, ErrCodeNotFound, "User not found", "")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, ok := userPathID(mux.Vars(r))
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid user id", "")
		return
	}
	var in models.CompanyUser
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.UpdateUser(r.Context(), id, in)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, ok := userPathID(mux.Vars(r))
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid user id", "")
		return
	}
	if err := h.svc.DeleteUser(r.Context(), id); err != nil {
		h.respondError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) ListAgentCommissions(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.ListAgentCommissions(r.Context(), models.AgentCommissionListQuery{
		ListQuery:       models.ListQuery{Limit: limit, Offset: offset},
		IDUser:          parseOptionalInt32Query(r, "idUser"),
		IDBu:            r.URL.Query().Get("idbu"),
		IDCustomer:      r.URL.Query().Get("idcustomer"),
		IDMaterialGroup: parseOptionalInt32Query(r, "idmaterialgroup"),
		IDMaterial:      r.URL.Query().Get("idmaterial"),
	})
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *UserHandler) AddAuthProfiles(w http.ResponseWriter, r *http.Request) {
	id, ok := parseInt32Path(mux.Vars(r), "idUser")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid user id", "")
		return
	}
	var profiles []string
	if err := json.NewDecoder(r.Body).Decode(&profiles); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	if err := h.svc.AddAuthProfiles(r.Context(), id, profiles); err != nil {
		h.respondError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func userPathID(vars map[string]string) (int32, bool) {
	if id, ok := parseInt32Path(vars, "id"); ok {
		return id, true
	}
	return parseInt32Path(vars, "idUser")
}

func (h *UserHandler) respondError(w http.ResponseWriter, err error) {
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
	if strings.Contains(msg, "unique") || strings.Contains(msg, "duplicate") {
		respondWithError(w, http.StatusConflict, ErrCodeConflict, "A resource with this identifier already exists.", "")
		return
	}
	status, code, m := parseDatabaseError(err)
	respondWithError(w, status, code, m, err.Error())
}
