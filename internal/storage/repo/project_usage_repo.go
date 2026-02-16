package repo

import (
	"context"
	"fmt"
	"superaib/internal/models"

	"gorm.io/gorm"
)

type ProjectUsageRepository interface {
	Create(ctx context.Context, tx *gorm.DB, usage *models.ProjectUsage) error
	GetByProjectID(ctx context.Context, projectUUID string) (*models.ProjectUsage, error)
	Update(ctx context.Context, usage *models.ProjectUsage) error
	IncrementField(ctx context.Context, projectUUID string, field string, value interface{}) error
}

type gormProjectUsageRepository struct {
	db *gorm.DB
}

func NewGormProjectUsageRepository(db *gorm.DB) ProjectUsageRepository {
	return &gormProjectUsageRepository{db: db}
}

func (r *gormProjectUsageRepository) Create(ctx context.Context, tx *gorm.DB, usage *models.ProjectUsage) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.WithContext(ctx).Create(usage).Error
}

func (r *gormProjectUsageRepository) GetByProjectID(ctx context.Context, projectUUID string) (*models.ProjectUsage, error) {
	var usage models.ProjectUsage
	err := r.db.WithContext(ctx).Where("project_id = ?", projectUUID).First(&usage).Error
	return &usage, err
}

func (r *gormProjectUsageRepository) Update(ctx context.Context, usage *models.ProjectUsage) error {
	return r.db.WithContext(ctx).Save(usage).Error
}

func (r *gormProjectUsageRepository) IncrementField(ctx context.Context, projectUUID string, field string, value interface{}) error {
	// field tusaale: "realtime_events_count"
	query := fmt.Sprintf("UPDATE project_usages SET %s = %s + ?, updated_at = NOW() WHERE project_id = ?", field, field)
	return r.db.WithContext(ctx).Exec(query, value, projectUUID).Error
}
