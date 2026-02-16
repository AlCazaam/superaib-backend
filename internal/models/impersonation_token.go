package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ImpersonationToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ProjectID string    `gorm:"type:uuid;index;not null" json:"project_id"`
	UserID    uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`

	// Token-ka dhabta ah (JWT)
	Token string `gorm:"type:text;not null" json:"token"`

	// Waqtiga uu dhacayo (Expiry)
	ExpiresAt time.Time `json:"expires_at"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Relation: Wuxuu ku xiran yahay AuthUser
	User AuthUser `gorm:"foreignKey:UserID" json:"user"`
}

func (t *ImpersonationToken) BeforeCreate(tx *gorm.DB) (err error) {
	t.ID = uuid.New()
	return
}
