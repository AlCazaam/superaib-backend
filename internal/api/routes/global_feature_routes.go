package routes

import (
	"superaib/internal/api/handlers"
	"superaib/internal/api/middleware"

	"github.com/gorilla/mux"
)

func GlobalFeatureRoutes(r *mux.Router, handler *handlers.GlobalFeatureHandler, authMiddleware *middleware.AuthMiddleware) {
	adminRouter := r.PathPrefix("/admin/features").Subrouter()
	adminRouter.Use(authMiddleware.Authenticate)

	adminRouter.HandleFunc("", handler.GetAllFeatures).Methods("GET")
	adminRouter.HandleFunc("", handler.CreateFeature).Methods("POST")
	adminRouter.HandleFunc("/{id}", handler.UpdateFeature).Methods("PUT")
	adminRouter.HandleFunc("/toggle/{type}", handler.ToggleFeature).Methods("PUT")
	adminRouter.HandleFunc("/{id}", handler.DeleteFeature).Methods("DELETE")
}
