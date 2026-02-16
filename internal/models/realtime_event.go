package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type RealtimeEventType string

const (
	EventTypeInsert    RealtimeEventType = "insert"
	EventTypeUpdate    RealtimeEventType = "update"
	EventTypeDelete    RealtimeEventType = "delete"
	EventTypeCustom    RealtimeEventType = "custom"
	EventTypeSystem    RealtimeEventType = "system"
	EventTypeBroadcast RealtimeEventType = "broadcast"
)

type RealtimeEvent struct {
	ID                uuid.UUID         `gorm:"type:uuid;primaryKey" json:"id"`
	ChannelID         uuid.UUID         `gorm:"type:uuid;index;not null" json:"channel_id"`
	EventType         RealtimeEventType `gorm:"type:varchar(50);default:'custom'" json:"event_type"`
	Payload           datatypes.JSON    `gorm:"type:jsonb;not null" json:"payload"`
	SenderID          *string           `gorm:"type:uuid" json:"sender_id,omitempty"`
	AckRequired       bool              `gorm:"default:false" json:"ack_required"`
	Delivered         bool              `gorm:"default:false" json:"delivered"`
	DeliveryTimestamp *time.Time        `json:"delivery_timestamp,omitempty"`
	Retries           int               `gorm:"default:0" json:"retries"`
	IPAddress         *string           `json:"ip_address,omitempty"`
	UserAgent         *string           `json:"user_agent,omitempty"`
	LatencyMs         float64           `json:"latency_ms"`
	ErrorMessage      *string           `json:"error_message,omitempty"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
}

func (e *RealtimeEvent) BeforeCreate(tx *gorm.DB) (err error) {
	e.ID = uuid.New()
	return
}
