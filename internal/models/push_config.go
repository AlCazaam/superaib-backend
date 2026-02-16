package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProjectPushConfig struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ProjectID   string    `gorm:"type:uuid;uniqueIndex;not null" json:"project_id"`
	ServiceJson string    `gorm:"type:text;not null" json:"service_json"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (p *ProjectPushConfig) BeforeCreate(tx *gorm.DB) (err error) {
	p.ID = uuid.New()
	return
}
