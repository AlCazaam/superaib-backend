package repo

import (
	"context"
	"strings"
	"superaib/internal/models"

	"gorm.io/gorm"
)

type ProjectAuthConfigRepository interface {
	Create(ctx context.Context, cfg *models.ProjectAuthConfig) error
	Update(ctx context.Context, cfg *models.ProjectAuthConfig) error
	GetByProjectAndProvider(ctx context.Context, projectID, providerID string) (*models.ProjectAuthConfig, error)
	GetByProjectAndProviderName(ctx context.Context, projectID, providerName string) (*models.ProjectAuthConfig, error)
	GetAllByProject(ctx context.Context, projectID string) ([]models.ProjectAuthConfig, error)
	Delete(ctx context.Context, projectID, providerID string) error
	GetGlobalProviderIDByName(ctx context.Context, name string) (string, error)
}

type gormProjectAuthConfigRepo struct {
	db *gorm.DB
}

func NewProjectAuthConfigRepo(db *gorm.DB) ProjectAuthConfigRepository {
	return &gormProjectAuthConfigRepo{db: db}
}

// ✅ 1. CREATE: Kaliya marka qofka la abuurayo
func (r *gormProjectAuthConfigRepo) Create(ctx context.Context, cfg *models.ProjectAuthConfig) error {
	return r.db.WithContext(ctx).Create(cfg).Error
}

// ✅ 2. UPDATE: Kaliya marka qofka la bedelayo (adoo isticmaalaya ID-ga rasmiga ah)
func (r *gormProjectAuthConfigRepo) Update(ctx context.Context, cfg *models.ProjectAuthConfig) error {
	return r.db.WithContext(ctx).Model(cfg).
		Select("Enabled", "Credentials", "UpdatedAt").
		Updates(cfg).Error
}

func (r *gormProjectAuthConfigRepo) GetGlobalProviderIDByName(ctx context.Context, name string) (string, error) {
	var provider models.GlobalAuthProvider
	err := r.db.WithContext(ctx).
		Where("LOWER(name) = ? OR LOWER(title) LIKE ?", strings.ToLower(name), "%"+strings.ToLower(name)+"%").
		First(&provider).Error
	if err != nil {
		return "", err
	}
	return provider.ID.String(), nil
}

func (r *gormProjectAuthConfigRepo) GetByProjectAndProvider(ctx context.Context, projectID, providerID string) (*models.ProjectAuthConfig, error) {
	var cfg models.ProjectAuthConfig
	err := r.db.WithContext(ctx).Preload("Provider").
		Where("project_id = ? AND provider_id = ?", projectID, providerID).First(&cfg).Error
	return &cfg, err
}

func (r *gormProjectAuthConfigRepo) GetByProjectAndProviderName(ctx context.Context, projectID, providerName string) (*models.ProjectAuthConfig, error) {
	var cfg models.ProjectAuthConfig
	err := r.db.WithContext(ctx).
		Joins("JOIN global_auth_providers ON global_auth_providers.id = project_auth_configs.provider_id").
		Where("project_auth_configs.project_id = ? AND LOWER(global_auth_providers.name) = ?", projectID, strings.ToLower(providerName)).
		First(&cfg).Error
	return &cfg, err
}

func (r *gormProjectAuthConfigRepo) GetAllByProject(ctx context.Context, projectID string) ([]models.ProjectAuthConfig, error) {
	var configs []models.ProjectAuthConfig
	err := r.db.WithContext(ctx).Preload("Provider").
		Where("project_id = ?", projectID).Find(&configs).Error
	return configs, err
}

func (r *gormProjectAuthConfigRepo) Delete(ctx context.Context, projectID, providerID string) error {
	return r.db.WithContext(ctx).
		Where("project_id = ? AND provider_id = ?", projectID, providerID).
		Delete(&models.ProjectAuthConfig{}).Error
}
