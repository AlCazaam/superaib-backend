package routes

import (
	"net/http"
	"superaib/internal/api/handlers"

	"github.com/gorilla/mux"
)

func PasswordResetRoutes(router *mux.Router, handler *handlers.PasswordResetHandler, apiKeyAuth func(http.Handler) http.Handler) {
	// Base URL: /projects/{reference_id}/auth/password-reset
	r := router.PathPrefix("/projects/{reference_id}/auth/password-reset").Subrouter()

	// API Key waa qasab si loo garto Project-ka
	r.Use(apiKeyAuth)

	r.HandleFunc("/request", handler.RequestReset).Methods("POST")
	r.HandleFunc("/verify", handler.VerifyOTP).Methods("POST")
	r.HandleFunc("/confirm", handler.ConfirmReset).Methods("POST")
	// ✅ NEW ROUTES FOR DEVELOPER DASHBOARD / HISTORY
	// GET history-ga token-ada user-ka gaarka ah
	r.HandleFunc("/history/user/{user_id}", handler.GetResetHistory).Methods("GET")
	// DELETE token gaar ah
	r.HandleFunc("/token/{token_id}", handler.DeleteResetToken).Methods("DELETE")
}
