package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// APIKey represents the security key for the Somali Firebase version
type APIKey struct {
	// Primary ID (UUID)
	ID string `gorm:"type:uuid;primaryKey" json:"id"`

	// Project-ka uu furuhu u gaarka yahay
	ProjectID string `gorm:"type:uuid;index;not null" json:"project_id"`

	// Key-ga rasmiga ah ee qarsoon (Unique Index si uusan u soo noqnoqon)
	Key string `gorm:"type:varchar(255);uniqueIndex;not null" json:"key"`

	// Magaca furaha (Tusaale: "Mobile App Key")
	Name string `gorm:"type:varchar(255);not null" json:"name"`

	// Permissions: Waxaan u isticmaalaynaa JSONB si uu ula mid noqdo Map-ka Flutter-ka
	// Tusaale: {"read": true, "write": false}
	Permissions datatypes.JSON `gorm:"type:jsonb" json:"permissions"`

	// Xaaladda furaha (Dami/Shid)
	Revoked bool `gorm:"default:false" json:"revoked"`

	// Inta jeer ee la isticmaalay furaha
	UsageCount int `gorm:"default:0" json:"usage_count"`

	// Qofka abuuray furaha (Admin-ka Project-ka)
	CreatedBy string `gorm:"type:uuid;not null" json:"created_by"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Relation: Furuhu wuxuu ka tirsan yahay Project
	Project *Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
}
