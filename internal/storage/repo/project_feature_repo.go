package repo

import (
	"context"
	"errors"

	"superaib/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProjectFeatureRepository interface {
	CreateFeature(ctx context.Context, feature *models.ProjectFeature) error
	CreateFeaturesInBatches(ctx context.Context, features []*models.ProjectFeature) error
	GetFeatureByID(ctx context.Context, id uuid.UUID) (*models.ProjectFeature, error)
	GetFeatureByProjectIDAndType(ctx context.Context, projectID string, featureType models.ProjectFeatureType) (*models.ProjectFeature, error)
	GetAllFeaturesByProject(ctx context.Context, projectID string) ([]models.ProjectFeature, error)
	UpdateFeature(ctx context.Context, feature *models.ProjectFeature) error
	UpdateFeatures(ctx context.Context, features []*models.ProjectFeature) error
	DeleteFeature(ctx context.Context, id uuid.UUID) error
}

type GormProjectFeatureRepo struct {
	db *gorm.DB
}

func NewGormProjectFeatureRepo(db *gorm.DB) ProjectFeatureRepository {
	return &GormProjectFeatureRepo{db: db}
}

func (r *GormProjectFeatureRepo) CreateFeature(ctx context.Context, feature *models.ProjectFeature) error {
	return r.db.WithContext(ctx).Create(feature).Error
}

func (r *GormProjectFeatureRepo) CreateFeaturesInBatches(ctx context.Context, features []*models.ProjectFeature) error {
	return r.db.WithContext(ctx).CreateInBatches(features, len(features)).Error
}

func (r *GormProjectFeatureRepo) GetFeatureByID(ctx context.Context, id uuid.UUID) (*models.ProjectFeature, error) {
	var feature models.ProjectFeature
	if err := r.db.WithContext(ctx).First(&feature, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &feature, nil
}

func (r *GormProjectFeatureRepo) GetFeatureByProjectIDAndType(ctx context.Context, projectID string, featureType models.ProjectFeatureType) (*models.ProjectFeature, error) {
	var feature models.ProjectFeature
	if err := r.db.WithContext(ctx).Where("project_id = ? AND type = ?", projectID, featureType).First(&feature).Error; err != nil {
		return nil, err
	}
	return &feature, nil
}

func (r *GormProjectFeatureRepo) GetAllFeaturesByProject(ctx context.Context, projectID string) ([]models.ProjectFeature, error) {
	var features []models.ProjectFeature
	if err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Find(&features).Error; err != nil {
		return nil, err
	}
	return features, nil
}

func (r *GormProjectFeatureRepo) UpdateFeature(ctx context.Context, feature *models.ProjectFeature) error {
	// ✅ SAX: Save wuxuu isticmaalaa ID-ga si uu u update-gareeyo.
	// Haddii ID aysan jirin, wuu abuuraa.
	return r.db.WithContext(ctx).Save(feature).Error
}
func (r *GormProjectFeatureRepo) UpdateFeatures(ctx context.Context, features []*models.ProjectFeature) error {
	return r.db.WithContext(ctx).Save(&features).Error
}

func (r *GormProjectFeatureRepo) DeleteFeature(ctx context.Context, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Delete(&models.ProjectFeature{}, "id = ?", id)
	if res.RowsAffected == 0 {
		return errors.New("feature not found")
	}
	return res.Error
}
