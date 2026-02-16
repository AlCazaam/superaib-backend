package services

import (
	"context"
	"errors"
	"superaib/internal/models"
	"superaib/internal/storage/repo"

	"gorm.io/gorm"
)

type RateLimitPolicyService interface {
	CreatePolicy(ctx context.Context, policy *models.RateLimitPolicy) (*models.RateLimitPolicy, error)
	UpdatePolicy(ctx context.Context, projectID string, updatedPolicy *models.RateLimitPolicy) (*models.RateLimitPolicy, error)
	GetPolicy(ctx context.Context, projectID string) (*models.RateLimitPolicy, error)
	DeletePolicy(ctx context.Context, projectID string) error
	TogglePolicy(ctx context.Context, projectID string, isEnabled bool) (*models.RateLimitPolicy, error)
}

type rateLimitPolicyService struct {
	repo repo.RateLimitPolicyRepository
}

func NewRateLimitPolicyService(r repo.RateLimitPolicyRepository) RateLimitPolicyService {
	return &rateLimitPolicyService{repo: r}
}

// 1. Create Policy (Waa inuu Policy kasta yeeshaa Project ID u gaar ah)

func (s *rateLimitPolicyService) CreatePolicy(ctx context.Context, policy *models.RateLimitPolicy) (*models.RateLimitPolicy, error) {

	// ✅ HAGAAJINTA: Isticmaal policy.ProjectID si aad u hesho kii hore
	existing, err := s.repo.GetByProjectID(ctx, policy.ProjectID)

	// Hubi haddii Policy hore u jiray (Unique constraint)
	if err == nil && existing != nil {
		return nil, errors.New("policy already exists for this project")
	}

	if createErr := s.repo.Create(ctx, policy); createErr != nil {
		return nil, createErr
	}
	return policy, nil
}

// 2. Update Policy
// 2. Update Policy (Hagaajinta line-kan)
func (s *rateLimitPolicyService) UpdatePolicy(ctx context.Context, projectID string, updatedPolicy *models.RateLimitPolicy) (*models.RateLimitPolicy, error) {
	// Wuxuu ahaa s.repo.GetPolicy(ctx, projectID)
	existing, err := s.repo.GetByProjectID(ctx, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("policy not found")
		}
		return nil, err
	}

	// Cusboonaysii goobaha la oggol yahay
	existing.MaxRequests = updatedPolicy.MaxRequests
	existing.WindowMinutes = updatedPolicy.WindowMinutes
	existing.LockoutMinutes = updatedPolicy.LockoutMinutes
	existing.IsEnabled = updatedPolicy.IsEnabled // Wuxuu u shaqayn karaa sidii toggle

	if updateErr := s.repo.Update(ctx, existing); updateErr != nil {
		return nil, updateErr
	}
	return existing, nil
}

// 3. Get Policy
// 3. Get Policy (Hagaajinta line-kan)
func (s *rateLimitPolicyService) GetPolicy(ctx context.Context, projectID string) (*models.RateLimitPolicy, error) {
	// Wuxuu ahaa s.repo.GetPolicy(ctx, projectID)
	return s.repo.GetByProjectID(ctx, projectID) // ✅ Hadda waa sax
}

// 4. Delete Policy
func (s *rateLimitPolicyService) DeletePolicy(ctx context.Context, projectID string) error {
	return s.repo.DeleteByProjectID(ctx, projectID)
}

// 5. Toggle Policy (Function gaar ah oo sahlan)
// 5. Toggle Policy (Hagaajinta line-kan)
func (s *rateLimitPolicyService) TogglePolicy(ctx context.Context, projectID string, isEnabled bool) (*models.RateLimitPolicy, error) {
	// Wuxuu ahaa s.repo.GetPolicy(ctx, projectID)
	existing, err := s.repo.GetByProjectID(ctx, projectID) // ✅ Hadda waa sax
	if err != nil {
		return nil, err
	}
	existing.IsEnabled = isEnabled
	if updateErr := s.repo.Update(ctx, existing); updateErr != nil {
		return nil, updateErr
	}
	return existing, nil
}
