// services/project_feature_service.go
package services

import (
	"context"
	"errors"
	"fmt"
	"superaib/internal/models"
	"superaib/internal/storage/repo"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ProjectFeatureService interface {
	CreateFeature(ctx context.Context, feature *models.ProjectFeature) error
	GenerateDefaultFeaturesWithTx(ctx context.Context, tx *gorm.DB, projectID string) ([]models.ProjectFeature, error)
	GenerateDefaultFeatures(ctx context.Context, projectID string) ([]models.ProjectFeature, error)
	GetFeatureByID(ctx context.Context, id uuid.UUID) (*models.ProjectFeature, error)
	GetAllFeaturesByProject(ctx context.Context, projectID string) ([]models.ProjectFeature, error)
	ToggleFeature(ctx context.Context, projectID string, featureType models.ProjectFeatureType, enable bool, config datatypes.JSON) (*models.ProjectFeature, error)
	UpdateFeature(ctx context.Context, feature *models.ProjectFeature) error
	DeleteFeature(ctx context.Context, id uuid.UUID) error
}

type projectFeatureService struct {
	repo     repo.ProjectFeatureRepository
	db       *gorm.DB
	validate *validator.Validate
}

func NewProjectFeatureService(r repo.ProjectFeatureRepository, db *gorm.DB) ProjectFeatureService {
	return &projectFeatureService{
		repo:     r,
		db:       db,
		validate: validator.New(),
	}
}

// 1. GenerateDefaultFeaturesWithTx - Waxaan ka saarnay PlanTier
func (s *projectFeatureService) GenerateDefaultFeaturesWithTx(ctx context.Context, tx *gorm.DB, projectID string) ([]models.ProjectFeature, error) {
	// Ka hel liiska default features-ka (Auth, DB, Storage, Analytics)
	features := models.SeedDefaultFeatures(projectID)

	// GORM wuxuu si toos ah u gelinayaa dhammaan liiska hal mar (Batch Create)
	if err := tx.WithContext(ctx).Create(&features).Error; err != nil {
		return nil, fmt.Errorf("failed to batch create default features: %w", err)
	}

	return features, nil
}

// 2. GenerateDefaultFeatures - Habka caadiga ah haddii aan Transaction weyn la dhex joogin
func (s *projectFeatureService) GenerateDefaultFeatures(ctx context.Context, projectID string) ([]models.ProjectFeature, error) {
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	features, err := s.GenerateDefaultFeaturesWithTx(ctx, tx, projectID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return features, nil
}

// 3. CreateFeature - Ku dar feature cusub oo aan ahayn kuwa default-ka ah
func (s *projectFeatureService) CreateFeature(ctx context.Context, feature *models.ProjectFeature) error {
	feature.ID = uuid.New()
	feature.CreatedAt = time.Now()
	feature.UpdatedAt = time.Now()

	if err := s.validate.Struct(feature); err != nil {
		return fmt.Errorf("feature validation failed: %w", err)
	}
	return s.repo.CreateFeature(ctx, feature)
}

// 4. GetFeatureByID - Raadi feature adigoo UUID-ga isticmaalaya
func (s *projectFeatureService) GetFeatureByID(ctx context.Context, id uuid.UUID) (*models.ProjectFeature, error) {
	feature, err := s.repo.GetFeatureByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("feature not found")
		}
		return nil, err
	}
	return feature, nil
}

// 5. GetAllFeaturesByProject - Liiska adeegyada u furan hal project
func (s *projectFeatureService) GetAllFeaturesByProject(ctx context.Context, projectID string) ([]models.ProjectFeature, error) {
	return s.repo.GetAllFeaturesByProject(ctx, projectID)
}
func (s *projectFeatureService) ToggleFeature(ctx context.Context, projectID string, fType models.ProjectFeatureType, enable bool, config datatypes.JSON) (*models.ProjectFeature, error) {
	if fType == "" {
		return nil, errors.New("feature type is required")
	}

	// 1. Marka hore raadi haddii uu feature-kani jiro (Aad u muhiim ah!)
	feature, err := s.repo.GetFeatureByProjectIDAndType(ctx, projectID, fType)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// MA JIRO: Dhis mid cusub (First time)
			feature = &models.ProjectFeature{
				ID:        uuid.New(),
				ProjectID: projectID,
				Type:      fType,
				Enabled:   enable,
				Config:    config,
				CreatedAt: time.Now(),
			}
		} else {
			return nil, err
		}
	} else {
		// WUU JIRAA: Kaliya xogta baddal
		feature.Enabled = enable
		// Kaliya update-garee config-ga haddii xog cusub la soo diray
		if config != nil && string(config) != "null" && string(config) != "{}" {
			feature.Config = config
		}
	}

	feature.UpdatedAt = time.Now()

	// 2. Save wuxuu qabanayaa Create ama Update (Upsert logic)
	if err := s.repo.UpdateFeature(ctx, feature); err != nil {
		return nil, err
	}

	return feature, nil
}

// 7. UpdateFeature - Isbedel guud
func (s *projectFeatureService) UpdateFeature(ctx context.Context, feature *models.ProjectFeature) error {
	feature.UpdatedAt = time.Now()
	return s.repo.UpdateFeature(ctx, feature)
}

// 8. DeleteFeature - Tirtir feature
func (s *projectFeatureService) DeleteFeature(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteFeature(ctx, id)
}
