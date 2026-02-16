package repo

import (
	"context"
	"superaib/internal/models"
	"time"

	"gorm.io/gorm"
)

type AnalyticsRepository interface {
	Create(ctx context.Context, tx *gorm.DB, record *models.Analytics) error
	GetByProjectForPeriod(ctx context.Context, projectID string, periodStart time.Time) ([]models.Analytics, error)
	IncrementMetric(ctx context.Context, projectID string, aType models.AnalyticsType, periodStart time.Time, key string, value float64) error
}

type GormAnalyticsRepository struct {
	db *gorm.DB
}

func NewGormAnalyticsRepository(db *gorm.DB) AnalyticsRepository {
	return &GormAnalyticsRepository{db: db}
}

func (r *GormAnalyticsRepository) Create(ctx context.Context, tx *gorm.DB, record *models.Analytics) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.WithContext(ctx).Create(record).Error
}

func (r *GormAnalyticsRepository) GetByProjectForPeriod(ctx context.Context, projectID string, periodStart time.Time) ([]models.Analytics, error) {
	var records []models.Analytics
	err := r.db.WithContext(ctx).Where("project_id = ? AND period_start = ?", projectID, periodStart).Find(&records).Error
	return records, err
}

func (r *GormAnalyticsRepository) IncrementMetric(ctx context.Context, projectID string, aType models.AnalyticsType, periodStart time.Time, key string, value float64) error {
	query := `
        INSERT INTO analytics (id, project_id, type, period_start, metrics, created_at, updated_at)
        VALUES (gen_random_uuid(), ?, ?, ?, jsonb_build_object(?::text, ?::numeric), NOW(), NOW())
        ON CONFLICT (project_id, type, period_start) 
        DO UPDATE SET 
            metrics = jsonb_set(
                analytics.metrics, 
                '{` + key + `}', 
                (COALESCE(analytics.metrics->>'` + key + `', '0')::numeric + ?::numeric)::text::jsonb
            ),
            updated_at = NOW();
    `
	return r.db.WithContext(ctx).Exec(query, projectID, aType, periodStart, key, value, value).Error
}
