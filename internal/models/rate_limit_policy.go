package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RateLimitPolicy: Waa xeerka Rate Limiting-ka ee Project kasta
type RateLimitPolicy struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ProjectID string    `gorm:"type:uuid;uniqueIndex:idx_project_policy;not null" json:"project_id"`

	// Shuruucda Asaasiga ah
	MaxRequests    int  `gorm:"default:3;not null" json:"max_requests"`
	WindowMinutes  int  `gorm:"default:1440;not null" json:"window_minutes"` // 24 hours
	LockoutMinutes int  `gorm:"default:30;not null" json:"lockout_minutes"`
	IsEnabled      bool `gorm:"default:true;not null" json:"is_enabled"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (p *RateLimitPolicy) BeforeCreate(tx *gorm.DB) (err error) {
	p.ID = uuid.New()
	return
}
