package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/services"
)

type AttributeHandler struct {
	svc services.AttributeService
}

func NewAttributeHandler(svc services.AttributeService) *AttributeHandler {
	return &AttributeHandler{svc: svc}
}

func (h *AttributeHandler) ListAttributes(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.ListAttributes(r.Context(), models.AttributeListQuery{
		ListQuery:      models.ListQuery{Limit: limit, Offset: offset},
		AttributeType:  r.URL.Query().Get("attributeType"),
		IDDoc:          r.URL.Query().Get("idDoc"),
		IDDocumentType: parseOptionalInt32Query(r, "idDocumentType"),
		IDLanguage:     r.URL.Query().Get("idLanguage"),
	})
	if err != nil {
		h.respondEntityError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *AttributeHandler) CreateAttributes(w http.ResponseWriter, r *http.Request) {
	var in models.ProductAttributesPayload
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	items, err := h.svc.CreateAttributeLinks(r.Context(), in)
	if err != nil {
		h.respondEntityError(w, err)
		return
	}
	writeJSONArray(w, http.StatusCreated, items)
}

func (h *AttributeHandler) ListProductCharacteristics(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListProductCharacteristics(r.Context())
	if err != nil {
		h.respondEntityError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *AttributeHandler) CreateProductCharacteristic(w http.ResponseWriter, r *http.Request) {
	var in models.ProductAttribute
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.CreateProductCharacteristic(r.Context(), in)
	if err != nil {
		h.respondEntityError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *AttributeHandler) UpdateProductCharacteristic(w http.ResponseWriter, r *http.Request) {
	var in models.ProductAttribute
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.UpdateProductCharacteristic(r.Context(), in)
	if err != nil {
		h.respondEntityError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *AttributeHandler) respondEntityError(w http.ResponseWriter, err error) {
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

func (h *AttributeHandler) respondDBError(w http.ResponseWriter, err error) {
	status, code, msg := parseDatabaseError(err)
	respondWithError(w, status, code, msg, err.Error())
}
