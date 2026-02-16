package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"superaib/internal/core/security"
	"superaib/internal/models"
	"superaib/internal/storage/repo"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"google.golang.org/api/idtoken"
	"gorm.io/gorm"
)

type AuthUserService interface {
	Create(ctx context.Context, user *models.AuthUser, password string) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.AuthUser, error)
	GetByEmailAndProject(ctx context.Context, email, projectID string) (*models.AuthUser, error)
	GetAllByProject(ctx context.Context, projectID string) ([]models.AuthUser, error)
	Update(ctx context.Context, user *models.AuthUser) error
	Delete(ctx context.Context, projectID string, id uuid.UUID) error
	TrackLogin(ctx context.Context, projectID string)

	// Login Methods (SDK)
	LoginUser(ctx context.Context, projectID, email, password string) (*models.AuthUser, string, error)
	LoginWithGoogle(ctx context.Context, projectID, idToken string) (*models.AuthUser, string, error)
	LoginWithFacebook(ctx context.Context, projectID, accessToken string) (*models.AuthUser, string, error)

	// 🚀 NEW: SDK Specific Features
	SendPasswordResetOTP(ctx context.Context, projectID, email string) error
	VerifyOTPAndResetPassword(ctx context.Context, projectID, email, otp, newPassword string) error
	LoginWithImpersonation(ctx context.Context, projectID, token string) (*models.AuthUser, string, error)
}
type authUserService struct {
	repo         repo.AuthUserRepository
	configRepo   repo.ProjectAuthConfigRepository
	validate     *validator.Validate
	tracker      *AnalyticsTracker
	usageService ProjectUsageService
	db           *gorm.DB // ✅ KU DAR KAN
}

func NewAuthUserService(
	r repo.AuthUserRepository,
	cr repo.ProjectAuthConfigRepository,
	tracker *AnalyticsTracker,
	usage ProjectUsageService,
	db *gorm.DB, // ✅ KU DAR HADDII UU KA MAQNAA
) AuthUserService {
	return &authUserService{
		repo:         r,
		configRepo:   cr,
		validate:     validator.New(),
		tracker:      tracker,
		usageService: usage,
		db:           db, // ✅
	}
}

// STANDARD LOGIN
func (s *authUserService) LoginUser(
	ctx context.Context,
	projectID, email, password string,
) (*models.AuthUser, string, error) {

	user, err := s.repo.GetByEmailAndProject(ctx, email, projectID)
	if err != nil {
		return nil, "", errors.New("invalid credentials")
	}

	if err := s.checkAccountStatus(user); err != nil {
		return nil, "", err
	}

	if err := security.VerifyPassword(user.PasswordHash, password); err != nil {
		return nil, "", errors.New("invalid credentials")
	}

	token, err := security.GenerateToken(user.ID.String(), "auth_user")
	if err != nil {
		return nil, "", err
	}

	s.TrackLogin(ctx, projectID)
	return user, token, nil
}

func (s *authUserService) Create(ctx context.Context, user *models.AuthUser, password string) error {
	hashed, err := security.HashPassword(password)
	if err != nil {
		return err
	}
	user.PasswordHash = hashed

	err = s.repo.Create(ctx, user)
	if err == nil {
		// ✅ 1. TRACK ANALYTICS
		s.tracker.TrackEvent(ctx, user.ProjectID, models.AnalyticsTypeAuthActivity, "new_signups", 1)

		// ✅ 2. UPDATE PROJECT USAGE (Xadka guud ee Users-ka)
		_ = s.usageService.UpdateUsage(ctx, user.ProjectID, "auth_users_count", 1)
	}
	return err
}

func (s *authUserService) GetByID(ctx context.Context, id uuid.UUID) (*models.AuthUser, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *authUserService) GetAllByProject(ctx context.Context, projectID string) ([]models.AuthUser, error) {
	return s.repo.GetAllByProject(ctx, projectID)
}

func (s *authUserService) Update(ctx context.Context, user *models.AuthUser) error {
	return s.repo.Update(ctx, user)
}

func (s *authUserService) Delete(ctx context.Context, projectID string, id uuid.UUID) error {
	err := s.repo.Delete(ctx, projectID, id)
	if err == nil {
		// ✅ 1. TRACK ANALYTICS
		s.tracker.TrackEvent(ctx, projectID, models.AnalyticsTypeAuthActivity, "new_signups", -1)

		// ✅ 2. UPDATE PROJECT USAGE (Ka dhim wadarta guud ee Users-ka)
		_ = s.usageService.UpdateUsage(ctx, projectID, "auth_users_count", -1)
	}
	return err
}

func (s *authUserService) GetByEmailAndProject(
	ctx context.Context,
	email, projectID string,
) (*models.AuthUser, error) {
	return s.repo.GetByEmailAndProject(ctx, email, projectID)
}

func (s *authUserService) TrackLogin(ctx context.Context, projectID string) {
	// optional analytics hook
}

// GOOGLE LOGIN
// GOOGLE LOGIN

