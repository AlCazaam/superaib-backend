package services

import (
	"context"
	"errors"
	"fmt"
	"superaib/internal/models"
	"superaib/internal/storage/repo"
	"superaib/pkg/utils"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProjectService interface {
	CreateProject(ctx context.Context, ownerID string, name string) (*models.Project, error)
	GetProjectByReferenceID(ctx context.Context, referenceID string, ownerID string) (*models.Project, error)
	GetProjectByInternalID(ctx context.Context, id string) (*models.Project, error)
	GetAllProjects(ctx context.Context, ownerID string, page, pageSize int) ([]models.Project, int64, error)
	UpdateProject(ctx context.Context, referenceID string, ownerID string, updates map[string]interface{}) (*models.Project, error)
	DeleteProject(ctx context.Context, referenceID string, ownerID string) error
	GetProjectByRefOrID(ctx context.Context, param string) (*models.Project, error)
}

type projectService struct {
	projectRepo           repo.ProjectRepository
	projectFeatureService ProjectFeatureService
	analyticsService      AnalyticsService
	projectUsageService   ProjectUsageService
	db                    *gorm.DB
	validate              *validator.Validate
}

func NewProjectService(
	projectRepo repo.ProjectRepository,
	projectFeatureService ProjectFeatureService,
	analyticsService AnalyticsService,
	projectUsageService ProjectUsageService,
	db *gorm.DB,
) ProjectService {
	return &projectService{
		projectRepo:           projectRepo,
		projectFeatureService: projectFeatureService,
		analyticsService:      analyticsService,
		projectUsageService:   projectUsageService,
		db:                    db,
		validate:              validator.New(),
	}
}

// CreateProject: Kani waa function-ka ugu muhiimsan ee abuura mashruuca iyo Limits-kiisa
func (s *projectService) CreateProject(ctx context.Context, ownerID string, name string) (*models.Project, error) {
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	// 1. Hel ID-ga "Free" Plan-ka si aan limits-ka u ogaano
	var freePlan models.Plan
	if err := tx.Where("LOWER(name) = ?", "free").First(&freePlan).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("default 'Free' plan not found. please create it in admin panel first")
	}

	projectSlug := utils.GenerateSlug(name)
	projectID := uuid.New().String()

	project := &models.Project{
		ID:          projectID,
		ReferenceID: fmt.Sprintf("%s-%s", projectSlug, utils.GenerateRandomString(6)),
		OwnerID:     ownerID,
		Name:        name,
		Slug:        projectSlug,
		PlanID:      freePlan.ID, // Ku xir Free Plan
		Active:      true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 2. Keydi Project-ka
	if err := s.projectRepo.CreateProject(ctx, tx, project); err != nil {
		tx.Rollback()
		return nil, err
	}

	// 3. SEED FEATURES (Auth, Storage, etc.)
	if _, err := s.projectFeatureService.GenerateDefaultFeaturesWithTx(ctx, tx, project.ID); err != nil {
		tx.Rollback()
		return nil, err
	}

	// 4. CREATE INITIAL ANALYTICS (Charts)
	if err := s.analyticsService.CreateInitialAnalyticsForProject(ctx, tx, project.ID); err != nil {
		tx.Rollback()
		return nil, err
	}

	// ✅ 5. CREATE USAGE RECORD WITH LIMIT SNAPSHOT
	// Halkan ayaan ku koobiyeeynaa Limits-ka Plan-ka kuna shubaynaa ProjectUsage
	// ✅ 2. Snapshot Limits (Koobiyeey xadka plan-ka)
	usage := &models.ProjectUsage{
		ID:                 uuid.New(),
		ProjectID:          project.ID,
		LimitApiCalls:      freePlan.MaxApiCalls,
		LimitAuthUsers:     freePlan.MaxAuthUsers,
		LimitStorageMB:     freePlan.MaxStorageMB,
		LimitDocuments:     freePlan.MaxDocuments,
		LimitNotifications: freePlan.MaxNotifications, // 👈 Halkan ayay xogtii ka soo gashay Plan-ka
	}
	if err := tx.Create(usage).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create usage record with limits: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return project, nil
}

func (s *projectService) GetProjectByReferenceID(ctx context.Context, referenceID string, ownerID string) (*models.Project, error) {
	project, err := s.projectRepo.GetProjectByAnyIDAndOwner(ctx, referenceID, ownerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("project not found")
		}
		return nil, err
	}
	return project, nil
}

func (s *projectService) GetAllProjects(ctx context.Context, ownerID string, page, pageSize int) ([]models.Project, int64, error) {
	return s.projectRepo.GetAllProjects(ctx, ownerID, page, pageSize)
}

func (s *projectService) UpdateProject(ctx context.Context, referenceID string, ownerID string, updates map[string]interface{}) (*models.Project, error) {
	project, err := s.projectRepo.GetProjectByReferenceIDAndOwner(ctx, referenceID, ownerID)
	if err != nil {
		return nil, err
	}
	if err := utils.ApplyUpdates(project, updates); err != nil {
		return nil, err
	}
	project.UpdatedAt = time.Now()
	if err := s.projectRepo.UpdateProject(ctx, project); err != nil {
		return nil, err
	}
	return project, nil
}

func (s *projectService) DeleteProject(ctx context.Context, idOrRef string, ownerID string) error {
	// 1. Marka hore soo saar UUID-ga dhabta ah ee project-ka
	project, err := s.projectRepo.GetProjectByRefOrID(ctx, idOrRef)
	if err != nil {
		return errors.New("project not found")
	}

	// 2. Hubi in qofka tirtiraya uu yahay owner-ka (Security check)
	if project.OwnerID != ownerID {
		return errors.New("unauthorized: you are not the owner of this project")
	}

	// 3. U dir Repo-ga si uu "Clean Wipe" ugu sameeyo
	return s.projectRepo.DeleteProject(ctx, project.ID, ownerID)
}

func (s *projectService) GetProjectByInternalID(ctx context.Context, id string) (*models.Project, error) {
	return s.projectRepo.GetProjectByInternalID(ctx, id)
}

func (s *projectService) GetProjectByRefOrID(ctx context.Context, param string) (*models.Project, error) {
	project, err := s.projectRepo.GetProjectByRefOrID(ctx, param)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("project not found")
		}
		return nil, err
	}
	return project, nil
}
