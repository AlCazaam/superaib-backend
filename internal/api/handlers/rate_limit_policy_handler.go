package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"superaib/internal/api/response"
	"superaib/internal/models"
	"superaib/internal/services"

	"gorm.io/gorm"
)

type RateLimitPolicyHandler struct {
	service services.RateLimitPolicyService
}

func NewRateLimitPolicyHandler(s services.RateLimitPolicyService) *RateLimitPolicyHandler {
	return &RateLimitPolicyHandler{service: s}
}

// Helper: Ka soo saar Project ID
func (h *RateLimitPolicyHandler) getProjectID(r *http.Request) string {
	// Waxaan u malaynaynaa in API Key middleware uu hore u sameeyay tan
	if ctxID, ok := r.Context().Value("projectID").(string); ok {
		return ctxID
	}
	return ""
}

// 1. CreatePolicy handles POST /policy
func (h *RateLimitPolicyHandler) CreatePolicy(w http.ResponseWriter, r *http.Request) {
	projectID := h.getProjectID(r)
	var policy models.RateLimitPolicy

	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	policy.ProjectID = projectID // Xaqiiji in Policy-ga uu leeyahay Project ID-ga

	newPolicy, err := h.service.CreatePolicy(r.Context(), &policy)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, errors.New("policy already exists for this project")) {
			status = http.StatusConflict
		}
		response.Error(w, status, "Failed to create policy", err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, "Rate limit policy created", newPolicy)
}

// 2. UpdatePolicy handles PUT /policy
func (h *RateLimitPolicyHandler) UpdatePolicy(w http.ResponseWriter, r *http.Request) {
	projectID := h.getProjectID(r)
	var policy models.RateLimitPolicy

	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	updatedPolicy, err := h.service.UpdatePolicy(r.Context(), projectID, &policy)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusNotFound
		}
		response.Error(w, status, "Failed to update policy", err.Error())
		return
	}
	response.JSON(w, http.StatusOK, "Rate limit policy updated", updatedPolicy)
}

// 3. GetPolicy handles GET /policy
func (h *RateLimitPolicyHandler) GetPolicy(w http.ResponseWriter, r *http.Request) {
	projectID := h.getProjectID(r)

	policy, err := h.service.GetPolicy(r.Context(), projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(w, http.StatusNotFound, "Policy not configured for this project")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Failed to retrieve policy", err.Error())
		return
	}
	response.JSON(w, http.StatusOK, "Rate limit policy retrieved", policy)
}

// 4. DeletePolicy handles DELETE /policy
func (h *RateLimitPolicyHandler) DeletePolicy(w http.ResponseWriter, r *http.Request) {
	projectID := h.getProjectID(r)

	if err := h.service.DeletePolicy(r.Context(), projectID); err != nil {
		response.Error(w, http.StatusNotFound, "Policy not found or delete failed", err.Error())
		return
	}
	response.JSON(w, http.StatusOK, "Rate limit policy deleted successfully", nil)
}

// 5. TogglePolicy handles PUT /policy/toggle
func (h *RateLimitPolicyHandler) TogglePolicy(w http.ResponseWriter, r *http.Request) {
	projectID := h.getProjectID(r)
	var req struct {
		IsEnabled bool `json:"is_enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	updatedPolicy, err := h.service.TogglePolicy(r.Context(), projectID, req.IsEnabled)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to toggle policy", err.Error())
		return
	}
	response.JSON(w, http.StatusOK, "Rate limit policy toggled successfully", updatedPolicy)
}
