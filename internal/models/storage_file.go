package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// StorageFile represents a single file linked to a project.
type StorageFile struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ProjectID string    `gorm:"type:uuid;index;not null" json:"project_id"`

	FileName       string         `gorm:"type:varchar(255);not null" json:"file_name"`
	FileType       string         `gorm:"type:varchar(100);not null" json:"file_type"` // MIME type
	SizeMB         float64        `gorm:"not null" json:"size_mb"`
	URL            string         `gorm:"not null" json:"url"`
	UploadedBy     *string        `gorm:"type:uuid" json:"uploaded_by,omitempty"`
	UploadedAt     time.Time      `json:"uploaded_at"`
	LastAccessedAt *time.Time     `json:"last_accessed_at,omitempty"`
	AccessControl  datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"access_control"`
	Version        int            `gorm:"default:1" json:"version"`
	Metadata       datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	Checksum       *string        `gorm:"type:varchar(255)" json:"checksum,omitempty"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"` // GORM soft delete support
}

func (sf *StorageFile) BeforeCreate(tx *gorm.DB) (err error) {
	sf.ID = uuid.New()
	sf.UploadedAt = time.Now()
	sf.Version = 1
	return
}

// TableName explicitly sets the table name.
func (StorageFile) TableName() string {
	return "storage_files"
}
