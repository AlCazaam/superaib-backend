package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// DeviceToken Model
type DeviceToken struct {
	ID        string `gorm:"type:uuid;primaryKey" json:"id"`
	ProjectID string `gorm:"type:uuid;index;not null" json:"project_id"`
	UserID    string `gorm:"type:uuid;uniqueIndex:idx_user_project;not null" json:"user_id"`
	Token     string `gorm:"not null" json:"token"`
	Platform  string `gorm:"type:varchar(20)" json:"platform"`
	// 🚀 XALKA CUSUB:
	Enabled   bool      `gorm:"default:true" json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
}

func (d *DeviceToken) BeforeCreate(tx *gorm.DB) (err error) {
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	return
}

// Notification Model
type Notification struct {
	ID          string         `gorm:"type:uuid;primaryKey" json:"id,omitempty"`
	ProjectID   string         `gorm:"type:uuid;index;not null" json:"project_id"`
	Title       string         `json:"title"`
	Body        string         `json:"body"`
	ImageURL    string         `json:"image_url,omitempty"`
	DeepLink    string         `json:"deep_link,omitempty"`
	CustomData  datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"custom_data,omitempty"`
	Target      string         `gorm:"type:varchar(50);default:'all'" json:"target"`
	UserID      *string        `gorm:"type:uuid;index" json:"user_id,omitempty"`
	Platform    string         `gorm:"type:varchar(20);default:'all'" json:"platform"`
	IsScheduled bool           `gorm:"default:false" json:"is_scheduled"`

	// 🚀 XALKA WAQTIYADA:
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Status      string     `gorm:"type:varchar(20);default:'pending'" json:"status"`
	SentCount   int        `json:"sent_count"`
}

func (n *Notification) BeforeCreate(tx *gorm.DB) (err error) {
	// Generate UUID haddii uu madhan yahay
	if n.ID == "" || n.ID == "null" {
		n.ID = uuid.New().String()
	}
	return
}
