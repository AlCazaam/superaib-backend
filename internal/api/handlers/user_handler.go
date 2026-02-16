package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"superaib/internal/api/response"
	"superaib/internal/models"
	"superaib/internal/services"

	"github.com/gorilla/mux"
)

type UserHandler struct {
	userService services.UserService
}

func NewUserHandler(us services.UserService) *UserHandler {
	return &UserHandler{userService: us}
}

func (h *UserHandler) GetCurrentUserProfile(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value("userID").(string)
	user, err := h.userService.GetUserByID(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "User not found")
		return
	}
	response.JSON(w, http.StatusOK, "Profile retrieved", user)
}

func (h *UserHandler) UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	// 1. Hel ID-ga (URL mise /me)
	targetID := mux.Vars(r)["id"]
	if targetID == "" {
		targetID, _ = r.Context().Value("userID").(string)
	}

	// 2. Decode Payload
	var req models.UserUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// 3. Wac Service-ka
	user, err := h.userService.UpdateUser(r.Context(), targetID, &req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to update profile", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "User profile updated successfully", user)
}

func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	user, err := h.userService.GetUserByID(r.Context(), idStr)
	if err != nil {
		response.Error(w, http.StatusNotFound, "User not found")
		return
	}
	response.JSON(w, http.StatusOK, "User found", user)
}

func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 10
	}
	users, err := h.userService.GetAllUsers(r.Context(), limit, 0)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to get users")
		return
	}
	response.JSON(w, http.StatusOK, "Users list", users)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	if err := h.userService.DeleteUser(r.Context(), idStr); err != nil {
		response.Error(w, http.StatusInternalServerError, "Delete failed")
		return
	}
	response.JSON(w, http.StatusOK, "User deleted", nil)
}

// DeleteAccount handles DELETE /users/account/wipe
func (h *UserHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Hel ID-ga qofka fadhiya (Session/JWT)
	userID, ok := ctx.Value("userID").(string)
	if !ok || userID == "" {
		response.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// 2. Wac Service-ka sifeeyaha ah
	if err := h.userService.DeleteAccount(ctx, userID); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to delete account", err.Error())
		return
	}

	// 3. U sheeg Flutter-ka inuu login-ka ka saaro (Logout)
	response.JSON(w, http.StatusOK, "Your account and all associated data have been permanently deleted.", nil)
}
