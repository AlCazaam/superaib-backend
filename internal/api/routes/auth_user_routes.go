package routes

import (
	"net/http"
	"superaib/internal/api/handlers"

	"github.com/gorilla/mux"
)

func AuthUserRoutes(router *mux.Router, handler *handlers.AuthUserHandler, apiKeyAuth func(http.Handler) http.Handler) {
	// /projects/{ref}/auth-users
	const prefix = "/projects/{reference_id}/auth-users"
	sub := router.PathPrefix(prefix).Subrouter()
	sub.Use(apiKeyAuth) // ✅ API Key requirement

	// 1. SDK Authentication Endpoints
	sub.HandleFunc("/login", handler.LoginAuthUser).Methods("POST")
	sub.HandleFunc("/google", handler.LoginWithGoogle).Methods("POST")
	sub.HandleFunc("/facebook", handler.LoginWithFacebook).Methods("POST")

	// 🚀 2. Password Management (OTP)
	sub.HandleFunc("/forgot-password", handler.SendOTP).Methods("POST")
	sub.HandleFunc("/reset-password", handler.ResetPassword).Methods("POST")

	// 🚀 3. Impersonation (Admin/Dev feature)
	sub.HandleFunc("/impersonate", handler.ImpersonateLogin).Methods("POST")

	// 4. CRUD (Developer Management)
	sub.HandleFunc("", handler.GetAllAuthUsers).Methods("GET")
	sub.HandleFunc("", handler.CreateAuthUser).Methods("POST")
	sub.HandleFunc("/{id}", handler.GetAuthUserByID).Methods("GET")
	sub.HandleFunc("/{id}", handler.UpdateAuthUser).Methods("PUT")
	sub.HandleFunc("/{id}", handler.DeleteAuthUser).Methods("DELETE")
}
