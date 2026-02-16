package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Project struct {
	ID          string `gorm:"type:uuid;primaryKey" json:"id"`
	ReferenceID string `gorm:"type:varchar(100);uniqueIndex;not null" json:"reference_id"`
	OwnerID     string `gorm:"type:uuid;index;not null" json:"owner_id"`
	Name        string `gorm:"type:varchar(255);not null" json:"name"`
	Slug        string `gorm:"type:varchar(255);uniqueIndex;not null" json:"slug"`
	Active      bool   `gorm:"default:true" json:"active"`

	// ✅ QAYBTA CUSUB: Isku xirka Plan-ka
	PlanID uuid.UUID `gorm:"type:uuid;index" json:"plan_id"`
	Plan   Plan      `gorm:"foreignKey:PlanID" json:"plan"` // Si aan u ogaano limits-ka

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	APIKeys         []APIKey         `gorm:"foreignKey:ProjectID" json:"api_keys,omitempty"`
	ProjectFeatures []ProjectFeature `gorm:"foreignKey:ProjectID" json:"project_features,omitempty"`
}
