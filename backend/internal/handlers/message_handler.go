package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/services"

	"github.com/gorilla/mux"
)

type MessageHandler struct {
	svc services.MessageService
}

func NewMessageHandler(svc services.MessageService) *MessageHandler {
	return &MessageHandler{svc: svc}
}

func (h *MessageHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	q := models.MessageListQuery{
		ListQuery:  models.ListQuery{Limit: limit, Offset: offset},
		OnlyUnread: parseBoolQuery(r, "onlyUnread", false),
	}
	if v := r.URL.Query().Get("idGroup"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			q.IDGroup = &id
		}
	}
	items, err := h.svc.ListMessages(r.Context(), q)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *MessageHandler) ListUnread(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListUnread(r.Context())
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *MessageHandler) GetMessage(w http.ResponseWriter, r *http.Request) {
	id, ok := parseInt64Path(mux.Vars(r), "idmessage")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid message id", "")
		return
	}
	item, err := h.svc.GetMessage(r.Context(), id)
	if err != nil {
		h.respondError(w, err)
		return
	}
	if item == nil {
		respondWithError(w, http.StatusNotFound, ErrCodeNotFound, "Message not found", "")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *MessageHandler) CreateMessage(w http.ResponseWriter, r *http.Request) {
	var in models.Message
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.CreateMessage(r.Context(), in)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *MessageHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	id, ok := parseInt64Path(mux.Vars(r), "id")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid message id", "")
		return
	}
	item, err := h.svc.MarkMessageRead(r.Context(), id)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *MessageHandler) ListContacts(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.ListContacts(r.Context(), limit, offset, r.URL.Query().Get("name"))
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *MessageHandler) ListGroups(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.ListGroups(r.Context(), limit, offset)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *MessageHandler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	id, ok := parseInt64Path(mux.Vars(r), "id")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid group id", "")
		return
	}
	if err := h.svc.DeleteGroup(r.Context(), id); err != nil {
		h.respondError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *MessageHandler) MarkGroupRead(w http.ResponseWriter, r *http.Request) {
	id, ok := parseInt64Path(mux.Vars(r), "idGroup")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid group id", "")
		return
	}
	if err := h.svc.MarkGroupRead(r.Context(), id); err != nil {
		h.respondError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *MessageHandler) respondError(w http.ResponseWriter, err error) {
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
