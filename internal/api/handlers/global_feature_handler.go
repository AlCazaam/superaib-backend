package handlers

import (
	"encoding/json"
	"net/http"
	"superaib/internal/api/response"
	"superaib/internal/models"
	"superaib/internal/services"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gorm.io/datatypes"
)

type GlobalFeatureHandler struct {
	service services.GlobalFeatureService
}

func NewGlobalFeatureHandler(s services.GlobalFeatureService) *GlobalFeatureHandler {
	return &GlobalFeatureHandler{service: s}
}

func (h *GlobalFeatureHandler) GetAllFeatures(w http.ResponseWriter, r *http.Request) {
	features, err := h.service.GetAllGlobalFeatures(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to fetch features")
		return
	}
	response.JSON(w, http.StatusOK, "Global features retrieved", features)
}

func (h *GlobalFeatureHandler) CreateFeature(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Type    string          `json:"type"`
		Enabled bool            `json:"enabled"`
		Config  json.RawMessage `json:"config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid body")
		return
	}
	feature, err := h.service.CreateGlobalFeature(r.Context(), models.ProjectFeatureType(req.Type), req.Enabled, datatypes.JSON(req.Config))
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, "Created", feature)
}

func (h *GlobalFeatureHandler) UpdateFeature(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, _ := uuid.Parse(idStr)
	var req struct {
		Type    string          `json:"type"`
		Enabled bool            `json:"enabled"`
		Config  json.RawMessage `json:"config"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	feature, err := h.service.UpdateGlobalFeature(r.Context(), id, models.ProjectFeatureType(req.Type), req.Enabled, datatypes.JSON(req.Config))
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, "Updated", feature)
}

func (h *GlobalFeatureHandler) ToggleFeature(w http.ResponseWriter, r *http.Request) {
	fType := mux.Vars(r)["type"]
	var req struct {
		Enabled bool `json:"enabled"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	feature, err := h.service.ToggleGlobalFeature(r.Context(), models.ProjectFeatureType(fType), req.Enabled)
	if err != nil {
		response.Error(w, http.StatusNotFound, "Not found")
		return
	}
	response.JSON(w, http.StatusOK, "Toggled", feature)
}

func (h *GlobalFeatureHandler) DeleteFeature(w http.ResponseWriter, r *http.Request) {
	id, _ := uuid.Parse(mux.Vars(r)["id"])
	h.service.DeleteGlobalFeature(r.Context(), id)
	response.JSON(w, http.StatusOK, "Deleted", nil)
}
