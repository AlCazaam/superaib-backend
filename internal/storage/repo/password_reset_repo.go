package repo

import (
	"context"
	"errors"
	"superaib/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PasswordResetRepository interface {
	Create(ctx context.Context, token *models.PasswordResetToken) error
	GetByEmailAndCode(ctx context.Context, email, code, projectID string) (*models.PasswordResetToken, error)
	GetByToken(ctx context.Context, tokenString string) (*models.PasswordResetToken, error)
	// ✅ CUSUB: Kani waa kii maqnaa ee Service-ku raadinayay
	Update(ctx context.Context, token *models.PasswordResetToken) error
	MarkAsUsed(ctx context.Context, id string) error
	// ✅ NEW: Soo qaado dhamaan reset token-ada User-ka gaarka ah
	GetHistoryByUserID(ctx context.Context, userID uuid.UUID) ([]models.PasswordResetToken, error)
	// ✅ NEW: Delete token-ka
	DeleteByID(ctx context.Context, tokenID uuid.UUID) error
}

type gormPasswordResetRepo struct {
	db *gorm.DB
}

func NewPasswordResetRepository(db *gorm.DB) PasswordResetRepository {
	return &gormPasswordResetRepo{db: db}
}

func (r *gormPasswordResetRepo) Create(ctx context.Context, token *models.PasswordResetToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

// GetByEmailAndCode: Raadi OTP-ga aan dhicin (ExpiresAt > Now) oo aan la isticmaalin (IsUsed = false)
func (r *gormPasswordResetRepo) GetByEmailAndCode(ctx context.Context, email, code, projectID string) (*models.PasswordResetToken, error) {
	var token models.PasswordResetToken
	err := r.db.WithContext(ctx).
		Where("email = ? AND code = ? AND project_id = ? AND is_used = ? AND expires_at > ?", email, code, projectID, false, time.Now()).
		First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid or expired OTP")
		}
		return nil, err
	}
	return &token, nil
}

// GetByToken: Raadi Token-ka kumeel-gaarka ah marka la rabo in password-ka la bedelo
func (r *gormPasswordResetRepo) GetByToken(ctx context.Context, tokenString string) (*models.PasswordResetToken, error) {
	var token models.PasswordResetToken
	err := r.db.WithContext(ctx).
		Where("token = ? AND is_used = ? AND expires_at > ?", tokenString, false, time.Now()).
		First(&token).Error
	if err != nil {
		return nil, errors.New("invalid or expired reset token")
	}
	return &token, nil
}

// ✅ IMPLEMENTATION: Function-ka Update
func (r *gormPasswordResetRepo) Update(ctx context.Context, token *models.PasswordResetToken) error {
	// Save wuxuu cusboonaysiiyaa record-ka jira haddii ID-gu jiro
	return r.db.WithContext(ctx).Save(token).Error
}

func (r *gormPasswordResetRepo) MarkAsUsed(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&models.PasswordResetToken{}).Where("id = ?", id).Update("is_used", true).Error
}

// ✅ IMPLEMENTATION: GetHistoryByUserID
func (r *gormPasswordResetRepo) GetHistoryByUserID(ctx context.Context, userID uuid.UUID) ([]models.PasswordResetToken, error) {
	var tokens []models.PasswordResetToken
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&tokens).Error
	return tokens, err
}

// ✅ IMPLEMENTATION: DeleteByID
func (r *gormPasswordResetRepo) DeleteByID(ctx context.Context, tokenID uuid.UUID) error {
	res := r.db.WithContext(ctx).Delete(&models.PasswordResetToken{}, "id = ?", tokenID)
	if res.RowsAffected == 0 {
		return errors.New("reset token not found")
	}
	return res.Error
}
