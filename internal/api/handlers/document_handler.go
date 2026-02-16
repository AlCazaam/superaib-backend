package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"superaib/internal/api/response"
	"superaib/internal/services"
	"superaib/internal/storage/repo"

	"github.com/gorilla/mux"
)

type DocumentHandler struct {
	service services.DocumentService
}

func NewDocumentHandler(s services.DocumentService) *DocumentHandler {
	return &DocumentHandler{service: s}
}

// 🟢 SMART HELPER: Mishiinka kala saaraya SDK (API Key) iyo Dashboard (URL Param)
func (h *DocumentHandler) getPID(r *http.Request) string {
	if uid, ok := r.Context().Value("projectID").(string); ok && uid != "" {
		return uid
	}
	vars := mux.Vars(r)
	pID := vars["project_id"]
	if pID == "" {
		pID = vars["reference_id"]
	}
	return pID
}

// --- 1. DOCUMENT CORE OPERATIONS ---

// Create: POST /db/{collection}
func (h *DocumentHandler) Create(w http.ResponseWriter, r *http.Request) {
	pID := h.getPID(r)
	vars := mux.Vars(r)
	collectionName := vars["collection"]

	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON body", nil)
		return
	}

	doc, err := h.service.Create(r.Context(), pID, collectionName, data)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Create failed", err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, "Created", doc)
}

// GetByID: GET /db/{collection}/{id}
func (h *DocumentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	pID := h.getPID(r)
	vars := mux.Vars(r)

	doc, err := h.service.Get(r.Context(), pID, vars["collection"], vars["id"])
	if err != nil {
		response.Error(w, http.StatusNotFound, "Document not found", err.Error())
		return
	}
	response.JSON(w, http.StatusOK, "Success", doc)
}

// Set: PUT /db/{collection}/{id}?merge=true
func (h *DocumentHandler) Set(w http.ResponseWriter, r *http.Request) {
	pID := h.getPID(r)
	vars := mux.Vars(r)
	merge, _ := strconv.ParseBool(r.URL.Query().Get("merge"))

	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON", nil)
		return
	}

	doc, err := h.service.Set(r.Context(), pID, vars["collection"], vars["id"], data, merge)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Set operation failed", err.Error())
		return
	}
	response.JSON(w, http.StatusOK, "Document Set Successfully", doc)
}

// Update: PATCH /db/{collection}/{id}
func (h *DocumentHandler) UpdateDocument(w http.ResponseWriter, r *http.Request) {
	pID := h.getPID(r)
	vars := mux.Vars(r)
	etag := r.Header.Get("If-Match")

	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON", nil)
		return
	}

	doc, err := h.service.Update(r.Context(), pID, vars["collection"], vars["id"], data, etag)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Update failed", err.Error())
		return
	}
	response.JSON(w, http.StatusOK, "Updated", doc)
}

// Upsert: POST /db/{collection}/{id}/upsert
func (h *DocumentHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	pID := h.getPID(r)
	vars := mux.Vars(r)

	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON", nil)
		return
	}

	doc, err := h.service.Upsert(r.Context(), pID, vars["collection"], vars["id"], data)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Upsert failed", err.Error())
		return
	}
	response.JSON(w, http.StatusOK, "Upsert Successful", doc)
}

// Delete: DELETE /db/{collection}/{id}
func (h *DocumentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	pID := h.getPID(r)
	vars := mux.Vars(r)

	err := h.service.Delete(r.Context(), pID, vars["collection"], vars["id"])
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Delete failed", err.Error())
		return
	}
	response.JSON(w, http.StatusOK, "Deleted", nil)
}

// Exists: GET /db/{collection}/{id}/exists
func (h *DocumentHandler) Exists(w http.ResponseWriter, r *http.Request) {
	pID := h.getPID(r)
	vars := mux.Vars(r)
	exists, err := h.service.Exists(r.Context(), pID, vars["collection"], vars["id"])
	if err != nil {
		response.Error(w, 500, "Error checking existence", err.Error())
		return
	}
	response.JSON(w, 200, "Success", map[string]bool{"exists": exists})
}

