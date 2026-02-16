package routes

import (
	"superaib/internal/api/handlers"
	"superaib/internal/api/middleware"

	"github.com/gorilla/mux"
)

func GlobalAuthProviderRoutes(r *mux.Router, h *handlers.GlobalAuthProviderHandler, auth *middleware.AuthMiddleware) {
	sub := r.PathPrefix("/admin/auth-providers").Subrouter()
	sub.Use(auth.Authenticate) // Kaliya Adminka ha galo

	sub.HandleFunc("", h.GetAll).Methods("GET")
	sub.HandleFunc("", h.Create).Methods("POST")
	sub.HandleFunc("/{id}", h.Update).Methods("PUT")
	sub.HandleFunc("/{id}", h.Delete).Methods("DELETE")
	sub.HandleFunc("/{id}/status", h.ToggleStatus).Methods("PATCH")
}
