package services

import (
	"context"
	"errors"
	"fmt"
	"superaib/internal/core/security"
	"superaib/internal/models"
	"superaib/internal/storage/repo"
	"time"

	"github.com/google/uuid"
)

type ImpersonationService interface {
	GenerateToken(ctx context.Context, projectID string, userID uuid.UUID, durationMinutes int) (*models.ImpersonationToken, error)
	ValidateToken(ctx context.Context, tokenString string) (*models.AuthUser, error)
	// ✅ CUSUB
	RevokeToken(ctx context.Context, tokenID string) error
	ExtendToken(ctx context.Context, tokenID string, extraMinutes int) (*models.ImpersonationToken, error)
	GetActiveTokens(ctx context.Context, userID string) ([]models.ImpersonationToken, error)
	// ✅ CUSUB: Kani waa kii maqnaa ee Interface-ka lagu xirayo
	GetTokenDetails(ctx context.Context, tokenString string) (*models.ImpersonationToken, error)
}

type impersonationService struct {
	repo     repo.ImpersonationRepository
	userRepo repo.AuthUserRepository
}

func NewImpersonationService(r repo.ImpersonationRepository, ur repo.AuthUserRepository) ImpersonationService {
	return &impersonationService{repo: r, userRepo: ur}
}

func (s *impersonationService) GenerateToken(ctx context.Context, projectID string, userID uuid.UUID, durationMinutes int) (*models.ImpersonationToken, error) {
	// 1. Hubi in User-ku jiro
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if user.ProjectID != projectID {
		return nil, errors.New("user does not belong to this project")
	}

	// 2. Dhali JWT Token dhab ah
	// Waxaan siinaynaa "auth_user" role si uu u galo App-ka
	jwtString, err := security.GenerateToken(user.ID.String(), "auth_user")
	if err != nil {
		return nil, fmt.Errorf("jwt generation failed: %w", err)
	}

	// 3. Xisaabi waqtiga uu dhacayo (Default: 60 daqiiqo haddii eber la soo diro)
	if durationMinutes <= 0 {
		durationMinutes = 60
	}
	expiresAt := time.Now().Add(time.Duration(durationMinutes) * time.Minute)

	// 4. Kaydi Database-ka
	impToken := &models.ImpersonationToken{
		ProjectID: projectID,
		UserID:    userID,
		Token:     jwtString,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, impToken); err != nil {
		return nil, err
	}

	// Soo celi record-ka oo wata User-ka
	impToken.User = *user
	return impToken, nil
}

// ValidateToken: Kani waa function aad mustaqbalka u isticmaali karto haddii aad rabto inaad token-ka hubiso
func (s *impersonationService) ValidateToken(ctx context.Context, tokenString string) (*models.AuthUser, error) {
	record, err := s.repo.GetByTokenString(ctx, tokenString)
	if err != nil {
		return nil, err
	}
	return &record.User, nil
}

func (s *impersonationService) RevokeToken(ctx context.Context, tokenID string) error {
	// Si toos ah u tirtir (Soft Delete)
	return s.repo.Revoke(ctx, tokenID)
}

func (s *impersonationService) ExtendToken(ctx context.Context, tokenID string, extraMinutes int) (*models.ImpersonationToken, error) {
	token, err := s.repo.GetByID(ctx, tokenID)
	if err != nil {
		return nil, errors.New("token not found")
	}

	// Kordhi waqtiga
	if extraMinutes <= 0 {
		extraMinutes = 60
	}
	token.ExpiresAt = token.ExpiresAt.Add(time.Duration(extraMinutes) * time.Minute)

	if err := s.repo.Update(ctx, token); err != nil {
		return nil, err
	}
	return token, nil
}

func (s *impersonationService) GetActiveTokens(ctx context.Context, userID string) ([]models.ImpersonationToken, error) {
	return s.repo.GetActiveTokensByUser(ctx, userID)
}

func (s *impersonationService) GetTokenDetails(ctx context.Context, tokenString string) (*models.ImpersonationToken, error) {
	// 1. Raadi Token-ka Database-ka
	tokenRecord, err := s.repo.GetByTokenString(ctx, tokenString)
	if err != nil {
		return nil, errors.New("token not found or revoked") // Haddii la waayo ama la revok-gareeyey
	}

	// 2. Hubi Expired
	if time.Now().After(tokenRecord.ExpiresAt) {
		return nil, errors.New("token has expired")
	}

	// 3. Hubi Signature-ka JWT (Security check)
	// Tani waxay xaqiijinaysaa inuusan qofna gacanta ku bedelin token-ka
	_, err = security.ValidateJWT(tokenString)
	if err != nil {
		return nil, errors.New("invalid token signature")
	}

	return tokenRecord, nil
}
