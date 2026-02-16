// repo/api_key_repo.go
package repo

import (
	"context"
	"superaib/internal/models"

	"gorm.io/gorm"
)

type APIKeyRepository interface {
	CreateAPIKey(ctx context.Context, key *models.APIKey) error
	GetAPIKeyByID(ctx context.Context, id string) (*models.APIKey, error) // ProjectID looma baahna halkan
	GetAPIKeyByKeyValue(ctx context.Context, keyValue string) (*models.APIKey, error)
	GetAllAPIKeysForProject(ctx context.Context, projectID string, page, pageSize int) ([]models.APIKey, int64, error)
	UpdateAPIKey(ctx context.Context, key *models.APIKey) error
	DeleteAPIKey(ctx context.Context, id string) error // ProjectID looma baahna halkan
	CountAPIKeysByProject(ctx context.Context, projectID string) (int64, error)
}

type GormAPIKeyRepository struct {
	db *gorm.DB
}

func NewGormAPIKeyRepository(db *gorm.DB) *GormAPIKeyRepository {
	return &GormAPIKeyRepository{db: db}
}

func (r *GormAPIKeyRepository) CreateAPIKey(ctx context.Context, key *models.APIKey) error {
	return r.db.WithContext(ctx).Create(key).Error
}

// GetAPIKeyByID retrieves an API key by its own ID.
func (r *GormAPIKeyRepository) GetAPIKeyByID(ctx context.Context, id string) (*models.APIKey, error) {
	var key models.APIKey
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&key).Error; err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *GormAPIKeyRepository) GetAPIKeyByKeyValue(ctx context.Context, keyValue string) (*models.APIKey, error) {
	var key models.APIKey
	if err := r.db.WithContext(ctx).Where("key = ?", keyValue).First(&key).Error; err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *GormAPIKeyRepository) GetAllAPIKeysForProject(ctx context.Context, projectID string, page, pageSize int) ([]models.APIKey, int64, error) {
	var keys []models.APIKey
	var total int64
	offset := (page - 1) * pageSize
	dbWithCtx := r.db.WithContext(ctx).Model(&models.APIKey{}).Where("project_id = ?", projectID)
	if err := dbWithCtx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := dbWithCtx.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&keys).Error; err != nil {
		return nil, 0, err
	}
	return keys, total, nil
}

func (r *GormAPIKeyRepository) UpdateAPIKey(ctx context.Context, key *models.APIKey) error {
	return r.db.WithContext(ctx).Save(key).Error
}

// DeleteAPIKey deletes an API key by its ID. Verification happens in the service layer.
func (r *GormAPIKeyRepository) DeleteAPIKey(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.APIKey{}).Error
}

func (r *GormAPIKeyRepository) CountAPIKeysByProject(ctx context.Context, projectID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.APIKey{}).Where("project_id = ?", projectID).Count(&count).Error
	return count, err
}
