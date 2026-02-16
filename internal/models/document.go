package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Document struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ProjectID string    `gorm:"type:uuid;index;not null" json:"project_id"`

	CollectionID uuid.UUID      `gorm:"type:uuid;index;not null" json:"collection_id"`
	Data         datatypes.JSON `gorm:"type:jsonb;default:'{}';not null" json:"data"`

	// ✅ KALIYA 'etag' (Database column name)
	ETag string `gorm:"column:etag;type:varchar(64);index" json:"etag"`

	Version   int       `gorm:"default:1;not null" json:"version"`
	IsDeleted bool      `gorm:"default:false;index" json:"is_deleted"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (d *Document) BeforeCreate(tx *gorm.DB) (err error) {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	d.ETag = uuid.New().String()
	d.Version = 1
	return
}
