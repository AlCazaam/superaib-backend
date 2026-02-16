package services

import (
	"context"
	"fmt"
	"superaib/internal/models"
	"superaib/internal/storage/repo"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SubscriptionService interface {
	UpgradeProject(ctx context.Context, req UpgradeRequest) error
}

type UpgradeRequest struct {
	ProjectID   string
	DeveloperID string
	PlanID      string
	WaafiID     string
	Amount      float64
	PhoneNumber string
}

type subscriptionService struct {
	db    *gorm.DB
	tRepo repo.TransactionRepository
	pRepo repo.ProjectRepository
}

func NewSubscriptionService(db *gorm.DB, tr repo.TransactionRepository, pr repo.ProjectRepository) SubscriptionService {
	return &subscriptionService{db: db, tRepo: tr, pRepo: pr}
}

// UpgradeProject: Function-ka rasmiga ah ee maamula Lacag bixinta iyo Limits-ka cusub
func (s *subscriptionService) UpgradeProject(ctx context.Context, req UpgradeRequest) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Hel Project-ka hadda jira
		var project models.Project
		if err := tx.Where("id = ?", req.ProjectID).First(&project).Error; err != nil {
			return fmt.Errorf("project not found")
		}

		// 2. Hel xogta Plan-ka cusub ee uu rabo inuu iibsado
		var nextPlan models.Plan
		if err := tx.Where("id = ?", req.PlanID).First(&nextPlan).Error; err != nil {
			return fmt.Errorf("selected plan not found")
		}

		// ✅ XALKA: FREE PLAN GUARD
		if nextPlan.Price <= 0 {
			// Haddii uu qofku horey u iibsaday Pro/Enterprise, diid inuu Free ku laabto.
			if project.PlanID != uuid.Nil {
				return fmt.Errorf("free plan is only for new projects. you cannot downgrade back to free after upgrading")
			}
		}

		// 3. Create Transaction Record (Cadeynta lacagta)
		transaction := &models.Transaction{
			ProjectID:     req.ProjectID,
			DeveloperID:   req.DeveloperID,
			PlanID:        nextPlan.ID,
			Amount:        req.Amount,
			PhoneNumber:   req.PhoneNumber,
			TransactionID: req.WaafiID,
			Status:        "completed",
		}
		if err := s.tRepo.Create(ctx, tx, transaction); err != nil {
			return err
		}

		// ✅ 4. UPDATE LIMITS (Snapshot logic)
		// Halkan ayaan ku daray labadii maqnaa si ay ugu dhacaan Database-ka (project_usages)
		err := tx.Model(&models.ProjectUsage{}).
			Where("project_id = ?", req.ProjectID).
			Updates(map[string]interface{}{
				"limit_api_calls":         nextPlan.MaxApiCalls,
				"limit_auth_users":        nextPlan.MaxAuthUsers,
				"limit_storage_mb":        nextPlan.MaxStorageMB,
				"limit_documents":         nextPlan.MaxDocuments,
				"limit_notifications":     nextPlan.MaxNotifications,
				"limit_realtime_channels": nextPlan.MaxRealtimeChannels, // 🚀 CUSUB (Fixed)
				"limit_realtime_events":   nextPlan.MaxRealtimeEvents,   // 🚀 CUSUB (Fixed)
				"updated_at":              gorm.Expr("NOW()"),
			}).Error
		if err != nil {
			return err
		}

		// 5. Link project to the new plan
		return tx.Model(&project).Update("plan_id", nextPlan.ID).Error
	})
}
