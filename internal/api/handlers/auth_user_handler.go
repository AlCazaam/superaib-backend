package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"superaib/internal/api/response"
	"superaib/internal/models"
	"superaib/internal/services"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gorm.io/datatypes"
)

type AuthUserHandler struct {
	service services.AuthUserService
}

func NewAuthUserHandler(s services.AuthUserService) *AuthUserHandler {
	return &AuthUserHandler{service: s}
}

// getProjectID: Wuxuu UUID-ga saxda ah ka soo saaraa Context-ka (Middleware) ama URL-ka
func (h *AuthUserHandler) getProjectID(r *http.Request) string {
	// 1. Marka hore ka eeg Context-ka (Kani waa kan ugu saxan ee Middleware-ku dhex dhigo)
	if ctxID, ok := r.Context().Value("projectID").(string); ok && ctxID != "" {
		return ctxID
	}

	// 2. Haddii la waayo, ka eeg URL-ka (labada magacba: project_id ama reference_id)
	vars := mux.Vars(r)
	projectID := vars["project_id"]
	if projectID == "" {
		projectID = vars["reference_id"]
	}
	return projectID
}

// 1. GetAllAuthUsers handles GET /projects/{reference_id}/auth-users
func (h *AuthUserHandler) GetAllAuthUsers(w http.ResponseWriter, r *http.Request) {
	projectID := h.getProjectID(r)
	if projectID == "" {
		response.Error(w, http.StatusBadRequest, "Project identifier is required")
		return
	}

	users, err := h.service.GetAllByProject(r.Context(), projectID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to fetch users", err.Error())
		return
	}

	// ✅ XALKA: Waxaan u soo celinaynaa List saafi ah gudaha "data"
	// si Flutter-ku uusan "Invalid data format" ugu dhihin.
	response.JSON(w, http.StatusOK, "Auth users retrieved successfully", users)
}

// 2. CreateAuthUser handles POST /projects/{reference_id}/auth-users
func (h *AuthUserHandler) CreateAuthUser(w http.ResponseWriter, r *http.Request) {
	projectID := h.getProjectID(r)
	if _, err := uuid.Parse(projectID); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid project ID format")
		return
	}

	var req struct {
		Email    string          `json:"email"`
		Password string          `json:"password"`
		Metadata json.RawMessage `json:"metadata"`
		Roles    json.RawMessage `json:"roles"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user := &models.AuthUser{
		ProjectID: projectID,
		Email:     req.Email,
		Metadata:  datatypes.JSON(req.Metadata),
		Roles:     datatypes.JSON(req.Roles),
		Status:    models.AuthUserActive,
	}

	if err := h.service.Create(r.Context(), user, req.Password); err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "exists") || strings.Contains(err.Error(), "duplicate") {
			status = http.StatusConflict
		}
		response.Error(w, status, "Failed to create user", err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, "Auth user created successfully", user)
}

// 3. GetAuthUserByID handles GET /projects/{reference_id}/auth-users/{id}
func (h *AuthUserHandler) GetAuthUserByID(w http.ResponseWriter, r *http.Request) {
	projectID := h.getProjectID(r)
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid auth user ID format")
		return
	}

	user, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "Auth user not found", err.Error())
		return
	}

	// Ammaanka: Hubi in user-kani uu ka tirsan yahay project-ka saxda ah
	if user.ProjectID != projectID {
		response.Error(w, http.StatusForbidden, "Access denied: user does not belong to this project")
		return
	}

	response.JSON(w, http.StatusOK, "Auth user retrieved successfully", user)
}

// 4. UpdateAuthUser handles PUT /projects/{reference_id}/auth-users/{id}
// 4. UpdateAuthUser handles PUT /projects/{reference_id}/auth-users/{id}
func (h *AuthUserHandler) UpdateAuthUser(w http.ResponseWriter, r *http.Request) {
	projectID := h.getProjectID(r)
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid user ID format")
		return
	}

	// 1. Decode Request-ka
	var updates models.AuthUser
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// 2. Soo qaado user-ka hadda jira (Existing User)
	existing, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "User not found", err.Error())
		return
	}

	// 3. Hubi inuu ka tirsan yahay Project-kan
	if existing.ProjectID != projectID {
		response.Error(w, http.StatusForbidden, "User does not belong to this project")
		return
	}

	// 4. ✅ SMART UPDATE: Kaliya bedel wixii la soo diray (oo aan maranayn)

	// Email: Kaliya bedel haddii uu cusub yahay oo uusan maranayn
	if updates.Email != "" && updates.Email != existing.Email {
		existing.Email = updates.Email
	}

	// Status: Bedel haddii la soo diray
	if updates.Status != "" {
		existing.Status = updates.Status
	}

	// Metadata: Bedel haddii la soo diray
	if len(updates.Metadata) > 0 {
		existing.Metadata = updates.Metadata
	}

	// Roles: Bedel haddii la soo diray
	if len(updates.Roles) > 0 {
		existing.Roles = updates.Roles
	}

	// Updates: Update Time
	existing.UpdatedAt = time.Now()

	// 5. Kaydi (Update)
	if err := h.service.Update(r.Context(), existing); err != nil {
		// Hubi haddii Email-ka cusub uu hore u jiray (Unique Constraint)
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			response.Error(w, http.StatusConflict, "Email already exists in this project")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Update failed", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "User updated successfully", existing)
}

// 5. DeleteAuthUser handles DELETE /projects/{reference_id}/auth-users/{id}
func (h *AuthUserHandler) DeleteAuthUser(w http.ResponseWriter, r *http.Request) {
	projectID := h.getProjectID(r)
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid user ID format")
		return
	}

	if err := h.service.Delete(r.Context(), projectID, id); err != nil {
		response.Error(w, http.StatusInternalServerError, "Delete failed", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "User deleted successfully", nil)
}

// --- Standard Handlers ---

// LoginAuthUser (Handler)
// 1. Email/Password Login
func (h *AuthUserHandler) LoginAuthUser(w http.ResponseWriter, r *http.Request) {
	projectID := h.getProjectID(r)
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, token, err := h.service.LoginUser(r.Context(), projectID, req.Email, req.Password)
	if err != nil {
		errMsg := err.Error()

		// ✅ HUBI STATUS-KA: Blocked, Invited, ama Not Active
		if strings.Contains(errMsg, "blocked") || strings.Contains(errMsg, "invited") || strings.Contains(errMsg, "not active") {
			response.Error(w, http.StatusForbidden, errMsg)
			return
		}

		response.Error(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	response.JSON(w, http.StatusOK, "Login successful", map[string]interface{}{
		"user":         user,
		"access_token": token,
	})
}

// 2. Google Login
func (h *AuthUserHandler) LoginWithGoogle(w http.ResponseWriter, r *http.Request) {
	projectID := h.getProjectID(r)
	var req struct {
		IdToken string `json:"idToken"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, token, err := h.service.LoginWithGoogle(r.Context(), projectID, req.IdToken)

	if err != nil {
		errMsg := err.Error()

		// ✅ HUBI STATUS-KA: Blocked ama Invited
		if strings.Contains(errMsg, "blocked") || strings.Contains(errMsg, "invited") {
			response.Error(w, http.StatusForbidden, errMsg)
			return
		}

		// Xaaladda Setup-ka (Not Configured)
		if strings.Contains(errMsg, "not configured") {
			response.Error(w, http.StatusNotImplemented, "Google Sign-In is not set up for this project")
			return
		}

		// Xaaladda Switch-ka (Disabled)
		if strings.Contains(errMsg, "disabled") {
			response.Error(w, http.StatusForbidden, "Google Sign-In is currently disabled")
			return
		}

		response.Error(w, http.StatusUnauthorized, "Google authentication failed", errMsg)
		return
	}

	response.JSON(w, http.StatusOK, "Google login successful", map[string]interface{}{
		"user":         user,
		"access_token": token,
	})
}

