package repo

import (
	"context"
	"superaib/internal/models"

	"gorm.io/gorm"
)

type PlanRepository interface {
	Create(ctx context.Context, plan *models.Plan) error
	GetAll(ctx context.Context) ([]models.Plan, error)
	GetByID(ctx context.Context, id string) (*models.Plan, error)
	Update(ctx context.Context, plan *models.Plan) error
	Delete(ctx context.Context, id string) error
}

type gormPlanRepository struct {
	db *gorm.DB
}

func NewPlanRepository(db *gorm.DB) PlanRepository {
	return &gormPlanRepository{db: db}
}

func (r *gormPlanRepository) Create(ctx context.Context, plan *models.Plan) error {
	return r.db.WithContext(ctx).Create(plan).Error
}

func (r *gormPlanRepository) GetAll(ctx context.Context) ([]models.Plan, error) {
	var plans []models.Plan
	err := r.db.WithContext(ctx).Order("price asc").Find(&plans).Error
	return plans, err
}

func (r *gormPlanRepository) GetByID(ctx context.Context, id string) (*models.Plan, error) {
	var plan models.Plan
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&plan).Error
	return &plan, err
}

func (r *gormPlanRepository) Update(ctx context.Context, plan *models.Plan) error {
	return r.db.WithContext(ctx).Save(plan).Error
}

func (r *gormPlanRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Plan{}, "id = ?", id).Error
}
