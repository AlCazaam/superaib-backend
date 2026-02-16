package repo

import (
	"context"
	"superaib/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RealtimeChannelRepository interface {
	// CRUD Operations
	Create(ctx context.Context, channel *models.RealtimeChannel) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.RealtimeChannel, error)
	GetAllByProject(ctx context.Context, projectID string) ([]models.RealtimeChannel, error)
	Update(ctx context.Context, channel *models.RealtimeChannel) error
	Delete(ctx context.Context, id uuid.UUID) error

	// 🚀 SDK & ENTERPRISE FUNCTIONS
	// GetByName: Waxaa loo isticmaalaa marka SDK-gu uu leeyahay .channel('chat_room')
	GetByName(ctx context.Context, projectID string, name string) (*models.RealtimeChannel, error)

	// UpdateConnectedCount: Kordhi ama ka dhim tirada dadka Online-ka ah (Atomic Operation)
	UpdateConnectedCount(ctx context.Context, id uuid.UUID, change int) error
}

type gormRealtimeChannelRepository struct {
	db *gorm.DB
}

func NewGormRealtimeChannelRepository(db *gorm.DB) RealtimeChannelRepository {
	return &gormRealtimeChannelRepository{db: db}
}

// 1. Create: Kaydi channel cusub
func (r *gormRealtimeChannelRepository) Create(ctx context.Context, channel *models.RealtimeChannel) error {
	return r.db.WithContext(ctx).Create(channel).Error
}

// 2. GetByID: Soo saar xogta hal channel isagoo UUID ah
func (r *gormRealtimeChannelRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.RealtimeChannel, error) {
	var channel models.RealtimeChannel
	if err := r.db.WithContext(ctx).First(&channel, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &channel, nil
}

// 3. GetByName: 🚀 MUHIIM! Kani waa kan SDK-gu isticmaalo mar kasta
func (r *gormRealtimeChannelRepository) GetByName(ctx context.Context, projectID string, name string) (*models.RealtimeChannel, error) {
	var channel models.RealtimeChannel
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND name = ?", projectID, name).
		First(&channel).Error
	if err != nil {
		return nil, err // Kani waa muhiim si Service-ku uu u abuuro mid cusub
	}
	return &channel, nil
}

// 4. GetAllByProject: Soo saar dhamaan Channels-ka uu mashruucu leeyahay
func (r *gormRealtimeChannelRepository) GetAllByProject(ctx context.Context, projectID string) ([]models.RealtimeChannel, error) {
	var channels []models.RealtimeChannel
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("created_at DESC").
		Find(&channels).Error
	return channels, err
}

// 5. Update: Bedel xogta channel-ka (Metadata, Privacy, iwm)
func (r *gormRealtimeChannelRepository) Update(ctx context.Context, channel *models.RealtimeChannel) error {
	return r.db.WithContext(ctx).Save(channel).Error
}

// 6. UpdateConnectedCount: 🚀 ATOMIC PRESENCE LOGIC
// Change wuxuu noqon karaa +1 (qof ayaa soo galay) ama -1 (qof ayaa ka baxay)
func (r *gormRealtimeChannelRepository) UpdateConnectedCount(ctx context.Context, id uuid.UUID, change int) error {
	// Waxaan isticmaalaynaa gorm.Expr si looga hortago "Race Conditions" (Atomic Update)
	return r.db.WithContext(ctx).Model(&models.RealtimeChannel{}).
		Where("id = ?", id).
		UpdateColumn("connected_clients", gorm.Expr("connected_clients + ?", change)).Error
}

// 7. Delete: Tirtir channel-ka gabi ahaanba
func (r *gormRealtimeChannelRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.RealtimeChannel{}, "id = ?", id).Error
}
