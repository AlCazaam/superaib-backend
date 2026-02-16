package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"superaib/internal/api/response"
	"superaib/internal/core/logger"
	"superaib/internal/models"
	"superaib/internal/services"
)

type AuthHandler struct {
	authService services.AuthService
}

func NewAuthHandler(as services.AuthService) *AuthHandler {
	return &AuthHandler{authService: as}
}

// Register handles POST /auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {

	var req models.UserCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// WAXAA LA BEDDELAY: U gudbi context-ka
	user, token, err := h.authService.Register(r.Context(), &req)
	if err != nil {
		logger.Log.Errorf("Failed to register user: %v", err)
		if strings.Contains(err.Error(), "already registered") || strings.Contains(err.Error(), "already taken") {
			response.Error(w, http.StatusConflict, err.Error())
		} else if strings.Contains(err.Error(), "validation failed") {
			response.Error(w, http.StatusBadRequest, err.Error())
		} else {
			response.Error(w, http.StatusInternalServerError, "Failed to register user", err.Error())
		}
		return
	}

	response.JSON(w, http.StatusCreated, "User registered successfully", map[string]interface{}{
		"user":         user,
		"access_token": token,
	})
}

// Login handles POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.UserLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// WAXAA LA BEDDELAY: U gudbi context-ka
	user, token, err := h.authService.Login(r.Context(), &req)
	if err != nil {
		logger.Log.Errorf("Failed login attempt for %s: %v", req.Email, err)
		if strings.Contains(err.Error(), "invalid credentials") {
			response.Error(w, http.StatusUnauthorized, err.Error())
		} else if strings.Contains(err.Error(), "validation failed") {
			response.Error(w, http.StatusBadRequest, err.Error())
		} else {
			response.Error(w, http.StatusInternalServerError, "Login failed", err.Error())
		}
		return
	}

	response.JSON(w, http.StatusOK, "Login successful", map[string]interface{}{
		"user":         user,
		"access_token": token,
	})
}
