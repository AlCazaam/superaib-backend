package repo

import (
	"context"
	"strings"
	"superaib/internal/models"

	"gorm.io/gorm"
)

type ProjectRepository interface {
	CreateProject(ctx context.Context, db *gorm.DB, project *models.Project) error
	GetProjectByID(ctx context.Context, id string, ownerID string) (*models.Project, error)
	GetProjectByInternalID(ctx context.Context, id string) (*models.Project, error)
	GetProjectByReferenceIDAndOwner(ctx context.Context, refID string, ownerID string) (*models.Project, error)
	GetProjectByAnyIDAndOwner(ctx context.Context, id string, ownerID string) (*models.Project, error)
	GetAllProjects(ctx context.Context, ownerID string, page, pageSize int) ([]models.Project, int64, error)
	UpdateProject(ctx context.Context, project *models.Project) error
	DeleteProject(ctx context.Context, refID string, ownerID string) error
	GetProjectByRefOrID(ctx context.Context, identifier string) (*models.Project, error)
	// ✅ KAN KU DAR SI SERVICE-KU U ARKO
	GetByID(ctx context.Context, id string) (*models.Project, error)
}

type GormProjectRepository struct {
	db *gorm.DB
}

func NewGormProjectRepository(db *gorm.DB) *GormProjectRepository {
	return &GormProjectRepository{db: db}
}

// 1. GetProjectByRefOrID - Kani waa kan Middleware-ka SDK-gu isticmaalo (MUHIIM!)
func (r *GormProjectRepository) GetProjectByRefOrID(ctx context.Context, identifier string) (*models.Project, error) {
	var project models.Project

	// ✅ SAX: Waxaan Preload ku samaynay Plan-ka si Middleware-ku u arko MaxStorage, MaxApiCalls, iwm.
	query := r.db.WithContext(ctx).Preload("Plan").Preload("ProjectFeatures")

	if len(identifier) == 36 && strings.Contains(identifier, "-") {
		// Haddii uu yahay UUID (ID)
		err := query.Where("id = ? OR reference_id = ?", identifier, identifier).First(&project).Error
		return &project, err
	} else {
		// Haddii uu yahay Slug (ReferenceID)
		err := query.Where("reference_id = ?", identifier).First(&project).Error
		return &project, err
	}
}

// 2. GetProjectByAnyIDAndOwner - Kani waxaa isticmaala Dashboard-ka
func (r *GormProjectRepository) GetProjectByAnyIDAndOwner(ctx context.Context, idOrRef string, ownerID string) (*models.Project, error) {
	var project models.Project
	// ✅ Preload Plan si Dashboard-ka loogu arko plan-ka uu haysto
	err := r.db.WithContext(ctx).Preload("Plan").
		Where("(id = ? OR reference_id = ?) AND owner_id = ?", idOrRef, idOrRef, ownerID).First(&project).Error
	return &project, err
}

// 3. GetAllProjects - Liiska mashaariicda Developer-ka
func (r *GormProjectRepository) GetAllProjects(ctx context.Context, ownerID string, page, pageSize int) ([]models.Project, int64, error) {
	var projects []models.Project
	var total int64
	offset := (page - 1) * pageSize

	dbWithCtx := r.db.WithContext(ctx).Model(&models.Project{}).Where("owner_id = ?", ownerID)
	dbWithCtx.Count(&total)

	// ✅ Preload Plan
	err := dbWithCtx.Preload("Plan").Preload("ProjectFeatures").
		Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&projects).Error
	return projects, total, err
}

// 4. CreateProject
func (r *GormProjectRepository) CreateProject(ctx context.Context, db *gorm.DB, project *models.Project) error {
	// Waxaan u isticmaalaynaa "db" parameter-ka haddii uu Transaction ku dhex jiro
	return db.WithContext(ctx).Create(project).Error
}

// 5. GetProjectByID
func (r *GormProjectRepository) GetProjectByID(ctx context.Context, id string, ownerID string) (*models.Project, error) {
	var project models.Project
	if err := r.db.WithContext(ctx).Preload("Plan").Preload("ProjectFeatures").Where("id = ? AND owner_id = ?", id, ownerID).First(&project).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

// 6. GetProjectByInternalID
func (r *GormProjectRepository) GetProjectByInternalID(ctx context.Context, id string) (*models.Project, error) {
	var project models.Project
	if err := r.db.WithContext(ctx).Preload("Plan").Preload("ProjectFeatures").First(&project, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

// 7. GetProjectByReferenceIDAndOwner
func (r *GormProjectRepository) GetProjectByReferenceIDAndOwner(ctx context.Context, refID string, ownerID string) (*models.Project, error) {
	var project models.Project
	if err := r.db.WithContext(ctx).Preload("Plan").Preload("ProjectFeatures").Where("reference_id = ? AND owner_id = ?", refID, ownerID).First(&project).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

// 8. UpdateProject
func (r *GormProjectRepository) UpdateProject(ctx context.Context, project *models.Project) error {
	return r.db.WithContext(ctx).Where("id = ? AND owner_id = ?", project.ID, project.OwnerID).Save(project).Error
}

// 9. DeleteProject
func (r *GormProjectRepository) DeleteProject(ctx context.Context, projectID string, ownerID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Hubi inuu qofku iska leeyahay project-ka
		var project models.Project
		if err := tx.Where("id = ? AND owner_id = ?", projectID, ownerID).First(&project).Error; err != nil {
			return err
		}

		// 2. TIRTIR XOGTA DHAMAAN TABLES-KA KU XIRAN (Cascade Delete Logic)

		// Tirtir API Keys
		tx.Where("project_id = ?", projectID).Delete(&models.APIKey{})

		// Tirtir Usage Stats
		tx.Where("project_id = ?", projectID).Delete(&models.ProjectUsage{})

		// Tirtir Analytics
		tx.Where("project_id = ?", projectID).Delete(&models.Analytics{})

		// Tirtir Documents (NoSQL data)
		tx.Where("project_id = ?", projectID).Delete(&models.Document{})

		// Tirtir Collections
		tx.Where("project_id = ?", projectID).Delete(&models.Collection{})

		// Tirtir Realtime Channels iyo Events
		// Fiiro gaar ah: Mararka qaar halkan waxaad u baahan kartaa inaad marka hore Events tirtirto
		var channels []models.RealtimeChannel
		tx.Where("project_id = ?", projectID).Find(&channels)
		for _, ch := range channels {
			tx.Where("channel_id = ?", ch.ID).Delete(&models.RealtimeEvent{})
		}
		tx.Where("project_id = ?", projectID).Delete(&models.RealtimeChannel{})

		// Tirtir Project Features settings
		tx.Where("project_id = ?", projectID).Delete(&models.ProjectFeature{})

		// Tirtir Transactions (Haddii aad rabto inaad hayso, waad iska dhaafi kartaa)
		tx.Where("project_id = ?", projectID).Delete(&models.Transaction{})

		// 3. Ugu dambeyntii, tirtir Project-ka laftiisa
		if err := tx.Delete(&project).Error; err != nil {
			return err
		}

		return nil
	})
}

// ✅ IMPLEMENTATION-KA GetByID (Kani waa kan xalinaya qaladka)
func (r *GormProjectRepository) GetByID(ctx context.Context, id string) (*models.Project, error) {
	var project models.Project
	err := r.db.WithContext(ctx).Preload("Plan").First(&project, "id = ?", id).Error
	return &project, err
}
