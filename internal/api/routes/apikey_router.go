// Faylka: routes/api_key_routes.go
package routes

import (
	"net/http"
	"superaib/internal/api/handlers"
	"superaib/internal/core/logger"

	"github.com/gorilla/mux"
)

func APIKeyRoutes(r *mux.Router, apiKeyHandler *handlers.APIKeyHandler, userAuth func(http.Handler) http.Handler) {
	// WAXAA LA BEDDELAY: project_id -> reference_id
	apiKeysRouter := r.PathPrefix("/projects/{reference_id}/keys").Subrouter()
	apiKeysRouter.Use(userAuth)

	apiKeysRouter.HandleFunc("", apiKeyHandler.CreateAPIKey).Methods("POST")
	apiKeysRouter.HandleFunc("", apiKeyHandler.GetAllAPIKeysForProject).Methods("GET")
	apiKeysRouter.HandleFunc("/{id}", apiKeyHandler.GetAPIKeyByID).Methods("GET")
	apiKeysRouter.HandleFunc("/{id}", apiKeyHandler.UpdateAPIKey).Methods("PUT")
	apiKeysRouter.HandleFunc("/{id}/revoke", apiKeyHandler.RevokeAPIKey).Methods("POST")
	apiKeysRouter.HandleFunc("/{id}", apiKeyHandler.DeleteAPIKey).Methods("DELETE")

	logger.Log.Info("✅ APIKey routes registered successfully.")
}
