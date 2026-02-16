package services

import (
	"context"
	"superaib/internal/core/logger"
	"superaib/internal/models"
	"superaib/internal/storage/repo"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AnalyticsTracker struct {
	repo repo.AnalyticsRepository
}

func (t *AnalyticsTracker) TrackEvent(ctx context.Context, projectID string, aType models.AnalyticsType, key string, value float64) {
	go func() {
		now := time.Now()
		periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		err := t.repo.IncrementMetric(context.Background(), projectID, aType, periodStart, key, value)
		if err != nil {
			logger.Log.Errorf("❌ Failed to track analytics: %v", err)
		}
	}()
}

type AnalyticsService interface {
	CreateInitialAnalyticsForProject(ctx context.Context, tx *gorm.DB, projectID string) error
	GetAnalyticsForCurrentMonth(ctx context.Context, projectID string) ([]models.Analytics, error)
	GetTracker() *AnalyticsTracker
}

type analyticsService struct {
	repo    repo.AnalyticsRepository
	tracker *AnalyticsTracker
}

func NewAnalyticsService(r repo.AnalyticsRepository) AnalyticsService {
	return &analyticsService{
		repo:    r,
		tracker: &AnalyticsTracker{repo: r},
	}
}

func (s *analyticsService) GetTracker() *AnalyticsTracker { return s.tracker }

func (s *analyticsService) CreateInitialAnalyticsForProject(ctx context.Context, tx *gorm.DB, projectID string) error {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	types := []models.AnalyticsType{
		models.AnalyticsTypeStorageUsage,
		models.AnalyticsTypeApiCalls,
		models.AnalyticsTypeAuthActivity,
		models.AnalyticsTypeDatabaseUsage,
		models.AnalyticsTypeNotifications,
		models.AnalyticsTypeRealtimeChannels, // ✅ KU DARNAY
		models.AnalyticsTypeRealtimeEvents,   // ✅ KU DARNAY
	}

	for _, aType := range types {
		record := &models.Analytics{
			ProjectID:   projectID,
			Type:        aType,
			PeriodStart: startOfMonth,
			Metrics:     datatypes.JSON([]byte("{}")),
		}
		if err := s.repo.Create(ctx, tx, record); err != nil {
			return err
		}
	}
	return nil
}

func (s *analyticsService) GetAnalyticsForCurrentMonth(ctx context.Context, projectID string) ([]models.Analytics, error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	records, err := s.repo.GetByProjectForPeriod(ctx, projectID, startOfMonth)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		s.CreateInitialAnalyticsForProject(ctx, nil, projectID)
		return s.repo.GetByProjectForPeriod(ctx, projectID, startOfMonth)
	}
	return records, nil
}
