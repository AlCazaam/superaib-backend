package services

import (
	"context"
	"errors"
	"fmt"
	"superaib/internal/core/config"
	"superaib/internal/core/logger"
	"superaib/internal/core/security"
	"superaib/internal/models"
	"superaib/internal/storage/repo"
	"time"

	"github.com/go-playground/validator/v10"
)

type AuthService interface {
	Register(ctx context.Context, req *models.UserCreateRequest) (*models.User, string, error)
	Login(ctx context.Context, req *models.UserLoginRequest) (*models.User, string, error)
}

type authService struct {
	userRepo  repo.UserRepository
	validator *validator.Validate
	cfg       *config.Config
}

func NewAuthService(userRepo repo.UserRepository, cfg *config.Config) AuthService {
	return &authService{
		userRepo:  userRepo,
		validator: validator.New(),
		cfg:       cfg,
	}
}

func (s *authService) Register(ctx context.Context, req *models.UserCreateRequest) (*models.User, string, error) {
	if err := s.validator.Struct(req); err != nil {
		return nil, "", fmt.Errorf("validation failed: %w", err)
	}

	if _, err := s.userRepo.GetUserByEmail(ctx, req.Email); err == nil {
		return nil, "", errors.New("email already registered")
	}
	if _, err := s.userRepo.GetUserByUsername(ctx, req.Username); err == nil {
		return nil, "", errors.New("username already taken")
	}

	hashedPassword, err := security.HashPassword(req.Password)
	if err != nil {
		return nil, "", errors.New("failed to process password")
	}

	user := &models.User{
		Name:         req.Name,
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Role:         models.RoleDeveloper,
		Status:       models.StatusPending,
	}

	createdUser, err := s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, "", fmt.Errorf("service failed to create user: %w", err)
	}

	token, err := security.GenerateToken(createdUser.ID.String(), string(createdUser.Role))
	if err != nil {
		return nil, "", errors.New("failed to generate token")
	}

	return createdUser, token, nil
}

func (s *authService) Login(ctx context.Context, req *models.UserLoginRequest) (*models.User, string, error) {
	if err := s.validator.Struct(req); err != nil {
		return nil, "", fmt.Errorf("validation failed: %w", err)
	}

	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, "", errors.New("invalid credentials")
	}

	if err := security.VerifyPassword(user.PasswordHash, req.Password); err != nil {
		return nil, "", errors.New("invalid credentials")
	}

	token, err := security.GenerateToken(user.ID.String(), string(user.Role))
	if err != nil {
		return nil, "", errors.New("failed to generate token")
	}

	// 🚀 XALKA QALADKA: Cusboonaysii LastLoginAt adigoo isticmaalaya Map
	now := time.Now()
	updateData := map[string]interface{}{
		"last_login_at": now,
	}

	// Wac UpdateUser adigoo u dhiibaya ID-ga iyo Map-ka
	_, _ = s.userRepo.UpdateUser(ctx, user.ID, updateData)

	logger.Log.WithField("userID", user.ID).Info("User logged in successfully.")
	return user, token, nil
}
