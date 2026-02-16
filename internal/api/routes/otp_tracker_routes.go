package routes

import (
	"net/http"
	"superaib/internal/api/handlers"

	"github.com/gorilla/mux"
)

func OtpTrackerRoutes(router *mux.Router, handler *handlers.OtpTrackerHandler, apiKeyAuth func(http.Handler) http.Handler) {
	r := router.PathPrefix("/projects/{reference_id}/auth-users/{id}/reset-limit").Subrouter()
	r.Use(apiKeyAuth)
	r.HandleFunc("", handler.ResetLimit).Methods("POST")
}
