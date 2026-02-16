package repo

import (
	"context"
	"errors"
	"superaib/internal/models"
	"time"

	"gorm.io/gorm"
)

type ImpersonationRepository interface {
	Create(ctx context.Context, token *models.ImpersonationToken) error
	GetByTokenString(ctx context.Context, tokenString string) (*models.ImpersonationToken, error)
	Revoke(ctx context.Context, id string) error

	// ✅ CUSUB
	GetByID(ctx context.Context, id string) (*models.ImpersonationToken, error)
	Update(ctx context.Context, token *models.ImpersonationToken) error
	GetActiveTokensByUser(ctx context.Context, userID string) ([]models.ImpersonationToken, error)
}

type gormImpersonationRepo struct {
	db *gorm.DB
}

func NewImpersonationRepository(db *gorm.DB) ImpersonationRepository {
	return &gormImpersonationRepo{db: db}
}

func (r *gormImpersonationRepo) Create(ctx context.Context, token *models.ImpersonationToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

// GetByTokenString: Wuxuu hubinayaa in token-ku jiro oo uusan weli dhicin (Expired)
func (r *gormImpersonationRepo) GetByTokenString(ctx context.Context, tokenString string) (*models.ImpersonationToken, error) {
	var token models.ImpersonationToken

	// Raadi token-ka, hubi inuusan tirtirnayn (DeletedAt) iyo in waqtigiisu uusan dhicin
	err := r.db.WithContext(ctx).Preload("User").
		Where("token = ? AND expires_at > ?", tokenString, time.Now()).
		First(&token).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("token invalid or expired")
		}
		return nil, err
	}
	return &token, nil
}

func (r *gormImpersonationRepo) Revoke(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.ImpersonationToken{}, "id = ?", id).Error
}

// Implementation:
func (r *gormImpersonationRepo) GetByID(ctx context.Context, id string) (*models.ImpersonationToken, error) {
	var token models.ImpersonationToken
	err := r.db.WithContext(ctx).First(&token, "id = ?", id).Error
	return &token, err
}

func (r *gormImpersonationRepo) Update(ctx context.Context, token *models.ImpersonationToken) error {
	return r.db.WithContext(ctx).Save(token).Error
}

func (r *gormImpersonationRepo) GetActiveTokensByUser(ctx context.Context, userID string) ([]models.ImpersonationToken, error) {
	var tokens []models.ImpersonationToken
	// Soo qaado kuwa aan dhicin (ExpiresAt > Now)
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND expires_at > ?", userID, time.Now()).
		Find(&tokens).Error
	return tokens, err
}
