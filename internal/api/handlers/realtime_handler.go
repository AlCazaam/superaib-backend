package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"superaib/internal/api/middleware"
	"superaib/internal/api/response"
	"superaib/internal/models"
	"superaib/internal/services"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"gorm.io/datatypes"
)

type Client struct {
	Conn      *websocket.Conn
	ProjectID string
	UserID    string
	Channels  map[string]bool // Qolalka uu ku jiro qofkan
	mu        sync.Mutex
	Send      chan []byte
}

type RealtimeHandler struct {
	service     services.RealtimeService
	projects    map[string]map[*Client]bool
	projectsMux sync.RWMutex
	upgrader    websocket.Upgrader
}

func NewRealtimeHandler(s services.RealtimeService) *RealtimeHandler {
	return &RealtimeHandler{
		service:  s,
		projects: make(map[string]map[*Client]bool),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
			// 🚀 XALKA SIMULATOR-KA: Dami wax kasta oo compression ah
			EnableCompression: false,
		},
	}
}

func (h *RealtimeHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Tirtir extensions-ka simulator-ka uu soo diro
	r.Header.Del("Sec-WebSocket-Extensions")

	vars := mux.Vars(r)
	projectID := vars["project_id"]
	if projectID == "" {
		projectID = r.URL.Query().Get("project_id")
	}
	userID := r.URL.Query().Get("user_id")

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := &Client{
		Conn:      conn,
		ProjectID: projectID,
		UserID:    userID,
		Channels:  make(map[string]bool),
		Send:      make(chan []byte, 256),
	}

	h.registerClient(client)
	fmt.Printf("\n🚀 [Realtime] Connected: User [%s]\n", userID)

	go h.writePump(client)
	h.readPump(client)
}

func (h *RealtimeHandler) readPump(c *Client) {
	defer func() {
		h.unregisterClient(c)
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		var msg struct {
			Action  string                 `json:"action"`
			Channel string                 `json:"channel"`
			Event   string                 `json:"event"`
			Payload map[string]interface{} `json:"payload"`
		}

		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		fmt.Printf("📩 [Realtime] Action: %s | Channel: %s\n", msg.Action, msg.Channel)

		switch msg.Action {
		case "SUBSCRIBE":
			c.mu.Lock()
			c.Channels[msg.Channel] = true
			c.mu.Unlock()

			// 💾 Database sync (SAVE TO PGADMIN)
			go h.service.JoinChannel(context.Background(), c.ProjectID, msg.Channel, c.UserID)

		case "BROADCAST":
			// 1. LIVE SEND (U dir qof kasta oo online ah)
			h.BroadcastToChannel(c.ProjectID, msg.Channel, msg.Event, msg.Payload, c.UserID)

			// 2. 💾 DATABASE SAVE (Inuu pgAdmin ka soo muuqdo)
			go func(p, ch, ev string, py map[string]interface{}) {
				channel, _ := h.service.JoinChannel(context.Background(), p, ch, c.UserID)
				if channel != nil {
					event := &models.RealtimeEvent{
						ChannelID: channel.ID,
						EventType: models.RealtimeEventType(ev),
						Payload:   h.mapToJSON(py),
					}
					if c.UserID != "" {
						uID := c.UserID
						event.SenderID = &uID
					}
					h.service.CreateEvent(context.Background(), p, event)
				}
			}(c.ProjectID, msg.Channel, msg.Event, msg.Payload)
		}
	}
}

func (h *RealtimeHandler) writePump(c *Client) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.Conn.WriteMessage(websocket.TextMessage, message)
		case <-ticker.C:
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *RealtimeHandler) BroadcastToChannel(pID, channel, event string, payload interface{}, senderID string) {
	h.projectsMux.RLock()
	clients := h.projects[pID]
	h.projectsMux.RUnlock()

	data, _ := json.Marshal(map[string]interface{}{
		"channel":    channel,
		"event_type": event,
		"payload":    payload,
		"sender_id":  senderID,
		"timestamp":  time.Now(),
	})

	for client := range clients {
		client.mu.Lock()
		in := client.Channels[channel]
		client.mu.Unlock()
		if in {
			select {
			case client.Send <- data:
			default:
			}
		}
	}
}

// Helpers
func (h *RealtimeHandler) registerClient(c *Client) {
	h.projectsMux.Lock()
	if h.projects[c.ProjectID] == nil {
		h.projects[c.ProjectID] = make(map[*Client]bool)
	}
	h.projects[c.ProjectID][c] = true
	h.projectsMux.Unlock()
}
func (h *RealtimeHandler) unregisterClient(c *Client) {
	h.projectsMux.Lock()
	if clients, ok := h.projects[c.ProjectID]; ok {
		delete(clients, c)
	}
	h.projectsMux.Unlock()
}
func (h *RealtimeHandler) mapToJSON(m map[string]interface{}) datatypes.JSON {
	b, _ := json.Marshal(m)
	return datatypes.JSON(b)
}
func (h *RealtimeHandler) getPID(r *http.Request) string {
	if uid, ok := r.Context().Value(middleware.ProjectIDKey).(string); ok && uid != "" {
		return uid
	}
	return mux.Vars(r)["project_id"]
}

