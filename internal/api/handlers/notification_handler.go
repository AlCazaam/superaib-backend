package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"superaib/internal/api/middleware"
	"superaib/internal/api/response"
	"superaib/internal/models"
	"superaib/internal/services"
	"time"

	"github.com/gorilla/mux"
)

type NotificationHandler struct {
	service services.NotificationService
}

func NewNotificationHandler(s services.NotificationService) *NotificationHandler {
	return &NotificationHandler{service: s}
}

// 1. REGISTER TOKEN (Waxaa soo waca SDK-ga taleefanka)
// 🚀 1. REGISTER TOKEN (Kani waa kan pgAdmin True/False ka dhigaya)
func (h *NotificationHandler) RegisterToken(w http.ResponseWriter, r *http.Request) {
	projectID, _ := r.Context().Value(middleware.ProjectIDKey).(string)

	var t models.DeviceToken
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	t.ProjectID = projectID

	// 📝 LOG: Inoo xaqiiji xogta soo gashay
	fmt.Printf("📱 [Incoming] User: %s | Enabled: %v\n", t.UserID, t.Enabled)

	if err := h.service.RegisterDevice(r.Context(), &t); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "Status synced with pgAdmin ✅", nil)
}

// 2. BROADCAST / CREATE (Waxaa soo waca Dashboard-ka)
func (h *NotificationHandler) Broadcast(w http.ResponseWriter, r *http.Request) {
	projectID := mux.Vars(r)["project_id"]

	// 1. Akhriso JSON-ka qaab Map ah si uusan u crash-gareyn
	var rawData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&rawData); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON", err.Error())
		return
	}

	var note models.Notification

	// 🚀 XALKA MUCJISADA AH: Manual Time Parsing
	// 🚀 XALKA CUSUB: In fariintaadu dhacdo waqtiga Soomaaliya (GMT+3)
	if schedStr, ok := rawData["scheduled_at"].(string); ok && schedStr != "" {
		layout := "2006-01-02T15:04:05.000"

		// 1. Samee Timezone-ka Soomaaliya (GMT+3)
		// Haddii "Africa/Mogadishu" aysan shaqayn, isticmaal FixedZone
		location := time.FixedZone("EAT", 3*3600)

		// 2. Waqtiga ka yimid Dashboard-ka u parse-garee sidii Local (Somalia)
		t, err := time.ParseInLocation(layout, schedStr, location)
		if err == nil {
			// 3. Marka xal loo helo Local Time, u beddel UTC si loogu keydiyo DB-ga
			utcTime := t.UTC()
			note.ScheduledAt = &utcTime
			note.IsScheduled = true
			fmt.Printf("📅 Scheduled Somalia Time: %v | Saved as UTC: %v\n", t, utcTime)
		} else {
			// Backup haddii format-ku yahay RFC3339
			t2, _ := time.Parse(time.RFC3339, schedStr)
			utcTime2 := t2.UTC()
			note.ScheduledAt = &utcTime2
		}
	}
	// 2. Buuxi inta kale ee note-ka
	note.ProjectID = projectID
	if title, ok := rawData["title"].(string); ok {
		note.Title = title
	}
	if body, ok := rawData["body"].(string); ok {
		note.Body = body
	}
	if img, ok := rawData["image_url"].(string); ok {
		note.ImageURL = img
	}
	if isSched, ok := rawData["is_scheduled"].(bool); ok {
		note.IsScheduled = isSched
	}

	if note.UserID != nil && *note.UserID == "" {
		note.UserID = nil
	}

	// 3. Wac Service-ka
	if err := h.service.SendBroadcast(r.Context(), &note); err != nil {
		response.Error(w, http.StatusInternalServerError, "Broadcast failed", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "Notification processed ✅", nil)
}
func (h *NotificationHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	projectID := mux.Vars(r)["project_id"]
	history, err := h.service.GetHistory(r.Context(), projectID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed", err.Error())
		return
	}
	response.JSON(w, http.StatusOK, "Success", history)
}

// 3. UPDATE NOTIFICATION (Dashboard - Waxaa loo isticmaalaa fariimaha Scheduled-ka ah)
func (h *NotificationHandler) Update(w http.ResponseWriter, r *http.Request) {
	projectID := mux.Vars(r)["project_id"]

	var note models.Notification
	if err := json.NewDecoder(r.Body).Decode(&note); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON body", nil)
		return
	}
	note.ProjectID = projectID
	// note.ID = uuid.Parse(noteID) // Haddii loo baahdo in la xaqiijiyo ID-ga

	if err := h.service.UpdateNotification(r.Context(), &note); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to update notification", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "Notification updated successfully", nil)
}

// 4. DELETE NOTIFICATION (Dashboard)
func (h *NotificationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	projectID := mux.Vars(r)["project_id"]
	noteID := mux.Vars(r)["note_id"]

	if err := h.service.DeleteNotification(r.Context(), projectID, noteID); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to delete notification", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "Notification deleted successfully", nil)
}

// 5. UPDATE PUSH CONFIG (JSON Upload-ka Dashboard-ka)
func (h *NotificationHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	projectID := mux.Vars(r)["project_id"]
	if projectID == "" {
		response.Error(w, http.StatusBadRequest, "Project ID is required", nil)
		return
	}

	var req struct {
		ServiceJson string `json:"service_json"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON body", nil)
		return
	}

	if err := h.service.UpdatePushConfig(r.Context(), projectID, req.ServiceJson); err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to update config", err.Error())
		return
	}

	response.JSON(w, http.StatusOK, "Push Configuration Updated ✅", nil)
}

// 6. GET PUSH CONFIG (Dashboard)
func (h *NotificationHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	projectID := mux.Vars(r)["project_id"]
	if projectID == "" {
		response.Error(w, http.StatusBadRequest, "Project ID is required", nil)
		return
	}

	config, err := h.service.GetPushConfig(r.Context(), projectID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "Push configuration not found", nil)
		return
	}

	response.JSON(w, http.StatusOK, "Success", config)
}

// 7. GET HISTORY (Liiska fariimihii hore ee Dashboard-ka)

// GetRegistrationStatus: Wuxuu inoo sheegayaa haddii qofku u diwaangashan yahay pgAdmin
// 🚀 2. GET STATUS (Kani wuxuu App-ka u sheegayaa inuu ahaa ON ama OFF)
func (h *NotificationHandler) GetRegistrationStatus(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["user_id"]

	tokens, err := h.service.GetTokensByUserID(r.Context(), userID)
	if err != nil || len(tokens) == 0 {
		response.JSON(w, 200, "Not Found", map[string]interface{}{"enabled": false})
		return
	}

	response.JSON(w, 200, "Success", map[string]interface{}{
		"enabled": tokens[0].Enabled, // 👈 Kani waa kan muhiimka ah
		"token":   tokens[0].Token,
	})
}
