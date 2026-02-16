package handlers

import (
	"net/http"
	"superaib/internal/api/response"
	"superaib/internal/services"

	"github.com/gorilla/mux"
)

type ProjectUsageHandler struct {
	service        services.ProjectUsageService
	projectService services.ProjectService
}

func NewProjectUsageHandler(s services.ProjectUsageService, ps services.ProjectService) *ProjectUsageHandler {
	return &ProjectUsageHandler{
		service:        s,
		projectService: ps,
	}
}

func (h *ProjectUsageHandler) GetUsage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectParam := vars["project_id"]

	project, err := h.projectService.GetProjectByRefOrID(r.Context(), projectParam)
	if err != nil {
		response.Error(w, http.StatusNotFound, "Project not found", err.Error())
		return
	}

	usage, err := h.service.GetUsage(r.Context(), project.ID)
	if err != nil {
		// Create if missing
		h.service.CreateInitialUsageRecord(r.Context(), nil, project.ID)
		usage, _ = h.service.GetUsage(r.Context(), project.ID)
	}

	response.JSON(w, http.StatusOK, "Usage retrieved", usage)
}
