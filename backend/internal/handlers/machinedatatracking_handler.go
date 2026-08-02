package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/services"

	"github.com/gorilla/mux"
)

type MachineDataTrackingHandler struct {
	svc services.MachineDataTrackingService
}

func NewMachineDataTrackingHandler(svc services.MachineDataTrackingService) *MachineDataTrackingHandler {
	return &MachineDataTrackingHandler{svc: svc}
}

func (h *MachineDataTrackingHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.List(r.Context(), models.MachineDataTrackingListQuery{
		ListQuery:       models.ListQuery{Limit: limit, Offset: offset},
		ID:              parseOptionalInt32Query(r, "id"),
		IDMachine:       r.URL.Query().Get("idmachine"),
		IDOperator:      r.URL.Query().Get("idoperator"),
		Materials:       r.URL.Query().Get("materials"),
		IDLot:           r.URL.Query().Get("idlot"),
		BoxQty:          parseOptionalFloatQuery(r, "boxqty"),
		BoxNumber:       parseOptionalFloatQuery(r, "boxnumber"),
		ProductionWaste: parseOptionalFloatQuery(r, "productionwaste"),
		Barcode:         r.URL.Query().Get("barcode"),
		StartTime:       r.URL.Query().Get("starttime"),
		EndTime:         r.URL.Query().Get("endtime"),
		Duration:        r.URL.Query().Get("duration"),
		MachineStatus:   r.URL.Query().Get("machinestatus"),
		AlarmCode:       r.URL.Query().Get("alarmcode"),
		AlarmTime:       r.URL.Query().Get("alarmtime"),
		Date:            r.URL.Query().Get("date"),
		IDAttachment:    r.URL.Query().Get("idattachment"),
		Deleted:         parseOptionalBoolQuery(r, "deleted"),
	})
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *MachineDataTrackingHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseInt32Path(mux.Vars(r), "id")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid id", "")
		return
	}
	item, err := h.svc.Get(r.Context(), id)
	if err != nil {
		h.respondError(w, err)
		return
	}
	if item == nil {
		respondWithError(w, http.StatusNotFound, ErrCodeNotFound, "Record not found", "")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *MachineDataTrackingHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in models.MachineDataTracking
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.Create(r.Context(), in)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *MachineDataTrackingHandler) CreateBatch(w http.ResponseWriter, r *http.Request) {
	var in []models.MachineDataTracking
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	items, err := h.svc.CreateBatch(r.Context(), in)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSONArray(w, http.StatusCreated, items)
}

func (h *MachineDataTrackingHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseInt32Path(mux.Vars(r), "id")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid id", "")
		return
	}
	var in models.MachineDataTracking
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.Update(r.Context(), id, in)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *MachineDataTrackingHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseInt32Path(mux.Vars(r), "id")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid id", "")
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		h.respondError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *MachineDataTrackingHandler) respondError(w http.ResponseWriter, err error) {
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
