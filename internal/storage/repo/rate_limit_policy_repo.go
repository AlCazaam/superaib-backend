package repo

import (
	"context"
	"errors"
	"superaib/internal/models"

	"gorm.io/gorm"
)

type RateLimitPolicyRepository interface {
	Create(ctx context.Context, policy *models.RateLimitPolicy) error
	Update(ctx context.Context, policy *models.RateLimitPolicy) error
	GetByProjectID(ctx context.Context, projectID string) (*models.RateLimitPolicy, error)
	DeleteByProjectID(ctx context.Context, projectID string) error
}

type gormRateLimitPolicyRepo struct {
	db *gorm.DB
}

func NewRateLimitPolicyRepository(db *gorm.DB) RateLimitPolicyRepository {
	return &gormRateLimitPolicyRepo{db: db}
}

func (r *gormRateLimitPolicyRepo) Create(ctx context.Context, policy *models.RateLimitPolicy) error {
	return r.db.WithContext(ctx).Create(policy).Error
}

func (r *gormRateLimitPolicyRepo) Update(ctx context.Context, policy *models.RateLimitPolicy) error {
	// Waxaan kaliya cusboonaysiinaynaa goobaha loo baahan yahay
	return r.db.WithContext(ctx).Save(policy).Error
}

func (r *gormRateLimitPolicyRepo) GetByProjectID(ctx context.Context, projectID string) (*models.RateLimitPolicy, error) {
	var policy models.RateLimitPolicy
	err := r.db.WithContext(ctx).First(&policy, "project_id = ?", projectID).Error
	return &policy, err
}

func (r *gormRateLimitPolicyRepo) DeleteByProjectID(ctx context.Context, projectID string) error {
	res := r.db.WithContext(ctx).Where("project_id = ?", projectID).Delete(&models.RateLimitPolicy{})
	if res.RowsAffected == 0 {
		return errors.New("rate limit policy not found for this project")
	}
	return res.Error
}
