package services

import (
	"context"
	"encoding/json"
	"fmt"
	"superaib/internal/models"
	"superaib/internal/storage/repo"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type RealtimeService interface {
	// --- CHANNEL MANAGEMENT ---
	CreateChannel(ctx context.Context, channel *models.RealtimeChannel) error
	GetChannelsByProject(ctx context.Context, projectID string) ([]models.RealtimeChannel, error)
	GetChannelByID(ctx context.Context, channelID string) (*models.RealtimeChannel, error)
	UpdateChannel(ctx context.Context, channel *models.RealtimeChannel) error
	DeleteChannel(ctx context.Context, channelID string) error

	// --- EVENT MANAGEMENT ---
	CreateEvent(ctx context.Context, projectID string, event *models.RealtimeEvent) error
	GetEventsByChannel(ctx context.Context, channelID string) ([]models.RealtimeEvent, error)
	UpdateEvent(ctx context.Context, eventID string, payload datatypes.JSON) error
	DeleteEvent(ctx context.Context, eventID string) error

	// --- 🚀 SDK & REALTIME LOGIC ---
	JoinChannel(ctx context.Context, projectID, channelName, userID string) (*models.RealtimeChannel, error)
	LeaveChannel(ctx context.Context, projectID, channelName, userID string) error
	BroadcastToChannel(ctx context.Context, projectID, channelName string, eventType models.RealtimeEventType, payload map[string]interface{}, senderID string) (*models.RealtimeEvent, error)
	TrackPresence(ctx context.Context, projectID, channelName string) (int, error)
	GetRecentHistory(ctx context.Context, channelID string, limit int) ([]models.RealtimeEvent, error)
}

type realtimeService struct {
	channelRepo  repo.RealtimeChannelRepository
	eventRepo    repo.RealtimeEventRepository
	tracker      *AnalyticsTracker
	usageService ProjectUsageService
}

func NewRealtimeService(cr repo.RealtimeChannelRepository, er repo.RealtimeEventRepository, tracker *AnalyticsTracker, usage ProjectUsageService) RealtimeService {
	return &realtimeService{
		channelRepo:  cr,
		eventRepo:    er,
		tracker:      tracker,
		usageService: usage,
	}
}

// =========================================================================
// ✅ 1. CHANNEL IMPLEMENTATION
// =========================================================================

func (s *realtimeService) CreateChannel(ctx context.Context, channel *models.RealtimeChannel) error {
	err := s.channelRepo.Create(ctx, channel)
	if err == nil {
		s.tracker.TrackEvent(ctx, channel.ProjectID, models.AnalyticsTypeRealtimeChannels, "total_channels", 1)
		_ = s.usageService.UpdateUsage(ctx, channel.ProjectID, "api_calls", 1)
	}
	return err
}

func (s *realtimeService) GetChannelsByProject(ctx context.Context, projectID string) ([]models.RealtimeChannel, error) {
	return s.channelRepo.GetAllByProject(ctx, projectID)
}

func (s *realtimeService) GetChannelByID(ctx context.Context, channelID string) (*models.RealtimeChannel, error) {
	id, err := uuid.Parse(channelID)
	if err != nil {
		return nil, err
	}
	return s.channelRepo.GetByID(ctx, id)
}

func (s *realtimeService) UpdateChannel(ctx context.Context, c *models.RealtimeChannel) error {
	return s.channelRepo.Update(ctx, c)
}

func (s *realtimeService) DeleteChannel(ctx context.Context, channelID string) error {
	id, err := uuid.Parse(channelID)
	if err != nil {
		return err
	}
	return s.channelRepo.Delete(ctx, id)
}

// =========================================================================
// ✅ 2. EVENT IMPLEMENTATION
// =========================================================================

func (s *realtimeService) CreateEvent(ctx context.Context, projectID string, event *models.RealtimeEvent) error {
	err := s.eventRepo.Create(ctx, event)
	if err == nil {
		s.tracker.TrackEvent(ctx, projectID, models.AnalyticsTypeRealtimeEvents, "total_messages", 1)
		_ = s.usageService.UpdateUsage(ctx, projectID, "api_calls", 1)

		// Update Channel timestamp
		channel, _ := s.channelRepo.GetByID(ctx, event.ChannelID)
		if channel != nil {
			now := time.Now()
			channel.LastMessageAt = &now
			_ = s.channelRepo.Update(ctx, channel)
		}
	}
	return err
}

func (s *realtimeService) GetEventsByChannel(ctx context.Context, channelID string) ([]models.RealtimeEvent, error) {
	id, err := uuid.Parse(channelID)
	if err != nil {
		return nil, err
	}
	return s.eventRepo.GetAllByChannel(ctx, id)
}

func (s *realtimeService) UpdateEvent(ctx context.Context, eventID string, payload datatypes.JSON) error {
	id, err := uuid.Parse(eventID)
	if err != nil {
		return err
	}
	event, err := s.eventRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	event.Payload = payload
	return s.eventRepo.Update(ctx, event)
}

func (s *realtimeService) DeleteEvent(ctx context.Context, eventID string) error {
	id, err := uuid.Parse(eventID)
	if err != nil {
		return err
	}
	return s.eventRepo.Delete(ctx, id)
}

// =========================================================================
// ✅ 3. SDK & REALTIME LOGIC
// =========================================================================
func (s *realtimeService) JoinChannel(ctx context.Context, projectID, channelName, userID string) (*models.RealtimeChannel, error) {
	// 🚀 1. Marka hore si degan u raadi qolka haddii uu jiro
	channel, err := s.channelRepo.GetByName(ctx, projectID, channelName)

	// 🚀 2. Haddii la waayo (RecordNotFound), markaas kaliya abuuro mid cusub
	if err != nil {
		fmt.Printf("📡 SDK: Channel [%s] not found, creating new one...\n", channelName)
		newChannel := &models.RealtimeChannel{
			ProjectID:        projectID,
			Name:             channelName,
			SubscriptionType: models.SubscriptionPublic,
			RetentionPolicy:  models.RetentionPersistent, // Kani ayaa pgAdmin ku xaraynaya
		}
		if err := s.channelRepo.Create(ctx, newChannel); err != nil {
			return nil, err
		}
		return newChannel, nil
	}

	// 🚀 3. Haddii uu hore u jiray, isaga si toos ah u soo celi (Ha isku dayin inaa INSERT gareyso)
	return channel, nil
}
func (s *realtimeService) LeaveChannel(ctx context.Context, projectID, channelName, userID string) error {
	channel, err := s.channelRepo.GetByName(ctx, projectID, channelName)
	if err != nil {
		return err
	}
	return s.channelRepo.UpdateConnectedCount(ctx, channel.ID, -1)
}

func (s *realtimeService) BroadcastToChannel(ctx context.Context, projectID, channelName string, eventType models.RealtimeEventType, payload map[string]interface{}, senderID string) (*models.RealtimeEvent, error) {
	channel, err := s.channelRepo.GetByName(ctx, projectID, channelName)
	if err != nil {
		return nil, err
	}

	pBytes, _ := json.Marshal(payload)
	event := &models.RealtimeEvent{
		ChannelID: channel.ID,
		EventType: eventType,
		Payload:   datatypes.JSON(pBytes),
		CreatedAt: time.Now(),
	}
	if senderID != "" {
		event.SenderID = &senderID
	}

	if channel.RetentionPolicy == models.RetentionPersistent {
		_ = s.CreateEvent(ctx, projectID, event)
	}
	return event, nil
}

func (s *realtimeService) TrackPresence(ctx context.Context, projectID, channelName string) (int, error) {
	channel, err := s.channelRepo.GetByName(ctx, projectID, channelName)
	if err != nil {
		return 0, err
	}
	return channel.ConnectedClients, nil
}

func (s *realtimeService) GetRecentHistory(ctx context.Context, channelID string, limit int) ([]models.RealtimeEvent, error) {
	id, err := uuid.Parse(channelID)
	if err != nil {
		return nil, err
	}
	return s.eventRepo.GetRecentEvents(ctx, id, limit)
}
