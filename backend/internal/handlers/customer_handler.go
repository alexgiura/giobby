package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/services"

	"github.com/gorilla/mux"
)

type CustomerHandler struct {
	svc services.CustomerService
}

func NewCustomerHandler(svc services.CustomerService) *CustomerHandler {
	return &CustomerHandler{svc: svc}
}

func (h *CustomerHandler) ListCustomers(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.ListCustomers(r.Context(), models.CustomerListQuery{
		ListQuery: models.ListQuery{Limit: limit, Offset: offset},
		Name:      r.URL.Query().Get("name"),
		FreeText:  r.URL.Query().Get("freeText"),
		Deleted:   parseOptionalBoolQuery(r, "deleted"),
	})
	if err != nil {
		h.respondDBError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *CustomerHandler) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	var in models.Customer
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.CreateCustomer(r.Context(), in)
	if err != nil {
		h.respondValidationOrDB(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *CustomerHandler) GetCustomer(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["idcustomer"]
	item, err := h.svc.GetCustomer(r.Context(), id, parseBoolQuery(r, "onlyCustomerData", false))
	if err != nil {
		h.respondValidationOrDB(w, err)
		return
	}
	if item == nil {
		respondWithError(w, http.StatusNotFound, ErrCodeNotFound, "Customer not found", "")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *CustomerHandler) UpdateCustomer(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["idcustomer"]
	var in models.Customer
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.UpdateCustomer(r.Context(), id, in)
	if err != nil {
		h.respondValidationOrDB(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *CustomerHandler) PatchCustomer(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["idcustomer"]
	var patch map[string]any
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.PatchCustomer(r.Context(), id, patch)
	if err != nil {
		h.respondValidationOrDB(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *CustomerHandler) DeleteCustomer(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["idcustomer"]
	if err := h.svc.DeleteCustomer(r.Context(), id); err != nil {
		h.respondValidationOrDB(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CustomerHandler) GetCustomerReport(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["idcustomer"]
	item, err := h.svc.GetCustomerReport(r.Context(), id)
	if err != nil {
		h.respondValidationOrDB(w, err)
		return
	}
	if item == nil {
		respondWithError(w, http.StatusNotFound, ErrCodeNotFound, "Customer not found", "")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *CustomerHandler) respondValidationOrDB(w http.ResponseWriter, err error) {
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
	h.respondDBError(w, err)
}

func (h *CustomerHandler) respondDBError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	status, code, msg := parseDatabaseError(err)
	respondWithError(w, status, code, msg, err.Error())
}
