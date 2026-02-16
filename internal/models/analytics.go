package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AnalyticsType string

const (
	AnalyticsTypeStorageUsage     AnalyticsType = "storage_usage"
	AnalyticsTypeApiCalls         AnalyticsType = "api_calls"
	AnalyticsTypeAuthActivity     AnalyticsType = "auth_activity"
	AnalyticsTypeDatabaseUsage    AnalyticsType = "database_usage"
	AnalyticsTypeNotifications    AnalyticsType = "notifications"
	AnalyticsTypeRealtimeChannels AnalyticsType = "realtime_channels" // ✅ CUSUB
	AnalyticsTypeRealtimeEvents   AnalyticsType = "realtime_events"   // ✅ CUSUB
)

type Analytics struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	ProjectID   string         `gorm:"type:uuid;not null;uniqueIndex:idx_project_type_period" json:"project_id"`
	Type        AnalyticsType  `gorm:"type:varchar(50);not null;uniqueIndex:idx_project_type_period" json:"type"`
	PeriodStart time.Time      `gorm:"not null;uniqueIndex:idx_project_type_period" json:"period_start"`
	Metrics     datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"metrics"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

func (a *Analytics) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return
}
