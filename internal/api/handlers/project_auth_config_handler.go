package handlers

import (
	"encoding/json"
	"net/http"
	"superaib/internal/api/response"
	"superaib/internal/services"

	"github.com/gorilla/mux"
)

type ProjectAuthConfigHandler struct {
	service services.ProjectAuthConfigService
}

func NewProjectAuthConfigHandler(s services.ProjectAuthConfigService) *ProjectAuthConfigHandler {
	return &ProjectAuthConfigHandler{service: s}
}

// POST /api/v1/projects/{project_id}/auth-configs
func (h *ProjectAuthConfigHandler) Configure(w http.ResponseWriter, r *http.Request) {
	projectID := mux.Vars(r)["project_id"]
	var req struct {
		ProviderID  string                 `json:"provider_id"`
		Credentials map[string]interface{} `json:"credentials"`
		Enabled     bool                   `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	res, err := h.service.ConfigureProvider(r.Context(), projectID, req.ProviderID, req.Credentials, req.Enabled)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, "Configuration saved", res)
}

// GET /api/v1/projects/{project_id}/auth-configs
func (h *ProjectAuthConfigHandler) GetConfigs(w http.ResponseWriter, r *http.Request) {
	projectID := mux.Vars(r)["project_id"]
	res, err := h.service.GetConfigsByProject(r.Context(), projectID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, "Configs retrieved", res)
}
