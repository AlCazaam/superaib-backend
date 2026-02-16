package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type RealtimeSubscriptionType string

const (
	SubscriptionPublic    RealtimeSubscriptionType = "public"
	SubscriptionProtected RealtimeSubscriptionType = "protected"
	SubscriptionPrivate   RealtimeSubscriptionType = "private"
)

type RealtimeRetentionPolicy string

const (
	RetentionEphemeral  RealtimeRetentionPolicy = "ephemeral"
	RetentionPersistent RealtimeRetentionPolicy = "persistent"
)

type RealtimeChannel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ProjectID string    `gorm:"type:uuid;index;not null" json:"project_id"`

	Name             string                   `gorm:"type:varchar(255);uniqueIndex:idx_project_channel_name;not null" json:"name"`
	Description      *string                  `json:"description,omitempty"`
	IsPrivate        bool                     `gorm:"default:false" json:"is_private"`
	ConnectedClients int                      `gorm:"default:0" json:"connected_clients"`
	MaxClients       int                      `gorm:"default:100" json:"max_clients"`
	SubscriptionType RealtimeSubscriptionType `gorm:"type:varchar(50);default:'public'" json:"subscription_type"`
	CreatedAt        time.Time                `json:"created_at"`
	UpdatedAt        time.Time                `json:"updated_at"`
	LastMessageAt    *time.Time               `json:"last_message_at,omitempty"`
	Metadata         datatypes.JSON           `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	RetentionPolicy  RealtimeRetentionPolicy  `gorm:"type:varchar(50);default:'ephemeral'" json:"retention_policy"`
	Archived         bool                     `gorm:"default:false" json:"archived"`
	ArchivedAt       *time.Time               `json:"archived_at,omitempty"`
	Tags             datatypes.JSON           `gorm:"type:jsonb;default:'[]'" json:"tags"`
	Events           []RealtimeEvent          `gorm:"foreignKey:ChannelID" json:"-"` // Relation
}

func (c *RealtimeChannel) BeforeCreate(tx *gorm.DB) (err error) {
	c.ID = uuid.New()
	return
}
