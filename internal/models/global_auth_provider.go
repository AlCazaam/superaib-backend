package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type GlobalAuthProvider struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Name           string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
	Title          string         `gorm:"type:varchar(255);not null" json:"title"`
	Description    string         `gorm:"type:text" json:"description"`
	IconURL        string         `gorm:"type:text" json:"icon_url"`
	RequiredFields datatypes.JSON `gorm:"type:jsonb;default:'[]'" json:"required_fields"` // e.g. ["client_id", "secret"]
	IsActive       bool           `gorm:"default:true" json:"is_active"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

func (m *GlobalAuthProvider) BeforeCreate(tx *gorm.DB) (err error) {
	m.ID = uuid.New()
	return
}
