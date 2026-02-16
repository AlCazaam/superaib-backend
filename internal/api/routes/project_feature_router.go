package routes

import (
	"net/http"
	"superaib/internal/api/handlers"

	"github.com/gorilla/mux"
)

// ProjectFeatureRoutes waxay maamushaa oggolaanshaha features-ka mashruuca (Dashboard & SDK)
func ProjectFeatureRoutes(router *mux.Router, handler *handlers.ProjectFeatureHandler, apiKeyAuth func(http.Handler) http.Handler) {

	// ===========================================================================
	// 1. PROJECT CONTEXT ROUTES (/projects/{project_id}/features)
	// Kuwani waa kuwa SDK App-ku isticmaalo si uu u helo config-ga (Cloudinary Keys)
	// Sidoo kale Dashboard-ka ayaa isticmaala si uu u Toggle-gareeyo.
	// ===========================================================================
	projectFeaturesRouter := router.PathPrefix("/projects/{project_id}/features").Subrouter()

	// Waxaan u isticmaalaynaa API Key si labada dhinacba (Dashboard & SDK) u wada isticmaalaan
	projectFeaturesRouter.Use(apiKeyAuth)

	// GET  /features -> Soo qaado dhamaan (SDK wuxuu halkan ka helayaa Storage config)
	projectFeaturesRouter.HandleFunc("", handler.GetAllFeatures).Methods("GET")

	// POST /features -> Abuuro feature cusub
	projectFeaturesRouter.HandleFunc("", handler.CreateFeature).Methods("POST")

	// PUT  /features/{feature_type}/toggle -> Shid ama Dami feature-ka (Dynamic Setup)
	projectFeaturesRouter.HandleFunc("/{feature_type}/toggle", handler.ToggleFeature).Methods("PUT")

	// ===========================================================================
	// 2. DIRECT FEATURE ROUTES (/features/{id})
	// Kuwani waa haddii si toos ah loo maamulayo hal Feature Record adigoo sita ID-giisa
	// ===========================================================================
	featuresRouter := router.PathPrefix("/features").Subrouter()
	featuresRouter.Use(apiKeyAuth)

	// GET    /features/{id} -> Soo qaado hal feature oo UUID leh
	featuresRouter.HandleFunc("/{id}", handler.GetFeatureByID).Methods("GET")

	// PUT    /features/{id} -> Update-gareey feature gaar ah
	featuresRouter.HandleFunc("/{id}", handler.UpdateFeature).Methods("PUT")

	// DELETE /features/{id} -> Tirtir feature-ka
	featuresRouter.HandleFunc("/{id}", handler.DeleteFeature).Methods("DELETE")
}
