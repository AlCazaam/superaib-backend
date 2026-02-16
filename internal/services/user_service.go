package services

import (
	"context"
	"errors"
	"fmt"
	"superaib/internal/core/logger"
	"superaib/internal/models"
	"superaib/internal/storage/repo"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetAllUsers(ctx context.Context, limit, offset int) ([]models.User, error)
	UpdateUser(ctx context.Context, id string, req *models.UserUpdateRequest) (*models.User, error)
	DeleteUser(ctx context.Context, id string) error

	// ✅ KAN KU DAR SI HANDLER-KU U ARKO
	DeleteAccount(ctx context.Context, id string) error
}

type userService struct {
	userRepo  repo.UserRepository
	validator *validator.Validate
}

func NewUserService(userRepo repo.UserRepository) UserService {
	return &userService{
		userRepo:  userRepo,
		validator: validator.New(),
	}
}

func (s *userService) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	userUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}
	return s.userRepo.GetUserByID(ctx, userUUID)
}

func (s *userService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return s.userRepo.GetUserByEmail(ctx, email)
}

func (s *userService) GetAllUsers(ctx context.Context, limit, offset int) ([]models.User, error) {
	return s.userRepo.GetAllUsers(ctx, limit, offset)
}

// 🚀 LOGIC-GA CUSUB: UpdateUser (Kani ayaa xalinaya dhibkaaga)
func (s *userService) UpdateUser(ctx context.Context, id string, req *models.UserUpdateRequest) (*models.User, error) {
	userUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	// 1. Dhis Map xogta cusub sita (Si GORM u dareemo isbeddelka)
	updateData := make(map[string]interface{})

	if req.Name != nil {
		updateData["name"] = *req.Name
	}
	if req.Username != nil {
		updateData["username"] = *req.Username
	}
	if req.Bio != nil {
		updateData["bio"] = *req.Bio
	}
	if req.Location != nil {
		updateData["location"] = *req.Location
	}
	if req.PhoneNumber != nil {
		updateData["phone_number"] = *req.PhoneNumber
	}
	if req.Website != nil {
		updateData["website"] = *req.Website
	}
	if req.GithubURL != nil {
		updateData["github_url"] = *req.GithubURL
	}
	if req.TwitterURL != nil {
		updateData["twitter_url"] = *req.TwitterURL
	}
	if req.LinkedinURL != nil {
		updateData["linkedin_url"] = *req.LinkedinURL
	}
	if req.PortfolioURL != nil {
		updateData["portfolio_url"] = *req.PortfolioURL
	}
	if req.ProfileImageURL != nil {
		updateData["profile_image_url"] = *req.ProfileImageURL
	}

	// 2. Haddii Password la soo diray, hash gareey
	if req.Password != nil && *req.Password != "" {
		hashed, _ := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		updateData["password_hash"] = string(hashed)
	}

	// 3. U dir Repository-ga (Kii aan Map-ka ka dhignay)
	updatedUser, err := s.userRepo.UpdateUser(ctx, userUUID, updateData)
	if err != nil {
		return nil, err
	}

	logger.Log.WithField("userID", id).Info("User profile fields updated successfully.")
	return updatedUser, nil
}

func (s *userService) DeleteUser(ctx context.Context, id string) error {
	userUUID, err := uuid.Parse(id)
	if err != nil {
		return errors.New("invalid user ID format")
	}
	return s.userRepo.DeleteUser(ctx, userUUID)
}

// Ku dar interface-ka dhexdiisa:
// DeleteAccount(ctx context.Context, id string) error

func (s *userService) DeleteAccount(ctx context.Context, id string) error {
	userUUID, err := uuid.Parse(id)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	// Wac Repo-ga weyn ee Global Wipe-ka ahaa
	if err := s.userRepo.WipeUserAccount(ctx, userUUID); err != nil {
		logger.Log.Errorf("Global wipe failed for user %s: %v", id, err)
		return fmt.Errorf("failed to wipe user account data: %w", err)
	}

	logger.Log.WithField("userID", id).Info("Developer account and all project data wiped successfully.")
	return nil
}
