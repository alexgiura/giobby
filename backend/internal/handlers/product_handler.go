package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/services"

	"github.com/gorilla/mux"
)

type ProductHandler struct {
	svc services.ProductService
}

func NewProductHandler(svc services.ProductService) *ProductHandler {
	return &ProductHandler{svc: svc}
}

func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.ListProducts(r.Context(), models.ProductListQuery{
		ListQuery:        models.ListQuery{Limit: limit, Offset: offset},
		Description:      r.URL.Query().Get("description"),
		SalesEnabled:     parseOptionalBoolQuery(r, "salesEnabled"),
		Locked:           parseOptionalBoolQuery(r, "locked"),
		StockEnabled:     parseOptionalBoolQuery(r, "stockEnabled"),
		PricelistEnabled: parseOptionalBoolQuery(r, "pricelistEnabled"),
		IDPlugin:         r.URL.Query().Get("idPlugin"),
		CreateDateMillis: r.URL.Query().Get("createDateMillis"),
		ModifyDateMillis: r.URL.Query().Get("modifyDateMillis"),
	})
	if err != nil {
		h.respondEntityError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	item, err := h.svc.GetProduct(r.Context(), id)
	if err != nil {
		h.respondEntityError(w, err)
		return
	}
	if item == nil {
		respondWithError(w, http.StatusNotFound, ErrCodeNotFound, "Product not found", "")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var in models.Product
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.CreateProduct(r.Context(), in)
	if err != nil {
		h.respondEntityError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	var in models.Product
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.UpdateProduct(r.Context(), in)
	if err != nil {
		h.respondEntityError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *ProductHandler) ListVariants(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	items, err := h.svc.ListVariants(r.Context(), id)
	if err != nil {
		h.respondEntityError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *ProductHandler) UpdateVariants(w http.ResponseWriter, r *http.Request) {
	var in models.ProductVariantsPayload
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	if err := h.svc.UpdateVariants(r.Context(), in); err != nil {
		h.respondEntityError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *ProductHandler) LinkMedia(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var in models.RepomediaRef
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	if err := h.svc.LinkMedia(r.Context(), id, in); err != nil {
		h.respondEntityError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *ProductHandler) respondEntityError(w http.ResponseWriter, err error) {
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

func (h *ProductHandler) respondDBError(w http.ResponseWriter, err error) {
	status, code, msg := parseDatabaseError(err)
	respondWithError(w, status, code, msg, err.Error())
}
