package repo

import (
	"context"
	"superaib/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OtpTrackerRepository interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*models.OtpUserTracker, error)
	Upsert(ctx context.Context, tracker *models.OtpUserTracker) error
	ResetTracker(ctx context.Context, userID uuid.UUID) error
}

type gormOtpTrackerRepo struct {
	db *gorm.DB
}

func NewOtpTrackerRepository(db *gorm.DB) OtpTrackerRepository {
	return &gormOtpTrackerRepo{db: db}
}

func (r *gormOtpTrackerRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.OtpUserTracker, error) {
	var tracker models.OtpUserTracker
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&tracker).Error
	return &tracker, err
}

func (r *gormOtpTrackerRepo) Upsert(ctx context.Context, tracker *models.OtpUserTracker) error {
	return r.db.WithContext(ctx).Save(tracker).Error
}

func (r *gormOtpTrackerRepo) ResetTracker(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.OtpUserTracker{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"request_count":    0,
			"window_starts_at": time.Now(),
			"lockout_until":    nil,
		}).Error
}
