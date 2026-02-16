package repo

import (
	"context"
	"superaib/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RealtimeEventRepository interface {
	// --- CORE CRUD OPERATIONS ---
	Create(ctx context.Context, event *models.RealtimeEvent) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.RealtimeEvent, error)
	Update(ctx context.Context, event *models.RealtimeEvent) error
	Delete(ctx context.Context, id uuid.UUID) error

	// --- 🚀 SDK & ENTERPRISE FUNCTIONS ---
	// GetRecentEvents: History support (Soo qaado fariimihii u dambeeyay)
	GetRecentEvents(ctx context.Context, channelID uuid.UUID, limit int) ([]models.RealtimeEvent, error)

	// GetAllByChannel: Dhamaan fariimaha qol gaar ah
	GetAllByChannel(ctx context.Context, channelID uuid.UUID) ([]models.RealtimeEvent, error)

	// DeleteByChannel: Nadiifinta xogta marka channel la tirtiro
	DeleteByChannel(ctx context.Context, channelID uuid.UUID) error

	// CountByChannel: Analytics (Immisa fariin ayaa qolkan dhex martay?)
	CountByChannel(ctx context.Context, channelID uuid.UUID) (int64, error)
}

type gormRealtimeEventRepository struct {
	db *gorm.DB
}

func NewGormRealtimeEventRepository(db *gorm.DB) RealtimeEventRepository {
	return &gormRealtimeEventRepository{db: db}
}

// 1. Create: Kaydi fariin/event cusub
func (r *gormRealtimeEventRepository) Create(ctx context.Context, event *models.RealtimeEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

// 2. GetByID: Soo saar hal fariin oo gaar ah
func (r *gormRealtimeEventRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.RealtimeEvent, error) {
	var event models.RealtimeEvent
	err := r.db.WithContext(ctx).First(&event, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// 3. GetRecentEvents: HISTORY LOGIC (Si qofka soo gala uu fariimihii hore u arko)
func (r *gormRealtimeEventRepository) GetRecentEvents(ctx context.Context, channelID uuid.UUID, limit int) ([]models.RealtimeEvent, error) {
	var events []models.RealtimeEvent
	if limit <= 0 {
		limit = 50
	}

	err := r.db.WithContext(ctx).
		Where("channel_id = ?", channelID).
		Order("created_at DESC").
		Limit(limit).
		Find(&events).Error

	return events, err
}

// 4. GetAllByChannel: Soo saar dhamaan xogta qolka
func (r *gormRealtimeEventRepository) GetAllByChannel(ctx context.Context, channelID uuid.UUID) ([]models.RealtimeEvent, error) {
	var events []models.RealtimeEvent
	err := r.db.WithContext(ctx).
		Where("channel_id = ?", channelID).
		Order("created_at ASC").
		Find(&events).Error
	return events, err
}

// 5. Update: Bedel xogta fariin hore u jirtay
func (r *gormRealtimeEventRepository) Update(ctx context.Context, event *models.RealtimeEvent) error {
	return r.db.WithContext(ctx).Save(event).Error
}

// 6. Delete: Tirtir hal fariin
func (r *gormRealtimeEventRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.RealtimeEvent{}, "id = ?", id).Error
}

// 7. DeleteByChannel: Sifee dhamaan fariimaha qol marka la tirtiro
func (r *gormRealtimeEventRepository) DeleteByChannel(ctx context.Context, channelID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("channel_id = ?", channelID).
		Delete(&models.RealtimeEvent{}).Error
}

// 8. CountByChannel: Analytics support
func (r *gormRealtimeEventRepository) CountByChannel(ctx context.Context, channelID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.RealtimeEvent{}).
		Where("channel_id = ?", channelID).
		Count(&count).Error
	return count, err
}
