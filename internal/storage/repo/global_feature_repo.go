package repo

import (
	"context"
	"superaib/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GlobalFeatureRepository interface {
	Create(ctx context.Context, feature *models.GlobalFeature) error
	GetAll(ctx context.Context) ([]models.GlobalFeature, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.GlobalFeature, error)
	GetByType(ctx context.Context, fType models.ProjectFeatureType) (*models.GlobalFeature, error)
	Update(ctx context.Context, feature *models.GlobalFeature) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type GormGlobalFeatureRepo struct {
	db *gorm.DB
}

func NewGormGlobalFeatureRepo(db *gorm.DB) GlobalFeatureRepository {
	return &GormGlobalFeatureRepo{db: db}
}

func (r *GormGlobalFeatureRepo) Create(ctx context.Context, feature *models.GlobalFeature) error {
	return r.db.WithContext(ctx).Create(feature).Error
}

func (r *GormGlobalFeatureRepo) GetAll(ctx context.Context) ([]models.GlobalFeature, error) {
	var features []models.GlobalFeature
	err := r.db.WithContext(ctx).Find(&features).Error
	return features, err
}

func (r *GormGlobalFeatureRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.GlobalFeature, error) {
	var feature models.GlobalFeature
	err := r.db.WithContext(ctx).First(&feature, "id = ?", id).Error
	return &feature, err
}

func (r *GormGlobalFeatureRepo) GetByType(ctx context.Context, fType models.ProjectFeatureType) (*models.GlobalFeature, error) {
	var feature models.GlobalFeature
	err := r.db.WithContext(ctx).Where("type = ?", fType).First(&feature).Error
	return &feature, err
}

func (r *GormGlobalFeatureRepo) Update(ctx context.Context, feature *models.GlobalFeature) error {
	return r.db.WithContext(ctx).Save(feature).Error
}

func (r *GormGlobalFeatureRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.GlobalFeature{}, "id = ?", id).Error
}
