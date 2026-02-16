package handlers

import (
	"encoding/json"
	"net/http"
	"superaib/internal/api/response"
	"superaib/internal/services"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type ImpersonationHandler struct {
	service services.ImpersonationService
}

func NewImpersonationHandler(s services.ImpersonationService) *ImpersonationHandler {
	return &ImpersonationHandler{service: s}
}

// CreateToken handles POST /projects/{reference_id}/auth-users/{user_id}/impersonate
func (h *ImpersonationHandler) CreateToken(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	// Soo qaado Project ID (Sida handler-yada kale)
	// 1. Project ID Extraction (Robust way)
	projectID := vars["reference_id"] // Si toos ah uga soo qaad route-ka
	if projectID == "" {
		if ctxID, ok := r.Context().Value("projectID").(string); ok {
			projectID = ctxID
		}
	}

	if projectID == "" {
		response.Error(w, http.StatusBadRequest, "Missing Project ID")
		return
	}

	userIDStr := vars["user_id"]
	userID, err := uuid.Parse(userIDStr)

	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid User ID")
		return
	}

	// Aqri Duration-ka (optional)
	var req struct {
		Duration int `json:"duration_minutes"` // e.g. 60, 1440 (1 day)
	}
	_ = json.NewDecoder(r.Body).Decode(&req) // Ignore error, default to 0 if empty

	// U yeer Service-ka
	tokenRecord, err := h.service.GenerateToken(r.Context(), projectID, userID, req.Duration)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to generate token", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "Impersonation token generated", map[string]interface{}{
		"access_token": tokenRecord.Token,
		"expires_at":   tokenRecord.ExpiresAt,
		"user_email":   tokenRecord.User.Email, // Halkan ayaad ka helaysaa email-ka saxda ah
	})
}

// 2. RevokeToken (POST /revoke/{token_id})
func (h *ImpersonationHandler) RevokeToken(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tokenID := vars["token_id"]

	if err := h.service.RevokeToken(r.Context(), tokenID); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to revoke token", err.Error())
		return
	}
	response.JSON(w, http.StatusOK, "Token revoked successfully", nil)
}

// 3. ExtendToken (POST /extend/{token_id})
func (h *ImpersonationHandler) ExtendToken(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tokenID := vars["token_id"]

	var req struct {
		Minutes int `json:"minutes"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	token, err := h.service.ExtendToken(r.Context(), tokenID, req.Minutes)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to extend token", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "Token extended", map[string]interface{}{
		"new_expires_at": token.ExpiresAt,
	})
}

// 4. GetActiveTokens (GET /active)
func (h *ImpersonationHandler) GetActiveTokens(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]

	tokens, err := h.service.GetActiveTokens(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to fetch tokens", err.Error())
		return
	}
	response.JSON(w, http.StatusOK, "Active tokens", tokens)
}

// ValidateToken handles POST /projects/.../impersonate/validate
func (h *ImpersonationHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// U yeer Service-ka
	details, err := h.service.GetTokenDetails(r.Context(), req.Token)
	if err != nil {
		// Halkan si cad ugu sheeg qaladka (Expired, Invalid, etc.)
		response.Error(w, http.StatusUnauthorized, "Validation failed", err.Error())
		return
	}

	// Soo celi xogta User-ka
	response.JSON(w, http.StatusOK, "Token is valid", map[string]interface{}{
		"valid":      true,
		"expires_at": details.ExpiresAt,
		"user": map[string]interface{}{
			"id":     details.User.ID,
			"email":  details.User.Email,
			"status": details.User.Status,
		},
	})
}