// Increment: POST /db/{collection}/{id}/increment
func (h *DocumentHandler) Increment(w http.ResponseWriter, r *http.Request) {
	pID := h.getPID(r)
	vars := mux.Vars(r)

	var req struct {
		Field  string  `json:"field"`
		Amount float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, 400, "Invalid JSON", nil)
		return
	}

	err := h.service.Increment(r.Context(), pID, vars["collection"], vars["id"], req.Field, req.Amount)
	if err != nil {
		response.Error(w, 500, "Increment failed", err.Error())
		return
	}
	response.JSON(w, 200, "Incremented", nil)
}

// Count: POST /db/{collection}/count
func (h *DocumentHandler) Count(w http.ResponseWriter, r *http.Request) {
	pID := h.getPID(r)
	vars := mux.Vars(r)

	var filters []repo.Filter
	_ = json.NewDecoder(r.Body).Decode(&filters) // Optional filters

	count, err := h.service.Count(r.Context(), pID, vars["collection"], filters)
	if err != nil {
		response.Error(w, 500, "Count failed", err.Error())
		return
	}
	response.JSON(w, 200, "Success", map[string]int64{"count": count})
}

// --- 2. ADVANCED QUERY & SEARCH ---

// AdvancedSearch: POST /db/{collection}/query
// internal/api/handlers/document_handler.go

func (h *DocumentHandler) AdvancedSearch(w http.ResponseWriter, r *http.Request) {
	pID := h.getPID(r)
	vars := mux.Vars(r)

	var req services.AdvancedQueryRequest

	// ✅ FIX: Haddii body-ga uu maran yahay, ha soo saarin error, isticmaal defaults
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.Error(w, 400, "Invalid Query Body", err.Error())
			return
		}
	}

	// Defaults haddii aan la soo dirin
	if req.Limit == 0 {
		req.Limit = 100
	}
	if req.OrderBy == "" {
		req.OrderBy = "created_at DESC"
	}

	docs, err := h.service.AdvancedSearch(r.Context(), pID, vars["collection"], req)
	if err != nil {
		response.Error(w, 500, "Query failed", err.Error())
		return
	}
	response.JSON(w, 200, "Success", docs)
}

// --- 3. COLLECTION MANAGEMENT ---

// CreateCollection: POST /collections
func (h *DocumentHandler) CreateCollection(w http.ResponseWriter, r *http.Request) {
	pID := h.getPID(r)
	var body struct {
		Name string `json:"name"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	coll, err := h.service.CreateColl(r.Context(), pID, body.Name)
	if err != nil {
		response.Error(w, 500, "Failed to create collection", err.Error())
		return
	}
	response.JSON(w, 201, "Collection Created", coll)
}

// ListCollections: GET /collections
func (h *DocumentHandler) GetCollections(w http.ResponseWriter, r *http.Request) {
	pID := h.getPID(r)
	colls, err := h.service.ListCollections(r.Context(), pID)
	if err != nil {
		response.Error(w, 500, "Failed to list collections", err.Error())
		return
	}
	response.JSON(w, 200, "Success", colls)
}

// RenameCollection: PATCH /collections/{collection}
func (h *DocumentHandler) RenameCollection(w http.ResponseWriter, r *http.Request) {
	pID := h.getPID(r)
	vars := mux.Vars(r)
	var body struct {
		NewName string `json:"new_name"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	err := h.service.RenameColl(r.Context(), pID, vars["collection"], body.NewName)
	if err != nil {
		response.Error(w, 500, "Rename failed", err.Error())
		return
	}
	response.JSON(w, 200, "Renamed", nil)
}

// DeleteCollection: DELETE /collections/{collection}
func (h *DocumentHandler) DeleteCollection(w http.ResponseWriter, r *http.Request) {
	pID := h.getPID(r)
	vars := mux.Vars(r)

	err := h.service.DeleteColl(r.Context(), pID, vars["collection"])
	if err != nil {
		response.Error(w, 500, "Delete failed", err.Error())
		return
	}
	response.JSON(w, 200, "Collection Deleted", nil)
}
