package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"superaib/internal/api/response"
	"superaib/internal/services"

	"github.com/gorilla/mux"
	"gorm.io/datatypes"
)

type APIKeyHandler struct {
	apiKeyService services.APIKeyService
}

func NewAPIKeyHandler(s services.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{apiKeyService: s}
}

// 1. CreateAPIKey handles POST /projects/{reference_id}/keys
func (h *APIKeyHandler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	projectRef := vars["reference_id"] // Waxaan isticmaalaynaa reference_id URL-ka

	var req struct {
		Name        string          `json:"name"`
		Permissions map[string]bool `json:"permissions"`
		CreatedBy   string          `json:"created_by"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if strings.TrimSpace(req.Name) == "" || req.CreatedBy == "" {
		response.Error(w, http.StatusBadRequest, "Name and CreatedBy are required")
		return
	}

	permsJSON, _ := json.Marshal(req.Permissions)

	apiKey, err := h.apiKeyService.CreateAPIKey(
		ctx,
		projectRef,
		req.CreatedBy,
		req.Name,
		datatypes.JSON(permsJSON),
	)

	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to create API key", err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, "API key generated successfully", apiKey)
}

// 2. HAGAAG: Kani waa kii maqnaa ee error-ka keenayay
// GetAPIKeyByID handles GET /projects/{reference_id}/keys/{id}
func (h *APIKeyHandler) GetAPIKeyByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	id := vars["id"]
	projectRef := vars["reference_id"]

	key, err := h.apiKeyService.GetAPIKeyByID(ctx, id, projectRef)
	if err != nil {
		response.Error(w, http.StatusNotFound, "API Key not found", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "API key retrieved successfully", key)
}

// 3. GetAllAPIKeysForProject handles GET /projects/{reference_id}/keys
func (h *APIKeyHandler) GetAllAPIKeysForProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	projectRef := vars["reference_id"]

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	keys, total, err := h.apiKeyService.GetAllAPIKeysForProject(ctx, projectRef, page, pageSize)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to retrieve API keys", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "API keys retrieved successfully", map[string]interface{}{
		"total":    total,
		"page":     page,
		"api_keys": keys,
	})
}

// 4. UpdateAPIKey handles PUT /projects/{reference_id}/keys/{id}
func (h *APIKeyHandler) UpdateAPIKey(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	id := vars["id"]
	projectRef := vars["reference_id"]

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	key, err := h.apiKeyService.UpdateAPIKey(ctx, id, projectRef, updates)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to update API key", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "API key updated successfully", key)
}

// 5. RevokeAPIKey handles POST /projects/{reference_id}/keys/{id}/revoke
func (h *APIKeyHandler) RevokeAPIKey(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	id := vars["id"]
	projectRef := vars["reference_id"]

	if err := h.apiKeyService.RevokeAPIKey(ctx, id, projectRef); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to revoke API key", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "API key revoked successfully", nil)
}

// 6. DeleteAPIKey handles DELETE /projects/{reference_id}/keys/{id}
func (h *APIKeyHandler) DeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	id := vars["id"]
	projectRef := vars["reference_id"]

	if err := h.apiKeyService.DeleteAPIKey(ctx, id, projectRef); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to delete API key", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "API key deleted successfully", nil)
}
