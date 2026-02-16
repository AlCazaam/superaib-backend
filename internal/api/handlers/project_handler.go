// handlers/project_handler.go
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"superaib/internal/api/response"
	"superaib/internal/services"

	"github.com/gorilla/mux"
)

type ProjectHandler struct {
	projectService services.ProjectService
}

func NewProjectHandler(s services.ProjectService) *ProjectHandler {
	return &ProjectHandler{projectService: s}
}

// Function-kan wuxuu Context-ka ka soo saaraa UserID-ga uu Middleware-ku geliyay
func getOwnerIDFromContext(r *http.Request) (string, error) {
	if v := r.Context().Value("userID"); v != nil {
		if id, ok := v.(string); ok && id != "" {
			return id, nil
		}
	}
	return "", fmt.Errorf("unauthorized: user session not found")
}

// 1. CreateProject handles POST /projects
func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	ownerID, err := getOwnerIDFromContext(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Request-ka hadda waa mid aad u fudud (Name oo kaliya)
	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		response.Error(w, http.StatusBadRequest, "Project name is required")
		return
	}

	// U yeer service-ka fududeeyay
	project, err := h.projectService.CreateProject(r.Context(), ownerID, req.Name)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "already exists") {
			status = http.StatusConflict
		}
		response.Error(w, status, "Failed to create project", err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, "Project created successfully", map[string]interface{}{"project": project})
}

// 2. GetAllProjects handles GET /projects
func (h *ProjectHandler) GetAllProjects(w http.ResponseWriter, r *http.Request) {
	ownerID, err := getOwnerIDFromContext(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Pagination
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	projects, total, err := h.projectService.GetAllProjects(r.Context(), ownerID, page, pageSize)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to retrieve projects", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "Projects retrieved successfully", map[string]interface{}{
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
		"projects": projects,
	})
}

// 3. GetProjectByReferenceID handles GET /projects/{reference_id}
func (h *ProjectHandler) GetProjectByReferenceID(w http.ResponseWriter, r *http.Request) {
	referenceID := mux.Vars(r)["reference_id"]
	ownerID, err := getOwnerIDFromContext(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	project, err := h.projectService.GetProjectByReferenceID(r.Context(), referenceID, ownerID)
	if err != nil {
		status := http.StatusNotFound
		response.Error(w, status, "Project not found", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "Project retrieved successfully", project)
}

// 4. UpdateProject handles PUT /projects/{reference_id}
func (h *ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	referenceID := mux.Vars(r)["reference_id"]
	ownerID, err := getOwnerIDFromContext(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	project, err := h.projectService.UpdateProject(r.Context(), referenceID, ownerID, updates)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to update project", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "Project updated successfully", project)
}

// 5. DeleteProject handles DELETE /projects/{reference_id}
func (h *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idOrRef := vars["reference_id"] // Kani waa URL-ka (Slug ama UUID)

	ownerID, err := getOwnerIDFromContext(r)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	if err := h.projectService.DeleteProject(r.Context(), idOrRef, ownerID); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to wipe project data", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "Project and all associated data deleted successfully", nil)
}
