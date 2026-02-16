package services

import (
	"context"
	"errors"
	"fmt"
	"superaib/internal/models"
	"superaib/internal/storage/repo"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProjectAuthConfigService interface {
	ConfigureProvider(ctx context.Context, projectID, providerInput string, credentials map[string]interface{}, enabled bool) (*models.ProjectAuthConfig, error)
	GetConfigsByProject(ctx context.Context, projectID string) ([]models.ProjectAuthConfig, error)
}

type projectAuthConfigService struct {
	repo repo.ProjectAuthConfigRepository
}

func NewProjectAuthConfigService(r repo.ProjectAuthConfigRepository) ProjectAuthConfigService {
	return &projectAuthConfigService{repo: r}
}
func (s *projectAuthConfigService) ConfigureProvider(ctx context.Context, projectID, providerInput string, credentials map[string]interface{}, enabled bool) (*models.ProjectAuthConfig, error) {

	var realProviderID string
	id, err := s.repo.GetGlobalProviderIDByName(ctx, providerInput)
	if err == nil {
		realProviderID = id
	} else {
		if _, uuidErr := uuid.Parse(providerInput); uuidErr == nil {
			realProviderID = providerInput
		} else {
			return nil, fmt.Errorf("invalid provider identity: %s", providerInput)
		}
	}

	// 2. Raadi haddii Config-gu jiro
	cfg, err := s.repo.GetByProjectAndProvider(ctx, projectID, realProviderID)

	isNew := false
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			isNew = true
			cfg = &models.ProjectAuthConfig{
				ID:         uuid.New(),
				ProjectID:  projectID,
				ProviderID: realProviderID,
				CreatedAt:  time.Now(),
			}
		} else {
			return nil, err
		}
	}

	// 3. Cusboonaysii xogta
	if len(credentials) > 0 {
		cfg.Credentials = models.ToJSONB(credentials)
	}
	cfg.Enabled = enabled
	cfg.UpdatedAt = time.Now()

	// 4. ✅ TALLAABADA CUSUB: Kala saar Update iyo Create
	if isNew {
		if err := s.repo.Create(ctx, cfg); err != nil {
			return nil, fmt.Errorf("failed to create config: %w", err)
		}
	} else {
		if err := s.repo.Update(ctx, cfg); err != nil {
			return nil, fmt.Errorf("failed to update config: %w", err)
		}
	}

	// 5. Reload (si xogtu u noqoto mid dhamaystiran)
	return s.repo.GetByProjectAndProvider(ctx, projectID, realProviderID)
}
func (s *projectAuthConfigService) GetConfigsByProject(ctx context.Context, projectID string) ([]models.ProjectAuthConfig, error) {
	return s.repo.GetAllByProject(ctx, projectID)
}

