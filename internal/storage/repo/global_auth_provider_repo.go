package repo

import (
	"context"
	"superaib/internal/models"

	"gorm.io/gorm"
)

type GlobalAuthProviderRepository interface {
	Create(ctx context.Context, p *models.GlobalAuthProvider) error
	GetAll(ctx context.Context) ([]models.GlobalAuthProvider, error)
	GetByID(ctx context.Context, id string) (*models.GlobalAuthProvider, error)
	Update(ctx context.Context, p *models.GlobalAuthProvider) error
	Delete(ctx context.Context, id string) error
}

type gormGlobalAuthProviderRepo struct {
	db *gorm.DB
}

func NewGlobalAuthProviderRepo(db *gorm.DB) GlobalAuthProviderRepository {
	return &gormGlobalAuthProviderRepo{db: db}
}

func (r *gormGlobalAuthProviderRepo) Create(ctx context.Context, p *models.GlobalAuthProvider) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *gormGlobalAuthProviderRepo) GetAll(ctx context.Context) ([]models.GlobalAuthProvider, error) {
	var providers []models.GlobalAuthProvider
	err := r.db.WithContext(ctx).Order("created_at desc").Find(&providers).Error
	return providers, err
}

func (r *gormGlobalAuthProviderRepo) GetByID(ctx context.Context, id string) (*models.GlobalAuthProvider, error) {
	var p models.GlobalAuthProvider
	err := r.db.WithContext(ctx).First(&p, "id = ?", id).Error
	return &p, err
}

func (r *gormGlobalAuthProviderRepo) Update(ctx context.Context, p *models.GlobalAuthProvider) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *gormGlobalAuthProviderRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.GlobalAuthProvider{}, "id = ?", id).Error
}
