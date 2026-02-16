package models

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type GlobalFeature struct {
	ID        uuid.UUID          `gorm:"type:uuid;primaryKey" json:"id"`
	Type      ProjectFeatureType `gorm:"type:varchar(50);uniqueIndex;not null" json:"type"`
	Enabled   bool               `gorm:"default:true" json:"enabled"`
	Config    datatypes.JSON     `gorm:"type:jsonb;default:'{}'" json:"config"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}

func (gf *GlobalFeature) BeforeSave(tx *gorm.DB) (err error) {
	// Waxaan ka dhigaynaa lowercase si uusan isku dhac u dhicin (e.g., Notification vs Notifications)
	gf.Type = ProjectFeatureType(strings.ToLower(strings.TrimSpace(string(gf.Type))))
	if gf.ID == uuid.Nil {
		gf.ID = uuid.New()
	}
	return
}

func SeedGlobalFeatures(db *gorm.DB) {
	featureTypes := []struct {
		Type   ProjectFeatureType
		Fields []string
	}{
		{FeatureTypeAuth, []string{}},
		{FeatureTypeDatabase, []string{}},
		{"realtime", []string{}},
		{FeatureTypeAnalytics, []string{}},
		{FeatureTypeStorage, []string{"api_key", "secret_key"}},
		{"notifications", []string{"fcm_server_key", "app_id"}},
	}

	for _, f := range featureTypes {
		var existing GlobalFeature
		if err := db.Where("LOWER(type) = LOWER(?)", f.Type).First(&existing).Error; err != nil {
			configMap := make(map[string]interface{})
			if len(f.Fields) > 0 {
				configMap["required_fields"] = f.Fields
			}
			configBytes, _ := json.Marshal(configMap)
			db.Create(&GlobalFeature{
				Type:    f.Type,
				Enabled: true,
				Config:  datatypes.JSON(configBytes),
			})
		}
	}
}
