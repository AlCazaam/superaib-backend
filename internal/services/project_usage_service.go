package services

import (
	"context"
	"superaib/internal/models"
	"superaib/internal/storage/repo"

	"gorm.io/gorm"
)

type ProjectUsageService interface {
	GetUsage(ctx context.Context, projectUUID string) (*models.ProjectUsage, error)
	UpdateUsage(ctx context.Context, projectUUID string, field string, value interface{}) error
	CreateInitialUsageRecord(ctx context.Context, tx *gorm.DB, projectID string) error
}

type projectUsageService struct {
	repo repo.ProjectUsageRepository
}

func NewProjectUsageService(r repo.ProjectUsageRepository) ProjectUsageService {
	return &projectUsageService{repo: r}
}

func (s *projectUsageService) GetUsage(ctx context.Context, projectUUID string) (*models.ProjectUsage, error) {
	return s.repo.GetByProjectID(ctx, projectUUID)
}

func (s *projectUsageService) UpdateUsage(ctx context.Context, projectUUID string, field string, value interface{}) error {
	return s.repo.IncrementField(ctx, projectUUID, field, value)
}

func (s *projectUsageService) CreateInitialUsageRecord(ctx context.Context, tx *gorm.DB, projectID string) error {
	usage := &models.ProjectUsage{
		ProjectID: projectID,
		// Dhamaan USED waa 0 default
	}
	return s.repo.Create(ctx, tx, usage)
}