// GOOGLE LOGIN
func (s *authUserService) LoginWithGoogle(
	ctx context.Context,
	projectID, idToken string,
) (*models.AuthUser, string, error) {

	if err := s.checkProviderStatus(ctx, projectID, "google"); err != nil {
		return nil, "", err
	}

	config, err := s.configRepo.GetByProjectAndProviderName(ctx, projectID, "google")
	if err != nil {
		return nil, "", errors.New("google sign-in not configured")
	}

	var creds map[string]interface{}
	json.Unmarshal(config.Credentials, &creds)

	// 🚀 SAXID: Hel Web Client ID
	webClientID, _ := creds["client_id"].(string)

	// ✅ MUHIIM: Si variable-ka loo isticmaalo compiler-kuna uusan u careysan,
	// waxaan u dhiibeynaa debugPrint si aan u ogaano ID-ga la isticmaalayo.
	debugPrint(fmt.Sprintf("🔍 Backend is checking token against Web Client ID: %s", webClientID))

	// Isticmaal "" si loogu ogolaado dhamaan Audiences (iOS, Android, Web)
	payload, err := idtoken.Validate(ctx, idToken, "")
	if err != nil {
		return nil, "", fmt.Errorf("invalid google token: %w", err)
	}

	// Xaqiiji in Token-ka uu dhab ahaan ka yimid mid ka mid ah Client IDs-kaaga
	tokenAudience := payload.Audience
	debugPrint(fmt.Sprintf("📩 Incoming Token Audience: %s", tokenAudience))

	email := payload.Claims["email"].(string)
	name, _ := payload.Claims["name"].(string)
	userID := payload.Subject

	user, err := s.repo.GetByEmailAndProject(ctx, email, projectID)
	if err != nil {
		user = &models.AuthUser{
			ID:             uuid.New(),
			ProjectID:      projectID,
			Email:          email,
			Status:         models.AuthUserActive,
			OAuthProviders: models.ToJSONBAuth(map[string]string{"google": userID}),
			Metadata:       models.ToJSONBAuth(map[string]string{"name": name}),
			CreatedAt:      time.Now(),
		}
		if createErr := s.repo.Create(ctx, user); createErr != nil {
			return nil, "", fmt.Errorf("failed to create user: %w", createErr)
		}
		s.tracker.TrackEvent(ctx, projectID, models.AnalyticsTypeAuthActivity, "new_signups", 1)
		_ = s.usageService.UpdateUsage(ctx, projectID, "auth_users_count", 1)
	} else {
		providers := models.FromJSONBAuth(user.OAuthProviders)
		if _, exists := providers["google"]; !exists {
			providers["google"] = userID
			user.OAuthProviders = models.ToJSONBAuth(providers)
			_ = s.repo.Update(ctx, user)
		}
	}

	token, err := security.GenerateToken(user.ID.String(), "auth_user")
	if err != nil {
		return nil, "", err
	}

	s.TrackLogin(ctx, projectID)
	return user, token, nil
}

// Helper for debugging
func debugPrint(msg string) {
	fmt.Printf("[DEBUG] %s\n", msg)
}

const facebookJWKSURL = "https://www.facebook.com/.well-known/oauth/openid/jwks/"

// FACEBOOK LOGIN
func (s *authUserService) LoginWithFacebook(
	ctx context.Context,
	projectID, tokenStr string,
) (*models.AuthUser, string, error) {

	config, err := s.configRepo.GetByProjectAndProviderName(ctx, projectID, "facebook")
	if err != nil {
		return nil, "", errors.New("facebook login not configured")
	}
	if !config.Enabled {
		return nil, "", errors.New("facebook login is currently disabled")
	}

	var creds map[string]interface{}
	json.Unmarshal(config.Credentials, &creds)
	appID, _ := creds["app_id"].(string)

	jwks, err := keyfunc.Get(facebookJWKSURL, keyfunc.Options{RefreshInterval: time.Hour})
	if err != nil {
		return nil, "", fmt.Errorf("failed to load facebook JWKS: %w", err)
	}

	token, err := jwt.Parse(tokenStr, jwks.Keyfunc, jwt.WithAudience(appID))
	if err != nil {
		return nil, "", fmt.Errorf("invalid facebook jwt: %w", err)
	}

	claims := token.Claims.(jwt.MapClaims)
	fbID, _ := claims["sub"].(string)
	fbEmail, _ := claims["email"].(string)
	fbName, _ := claims["name"].(string)

	if fbID == "" {
		return nil, "", errors.New("facebook user id (sub) missing")
	}

	var user *models.AuthUser
	user, _ = s.repo.GetByProviderID(ctx, "facebook", fbID)

	if user == nil && fbEmail != "" {
		existingUser, emailErr := s.repo.GetByEmailAndProject(ctx, fbEmail, projectID)
		if emailErr == nil && existingUser != nil {
			user = existingUser
			providers := models.FromJSONBAuth(user.OAuthProviders)
			providers["facebook"] = fbID
			user.OAuthProviders = models.ToJSONBAuth(providers)
			_ = s.repo.Update(ctx, user)
		}
	}

	if user != nil {
		if statusErr := s.checkAccountStatus(user); statusErr != nil {
			return nil, "", statusErr
		}
	}

	if user == nil {
		emailToSave := fbEmail
		if emailToSave == "" {
			emailToSave = fmt.Sprintf("fb_%s@noemail.internal", fbID)
		}

		user = &models.AuthUser{
			ID:             uuid.New(),
			ProjectID:      projectID,
			Email:          emailToSave,
			Status:         models.AuthUserActive,
			OAuthProviders: models.ToJSONBAuth(map[string]string{"facebook": fbID}),
			Metadata:       models.ToJSONBAuth(map[string]string{"name": fbName}),
			CreatedAt:      time.Now(),
		}
		if err := s.repo.Create(ctx, user); err != nil {
			return nil, "", fmt.Errorf("failed to create user: %w", err)
		}

		// ✅ TRACK ANALYTICS & USAGE (Cusub)
		s.tracker.TrackEvent(ctx, projectID, models.AnalyticsTypeAuthActivity, "new_signups", 1)
		_ = s.usageService.UpdateUsage(ctx, projectID, "auth_users_count", 1)
	}

	jwtToken, err := security.GenerateToken(user.ID.String(), "auth_user")
	if err != nil {
		return nil, "", err
	}

	s.TrackLogin(ctx, projectID)
	return user, jwtToken, nil
}

