package repo

import (
	"context"
	"superaib/internal/models"

	"gorm.io/gorm"
)

type PushConfigRepository interface {
	SaveConfig(ctx context.Context, config *models.ProjectPushConfig) error
	GetConfig(ctx context.Context, projectID string) (*models.ProjectPushConfig, error)
}

type gormPushConfigRepository struct {
	db *gorm.DB
}

func NewPushConfigRepository(db *gorm.DB) PushConfigRepository {
	return &gormPushConfigRepository{db: db}
}

func (r *gormPushConfigRepository) SaveConfig(ctx context.Context, c *models.ProjectPushConfig) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *gormPushConfigRepository) GetConfig(ctx context.Context, projectID string) (*models.ProjectPushConfig, error) {
	var config models.ProjectPushConfig
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).First(&config).Error
	return &config, err
}
