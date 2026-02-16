package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ProjectAuthConfig struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`

	// ✅ SAXID: Labadan column waxay wadaagaan hal magac oo index ah (idx_project_provider)
	// Tani waxay database-ka ku dhex abuuraysaa UNIQUE(project_id, provider_id)
	ProjectID string `gorm:"type:uuid;uniqueIndex:idx_project_provider;not null" json:"project_id"`

	ProviderID string `gorm:"type:uuid;uniqueIndex:idx_project_provider;not null" json:"provider_id"`

	Credentials datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"credentials"`
	Enabled     bool           `gorm:"default:false" json:"enabled"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`

	// ✅ CASCADE: Haddii Global Provider-ka la tirtiro, kanina wuu raacayaa
	Provider GlobalAuthProvider `gorm:"foreignKey:ProviderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"provider"`
}

// Hook: ID-ga si otomaatig ah u dhali ka hor intaan la keydin
func (m *ProjectAuthConfig) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return
}

// ToJSONB Helper: Wuxuu interface u bedelaa JSONB
func ToJSONB(v interface{}) datatypes.JSON {
	bytes, err := json.Marshal(v)
	if err != nil {
		return datatypes.JSON([]byte("{}"))
	}
	return datatypes.JSON(bytes)
}
