package handlers

import (
	"encoding/json"
	"net/http"
	"superaib/internal/api/response"
	"superaib/internal/models"
	"superaib/internal/services"

	"github.com/gorilla/mux"
)

type PlanHandler struct {
	service services.PlanService
}

func NewPlanHandler(s services.PlanService) *PlanHandler {
	return &PlanHandler{service: s}
}

func (h *PlanHandler) CreatePlan(w http.ResponseWriter, r *http.Request) {
	var plan models.Plan
	json.NewDecoder(r.Body).Decode(&plan)
	if err := h.service.CreatePlan(r.Context(), &plan); err != nil {
		response.Error(w, 500, "Create failed", err.Error())
		return
	}
	response.JSON(w, 201, "Plan Created", plan)
}

func (h *PlanHandler) GetAllPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.service.GetAllPlans(r.Context())
	if err != nil {
		response.Error(w, 500, "Fetch failed", err.Error())
		return
	}
	response.JSON(w, 200, "Success", plans)
}

func (h *PlanHandler) UpdatePlan(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var plan models.Plan
	json.NewDecoder(r.Body).Decode(&plan)
	updated, err := h.service.UpdatePlan(r.Context(), id, &plan)
	if err != nil {
		response.Error(w, 500, "Update failed", err.Error())
		return
	}
	response.JSON(w, 200, "Updated", updated)
}

func (h *PlanHandler) DeletePlan(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	h.service.DeletePlan(r.Context(), id)
	response.JSON(w, 200, "Deleted", nil)
}
