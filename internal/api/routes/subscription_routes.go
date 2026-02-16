package routes

import (
	"superaib/internal/api/handlers"
	"superaib/internal/api/middleware"

	"github.com/gorilla/mux"
)

func SubscriptionRoutes(router *mux.Router, h *handlers.SubscriptionHandler, auth *middleware.AuthMiddleware) {
	subRouter := router.PathPrefix("/projects/{project_id}/upgrade").Subrouter()
	subRouter.Use(auth.Authenticate)

	subRouter.HandleFunc("", h.Upgrade).Methods("POST")
}
