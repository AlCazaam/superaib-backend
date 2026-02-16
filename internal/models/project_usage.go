package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProjectUsage struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ProjectID string    `gorm:"type:uuid;not null;index;unique" json:"project_id"`

	// 📊 USED (Wixii hadda la isticmaalay)
	ApiCalls              int     `gorm:"default:0" json:"api_calls"`
	AuthUsersCount        int     `gorm:"default:0" json:"auth_users_count"`
	StorageUsedMB         float64 `gorm:"type:decimal(10,2);default:0" json:"storage_used_mb"`
	DocumentsCount        int     `gorm:"default:0" json:"documents_count"`
	NotificationsCount    int     `gorm:"default:0" json:"notifications_count"`
	RealtimeChannelsCount int     `gorm:"default:0" json:"realtime_channels_count"` // ✅ CUSUB
	RealtimeEventsCount   int     `gorm:"default:0" json:"realtime_events_count"`   // ✅ CUSUB

	// 🛡️ LIMITS (Snapshot-ka Plan-ka uu iibsaday)
	LimitApiCalls         int     `gorm:"default:0" json:"limit_api_calls"`
	LimitAuthUsers        int     `gorm:"default:0" json:"limit_auth_users"`
	LimitStorageMB        float64 `gorm:"type:decimal(10,2);default:0" json:"limit_storage_mb"`
	LimitDocuments        int     `gorm:"default:0" json:"limit_documents"`
	LimitNotifications    int     `gorm:"default:0" json:"limit_notifications"`
	LimitRealtimeChannels int     `gorm:"default:0" json:"limit_realtime_channels"` // 👈 Hubi inuu yahay 'limit_realtime_channels'
	LimitRealtimeEvents   int     `gorm:"default:0" json:"limit_realtime_events"`   // 👈 Hubi inuu yahay 'limit_realtime_events'

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (p *ProjectUsage) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return
}
