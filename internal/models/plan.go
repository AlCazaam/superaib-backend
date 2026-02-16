package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Plan struct {
	ID                  uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name                string    `gorm:"type:varchar(100);unique;not null" json:"name"`
	Price               float64   `gorm:"type:decimal(10,2);default:0" json:"price"`
	MaxAuthUsers        int       `gorm:"default:0" json:"max_auth_users"`
	MaxStorageMB        float64   `gorm:"type:decimal(10,2);default:0" json:"max_storage_mb"`
	MaxApiCalls         int       `gorm:"default:0" json:"max_api_calls"`
	MaxDocuments        int       `gorm:"default:0" json:"max_documents"`
	MaxNotifications    int       `gorm:"default:0" json:"max_notifications"`
	MaxRealtimeChannels int       `gorm:"default:0" json:"max_realtime_channels"` // Xadka qolalka
	MaxRealtimeEvents   int       `gorm:"default:0" json:"max_realtime_events"`   // Xadka fariimaha
	IsActive            bool      `gorm:"default:true" json:"is_active"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

func (p *Plan) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return
}
