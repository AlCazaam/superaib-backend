package routes

import (
	"superaib/internal/api/handlers"
	"superaib/internal/api/middleware"

	"github.com/gorilla/mux"
)

func AnalyticsRoutes(router *mux.Router, handler *handlers.AnalyticsHandler, authMiddleware *middleware.AuthMiddleware) {
	analyticsRouter := router.PathPrefix("/projects/{project_id}/analytics").Subrouter()
	analyticsRouter.Use(authMiddleware.Authenticate)
	analyticsRouter.HandleFunc("", handler.GetAnalytics).Methods("GET")
}
