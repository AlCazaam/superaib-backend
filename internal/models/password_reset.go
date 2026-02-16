package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PasswordResetToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ProjectID string    `gorm:"type:uuid;index;not null" json:"project_id"`

	UserID uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	Email  string    `gorm:"type:varchar(255);not null" json:"email"`

	Code  string `gorm:"type:varchar(6);not null" json:"code"` // OTP 6-lambar ah
	Token string `gorm:"type:text" json:"token"`               // Token-ka sirta ah ee la isticmaalo marka la verify-gareeyo

	ExpiresAt time.Time `json:"expires_at"`
	IsUsed    bool      `gorm:"default:false" json:"is_used"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (t *PasswordResetToken) BeforeCreate(tx *gorm.DB) (err error) {
	t.ID = uuid.New()
	return
}
