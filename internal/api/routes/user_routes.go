package routes

import (
	"net/http"
	"superaib/internal/api/handlers"
	"superaib/internal/core/logger"

	"github.com/gorilla/mux"
)

func UserRoutes(r *mux.Router, userHandler *handlers.UserHandler, userAuth func(http.Handler) http.Handler) {
	usersRouter := r.PathPrefix("/users").Subrouter()
	usersRouter.Use(userAuth)

	// User Self-Management
	usersRouter.HandleFunc("/me", userHandler.GetCurrentUserProfile).Methods("GET")
	usersRouter.HandleFunc("/me", userHandler.UpdateUserProfile).Methods("PUT")

	// User Administration (Optional/Admin)
	usersRouter.HandleFunc("/{id}", userHandler.GetUserByID).Methods("GET")
	usersRouter.HandleFunc("/{id}", userHandler.UpdateUserProfile).Methods("PUT")
	usersRouter.HandleFunc("/{id}", userHandler.DeleteUser).Methods("DELETE")
	usersRouter.HandleFunc("", userHandler.GetAllUsers).Methods("GET")

	// 🚀 GLOBAL WIPE ROUTE
	usersRouter.HandleFunc("/account/wipe", userHandler.DeleteAccount).Methods("DELETE")

	logger.Log.Info("✅ User routes registered successfully.")
}
