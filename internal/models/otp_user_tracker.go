package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OtpUserTracker struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"` // User-ka aan xisaabinayno
	ProjectID string    `gorm:"type:uuid;index;not null" json:"project_id"`

	RequestCount   int        `gorm:"default:0" json:"request_count"` // Immisa jeer ayuu codsaday?
	WindowStartsAt time.Time  `json:"window_starts_at"`               // Goorta tirintu bilaabatay
	LockoutUntil   *time.Time `json:"lockout_until"`                  // Goorta uu xor noqonayo

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (t *OtpUserTracker) BeforeCreate(tx *gorm.DB) (err error) {
	t.ID = uuid.New()
	return
}
