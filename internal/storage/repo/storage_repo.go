package repo

import (
	"context"
	"superaib/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StorageRepository interface {
	Create(ctx context.Context, file *models.StorageFile) error
	GetByProject(ctx context.Context, projectID string, page, pageSize int) ([]models.StorageFile, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.StorageFile, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type GormStorageRepository struct {
	db *gorm.DB
}

func NewGormStorageRepository(db *gorm.DB) StorageRepository {
	return &GormStorageRepository{db: db}
}

func (r *GormStorageRepository) Create(ctx context.Context, file *models.StorageFile) error {
	return r.db.WithContext(ctx).Create(file).Error
}

func (r *GormStorageRepository) GetByProject(ctx context.Context, projectID string, page, pageSize int) ([]models.StorageFile, int64, error) {
	var files []models.StorageFile
	var total int64
	offset := (page - 1) * pageSize

	query := r.db.WithContext(ctx).Model(&models.StorageFile{}).Where("project_id = ?", projectID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(pageSize).Order("uploaded_at DESC").Find(&files).Error; err != nil {
		return nil, 0, err
	}
	return files, total, nil
}

func (r *GormStorageRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.StorageFile, error) {
	var file models.StorageFile
	if err := r.db.WithContext(ctx).First(&file, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *GormStorageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// GORM's soft delete will be used because of `gorm.DeletedAt` in the model
	return r.db.WithContext(ctx).Delete(&models.StorageFile{}, "id = ?", id).Error
}
