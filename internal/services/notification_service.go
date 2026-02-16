package services

import (
	"context"
	"fmt"
	"superaib/internal/models"
	"superaib/internal/storage/repo"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

// Broadcaster Interface
type Broadcaster interface {
	BroadcastToProject(projectID string, eventType string, payload interface{})
	BroadcastToChannel(pID, channel, event string, payload interface{}, senderID string)
}

type NotificationService interface {
	// Device Registration
	RegisterDevice(ctx context.Context, token *models.DeviceToken) error
	GetTokensByUserID(ctx context.Context, userID string) ([]models.DeviceToken, error) // 🚀 Hadda waa lagu daray!

	// Notification Management (CRUD)
	SendBroadcast(ctx context.Context, note *models.Notification) error
	UpdateNotification(ctx context.Context, note *models.Notification) error
	DeleteNotification(ctx context.Context, projectID string, noteID string) error
	GetNotification(ctx context.Context, projectID string, noteID string) (*models.Notification, error)
	GetHistory(ctx context.Context, projectID string) ([]models.Notification, error)
	StartScheduler(ctx context.Context) // 🚀 Mishiinka kicinaya

	// Push Configuration
	UpdatePushConfig(ctx context.Context, projectID string, jsonStr string) error
	GetPushConfig(ctx context.Context, projectID string) (*models.ProjectPushConfig, error)
}

type notificationService struct {
	repo          repo.NotificationRepository
	configRepo    repo.PushConfigRepository
	tracker       *AnalyticsTracker
	usageService  ProjectUsageService
	rtBroadcaster Broadcaster
}

func NewNotificationService(r repo.NotificationRepository, cr repo.PushConfigRepository, t *AnalyticsTracker, u ProjectUsageService, b Broadcaster) NotificationService {
	return &notificationService{
		repo:          r,
		configRepo:    cr,
		tracker:       t,
		usageService:  u,
		rtBroadcaster: b,
	}
}

// 1. REGISTER DEVICE TOKEN
func (s *notificationService) RegisterDevice(ctx context.Context, t *models.DeviceToken) error {
	return s.repo.SaveToken(ctx, t)
}

// 🚀 2. GET TOKENS BY USER ID (Kani waa kii maanta loo baahnaa!)
func (s *notificationService) GetTokensByUserID(ctx context.Context, userID string) ([]models.DeviceToken, error) {
	return s.repo.GetTokensByUserID(ctx, userID)
}

// 3. SEND BROADCAST (Smart Filtering)
func (s *notificationService) SendBroadcast(ctx context.Context, note *models.Notification) error {
	tokens, _ := s.repo.GetActiveTokensByProject(ctx, note.ProjectID)
	note.SentCount = len(tokens)

	// 🚀 XALKA 2: Haddii ay tahay Scheduled, kaliya DB-ga ku kaydi (Status: pending)
	// Waxaan siinaynaa margin 5 ilbiriqsi ah.
	if note.IsScheduled && note.ScheduledAt != nil {
		if note.ScheduledAt.After(time.Now().Add(5 * time.Second)) {
			note.Status = "pending"
			fmt.Printf("📅 [Scheduler] Note '%s' queued for: %v\n", note.Title, note.ScheduledAt)
			return s.repo.CreateLog(ctx, note)
		}
	}

	// Immediate Send (Haddii waqtigu hadda yahay)
	note.Status = "sent"
	if err := s.repo.CreateLog(ctx, note); err != nil {
		return err
	}

	payload := map[string]interface{}{
		"id": note.ID, "title": note.Title, "body": note.Body, "timestamp": time.Now(),
	}
	s.rtBroadcaster.BroadcastToProject(note.ProjectID, "PUSH_NOTIFICATION", payload)
	go s.sendToFCM(note)

	return nil
}

// 🚀 XALKA 2: MISHIINKA DHAGAYSTIGA WAQTIGA (The Scheduler)
func (s *notificationService) StartScheduler(ctx context.Context) {
	// Wuxuu isbaaraa 30-kii ilbiriqsiba fariimaha dhimman
	ticker := time.NewTicker(30 * time.Second)

	go func() {
		fmt.Println("⏰ [SYSTEM] Notification Scheduler is running (UTC Mode)...")
		for range ticker.C {
			// 🚀 MUHIIM: Had iyo jeer isticmaal UTC isbarbardhigga
			now := time.Now().UTC()

			pendingNotes, err := s.repo.GetPendingNotifications(ctx, now)
			if err != nil {
				continue
			}

			if len(pendingNotes) > 0 {
				fmt.Printf("🎯 [TRIGGER] Found %d notifications to send.\n", len(pendingNotes))
			}

			for _, note := range pendingNotes {
				fmt.Printf("🚀 Sending Scheduled Note: %s (ID: %s)\n", note.Title, note.ID)

				// 1. Live Broadcast (Websocket/Realtime)
				payload := map[string]interface{}{
					"id":        note.ID,
					"title":     note.Title,
					"body":      note.Body,
					"image_url": note.ImageURL,
					"timestamp": now,
				}
				s.rtBroadcaster.BroadcastToProject(note.ProjectID, "PUSH_NOTIFICATION", payload)

				// 2. U dir Firebase (FCM) - haddii uu u jiro function-kaas
				go s.sendToFCM(&note)

				// 3. 🚀 CUSBOONAYSIIN STATUS-KA (Tani waa tallaabada ugu muhiimsan)
				// Waxaan u beddelaynaa 'sent' si uusan mishiinku mar labaad u soo qaadin
				note.Status = "sent"
				if err := s.repo.Update(ctx, &note); err != nil {
					fmt.Printf("❌ Error updating status for note %s: %v\n", note.ID, err)
				} else {
					fmt.Printf("✅ [SUCCESS] Note '%s' status updated to SENT in DB.\n", note.Title)
				}
			}
		}
	}()
}

// 4. UPDATE NOTIFICATION
func (s *notificationService) UpdateNotification(ctx context.Context, note *models.Notification) error {
	return s.repo.Update(ctx, note)
}

// 5. DELETE NOTIFICATION
func (s *notificationService) DeleteNotification(ctx context.Context, projectID string, noteID string) error {
	return s.repo.Delete(ctx, projectID, noteID)
}

// 6. GET SINGLE NOTIFICATION
func (s *notificationService) GetNotification(ctx context.Context, projectID string, noteID string) (*models.Notification, error) {
	return s.repo.GetByID(ctx, projectID, noteID)
}

// 7. GET HISTORY
func (s *notificationService) GetHistory(ctx context.Context, projectID string) ([]models.Notification, error) {
	return s.repo.GetHistory(ctx, projectID)
}

// 8. UPDATE PUSH CONFIG
func (s *notificationService) UpdatePushConfig(ctx context.Context, projectID string, jsonStr string) error {
	config := &models.ProjectPushConfig{
		ProjectID:   projectID,
		ServiceJson: jsonStr,
	}
	return s.configRepo.SaveConfig(ctx, config)
}

// 9. GET PUSH CONFIG
func (s *notificationService) GetPushConfig(ctx context.Context, projectID string) (*models.ProjectPushConfig, error) {
	return s.configRepo.GetConfig(ctx, projectID)
}

// 🚀 INTERNAL: FCM BRIDGE
func (s *notificationService) sendToFCM(note *models.Notification) {
	ctx := context.Background()
	cfg, err := s.configRepo.GetConfig(ctx, note.ProjectID)
	if err != nil || cfg.ServiceJson == "" {
		return
	}

	opt := option.WithCredentialsJSON([]byte(cfg.ServiceJson))
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return
	}

	// Kaliya kuwa Enabled ah
	tokens, _ := s.repo.GetActiveTokensByProject(ctx, note.ProjectID)
	var regTokens []string
	for _, t := range tokens {
		regTokens = append(regTokens, t.Token)
	}

	if len(regTokens) > 0 {
		msg := &messaging.MulticastMessage{
			Tokens: regTokens,
			Notification: &messaging.Notification{
				Title: note.Title,
				Body:  note.Body,
			},
		}
		_, _ = client.SendEachForMulticast(ctx, msg)
	}
}
