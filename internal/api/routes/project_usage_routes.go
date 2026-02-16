package routes

import (
	"superaib/internal/api/handlers"
	"superaib/internal/api/middleware"

	"github.com/gorilla/mux"
)

func ProjectUsageRoutes(router *mux.Router, h *handlers.ProjectUsageHandler, auth *middleware.AuthMiddleware) {
	usageRouter := router.PathPrefix("/projects/{project_id}/usage").Subrouter()
	usageRouter.Use(auth.Authenticate)
	usageRouter.HandleFunc("", h.GetUsage).Methods("GET")
}
