package routes

import (
	"net/http"
	"superaib/internal/api/handlers"

	"github.com/gorilla/mux"
)

func RateLimitRoutes(router *mux.Router, handler *handlers.RateLimitPolicyHandler, apiKeyAuth func(http.Handler) http.Handler) {
	// Base URL: /projects/{reference_id}/auth/rate-limit
	r := router.PathPrefix("/projects/{reference_id}/auth/rate-limit").Subrouter()

	// API Key waa qasab si loo garto Project-ka
	r.Use(apiKeyAuth)

	// Standard CRUD Routes
	r.HandleFunc("/policy", handler.GetPolicy).Methods("GET")
	r.HandleFunc("/policy", handler.CreatePolicy).Methods("POST")
	r.HandleFunc("/policy", handler.UpdatePolicy).Methods("PUT")
	r.HandleFunc("/policy", handler.DeletePolicy).Methods("DELETE")

	// Toggle (Switch Button)
	r.HandleFunc("/policy/toggle", handler.TogglePolicy).Methods("PUT")
}
