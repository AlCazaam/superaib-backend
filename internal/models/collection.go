package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Collection struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ProjectID string    `gorm:"type:uuid;index;not null" json:"project_id"`

	Name string `gorm:"type:varchar(100);index:idx_project_coll_name;not null" json:"name"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (c *Collection) BeforeCreate(tx *gorm.DB) (err error) {
	c.ID = uuid.New()
	return
}
