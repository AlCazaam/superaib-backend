package routes

import (
	"superaib/internal/api/handlers"
	"superaib/internal/api/middleware"

	"github.com/gorilla/mux"
)

func ProjectAuthConfigRoutes(r *mux.Router, h *handlers.ProjectAuthConfigHandler, auth *middleware.AuthMiddleware) {
	sub := r.PathPrefix("/projects/{project_id}/auth-configs").Subrouter()
	sub.Use(auth.Authenticate) // Xaqiiji inuu qofku login yahay

	sub.HandleFunc("", h.GetConfigs).Methods("GET")
	sub.HandleFunc("", h.Configure).Methods("POST")
}
