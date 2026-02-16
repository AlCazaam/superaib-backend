package services

import (
	"context"
	"superaib/internal/models"
	"superaib/internal/storage/repo"
)

type PlanService interface {
	CreatePlan(ctx context.Context, plan *models.Plan) error
	GetAllPlans(ctx context.Context) ([]models.Plan, error)
	UpdatePlan(ctx context.Context, id string, planData *models.Plan) (*models.Plan, error)
	DeletePlan(ctx context.Context, id string) error
}

type planService struct {
	repo repo.PlanRepository
}

func NewPlanService(r repo.PlanRepository) PlanService {
	return &planService{repo: r}
}

func (s *planService) CreatePlan(ctx context.Context, plan *models.Plan) error {
	return s.repo.Create(ctx, plan)
}

func (s *planService) GetAllPlans(ctx context.Context) ([]models.Plan, error) {
	return s.repo.GetAll(ctx)
}

func (s *planService) UpdatePlan(ctx context.Context, id string, planData *models.Plan) (*models.Plan, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	planData.ID = existing.ID
	if err := s.repo.Update(ctx, planData); err != nil {
		return nil, err
	}
	return planData, nil
}

func (s *planService) DeletePlan(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
