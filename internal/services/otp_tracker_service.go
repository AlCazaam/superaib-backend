package services

import (
	"context"
	"fmt"
	"superaib/internal/models"
	"superaib/internal/storage/repo"
	"time"

	"github.com/google/uuid"
)

type OtpTrackerService interface {
	CheckAndTrack(ctx context.Context, projectID string, userID uuid.UUID) error
	ResetLimit(ctx context.Context, userID uuid.UUID) error
}

type otpTrackerService struct {
	trackerRepo repo.OtpTrackerRepository
	policyRepo  repo.RateLimitPolicyRepository
	userRepo    repo.AuthUserRepository
}

func NewOtpTrackerService(tr repo.OtpTrackerRepository, pr repo.RateLimitPolicyRepository, ur repo.AuthUserRepository) OtpTrackerService {
	return &otpTrackerService{trackerRepo: tr, policyRepo: pr, userRepo: ur}
}

func (s *otpTrackerService) CheckAndTrack(ctx context.Context, projectID string, userID uuid.UUID) error {
	// 1. Soo qaado Policy-ga. Haddii uusan jirin ama uu dansan yahay, ha xadidin.
	policy, err := s.policyRepo.GetByProjectID(ctx, projectID)
	if err != nil || policy == nil || !policy.IsEnabled {
		fmt.Printf("⚠️ Rate Limit Policy not found or disabled for Project: %s\n", projectID)
		return nil
	}

	// 2. Soo qaado Tracker-ka isticmaalaha
	tracker, err := s.trackerRepo.GetByUserID(ctx, userID)
	if err != nil {
		tracker = &models.OtpUserTracker{
			ID: uuid.New(), UserID: userID, ProjectID: projectID,
			WindowStartsAt: time.Now(), RequestCount: 0,
		}
	}

	// 3. AUTO-UNLOCK LOGIC: Haddii waqtigii xirnaanshiyaha uu dhamaaday
	if tracker.LockoutUntil != nil {
		if time.Now().After(*tracker.LockoutUntil) {
			// Waqtigu waa dhamaaday - Fur user-ka (Set to Active)
			tracker.LockoutUntil = nil
			tracker.RequestCount = 0
			tracker.WindowStartsAt = time.Now()

			user, _ := s.userRepo.GetByID(ctx, userID)
			if user != nil {
				user.Status = models.AuthUserActive
				_ = s.userRepo.Update(ctx, user)
			}
		} else {
			// Wali waa xiran yahay
			remaining := time.Until(*tracker.LockoutUntil).Minutes()
			return fmt.Errorf("limit reached: account locked for %.0f more minutes", remaining)
		}
	}

	// 4. Reset Window: Haddii waqtigii loo qabtay (tusaale 24h) uu dhaafay
	if time.Now().Sub(tracker.WindowStartsAt) > time.Duration(policy.WindowMinutes)*time.Minute {
		tracker.RequestCount = 0
		tracker.WindowStartsAt = time.Now()
	}

	// 5. Kordhi Tirada (Increment)
	tracker.RequestCount++
	fmt.Printf("📈 OTP Attempt: %d/%d for User: %s\n", tracker.RequestCount, policy.MaxRequests, userID)

	// 6. AUTO-BLOCK LOGIC: Haddii xadkii la dhaafay
	if tracker.RequestCount > policy.MaxRequests {
		lockoutDuration := time.Now().Add(time.Duration(policy.LockoutMinutes) * time.Minute)
		tracker.LockoutUntil = &lockoutDuration

		// Iska xir user-ka asalkiisa (Status = Blocked)
		user, _ := s.userRepo.GetByID(ctx, userID)
		if user != nil {
			user.Status = models.AuthUserBlocked
			_ = s.userRepo.Update(ctx, user)
		}

		_ = s.trackerRepo.Upsert(ctx, tracker)
		return fmt.Errorf("limit reached: account locked for %d minutes", policy.LockoutMinutes)
	}

	return s.trackerRepo.Upsert(ctx, tracker)
}

func (s *otpTrackerService) ResetLimit(ctx context.Context, userID uuid.UUID) error {
	user, _ := s.userRepo.GetByID(ctx, userID)
	if user != nil {
		user.Status = models.AuthUserActive
		_ = s.userRepo.Update(ctx, user)
	}
	return s.trackerRepo.ResetTracker(ctx, userID)
}