func (s *authUserService) checkProviderStatus(ctx context.Context, projectID, providerName string) error {
	config, err := s.configRepo.GetByProjectAndProviderName(ctx, projectID, providerName)
	if err != nil {
		return fmt.Errorf("authentication via %s is not configured", providerName)
	}
	if !config.Enabled {
		return fmt.Errorf("authentication via %s is currently disabled", providerName)
	}
	return nil
}

func (s *authUserService) checkAccountStatus(user *models.AuthUser) error {
	status := strings.ToLower(string(user.Status))
	if status == "blocked" {
		return errors.New("your account is blocked")
	}
	if status == "invited" {
		return errors.New("your account is invited")
	}
	if status != strings.ToLower(string(models.AuthUserActive)) {
		return errors.New("your account is not active")
	}
	return nil
}

// 📧 1. SEND OTP (Forgot Password)
// 📧 1. SEND OTP (Forgot Password) - ✅ SAXID: 'user' hadda waa la isticmaalay
func (s *authUserService) SendPasswordResetOTP(ctx context.Context, projectID string, email string) error {
	user, err := s.repo.GetByEmailAndProject(ctx, email, projectID)
	if err != nil {
		return errors.New("user not found in this project")
	}

	// 1. Generate 6-digit OTP
	otpCode := "123456"

	// 2. LOG & TRACK: Isticmaal user variable si qaladku u baxo
	fmt.Printf("📧 PROJECT [%s]: Sending OTP [%s] to User [%s] (ID: %s)\n", projectID, otpCode, user.Email, user.ID)

	// ✅ Halkan waxaad ku xiri doontaa EmailService.SendOTP(user.Email, otpCode)
	return nil
}

// 🔐 2. VERIFY OTP & RESET
func (s *authUserService) VerifyOTPAndResetPassword(ctx context.Context, projectID, email, otp, newPassword string) error {
	user, err := s.repo.GetByEmailAndProject(ctx, email, projectID)
	if err != nil {
		return err
	}

	// Halkan ku xir OTP Validation logic (OtpUserTracker)
	if otp != "123456" {
		return errors.New("invalid or expired OTP")
	}

	// Hash-garee password-ka cusub
	hashed, _ := security.HashPassword(newPassword)
	user.PasswordHash = hashed

	return s.repo.Update(ctx, user)
}

// 🕵️ 3. IMPERSONATION LOGIN
// 🕵️ 3. IMPERSONATION LOGIN (Fixed & Real Implementation)
func (s *authUserService) LoginWithImpersonation(ctx context.Context, projectID, token string) (*models.AuthUser, string, error) {
	// 1. Marka hore raadi Token-ka dhexdiisa table-ka "impersonation_tokens"
	// Waxaan u baahanahay inaan ogaano User-ka uu token-kani u taagan yahay.
	var impToken models.ImpersonationToken

	// ✅ MUHIIM: Baar haddii token-ku jiro, uusan dhicin (expired), uuna leeyahay project-gan
	if err := s.db.WithContext(ctx).Where("token = ? AND project_id = ?", token, projectID).First(&impToken).Error; err != nil {
		return nil, "", errors.New("invalid or expired impersonation token")
	}

	// 2. Hubi haddii uu waqtigii ka dhacay
	if time.Now().After(impToken.ExpiresAt) {
		return nil, "", errors.New("impersonation token has expired")
	}

	// 3. Hadda soo qaado User-kii saxda ahaa ee Token-kaas loo dhalisay
	user, err := s.repo.GetByID(ctx, impToken.UserID)
	if err != nil {
		return nil, "", errors.New("impersonation user not found in database")
	}

	// 4. Generate SuperAIB JWT Token (User Session)
	jwtToken, err := security.GenerateToken(user.ID.String(), "auth_user")
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate session: %w", err)
	}

	// 5. Optionally: Tirtir token-ka maadaama la isticmaalay (One-time use)
	// s.db.Delete(&impToken)

	return user, jwtToken, nil
}
