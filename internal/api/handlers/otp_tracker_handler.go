package handlers

import (
	"net/http"
	"superaib/internal/api/response"
	"superaib/internal/services"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type OtpTrackerHandler struct {
	service services.OtpTrackerService
}

func NewOtpTrackerHandler(s services.OtpTrackerService) *OtpTrackerHandler {
	return &OtpTrackerHandler{service: s}
}

// ResetLimit handles POST /projects/.../auth-users/{id}/reset-limit
func (h *OtpTrackerHandler) ResetLimit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["id"]
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid User ID")
		return
	}

	if err := h.service.ResetLimit(r.Context(), userID); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to reset limit", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "User rate limit reset successfully", nil)
}
