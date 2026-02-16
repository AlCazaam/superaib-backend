package services

import (
	"context"
	"errors"
	"superaib/internal/models"
	"superaib/internal/storage/repo"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type GlobalFeatureService interface {
	CreateGlobalFeature(ctx context.Context, fType models.ProjectFeatureType, enabled bool, config datatypes.JSON) (*models.GlobalFeature, error)
	GetAllGlobalFeatures(ctx context.Context) ([]models.GlobalFeature, error)
	ToggleGlobalFeature(ctx context.Context, fType models.ProjectFeatureType, enabled bool) (*models.GlobalFeature, error)
	UpdateGlobalFeature(ctx context.Context, id uuid.UUID, fType models.ProjectFeatureType, enabled bool, config datatypes.JSON) (*models.GlobalFeature, error)
	DeleteGlobalFeature(ctx context.Context, id uuid.UUID) error
}

type globalFeatureService struct {
	repo repo.GlobalFeatureRepository
}

func NewGlobalFeatureService(r repo.GlobalFeatureRepository) GlobalFeatureService {
	return &globalFeatureService{repo: r}
}

func (s *globalFeatureService) CreateGlobalFeature(ctx context.Context, fType models.ProjectFeatureType, enabled bool, config datatypes.JSON) (*models.GlobalFeature, error) {
	feature := &models.GlobalFeature{
		Type:    fType,
		Enabled: enabled,
		Config:  config,
	}
	if err := s.repo.Create(ctx, feature); err != nil {
		return nil, err
	}
	return feature, nil
}

func (s *globalFeatureService) GetAllGlobalFeatures(ctx context.Context) ([]models.GlobalFeature, error) {
	return s.repo.GetAll(ctx)
}

func (s *globalFeatureService) ToggleGlobalFeature(ctx context.Context, fType models.ProjectFeatureType, enabled bool) (*models.GlobalFeature, error) {
	feature, err := s.repo.GetByType(ctx, fType)
	if err != nil {
		return nil, errors.New("global feature type not found")
	}
	feature.Enabled = enabled
	feature.UpdatedAt = time.Now()
	return feature, s.repo.Update(ctx, feature)
}

func (s *globalFeatureService) UpdateGlobalFeature(ctx context.Context, id uuid.UUID, fType models.ProjectFeatureType, enabled bool, config datatypes.JSON) (*models.GlobalFeature, error) {
	feature, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	feature.Type = fType
	feature.Enabled = enabled
	feature.Config = config
	feature.UpdatedAt = time.Now()
	return feature, s.repo.Update(ctx, feature)
}

func (s *globalFeatureService) DeleteGlobalFeature(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
