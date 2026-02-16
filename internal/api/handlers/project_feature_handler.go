package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"superaib/internal/api/response"
	"superaib/internal/models"
	"superaib/internal/services"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gorm.io/datatypes"
)

// ProjectFeatureHandler handles HTTP requests for project features
type ProjectFeatureHandler struct {
	service services.ProjectFeatureService
}

// NewProjectFeatureHandler creates a new ProjectFeatureHandler
func NewProjectFeatureHandler(s services.ProjectFeatureService) *ProjectFeatureHandler {
	return &ProjectFeatureHandler{service: s}
}

// 1. CreateFeature handles POST /projects/{project_id}/features
// POST /projects/{project_id}/features
func (h *ProjectFeatureHandler) CreateFeature(w http.ResponseWriter, r *http.Request) {
	projectID := mux.Vars(r)["project_id"]

	var req struct {
		Type    string          `json:"type"`
		Enabled bool            `json:"enabled"`
		Config  json.RawMessage `json:"config"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	feature := &models.ProjectFeature{
		ProjectID: projectID,
		Type:      models.ProjectFeatureType(req.Type),
		Enabled:   req.Enabled,
		Config:    datatypes.JSON(req.Config),
	}

	if err := h.service.CreateFeature(r.Context(), feature); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, "Created", feature)
}

// PUT /projects/{project_id}/features/{type}/toggle
// PUT /projects/{project_id}/features/{feature_type}/toggle

// 1. ToggleFeature handles PUT /projects/{project_id}/features/{type}/toggle
// Kani waa mishiinka xallinaya inuu Shido ama Damiye adeegyada (Auth, Storage, etc.)
func (h *ProjectFeatureHandler) ToggleFeature(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["project_id"]

	// ✅ FIX 1: Force Lowercase si database-ka uu u garto (e.g. "Storage" -> "storage")
	// Waxaan iska hubinaynaa "type" ama "feature_type" hadba midka router-ka ku jira
	rawType := vars["type"]
	if rawType == "" {
		rawType = vars["feature_type"]
	}
	fType := models.ProjectFeatureType(strings.ToLower(rawType))

	// ✅ FIX 2: Query parameter-ka "enable" (URL?enable=true)
	enable := r.URL.Query().Get("enable") == "true"

	// ✅ FIX 3: Decode JSON Config (Haddii uu Developer-ku soo diray Keys sida Storage)
	var body struct {
		Config datatypes.JSON `json:"config"`
	}

	// Ha u ogolaan inuu istaago haddii body-gu maran yahay (default {} ayaa galaya)
	_ = json.NewDecoder(r.Body).Decode(&body)

	// U dir Service-ka "Smart" ah
	feature, err := h.service.ToggleFeature(
		r.Context(),
		projectID,
		fType,
		enable,
		body.Config,
	)

	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to toggle feature", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "Feature updated successfully", feature)
}

// 2. GetAllFeatures handles GET /projects/{project_id}/features
func (h *ProjectFeatureHandler) GetAllFeatures(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["project_id"]

	if _, err := uuid.Parse(projectID); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid project ID format")
		return
	}

	features, err := h.service.GetAllFeaturesByProject(r.Context(), projectID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to retrieve project features", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "Project features retrieved successfully", features)
}

// 3. HAGAAG: Kani waa kii maqnaa ee error-ka keenayay
// GetFeatureByID handles GET /features/{id}
func (h *ProjectFeatureHandler) GetFeatureByID(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid feature ID format")
		return
	}

	feature, err := h.service.GetFeatureByID(r.Context(), id)
	if err != nil {
		status := http.StatusNotFound
		response.Error(w, status, "Project feature not found", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "Project feature retrieved successfully", feature)
}

// 4. ToggleFeature handles PUT /projects/{project_id}/features/{feature_type}/toggle

// 5. UpdateFeature handles PUT /features/{id}
func (h *ProjectFeatureHandler) UpdateFeature(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid feature ID format")
		return
	}

	var feature models.ProjectFeature
	if err := json.NewDecoder(r.Body).Decode(&feature); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	feature.ID = id
	if err := h.service.UpdateFeature(r.Context(), &feature); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to update feature", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "Project feature updated successfully", feature)
}

// 6. DeleteFeature handles DELETE /features/{id}
func (h *ProjectFeatureHandler) DeleteFeature(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid feature ID format")
		return
	}

	if err := h.service.DeleteFeature(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to delete feature", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "Project feature deleted successfully", nil)
}
