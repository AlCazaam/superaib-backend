package handlers

import (
	"encoding/json"
	"net/http"
	"superaib/internal/api/response"
	"superaib/internal/services"

	"github.com/gorilla/mux"
)

type SubscriptionHandler struct {
	service services.SubscriptionService
}

func NewSubscriptionHandler(s services.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{service: s}
}

func (h *SubscriptionHandler) Upgrade(w http.ResponseWriter, r *http.Request) {
	projectID := mux.Vars(r)["project_id"]
	developerID := r.Context().Value("userID").(string)

	var body struct {
		PlanID      string  `json:"plan_id"`
		WaafiID     string  `json:"waafi_id"`
		Amount      float64 `json:"amount"`
		PhoneNumber string  `json:"phone_number"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	req := services.UpgradeRequest{
		ProjectID:   projectID,
		DeveloperID: developerID,
		PlanID:      body.PlanID,
		WaafiID:     body.WaafiID,
		Amount:      body.Amount,
		PhoneNumber: body.PhoneNumber,
	}

	if err := h.service.UpgradeProject(r.Context(), req); err != nil {
		response.Error(w, http.StatusInternalServerError, "Upgrade failed", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "Project upgraded successfully", nil)
}
