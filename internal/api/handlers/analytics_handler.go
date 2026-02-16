package handlers

import (
	"net/http"
	"superaib/internal/api/response"
	"superaib/internal/services"

	"github.com/gorilla/mux"
)

type AnalyticsHandler struct {
	service services.AnalyticsService
}

func NewAnalyticsHandler(s services.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{service: s}
}

func (h *AnalyticsHandler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	projectID := mux.Vars(r)["project_id"]
	analyticsData, err := h.service.GetAnalyticsForCurrentMonth(r.Context(), projectID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to retrieve analytics", err.Error())
		return
	}
	response.JSON(w, http.StatusOK, "Analytics retrieved", analyticsData)
}
