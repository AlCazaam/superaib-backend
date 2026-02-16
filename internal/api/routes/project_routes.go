// Faylka: routes/project_routes.go
package routes

import (
	"net/http"
	"superaib/internal/api/handlers"
	"superaib/internal/core/logger"

	"github.com/gorilla/mux"
)

func ProjectRoutes(r *mux.Router, projectHandler *handlers.ProjectHandler, userAuth func(http.Handler) http.Handler) {
	projectsRouter := r.PathPrefix("/projects").Subrouter()
	projectsRouter.Use(userAuth)

	projectsRouter.HandleFunc("", projectHandler.CreateProject).Methods("POST")
	projectsRouter.HandleFunc("", projectHandler.GetAllProjects).Methods("GET")
	projectsRouter.HandleFunc("/{reference_id}", projectHandler.GetProjectByReferenceID).Methods("GET")
	projectsRouter.HandleFunc("/{reference_id}", projectHandler.UpdateProject).Methods("PUT")
	projectsRouter.HandleFunc("/{reference_id}", projectHandler.DeleteProject).Methods("DELETE")

	logger.Log.Info("✅ Project routes registered successfully.")
}