// 3. Facebook Login
func (h *AuthUserHandler) LoginWithFacebook(w http.ResponseWriter, r *http.Request) {
	projectID := h.getProjectID(r)
	var req struct {
		AccessToken string `json:"accessToken"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, token, err := h.service.LoginWithFacebook(r.Context(), projectID, req.AccessToken)

	if err != nil {
		errMsg := err.Error()

		// ✅ HUBI STATUS-KA: Blocked ama Invited
		if strings.Contains(errMsg, "blocked") || strings.Contains(errMsg, "invited") {
			response.Error(w, http.StatusForbidden, errMsg)
			return
		}

		// Xaaladda Setup-ka (Not Configured)
		if strings.Contains(errMsg, "not configured") {
			response.Error(w, http.StatusNotImplemented, "Facebook Login is not set up for this project")
			return
		}

		// Xaaladda Switch-ka (Disabled)
		if strings.Contains(errMsg, "disabled") {
			response.Error(w, http.StatusForbidden, "Facebook Login is currently disabled")
			return
		}

		response.Error(w, http.StatusUnauthorized, "Facebook authentication failed", errMsg)
		return
	}

	response.JSON(w, http.StatusOK, "Facebook login successful", map[string]interface{}{
		"user":         user,
		"access_token": token,
	})
}

// 📧 HANDLER: Send OTP
func (h *AuthUserHandler) SendOTP(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	projectID := h.getProjectID(r)
	if err := h.service.SendPasswordResetOTP(r.Context(), projectID, req.Email); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, "OTP sent successfully", nil)
}

// 🔐 HANDLER: Reset Password
func (h *AuthUserHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		OTP      string `json:"otp"`
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if err := h.service.VerifyOTPAndResetPassword(r.Context(), h.getProjectID(r), req.Email, req.OTP, req.Password); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, "Password updated successfully", nil)
}

// 🕵️ HANDLER: Impersonate Login
func (h *AuthUserHandler) ImpersonateLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	user, jwtToken, err := h.service.LoginWithImpersonation(r.Context(), h.getProjectID(r), req.Token)
	if err != nil {
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, "Impersonated successfully", map[string]interface{}{
		"user": user, "access_token": jwtToken,
	})
}
