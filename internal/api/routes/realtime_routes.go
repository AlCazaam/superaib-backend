package routes

import (
	"net/http"
	"superaib/internal/api/handlers"
	"superaib/internal/core/logger"

	"github.com/gorilla/mux"
)

// RealtimeRoutes: Waxay diwaangelisaa WebSocket-ka iyo REST APIs ee Realtime-ka
func RealtimeRoutes(router *mux.Router, h *handlers.RealtimeHandler, apiKeyAuth func(http.Handler) http.Handler) {

	// =========================================================================
	// 📡 1. WEBSOCKET ENTRY POINT (SDK-ga Flutter wuxuu ka soo galaa halkan)
	// =========================================================================
	// ⚠️ MUHIIM: WebSocket route-ka halkan ayaan ku qoreynaa si uu u noqdo mid furan.
	// Hubinta API Key-ga waxaa lagu dhex qabtaa HandleWebSocket dhexdiisa.
	router.HandleFunc("/ws/{project_id}", h.HandleWebSocket).Methods("GET")
	router.HandleFunc("/ws", h.HandleWebSocket).Methods("GET") // Fallback route

	// =========================================================================
	// 🛠️ 2. REALTIME MANAGEMENT API (DASHBOARD & ADMIN)
	// =========================================================================
	// Wadada (Path): /api/v1/projects/{project_id}/realtime
	rt := router.PathPrefix("/projects/{project_id}/realtime").Subrouter()

	// Dhamaan API-yada hoose waxay u baahan yihiin API Key hubaal ah (X-API-KEY)
	rt.Use(apiKeyAuth)

	// --- 📺 CHANNEL MANAGEMENT (CRUD) ---
	// Soo saar dhamaan channels-ka mashruuca
	rt.HandleFunc("/channels", h.GetChannels).Methods("GET")

	// Abuur channel cusub (Manual Creation)
	rt.HandleFunc("/channels", h.CreateChannel).Methods("POST")

	// Wax ka bedel channel jira (Privacy, Name, Description)
	rt.HandleFunc("/channels/{id}", h.UpdateChannel).Methods("PUT", "PATCH")

	// Tirtir channel gabi ahaanba
	rt.HandleFunc("/channels/{id}", h.DeleteChannel).Methods("DELETE")

	// --- 📩 MESSAGE & EVENT HISTORY ---
	// U dir fariin/event channel gaar ah (Broadcast via HTTP)
	rt.HandleFunc("/channels/{channel_id}/events", h.CreateEvent).Methods("POST")

	// Soo saar taariikhda fariimaha (Message History) ee channel-kaas
	rt.HandleFunc("/channels/{channel_id}/events", h.GetEvents).Methods("GET")

	// --- 🔧 SPECIFIC EVENT MANAGEMENT ---
	// Wax ka bedel fariin hore u jirtay (Payload update)
	rt.HandleFunc("/events/{event_id}", h.UpdateEvent).Methods("PUT", "PATCH")

	// Tirtir fariin gaar ah database-ka
	rt.HandleFunc("/events/{event_id}", h.DeleteEvent).Methods("DELETE")

	logger.Log.Info("🚀 Realtime routes optimized for SDK & Dashboard successfully.")
}
