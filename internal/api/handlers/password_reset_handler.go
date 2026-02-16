package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"superaib/internal/api/response"
	"superaib/internal/services"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type PasswordResetHandler struct {
	service services.PasswordResetService
}

func NewPasswordResetHandler(s services.PasswordResetService) *PasswordResetHandler {
	return &PasswordResetHandler{service: s}
}

func (h *PasswordResetHandler) getProjectID(r *http.Request) string {
	if ctxID, ok := r.Context().Value("projectID").(string); ok && ctxID != "" {
		return ctxID
	}
	vars := mux.Vars(r)
	projectID := vars["project_id"]
	if projectID == "" {
		projectID = vars["reference_id"]
	}
	return projectID
}

// 1. POST /request (OTP Request)

func (h *PasswordResetHandler) RequestReset(w http.ResponseWriter, r *http.Request) {
	projectID := h.getProjectID(r)
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	err := h.service.RequestReset(r.Context(), projectID, req.Email)
	if err != nil {
		errMsg := err.Error()

		// 1. Account Status Issues (403 Forbidden)
		if strings.Contains(errMsg, "blocked") || strings.Contains(errMsg, "invited") {
			response.Error(w, http.StatusForbidden, errMsg)
			return
		}

		// 2. Rate Limit Issues (429 Too Many Requests)
		if strings.Contains(errMsg, "limit reached") || strings.Contains(errMsg, "locked") {
			response.Error(w, http.StatusTooManyRequests, errMsg)
			return
		}

		// 3. Configuration Issues (501 or 403)
		if strings.Contains(errMsg, "email service") {
			status := http.StatusNotImplemented
			if strings.Contains(errMsg, "disabled") {
				status = http.StatusForbidden
			}
			response.Error(w, status, errMsg)
			return
		}

		response.Error(w, http.StatusInternalServerError, "An unexpected error occurred")
		return
	}

	response.JSON(w, http.StatusOK, "If an account exists, an OTP has been sent.", nil)
}

// 2. POST /verify

func (h *PasswordResetHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	projectID := h.getProjectID(r)
	var req struct {
		Email string `json:"email"`
		OTP   string `json:"otp"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	token, err := h.service.VerifyOTP(r.Context(), projectID, req.Email, req.OTP)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, "OTP Verified", map[string]string{"reset_token": token})
}

func (h *PasswordResetHandler) ConfirmReset(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if err := h.service.ConfirmReset(r.Context(), req.Token, req.NewPassword); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, "Password changed successfully", nil)
}

func (h *PasswordResetHandler) GetResetHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, _ := uuid.Parse(vars["user_id"])
	tokens, err := h.service.GetResetHistory(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, "History retrieved", tokens)
}

func (h *PasswordResetHandler) DeleteResetToken(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tokenID, _ := uuid.Parse(vars["token_id"])
	h.service.DeleteToken(r.Context(), tokenID)
	response.JSON(w, http.StatusOK, "Token deleted", nil)
}
