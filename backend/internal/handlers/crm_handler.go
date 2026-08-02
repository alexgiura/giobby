package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"dnsc_microservice/internal/models"
	"dnsc_microservice/internal/services"

	"github.com/gorilla/mux"
)

type CrmHandler struct {
	svc services.CrmService
}

func NewCrmHandler(svc services.CrmService) *CrmHandler {
	return &CrmHandler{svc: svc}
}

func (h *CrmHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var in models.CrmAccount
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.CreateAccount(r.Context(), in)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *CrmHandler) GlobalReport(w http.ResponseWriter, r *http.Request) {
	out, err := h.svc.GlobalReport(r.Context())
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *CrmHandler) AccountReport(w http.ResponseWriter, r *http.Request) {
	id, ok := crmPathInt32(mux.Vars(r), "idcrmaccount")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid crm account id", "")
		return
	}
	out, err := h.svc.AccountReport(r.Context(), id)
	if err != nil {
		h.respondError(w, err)
		return
	}
	if out == nil {
		respondWithError(w, http.StatusNotFound, ErrCodeNotFound, "CRM account not found", "")
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *CrmHandler) ListAllTasks(w http.ResponseWriter, r *http.Request) {
	h.listTasks(w, r, nil)
}

func (h *CrmHandler) ListAccountTasks(w http.ResponseWriter, r *http.Request) {
	id, ok := crmPathInt32(mux.Vars(r), "idcrmaccount")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid crm account id", "")
		return
	}
	h.listTasks(w, r, &id)
}

func (h *CrmHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	id, ok := crmPathInt32(mux.Vars(r), "idTask")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid task id", "")
		return
	}
	item, err := h.svc.GetTask(r.Context(), id)
	if err != nil {
		h.respondError(w, err)
		return
	}
	if item == nil {
		respondWithError(w, http.StatusNotFound, ErrCodeNotFound, "Task not found", "")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *CrmHandler) ListActivities(w http.ResponseWriter, r *http.Request) {
	idParent, ok := crmPathInt32(mux.Vars(r), "idParent")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid parent id", "")
		return
	}
	limit, offset := parseLimitOffset(r)
	items, err := h.svc.ListActivities(r.Context(), idParent, models.CrmActivityListQuery{
		ListQuery: models.ListQuery{Limit: limit, Offset: offset},
		Type:      r.URL.Query().Get("type"),
		IDDoc:     r.URL.Query().Get("idDoc"),
		IDDocType: parseOptionalInt32Query(r, "idDocType"),
	})
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func (h *CrmHandler) CreateActivity(w http.ResponseWriter, r *http.Request) {
	idParent, ok := crmPathInt32(mux.Vars(r), "idParent")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid parent id", "")
		return
	}
	var in models.CrmActivity
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.CreateActivity(r.Context(), idParent, in)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *CrmHandler) GetActivity(w http.ResponseWriter, r *http.Request) {
	idParent, id, ok := crmParentAndID(mux.Vars(r))
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid ids", "")
		return
	}
	item, err := h.svc.GetActivity(r.Context(), idParent, id)
	if err != nil {
		h.respondError(w, err)
		return
	}
	if item == nil {
		respondWithError(w, http.StatusNotFound, ErrCodeNotFound, "Activity not found", "")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *CrmHandler) UpdateActivity(w http.ResponseWriter, r *http.Request) {
	idParent, id, ok := crmParentAndID(mux.Vars(r))
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid ids", "")
		return
	}
	var in models.CrmActivity
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.UpdateActivity(r.Context(), idParent, id, in)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *CrmHandler) PatchActivity(w http.ResponseWriter, r *http.Request) {
	idParent, id, ok := crmParentAndID(mux.Vars(r))
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid ids", "")
		return
	}
	var fields map[string]any
	if err := json.NewDecoder(r.Body).Decode(&fields); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.PatchActivity(r.Context(), idParent, id, fields)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *CrmHandler) DeleteActivity(w http.ResponseWriter, r *http.Request) {
	idParent, id, ok := crmParentAndID(mux.Vars(r))
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid ids", "")
		return
	}
	if err := h.svc.DeleteActivity(r.Context(), idParent, id); err != nil {
		h.respondError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CrmHandler) UploadActivityAttachment(w http.ResponseWriter, r *http.Request) {
	idParent, ok := crmPathInt32(mux.Vars(r), "idParent")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid parent id", "")
		return
	}
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid multipart form", err.Error())
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		file, header, err = r.FormFile("attachment")
	}
	if err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Attachment file required", err.Error())
		return
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Failed to read attachment", err.Error())
		return
	}
	mime := header.Header.Get("Content-Type")
	if mime == "" {
		mime = "application/octet-stream"
	}
	item, err := h.svc.SaveAttachment(r.Context(), idParent, header.Filename, mime, data)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *CrmHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	idParent, ok := crmPathInt32(mux.Vars(r), "idParent")
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid parent id", "")
		return
	}
	var in models.CrmTask
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.CreateTask(r.Context(), idParent, in)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *CrmHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	idParent, id, ok := crmParentAndID(mux.Vars(r))
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid ids", "")
		return
	}
	var in models.CrmTask
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.UpdateTask(r.Context(), idParent, id, in)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *CrmHandler) PatchTask(w http.ResponseWriter, r *http.Request) {
	idParent, id, ok := crmParentAndID(mux.Vars(r))
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid ids", "")
		return
	}
	var fields map[string]any
	if err := json.NewDecoder(r.Body).Decode(&fields); err != nil {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}
	item, err := h.svc.PatchTask(r.Context(), idParent, id, fields)
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *CrmHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	idParent, id, ok := crmParentAndID(mux.Vars(r))
	if !ok {
		respondWithError(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid ids", "")
		return
	}
	if err := h.svc.DeleteTask(r.Context(), idParent, id); err != nil {
		h.respondError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CrmHandler) listTasks(w http.ResponseWriter, r *http.Request, idCrmAccount *int32) {
	limit, offset := parseLimitOffset(r)
	include := parseBoolQuery(r, "include_activities", false) || parseBoolQuery(r, "alsoActivities", false)
	items, err := h.svc.ListTasks(r.Context(), models.CrmTaskListQuery{
		ListQuery:         models.ListQuery{Limit: limit, Offset: offset},
		Types:             r.URL.Query().Get("types"),
		Status:            r.URL.Query().Get("status"),
		CreateDateStart:   r.URL.Query().Get("createDateStart"),
		CreateDateEnd:     r.URL.Query().Get("createDateEnd"),
		IncludeActivities: include,
		IDCrmAccount:      idCrmAccount,
	})
	if err != nil {
		h.respondError(w, err)
		return
	}
	writeJSONArray(w, http.StatusOK, items)
}

func crmPathInt32(vars map[string]string, key string) (int32, bool) {
	return parseInt32Path(vars, key)
}

func crmParentAndID(vars map[string]string) (idParent, id int32, ok bool) {
	idParent, ok = parseInt32Path(vars, "idParent")
	if !ok {
		return 0, 0, false
	}
	id, ok = parseInt32Path(vars, "id")
	return idParent, id, ok
}

func (h *CrmHandler) respondError(w http.ResponseWriter, err error) {
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
