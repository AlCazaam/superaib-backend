package services

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"superaib/internal/core/security"
	"superaib/internal/models"
	"superaib/internal/storage/repo"
	"time"

	"github.com/google/uuid"
)

type PasswordResetService interface {
	RequestReset(ctx context.Context, projectID, email string) error
	VerifyOTP(ctx context.Context, projectID, email, otp string) (string, error)
	ConfirmReset(ctx context.Context, resetToken, newPassword string) error
	GetResetHistory(ctx context.Context, userID uuid.UUID) ([]models.PasswordResetToken, error)
	DeleteToken(ctx context.Context, tokenID uuid.UUID) error
}

type passwordResetService struct {
	resetRepo    repo.PasswordResetRepository
	userRepo     repo.AuthUserRepository
	configRepo   repo.ProjectAuthConfigRepository
	emailService EmailService
	otpTracker   OtpTrackerService
}

func NewPasswordResetService(rr repo.PasswordResetRepository, ur repo.AuthUserRepository, cr repo.ProjectAuthConfigRepository, es EmailService, ot OtpTrackerService) PasswordResetService {
	return &passwordResetService{resetRepo: rr, userRepo: ur, configRepo: cr, emailService: es, otpTracker: ot}
}

func (s *passwordResetService) RequestReset(ctx context.Context, projectID, email string) error {
	user, err := s.userRepo.GetByEmailAndProject(ctx, email, projectID)
	if err != nil {
		return nil // Silent return ammaanka awgiis
	}

	// ✅ TALLAABADA 1: RATE LIMIT CHECK (Ugu soo horeysii)
	// Tani waxay xaqiijinaysaa in isku daygaaga la xisaabiyo xataa haddii SMTP fashilmo.
	if err := s.otpTracker.CheckAndTrack(ctx, projectID, user.ID); err != nil {
		return err
	}

	// ✅ TALLAABADA 2: SMTP STATUS CHECK
	if err := s.checkSmtpStatus(ctx, projectID); err != nil {
		return err
	}

	// ✅ TALLAABADA 3: USER ACCOUNT STATUS CHECK
	if err := s.checkAccountStatus(user); err != nil {
		return err
	}

	// 4. Dhali OTP
	rand.Seed(time.Now().UnixNano())
	otp := fmt.Sprintf("%06d", rand.Intn(1000000))

	token := &models.PasswordResetToken{
		ProjectID: projectID, UserID: user.ID, Email: email, Code: otp,
		ExpiresAt: time.Now().Add(15 * time.Minute), IsUsed: false,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}

	if err := s.resetRepo.Create(ctx, token); err != nil {
		return err
	}

	// 5. Dir Email (Async)
	go func() {
		subject := "Reset Your Password"
		body := fmt.Sprintf("Your OTP code is: <b>%s</b>", otp)
		_ = s.emailService.SendEmail(context.Background(), projectID, email, subject, body)
	}()

	return nil
}

func (s *passwordResetService) checkSmtpStatus(ctx context.Context, projectID string) error {
	// MAGACA DATABASE-KA KU QORAN WAA "smtp_email" EE MA AHA "smtp"
	config, err := s.configRepo.GetByProjectAndProviderName(ctx, projectID, "smtp_email")
	if err != nil {
		return errors.New("email service is not configured")
	}
	if !config.Enabled {
		return errors.New("email service is currently disabled")
	}
	return nil
}
func (s *passwordResetService) checkAccountStatus(user *models.AuthUser) error {
	status := strings.ToLower(string(user.Status))
	if status == "blocked" {
		return errors.New("your account is blocked")
	}
	if status == "invited" {
		return errors.New("your account is invited")
	}
	return nil
}

// Implementations for VerifyOTP, ConfirmReset, etc.
func (s *passwordResetService) VerifyOTP(ctx context.Context, projectID, email, otp string) (string, error) {
	record, err := s.resetRepo.GetByEmailAndCode(ctx, email, otp, projectID)
	if err != nil {
		return "", errors.New("invalid or expired otp")
	}
	magicToken := uuid.New().String()
	record.Token = magicToken
	record.ExpiresAt = time.Now().Add(5 * time.Minute)
	s.resetRepo.Update(ctx, record)
	return magicToken, nil
}

func (s *passwordResetService) ConfirmReset(ctx context.Context, resetToken, newPassword string) error {
	record, err := s.resetRepo.GetByToken(ctx, resetToken)
	if err != nil {
		return errors.New("invalid or expired reset token")
	}
	user, _ := s.userRepo.GetByID(ctx, record.UserID)
	hashed, _ := security.HashPassword(newPassword)
	user.PasswordHash = hashed
	s.userRepo.Update(ctx, user)
	return s.resetRepo.MarkAsUsed(ctx, record.ID.String())
}

func (s *passwordResetService) GetResetHistory(ctx context.Context, userID uuid.UUID) ([]models.PasswordResetToken, error) {
	return s.resetRepo.GetHistoryByUserID(ctx, userID)
}

func (s *passwordResetService) DeleteToken(ctx context.Context, tokenID uuid.UUID) error {
	return s.resetRepo.DeleteByID(ctx, tokenID)
}
