// Faylka: routes/auth_routes.go
package routes

import (
	"superaib/internal/api/handlers"
	"superaib/internal/core/logger"

	"github.com/gorilla/mux"
)

func AuthRoutes(r *mux.Router, authHandler *handlers.AuthHandler) {
	authRouter := r.PathPrefix("/auth").Subrouter()

	authRouter.HandleFunc("/register", authHandler.Register).Methods("POST")
	authRouter.HandleFunc("/login", authHandler.Login).Methods("POST")
	logger.Log.Info("✅ Auth routes registered successfully.")
}
