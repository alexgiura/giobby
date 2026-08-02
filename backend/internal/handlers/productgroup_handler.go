package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/services"

	"github.com/gorilla/mux"
)

type ProductGroupHandler struct {
	svc services.ProductGroupService
}

func NewProductGroupHandler(svc services.ProductGroupService) *ProductGroupHandler {
	return &ProductGroupHandler{svc: svc}
}

func (h *ProductGroupHandler) ListProductGroups(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.ListProductGroups(r.Context(), models.ProductGroupListQuery{
		ListQuery:   models.ListQuery{Limit: limit, Offset: offset},
		Description: r.URL.Query().Get("description"),
	})
	if err != nil {
		h.respondEntityError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *ProductGroupHandler) GetProductGroup(w http.ResponseWriter, r *http.Request) {
	id, ok := parseInt32Path(mux.Vars(r), "id")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid product group id", "")
		return
	}
	item, err := h.svc.GetProductGroup(r.Context(), id)
	if err != nil {
		h.respondEntityError(w, err)
		return
	}
	if item == nil {
		respondWithError(w, http.StatusNotFound, ErrCodeNotFound, "Product group not found", "")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *ProductGroupHandler) CreateProductGroup(w http.ResponseWriter, r *http.Request) {
	var in models.ProductGroup
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.CreateProductGroup(r.Context(), in)
	if err != nil {
		h.respondEntityError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *ProductGroupHandler) UpdateProductGroup(w http.ResponseWriter, r *http.Request) {
	id, ok := parseInt32Path(mux.Vars(r), "id")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid product group id", "")
		return
	}
	var in models.ProductGroup
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.UpdateProductGroup(r.Context(), id, in)
	if err != nil {
		h.respondEntityError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *ProductGroupHandler) DeleteProductGroup(w http.ResponseWriter, r *http.Request) {
	id, ok := parseInt32Path(mux.Vars(r), "id")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid product group id", "")
		return
	}
	if err := h.svc.DeleteProductGroup(r.Context(), id); err != nil {
		h.respondEntityError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProductGroupHandler) respondEntityError(w http.ResponseWriter, err error) {
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

func (h *ProductGroupHandler) respondDBError(w http.ResponseWriter, err error) {
	status, code, msg := parseDatabaseError(err)
	respondWithError(w, status, code, msg, err.Error())
}