// DASHBOARD CRUD... (Dhamaan functions-ka kale halkooda ha u joogaan)
func (h *RealtimeHandler) GetChannels(w http.ResponseWriter, r *http.Request) {
	colls, _ := h.service.GetChannelsByProject(r.Context(), h.getPID(r))
	response.JSON(w, 200, "Success", colls)
}

func (h *RealtimeHandler) UpdateChannel(w http.ResponseWriter, r *http.Request) {
	var c models.RealtimeChannel
	json.NewDecoder(r.Body).Decode(&c)
	c.ID, _ = uuid.Parse(mux.Vars(r)["id"])
	h.service.UpdateChannel(r.Context(), &c)
	response.JSON(w, 200, "Updated", c)
}
func (h *RealtimeHandler) DeleteChannel(w http.ResponseWriter, r *http.Request) {
	h.service.DeleteChannel(r.Context(), mux.Vars(r)["id"])
	response.JSON(w, 200, "Deleted", nil)
}
func (h *RealtimeHandler) GetEvents(w http.ResponseWriter, r *http.Request) {
	events, _ := h.service.GetEventsByChannel(r.Context(), mux.Vars(r)["channel_id"])
	response.JSON(w, 200, "Success", events)
}
func (h *RealtimeHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Payload datatypes.JSON `json:"payload"`
	}
	json.NewDecoder(r.Body).Decode(&b)
	h.service.UpdateEvent(r.Context(), mux.Vars(r)["event_id"], b.Payload)
	response.JSON(w, 200, "Updated", nil)
}
func (h *RealtimeHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	h.service.DeleteEvent(r.Context(), mux.Vars(r)["event_id"])
	response.JSON(w, 200, "Deleted", nil)
}
func (h *RealtimeHandler) BroadcastToProject(pID, event string, payload interface{}) {
	h.projectsMux.RLock()
	clients := h.projects[pID]
	h.projectsMux.RUnlock()
	data, _ := json.Marshal(map[string]interface{}{"event_type": event, "payload": payload, "timestamp": time.Now()})
	for client := range clients {
		select {
		case client.Send <- data:
		default:
		}
	}
}

// 📺 1. CREATE CHANNEL (Kani waa kan SDK-gu waco marka uu bilaabanayo)
func (h *RealtimeHandler) CreateChannel(w http.ResponseWriter, r *http.Request) {
	var ch models.RealtimeChannel
	if err := json.NewDecoder(r.Body).Decode(&ch); err != nil {
		response.Error(w, 400, "Invalid request")
		return
	}
	pID := h.getPID(r)

	// 🚀 XALKA: Halkii aan h.channelRepo.Create wici lahayn, waxaan wacaynaa JoinChannel
	// JoinChannel ayaa isagu aqoon u leh inuu "Get or Create" sameeyo
	finalChannel, err := h.service.JoinChannel(r.Context(), pID, ch.Name, "")
	if err != nil {
		response.Error(w, 500, "Failed to initialize channel", err.Error())
		return
	}

	response.JSON(w, 201, "Success", finalChannel)
}

// 📩 2. CREATE EVENT (Kani waa kan fariimaha pgAdmin ku ridaya)
func (h *RealtimeHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	pID := h.getPID(r)
	channelIDStr := mux.Vars(r)["channel_id"]

	// Hubi ID-ga uu SDK-gu soo diray
	chanID, err := uuid.Parse(channelIDStr)
	if err != nil {
		response.Error(w, 400, "Invalid Channel UUID")
		return
	}

	var event models.RealtimeEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		response.Error(w, 400, "Invalid JSON payload")
		return
	}
	event.ChannelID = chanID // 👈 Ku xir ID-ga saxda ah

	// 💾 SAVE TO pgAdmin
	if err := h.service.CreateEvent(r.Context(), pID, &event); err != nil {
		fmt.Printf("❌ [DB ERROR]: Event save failed: %v\n", err)
		response.Error(w, 500, "Failed to save event")
		return
	}

	fmt.Printf("✅ [DB SUCCESS]: Event [%s] saved to pgAdmin\n", event.EventType)

	// 📢 LIVE BROADCAST
	channel, _ := h.service.GetChannelByID(r.Context(), channelIDStr)
	if channel != nil {
		h.BroadcastToChannel(pID, channel.Name, string(event.EventType), event.Payload, "sdk_user")
	}

	response.JSON(w, 201, "Created & Saved", event)
}
