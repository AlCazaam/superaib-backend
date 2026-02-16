package services

import (
	"context"
	"superaib/internal/models"
	"superaib/internal/storage/repo"
)

type GlobalAuthProviderService interface {
	Create(ctx context.Context, p *models.GlobalAuthProvider) (*models.GlobalAuthProvider, error)
	GetAll(ctx context.Context) ([]models.GlobalAuthProvider, error)
	Update(ctx context.Context, id string, p *models.GlobalAuthProvider) (*models.GlobalAuthProvider, error)
	Delete(ctx context.Context, id string) error
	ToggleStatus(ctx context.Context, id string, isActive bool) error
}

type globalAuthProviderService struct {
	repo repo.GlobalAuthProviderRepository
}

func NewGlobalAuthProviderService(r repo.GlobalAuthProviderRepository) GlobalAuthProviderService {
	return &globalAuthProviderService{repo: r}
}

func (s *globalAuthProviderService) Create(ctx context.Context, p *models.GlobalAuthProvider) (*models.GlobalAuthProvider, error) {
	err := s.repo.Create(ctx, p)
	return p, err
}

func (s *globalAuthProviderService) GetAll(ctx context.Context) ([]models.GlobalAuthProvider, error) {
	return s.repo.GetAll(ctx)
}

func (s *globalAuthProviderService) Update(ctx context.Context, id string, p *models.GlobalAuthProvider) (*models.GlobalAuthProvider, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	existing.Title = p.Title
	existing.Description = p.Description
	existing.IconURL = p.IconURL
	existing.RequiredFields = p.RequiredFields

	err = s.repo.Update(ctx, existing)
	return existing, err
}

func (s *globalAuthProviderService) ToggleStatus(ctx context.Context, id string, isActive bool) error {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	p.IsActive = isActive
	return s.repo.Update(ctx, p)
}

func (s *globalAuthProviderService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
