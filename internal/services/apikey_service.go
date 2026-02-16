package services

import (
	"context"
	"errors"
	"fmt"
	"superaib/internal/models"
	"superaib/internal/storage/repo"
	"superaib/pkg/utils"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	APIKeyLength = 32
	APIKeyPrefix = "sk-"
)

type APIKeyService interface {
	CreateAPIKey(ctx context.Context, projectIdentifier, createdBy, name string, permissions datatypes.JSON) (*models.APIKey, error)
	GetAPIKeyByID(ctx context.Context, id, projectIdentifier string) (*models.APIKey, error)
	GetAPIKeyByKeyValue(ctx context.Context, keyValue string) (*models.APIKey, error)
	GetAllAPIKeysForProject(ctx context.Context, projectIdentifier string, page, pageSize int) ([]models.APIKey, int64, error)
	UpdateAPIKey(ctx context.Context, id, projectIdentifier string, updates map[string]interface{}) (*models.APIKey, error)
	RevokeAPIKey(ctx context.Context, id, projectIdentifier string) error
	DeleteAPIKey(ctx context.Context, id, projectIdentifier string) error
	UpdateAPIKeyUsage(ctx context.Context, keyValue string) error
}

type apiKeyService struct {
	apiKeyRepo   repo.APIKeyRepository
	projectRepo  repo.ProjectRepository
	tracker      *AnalyticsTracker
	usageService ProjectUsageService // ✅ SI SAX AH LOOGU DARAY
}

// ✅ Constructor-ka hadda wuxuu aqbalayaa 4 dependencies sidii main.go loogu baahnaa
func NewAPIKeyService(ak repo.APIKeyRepository, pr repo.ProjectRepository, tracker *AnalyticsTracker, usage ProjectUsageService) APIKeyService {
	return &apiKeyService{
		apiKeyRepo:   ak,
		projectRepo:  pr,
		tracker:      tracker,
		usageService: usage,
	}
}

// UpdateAPIKeyUsage: Function-ka ugu muhiimsan ee maamula Billing-ka iyo Analytics-ka
func (s *apiKeyService) UpdateAPIKeyUsage(ctx context.Context, keyValue string) error {
	apiKey, err := s.apiKeyRepo.GetAPIKeyByKeyValue(ctx, keyValue)
	if err != nil {
		return err
	}

	// 1. Kordhi Usage-ka gudaha API Key Table (Local count)
	apiKey.UsageCount += 1
	apiKey.UpdatedAt = time.Now()
	if err := s.apiKeyRepo.UpdateAPIKey(ctx, apiKey); err != nil {
		return err
	}

	// ✅ 2. KORDHI ANALYTICS (Si uu ugu soo muuqdo Charts-ka Analytics-ka)
	s.tracker.TrackEvent(ctx, apiKey.ProjectID, models.AnalyticsTypeApiCalls, "total_calls", 1)

	// ✅ 3. KORDHI PROJECT USAGE (Si loogu xakameeyo Quota/Limits-ka Plan-ka uu haysto)
	// Field-ka 'api_calls' waa kan ku dhex jira models.ProjectUsage
	err = s.usageService.UpdateUsage(ctx, apiKey.ProjectID, "api_calls", 1)
	if err != nil {
		return fmt.Errorf("failed to update project usage: %w", err)
	}

	return nil
}

// -------------------------------------------------------------------------
// STANDARD CRUD FUNCTIONS (Waa sidoodii)
// -------------------------------------------------------------------------

func (s *apiKeyService) CreateAPIKey(ctx context.Context, projectIdentifier, createdBy, name string, permissions datatypes.JSON) (*models.APIKey, error) {
	project, err := s.projectRepo.GetProjectByAnyIDAndOwner(ctx, projectIdentifier, createdBy)
	if err != nil {
		return nil, errors.New("project not found or unauthorized")
	}

	generatedKey, err := s.generateUniqueAPIKey(ctx)
	if err != nil {
		return nil, err
	}

	apiKey := &models.APIKey{
		ID:          uuid.New().String(),
		ProjectID:   project.ID,
		Key:         generatedKey,
		Name:        name,
		Permissions: permissions,
		Revoked:     false,
		UsageCount:  0,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.apiKeyRepo.CreateAPIKey(ctx, apiKey); err != nil {
		return nil, fmt.Errorf("failed to save api key: %w", err)
	}

	return apiKey, nil
}

func (s *apiKeyService) GetAllAPIKeysForProject(ctx context.Context, projectIdentifier string, page, pageSize int) ([]models.APIKey, int64, error) {
	ownerID := ctx.Value("userID").(string)
	project, err := s.projectRepo.GetProjectByAnyIDAndOwner(ctx, projectIdentifier, ownerID)
	if err != nil {
		return nil, 0, errors.New("unauthorized access to project keys")
	}
	return s.apiKeyRepo.GetAllAPIKeysForProject(ctx, project.ID, page, pageSize)
}

func (s *apiKeyService) RevokeAPIKey(ctx context.Context, id, projectIdentifier string) error {
	apiKey, err := s.apiKeyRepo.GetAPIKeyByID(ctx, id)
	if err != nil {
		return err
	}
	apiKey.Revoked = true
	apiKey.UpdatedAt = time.Now()
	return s.apiKeyRepo.UpdateAPIKey(ctx, apiKey)
}

func (s *apiKeyService) generateUniqueAPIKey(ctx context.Context) (string, error) {
	for i := 0; i < 5; i++ {
		key := APIKeyPrefix + utils.GenerateRandomString(APIKeyLength)
		_, err := s.apiKeyRepo.GetAPIKeyByKeyValue(ctx, key)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return key, nil
		}
	}
	return "", errors.New("failed to generate unique key")
}

func (s *apiKeyService) GetAPIKeyByKeyValue(ctx context.Context, keyValue string) (*models.APIKey, error) {
	return s.apiKeyRepo.GetAPIKeyByKeyValue(ctx, keyValue)
}

func (s *apiKeyService) DeleteAPIKey(ctx context.Context, id, projectIdentifier string) error {
	return s.apiKeyRepo.DeleteAPIKey(ctx, id)
}

func (s *apiKeyService) UpdateAPIKey(ctx context.Context, id, projectIdentifier string, updates map[string]interface{}) (*models.APIKey, error) {
	apiKey, err := s.apiKeyRepo.GetAPIKeyByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := utils.ApplyUpdates(apiKey, updates); err != nil {
		return nil, err
	}
	apiKey.UpdatedAt = time.Now()
	if err := s.apiKeyRepo.UpdateAPIKey(ctx, apiKey); err != nil {
		return nil, err
	}
	return apiKey, nil
}

func (s *apiKeyService) GetAPIKeyByID(ctx context.Context, id, projectIdentifier string) (*models.APIKey, error) {
	return s.apiKeyRepo.GetAPIKeyByID(ctx, id)
}
