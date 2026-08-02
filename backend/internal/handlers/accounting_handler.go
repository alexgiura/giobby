package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/services"

	"github.com/gorilla/mux"
)

type AccountingHandler struct {
	svc services.AccountingService
}

func NewAccountingHandler(svc services.AccountingService) *AccountingHandler {
	return &AccountingHandler{svc: svc}
}

func (h *AccountingHandler) ListAccountMovements(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.ListAccountMovements(r.Context(), models.AccountMovementListQuery{
		ListQuery:      models.ListQuery{Limit: limit, Offset: offset},
		EntryType:      r.URL.Query().Get("entryType"),
		IDBu:           r.URL.Query().Get("idBu"),
		IDBup:          r.URL.Query().Get("idBup"),
		IDDoc:          parseOptionalInt32Query(r, "idDoc"),
		AccountCenter:  parseOptionalInt32Query(r, "accountCenter"),
		FreeText:       r.URL.Query().Get("freeText"),
		ShowZero:       parseBoolQuery(r, "showZero", false),
		FromDateMillis: r.URL.Query().Get("fromDateMillis"),
		ToDateMillis:   r.URL.Query().Get("toDateMillis"),
	})
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *AccountingHandler) CreateAccountMovement(w http.ResponseWriter, r *http.Request) {
	var in models.AccountMovementRegistration
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.CreateAccountMovement(r.Context(), in)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *AccountingHandler) UpdateAccountMovement(w http.ResponseWriter, r *http.Request) {
	var in models.AccountMovementRegistration
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.UpdateAccountMovement(r.Context(), in)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *AccountingHandler) DeleteAccountMovement(w http.ResponseWriter, r *http.Request) {
	idDoc, ok := parseInt32Path(mux.Vars(r), "idDoc")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid document id", "")
		return
	}
	if err := h.svc.DeleteAccountMovement(r.Context(), idDoc); err != nil {
		h.respondError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AccountingHandler) respondError(w http.ResponseWriter, err error) {
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
