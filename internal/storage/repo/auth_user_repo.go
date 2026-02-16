package repo

import (
	"context"
	"errors"
	"fmt"
	"superaib/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthUserRepository interface {
	Create(ctx context.Context, user *models.AuthUser) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.AuthUser, error)
	GetByEmailAndProject(ctx context.Context, email, projectID string) (*models.AuthUser, error)
	// ✅ NEW: Raadi user-ka isagoo isticmaalaya Provider ID-ga (tusaale: Facebook ID)
	GetByProviderID(ctx context.Context, provider, providerID string) (*models.AuthUser, error)

	GetAllByProject(ctx context.Context, projectID string) ([]models.AuthUser, error)
	Update(ctx context.Context, user *models.AuthUser) error
	Delete(ctx context.Context, projectID string, id uuid.UUID) error
}

type GormAuthUserRepository struct {
	db *gorm.DB
}

func NewAuthUserRepository(db *gorm.DB) AuthUserRepository {
	return &GormAuthUserRepository{db: db}
}

// Create: Kaydi isticmaale cusub
func (r *GormAuthUserRepository) Create(ctx context.Context, user *models.AuthUser) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetByID: Raadi user adigoo isticmaalaya UUID
func (r *GormAuthUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.AuthUser, error) {
	var u models.AuthUser
	if err := r.db.WithContext(ctx).First(&u, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

// GetByEmailAndProject: Aad u muhiim ah xilliga Login-ka
func (r *GormAuthUserRepository) GetByEmailAndProject(ctx context.Context, email, projectID string) (*models.AuthUser, error) {
	var u models.AuthUser
	err := r.db.WithContext(ctx).Where("email = ? AND project_id = ?", email, projectID).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GetAllByProject: Liiska users-ka ee Dashboard-ka Developer-ka
func (r *GormAuthUserRepository) GetAllByProject(ctx context.Context, projectID string) ([]models.AuthUser, error) {
	var users []models.AuthUser
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("created_at DESC").
		Find(&users).Error
	return users, err
}

// Update: Bedel xogta user-ka
func (r *GormAuthUserRepository) Update(ctx context.Context, user *models.AuthUser) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// Delete: Tirtir user-ka (Iska hubi inuu ka tirsan yahay project-kaas)
func (r *GormAuthUserRepository) Delete(ctx context.Context, projectID string, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Where("project_id = ?", projectID).Delete(&models.AuthUser{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("auth user not found in this project")
	}
	return nil
}

// ✅ IMPLEMENTATION: GetByProviderID
func (r *GormAuthUserRepository) GetByProviderID(ctx context.Context, provider, providerID string) (*models.AuthUser, error) {
	var user models.AuthUser
	// ✅ SAX: Isticmaal column-ka "o_auth_providers" oo loogu talagalay Postgres JSONB search
	query := fmt.Sprintf(`o_auth_providers @> '{"%s": "%s"}'`, provider, providerID)
	err := r.db.WithContext(ctx).Where(query).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
