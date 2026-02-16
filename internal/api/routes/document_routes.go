package routes

import (
	"net/http"
	"superaib/internal/api/handlers"

	"github.com/gorilla/mux"
)

func DocumentRoutes(router *mux.Router, h *handlers.DocumentHandler, apiKeyAuth func(http.Handler) http.Handler) {
	// 1. Root subrouter for projects
	// Dhammaan waddooyinkani waxay u baahan yihiin API Key ama Session Auth
	projectRouter := router.PathPrefix("/projects/{project_id}").Subrouter()
	projectRouter.Use(apiKeyAuth)

	// ===========================================================================
	// 🖥️ DASHBOARD ROUTES (Admin Panel Logic)
	// Waxay isticmaalaan URL-ka: /collections/{collection_id}/...
	// ===========================================================================

	// Collection management
	projectRouter.HandleFunc("/collections", h.GetCollections).Methods("GET")
	projectRouter.HandleFunc("/collections", h.CreateCollection).Methods("POST")
	projectRouter.HandleFunc("/collections/{collection}", h.RenameCollection).Methods("PUT", "PATCH")
	projectRouter.HandleFunc("/collections/{collection}", h.DeleteCollection).Methods("DELETE")

	// Document Management (Dashboard UI)
	projectRouter.HandleFunc("/collections/{collection}/documents", h.Create).Methods("POST")
	projectRouter.HandleFunc("/collections/{collection}/documents", h.AdvancedSearch).Methods("GET", "POST")
	projectRouter.HandleFunc("/collections/{collection}/documents/{id}", h.GetByID).Methods("GET")
	projectRouter.HandleFunc("/collections/{collection}/documents/{id}", h.UpdateDocument).Methods("PUT", "PATCH")
	projectRouter.HandleFunc("/collections/{collection}/documents/{id}", h.Delete).Methods("DELETE")

	// ===========================================================================
	// 🚀 SDK / APP ROUTES (Developer Friendly Logic)
	// Waxay isticmaalaan habka gaaban: /db/{collection_name}/...
	// ===========================================================================

	// 1. Basic CRUD & List
	projectRouter.HandleFunc("/db/{collection}", h.Create).Methods("POST")               // .add({...})
	projectRouter.HandleFunc("/db/{collection}", h.AdvancedSearch).Methods("GET")        // .get()
	projectRouter.HandleFunc("/db/{collection}/query", h.AdvancedSearch).Methods("POST") // .where(...).get()

	// 2. Single Document Operations
	projectRouter.HandleFunc("/db/{collection}/{id}", h.GetByID).Methods("GET")          // .doc(id).get()
	projectRouter.HandleFunc("/db/{collection}/{id}", h.UpdateDocument).Methods("PATCH") // .doc(id).update({...})
	projectRouter.HandleFunc("/db/{collection}/{id}", h.Delete).Methods("DELETE")        // .doc(id).delete()

	// 3. ⚡ ADVANCED OPERATIONS (The Powerhouse)

	// Set (Overwrite or Merge)
	// .doc(id).set({...}, merge: true)
	projectRouter.HandleFunc("/db/{collection}/{id}", h.Set).Methods("PUT")

	// Upsert (Update or Create)
	// .doc(id).upsert({...})
	projectRouter.HandleFunc("/db/{collection}/{id}/upsert", h.Upsert).Methods("POST")

	// Exists (Check if document exists)
	// .doc(id).exists()
	projectRouter.HandleFunc("/db/{collection}/{id}/exists", h.Exists).Methods("GET")

	// Increment (Atomic Counter)
	// .doc(id).increment("views", 1)
	projectRouter.HandleFunc("/db/{collection}/{id}/increment", h.Increment).Methods("POST")

	// Count (Get total number of documents with optional filters)
	// .collection("notes").count()
	projectRouter.HandleFunc("/db/{collection}/count", h.Count).Methods("GET", "POST")

	// 4. SDK Collection Config (Rename/Delete via SDK)
	projectRouter.HandleFunc("/db/{collection}/config", h.RenameCollection).Methods("PATCH")
	projectRouter.HandleFunc("/db/{collection}/config", h.DeleteCollection).Methods("DELETE")

}
