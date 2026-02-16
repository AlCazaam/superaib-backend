package routes

import (
	"net/http"
	"superaib/internal/api/handlers"

	"github.com/gorilla/mux"
)

func NotificationRoutes(router *mux.Router, h *handlers.NotificationHandler, apiKeyAuth func(http.Handler) http.Handler) {
	sub := router.PathPrefix("/projects/{project_id}/notifications").Subrouter()
	sub.Use(apiKeyAuth)
	// 🚀 KAN KU DAR:
	sub.HandleFunc("/status/{user_id}", h.GetRegistrationStatus).Methods("GET")

	// Device Registration (SDK)
	sub.HandleFunc("/register", h.RegisterToken).Methods("POST")

	// Management (Dashboard)
	sub.HandleFunc("/broadcast", h.Broadcast).Methods("POST")
	sub.HandleFunc("/history", h.GetHistory).Methods("GET")
	sub.HandleFunc("/{note_id}", h.Update).Methods("PUT")
	sub.HandleFunc("/{note_id}", h.Delete).Methods("DELETE")

	// Config
	sub.HandleFunc("/config", h.UpdateConfig).Methods("POST", "OPTIONS")
	sub.HandleFunc("/config", h.GetConfig).Methods("GET", "OPTIONS")
}
