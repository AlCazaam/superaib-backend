package handlers

import (
	"encoding/json"
	"net/http"
	"superaib/internal/api/response"
	"superaib/internal/models"
	"superaib/internal/services"

	"github.com/gorilla/mux"
)

type GlobalAuthProviderHandler struct {
	service services.GlobalAuthProviderService
}

func NewGlobalAuthProviderHandler(s services.GlobalAuthProviderService) *GlobalAuthProviderHandler {
	return &GlobalAuthProviderHandler{service: s}
}

func (h *GlobalAuthProviderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var p models.GlobalAuthProvider
	json.NewDecoder(r.Body).Decode(&p)
	res, err := h.service.Create(r.Context(), &p)
	if err != nil {
		response.Error(w, 500, err.Error())
		return
	}
	response.JSON(w, 201, "Created", res)
}

func (h *GlobalAuthProviderHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	res, err := h.service.GetAll(r.Context())
	if err != nil {
		response.Error(w, 500, err.Error())
		return
	}
	response.JSON(w, 200, "Success", res)
}

func (h *GlobalAuthProviderHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var p models.GlobalAuthProvider
	json.NewDecoder(r.Body).Decode(&p)
	res, err := h.service.Update(r.Context(), id, &p)
	if err != nil {
		response.Error(w, 500, err.Error())
		return
	}
	response.JSON(w, 200, "Updated", res)
}

func (h *GlobalAuthProviderHandler) ToggleStatus(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var req struct {
		IsActive bool `json:"is_active"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	err := h.service.ToggleStatus(r.Context(), id, req.IsActive)
	if err != nil {
		response.Error(w, 500, err.Error())
		return
	}
	response.JSON(w, 200, "Status Toggled", nil)
}

func (h *GlobalAuthProviderHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	err := h.service.Delete(r.Context(), id)
	if err != nil {
		response.Error(w, 500, err.Error())
		return
	}
	response.JSON(w, 200, "Deleted", nil)
}
