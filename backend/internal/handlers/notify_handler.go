package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/services"

	"github.com/gorilla/mux"
)

type NotifyHandler struct {
	svc services.NotifyService
}

func NewNotifyHandler(svc services.NotifyService) *NotifyHandler {
	return &NotifyHandler{svc: svc}
}

func (h *NotifyHandler) listChannel(w http.ResponseWriter, r *http.Request, channel string) {
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.ListNotifies(r.Context(), channel, models.NotifyListQuery{
		ListQuery:       models.ListQuery{Limit: limit, Offset: offset},
		OnlyUnread:      parseBoolQuery(r, "onlyUnread", false),
		DescendingOrder: parseBoolQuery(r, "descendingOrder", false),
	})
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *NotifyHandler) unreadChannel(w http.ResponseWriter, r *http.Request, channel string) {
	items, err := h.svc.ListUnreadNotifies(r.Context(), channel)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *NotifyHandler) getChannel(w http.ResponseWriter, r *http.Request, channel string) {
	id, ok := parseInt64Path(mux.Vars(r), "id")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid notify id", "")
		return
	}
	item, err := h.svc.GetNotify(r.Context(), channel, id)
	if err != nil {
		h.respondError(w, err)
		return
	}
	if item == nil {
		respondWithError(w, http.StatusNotFound, ErrCodeNotFound, "Notify not found", "")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *NotifyHandler) markChannel(w http.ResponseWriter, r *http.Request, channel string) {
	id, ok := parseInt64Path(mux.Vars(r), "id")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid notify id", "")
		return
	}
	item, err := h.svc.MarkNotifyRead(r.Context(), channel, id)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *NotifyHandler) ListEcommerce(w http.ResponseWriter, r *http.Request) {
	h.listChannel(w, r, "ecommerce")
}
func (h *NotifyHandler) UnreadEcommerce(w http.ResponseWriter, r *http.Request) {
	h.unreadChannel(w, r, "ecommerce")
}
func (h *NotifyHandler) GetEcommerce(w http.ResponseWriter, r *http.Request) {
	h.getChannel(w, r, "ecommerce")
}
func (h *NotifyHandler) MarkEcommerce(w http.ResponseWriter, r *http.Request) {
	h.markChannel(w, r, "ecommerce")
}

func (h *NotifyHandler) ListSocialNotifies(w http.ResponseWriter, r *http.Request) {
	h.listChannel(w, r, "social")
}
func (h *NotifyHandler) UnreadSocial(w http.ResponseWriter, r *http.Request) {
	h.unreadChannel(w, r, "social")
}
func (h *NotifyHandler) GetSocialNotify(w http.ResponseWriter, r *http.Request) {
	h.getChannel(w, r, "social")
}

func (h *NotifyHandler) ListTaskNotifies(w http.ResponseWriter, r *http.Request) {
	h.listChannel(w, r, "task")
}
func (h *NotifyHandler) UnreadTask(w http.ResponseWriter, r *http.Request) {
	h.unreadChannel(w, r, "task")
}
func (h *NotifyHandler) GetTaskNotify(w http.ResponseWriter, r *http.Request) {
	h.getChannel(w, r, "task")
}

func (h *NotifyHandler) ListSocialPosts(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.ListSocialPosts(r.Context(), limit, offset, r.URL.Query().Get("idLang"))
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *NotifyHandler) ListEmails(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.ListEmails(r.Context(), limit, offset, r.URL.Query().Get("hashcode"))
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *NotifyHandler) ListGlobalNotifications(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListGlobalNotifications(r.Context())
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *NotifyHandler) BindCommerceUpdateStock(w http.ResponseWriter, r *http.Request) {
	var in models.BindCommerceRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	ack, err := h.svc.BindCommerceUpdateStock(r.Context(), in)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ack)
}

func (h *NotifyHandler) TilbySales(w http.ResponseWriter, r *http.Request) {
	var payload any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	ack, err := h.svc.TilbySales(r.Context(), payload)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ack)
}

func (h *NotifyHandler) respondError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "not found") {
		respondWithError(w, http.StatusNotFound, ErrCodeNotFound, err.Error(), "")
		return
	}
	status, code, m := parseDatabaseError(err)
	respondWithError(w, status, code, m, err.Error())
}
