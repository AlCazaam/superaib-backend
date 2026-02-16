package repo

import (
	"context"
	"errors"
	"fmt"
	"superaib/internal/core/logger"
	"superaib/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	GetAllUsers(ctx context.Context, limit, offset int) ([]models.User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, data map[string]interface{}) (*models.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error

	// ✅ KAN AYAA MAQNAA: Ku dar interface-ka
	WipeUserAccount(ctx context.Context, userID uuid.UUID) error
}

type GormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) UserRepository {
	return &GormUserRepository{db: db}
}

func (r *GormUserRepository) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		logger.Log.Errorf("Error creating user: %v", err)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return user, nil
}

func (r *GormUserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// 🚀 LOGIC-GA SAXDA AH EE UPDATE-KA:
// Waxaan isticmaalaynaa Updates(map) si GORM u beddelo kaliya waxa Flutter-ka ka yimid
func (r *GormUserRepository) UpdateUser(ctx context.Context, id uuid.UUID, data map[string]interface{}) (*models.User, error) {
	var user models.User

	// 1. Marka hore soo hel User-ka si aan u xaqiijino inuu jiro
	if err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}

	// 2. Ku samee Update xogta cusub (Updates method-ka Map wuu u fiican yahay)
	if err := r.db.WithContext(ctx).Model(&user).Updates(data).Error; err != nil {
		logger.Log.Errorf("Error updating user fields for %s: %v", id, err)
		return nil, fmt.Errorf("failed to update user fields: %w", err)
	}

	// 3. Dib u soo aqri User-ka si loogu celiyo xogta dhamaystiran ee saxda ah
	r.db.WithContext(ctx).First(&user, "id = ?", id)
	return &user, nil
}

func (r *GormUserRepository) SaveUser(ctx context.Context, user *models.User) (*models.User, error) {
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (r *GormUserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).First(&user, "email = ?", email).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).First(&user, "username = ?", username).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) GetAllUsers(ctx context.Context, limit, offset int) ([]models.User, error) {
	var users []models.User
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&users).Error
	return users, err
}

func (r *GormUserRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.User{}, "id = ?", id).Error
}

func (r *GormUserRepository) WipeUserAccount(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		userIDStr := userID.String()

		// 1. Hel dhammaan ID-yada mashaariicda uu leeyahay developer-ku
		var projectIDs []string
		if err := tx.Table("projects").Where("owner_id = ?", userIDStr).Pluck("id", &projectIDs).Error; err != nil {
			return err
		}

		// --- THE NUCLEAR OPTION: Disable Foreign Key Checks ---
		// Tani waxay u sheegaysaa PostgreSQL inuusan hubin FK mudada transaction-ku socdo
		tx.Exec("SET LOCAL session_replication_role = 'replica'")

		if len(projectIDs) > 0 {
			// --- A. NESTED TABLES ---
			// 1. Realtime Events (Waxay ku xiran yihiin Channel)
			tx.Exec("DELETE FROM realtime_events WHERE channel_id IN (SELECT id FROM realtime_channels WHERE project_id IN ?)", projectIDs)

			// 2. Documents (Waxay ku xiran yihiin Collection)
			tx.Exec("DELETE FROM documents WHERE collection_id IN (SELECT id FROM collections WHERE project_id IN ?)", projectIDs)

			// --- B. DIRECT TABLES ---
			// Halkan waxaa laga saaray "transactions" si xogtaas ay u harto
			directTables := []string{
				"realtime_channels",
				"collections",
				"api_keys",
				"project_usages",
				"analytics",
				"project_features",
				"notifications",
				"storage_files",
				"project_push_configs",
				"project_auth_configs",
				"rate_limit_policies",
				"device_tokens",
				"impersonation_tokens",
				"otp_user_trackers",
				"password_reset_tokens",
				"auth_users",
			}

			for _, table := range directTables {
				// Waxaan isticmaalaynaa Exec si aan u hubino Hard Delete
				if err := tx.Exec(fmt.Sprintf("DELETE FROM %s WHERE project_id IN ?", table), projectIDs).Error; err != nil {
					// Haddii table uusan jirin ama uu error dhaco, waan iska dhaafaynaa
					continue
				}
			}

			// --- C. DELETE PROJECTS ---
			// Hadda oo dhamaan xogta kale la sifeeyay, project-ga tirtir
			if err := tx.Exec("DELETE FROM projects WHERE owner_id = ?", userIDStr).Error; err != nil {
				return err
			}
		}

		// --- D. FINAL USER WIPE ---
		// Ugu dambeyn tirtir Developer-ka dhabta ah
		if err := tx.Exec("DELETE FROM users WHERE id = ?", userID).Error; err != nil {
			return err
		}

		// --- Restore Foreign Key Checks ---
		tx.Exec("SET LOCAL session_replication_role = 'origin'")

		return nil
	})
}
