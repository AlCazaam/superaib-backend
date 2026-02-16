package routes

import (
	"net/http"
	"superaib/internal/api/handlers"
	"superaib/internal/core/logger"

	"github.com/gorilla/mux"
)

// StorageRoutes waxay maamushaa dhamaan adeegyada File Storage-ka ee SDK-ga iyo Dashboard-ka
func StorageRoutes(router *mux.Router, handler *handlers.StorageHandler, apiKeyAuth func(http.Handler) http.Handler) {

	// 1. Project Context Router (/api/v1/projects/{project_id}/storage)
	const projectPrefix = "/projects/{project_id}/storage"
	storageRouter := router.PathPrefix(projectPrefix).Subrouter()

	// 2. Middleware Security (Si loo xaqiijiyo API Key-ga SDK-ga ama Dashboard-ka)
	storageRouter.Use(apiKeyAuth)

	// --- 🚀 BINARY FILE OPERATIONS ---

	// POST /api/v1/projects/{project_id}/storage/upload
	// Kani waa mishiinka sawirka dhabta ah (Binary) ka aqbalaya SDK-ga kuna ridayo Cloudinary
	storageRouter.HandleFunc("/upload", handler.UploadFile).Methods("POST")

	// --- 📊 METADATA OPERATIONS (pgAdmin) ---

	// POST /api/v1/projects/{project_id}/storage/files
	// Kani wuxuu diwaangelinayaa Metadata-ga file-ka (Record creation)
	storageRouter.HandleFunc("/files", handler.CreateFile).Methods("POST")

	// GET /api/v1/projects/{project_id}/storage/files
	// Kani wuxuu soo saarayaa dhamaan files-ka mashruuca u xareysan (List view)
	storageRouter.HandleFunc("/files", handler.GetFiles).Methods("GET")

	// DELETE /api/v1/projects/{project_id}/storage/files/{id}
	// Kani wuxuu tirtirayaa file gaar ah isagoo isticmaalaya UUID
	storageRouter.HandleFunc("/files/{id}", handler.DeleteFile).Methods("DELETE")

	logger.Log.Info("✅ Storage routes (including /upload) registered with API Key security.")
}
