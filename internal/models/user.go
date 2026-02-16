package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRole defines the roles a user can have
type UserRole string

const (
	RoleDeveloper UserRole = "developer"
	RoleAdmin     UserRole = "admin"
	RoleSupport   UserRole = "support"
)

// UserStatus defines the status of a user account
type UserStatus string

const (
	StatusActive    UserStatus = "active"
	StatusSuspended UserStatus = "suspended"
	StatusDeleted   UserStatus = "deleted"
	StatusPending   UserStatus = "pending"
)

// AccountPlan defines the subscription plan of a user
type AccountPlan string

const (
	PlanFree       AccountPlan = "free"
	PlanPro        AccountPlan = "pro"
	PlanEnterprise AccountPlan = "enterprise"
)

// ThemePreference defines the UI theme preference
type ThemePreference string

const (
	ThemeLight ThemePreference = "light"
	ThemeDark  ThemePreference = "dark"
	ThemeAuto  ThemePreference = "auto"
)

// User represents the developer model in the system
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()" json:"id"`
	Name         string    `gorm:"type:varchar(255);not null" json:"name" validate:"required"`
	Username     string    `gorm:"type:varchar(255);unique;not null" json:"username" validate:"required,alphanum,min=3,max=30"`
	Email        string    `gorm:"type:varchar(255);unique;not null" json:"email" validate:"required,email"`
	PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"` // Exclude from JSON output

	ProfileImageURL *string `gorm:"type:text" json:"profile_image_url"`
	Bio             *string `gorm:"type:text" json:"bio"`
	Website         *string `gorm:"type:text" json:"website"`
	GithubURL       *string `gorm:"type:text" json:"github_url"`
	TwitterURL      *string `gorm:"type:text" json:"twitter_url"`
	LinkedinURL     *string `gorm:"type:text" json:"linkedin_url"`
	PortfolioURL    *string `gorm:"type:text" json:"portfolio_url"`
	Location        *string `gorm:"type:varchar(255)" json:"location"`
	PhoneNumber     *string `gorm:"type:varchar(50)" json:"phone_number"`

	Verified          bool            `gorm:"default:false" json:"verified"`
	VerificationToken *string         `gorm:"type:varchar(255)" json:"-"` // Hidden
	TwoFactorEnabled  bool            `gorm:"default:false" json:"two_factor_enabled"`
	TwoFactorSecret   *string         `gorm:"type:varchar(255)" json:"-"`                    // Hidden
	RecoveryCodes     json.RawMessage `gorm:"type:jsonb;default:'{}'" json:"recovery_codes"` // Hidden, manage via specific endpoints

	Role                    UserRole        `gorm:"type:user_role;default:'developer';not null" json:"role"`
	Status                  UserStatus      `gorm:"type:user_status;default:'pending';not null" json:"status"` // Changed default to pending for verification flow
	PreferredLanguage       string          `gorm:"type:varchar(10);default:'en'" json:"preferred_language"`
	Timezone                string          `gorm:"type:varchar(50);default:'UTC'" json:"timezone"`
	NotificationPreferences json.RawMessage `gorm:"type:jsonb;default:'{}'" json:"notification_preferences"`
	ThemePreference         ThemePreference `gorm:"type:theme_preference;default:'auto';not null" json:"theme_preference"`
	AccountPlan             AccountPlan     `gorm:"type:account_plan;default:'free';not null" json:"account_plan"`

	ProjectsCount int        `gorm:"default:0" json:"projects_count"`
	LastLoginIP   *string    `gorm:"type:varchar(45)" json:"last_login_ip"`
	LastLoginAt   *time.Time `json:"last_login_at"`

	CreatedAt time.Time      `gorm:"default:now();not null" json:"created_at"`
	UpdatedAt time.Time      `gorm:"default:now();not null" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // GORM soft delete
}

// UserCreateRequest is used for creating new users (registration)
type UserCreateRequest struct {
	Name     string   `json:"name" validate:"required"`
	Username string   `json:"username" validate:"required,alphanum,min=3,max=30"`
	Email    string   `json:"email" validate:"required,email"`
	Password string   `json:"password" validate:"required,min=8,max=72"` // Max for bcrypt
	Role     UserRole `json:"role"`                                      // Optional, defaults to developer
}

// UserLoginRequest for login
type UserLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// UserUpdateRequest for update
type UserUpdateRequest struct {
	Name                    *string          `json:"name"`
	Username                *string          `json:"username" validate:"omitempty,alphanum,min=3,max=30"`
	Email                   *string          `json:"email" validate:"omitempty,email"`
	Password                *string          `json:"password" validate:"omitempty,min=8,max=72"` // Optional password change
	ProfileImageURL         *string          `json:"profile_image_url"`
	Bio                     *string          `json:"bio"`
	Website                 *string          `json:"website"`
	GithubURL               *string          `json:"github_url"`
	TwitterURL              *string          `json:"twitter_url"`
	LinkedinURL             *string          `json:"linkedin_url"`
	PortfolioURL            *string          `json:"portfolio_url"`
	Location                *string          `json:"location"`
	PhoneNumber             *string          `json:"phone_number"`
	Status                  *UserStatus      `json:"status"` // Admin-only updates
	PreferredLanguage       *string          `json:"preferred_language"`
	Timezone                *string          `json:"timezone"`
	NotificationPreferences json.RawMessage  `json:"notification_preferences"`
	ThemePreference         *ThemePreference `json:"theme_preference"`
	AccountPlan             *AccountPlan     `json:"account_plan"`       // Admin-only updates
	Role                    *UserRole        `json:"role"`               // Admin-only updates
	Verified                *bool            `json:"verified"`           // Admin-only updates
	TwoFactorEnabled        *bool            `json:"two_factor_enabled"` // Admin-only updates, or specific endpoint
}

// ToPointer is a helper to get a pointer to a value
func ToPointer[T any](v T) *T {
	return &v
}

// Ensure User implements json.Marshaller to hide sensitive fields during general API calls
func (u User) MarshalJSON() ([]byte, error) {
	type Alias User // Create an alias to avoid infinite recursion
	return json.Marshal(&struct {
		Alias
		PasswordHash      string          `json:"-"` // Explicitly hide
		VerificationToken *string         `json:"-"`
		TwoFactorSecret   *string         `json:"-"`
		RecoveryCodes     json.RawMessage `json:"-"`
	}{
		Alias: Alias(u),
	})
}
