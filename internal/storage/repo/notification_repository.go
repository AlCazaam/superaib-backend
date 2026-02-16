package repo

import (
	"context"
	"errors"
	"superaib/internal/models"
	"time"

	"gorm.io/gorm"
)

type NotificationRepository interface {
	// --- Device Token Management ---
	SaveToken(ctx context.Context, token *models.DeviceToken) error
	GetTokensByProject(ctx context.Context, projectID string) ([]models.DeviceToken, error)
	GetActiveTokensByProject(ctx context.Context, projectID string) ([]models.DeviceToken, error) // 🚀 CUSUB
	GetTokensByUserID(ctx context.Context, userID string) ([]models.DeviceToken, error)

	// --- Notification CRUD (pgAdmin History) ---
	CreateLog(ctx context.Context, note *models.Notification) error
	Update(ctx context.Context, note *models.Notification) error
	Delete(ctx context.Context, projectID string, noteID string) error
	GetByID(ctx context.Context, projectID string, noteID string) (*models.Notification, error)
	GetHistory(ctx context.Context, projectID string) ([]models.Notification, error)
	// 🚀 SCHEDULER ENGINE: Kani wuxuu soo saaraa fariimaha u dhimman in la diro
	GetPendingNotifications(ctx context.Context, now time.Time) ([]models.Notification, error)
}

type gormNotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &gormNotificationRepository{db: db}
}

// 🚀 1. SAVE TOKEN (UPSERT LOGIC)
func (r *gormNotificationRepository) SaveToken(ctx context.Context, t *models.DeviceToken) error {
	var existing models.DeviceToken

	// 🚀 1. Raadi haddii uu User-kan hore u jiray
	err := r.db.WithContext(ctx).Where("user_id = ?", t.UserID).First(&existing).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 🚀 2. MA JIRO: Markaas kaliya abuuro (Create)
			return r.db.WithContext(ctx).Create(t).Error
		}
		return err
	}

	// 🚀 3. WUU JIRAA: Cusboonaysii (Update)
	// Waxaan isticmaalaynaa Map si aan u hubino in 'false' la kaydiyo (GORM fix)
	return r.db.WithContext(ctx).Model(&existing).Updates(map[string]interface{}{
		"token":      t.Token,
		"platform":   t.Platform,
		"enabled":    t.Enabled, // 👈 Hadda 'false' waa la kaydinayaa!
		"project_id": t.ProjectID,
	}).Error
}
func (r *gormNotificationRepository) GetPendingNotifications(ctx context.Context, now time.Time) ([]models.Notification, error) {
	var notes []models.Notification
	// Hubi in 'now' loo dirayo sidii UTC saafi ah
	err := r.db.WithContext(ctx).
		Where("status = ? AND is_scheduled = ? AND scheduled_at <= ?", "pending", true, now.UTC()).
		Find(&notes).Error
	return notes, err
}

// 🚀 2. GET ACTIVE TOKENS (Filtering for Broadcast)
func (r *gormNotificationRepository) GetActiveTokensByProject(ctx context.Context, projectID string) ([]models.DeviceToken, error) {
	var tokens []models.DeviceToken
	// Kaliya soo qaado kuwa 'enabled = true' u tahay
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND enabled = ?", projectID, true).
		Find(&tokens).Error
	return tokens, err
}

func (r *gormNotificationRepository) GetTokensByProject(ctx context.Context, projectID string) ([]models.DeviceToken, error) {
	var tokens []models.DeviceToken
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Find(&tokens).Error
	return tokens, err
}

func (r *gormNotificationRepository) GetTokensByUserID(ctx context.Context, userID string) ([]models.DeviceToken, error) {
	var tokens []models.DeviceToken
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&tokens).Error
	return tokens, err
}

func (r *gormNotificationRepository) CreateLog(ctx context.Context, n *models.Notification) error {
	return r.db.WithContext(ctx).Create(n).Error
}

func (r *gormNotificationRepository) Update(ctx context.Context, n *models.Notification) error {
	return r.db.WithContext(ctx).Save(n).Error
}

func (r *gormNotificationRepository) Delete(ctx context.Context, projectID string, noteID string) error {
	return r.db.WithContext(ctx).Where("project_id = ? AND id = ?", projectID, noteID).Delete(&models.Notification{}).Error
}

func (r *gormNotificationRepository) GetByID(ctx context.Context, projectID string, noteID string) (*models.Notification, error) {
	var note models.Notification
	err := r.db.WithContext(ctx).Where("project_id = ? AND id = ?", projectID, noteID).First(&note).Error
	return &note, err
}

func (r *gormNotificationRepository) GetHistory(ctx context.Context, projectID string) ([]models.Notification, error) {
	var notes []models.Notification
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Order("created_at DESC").Find(&notes).Error
	return notes, err
}
