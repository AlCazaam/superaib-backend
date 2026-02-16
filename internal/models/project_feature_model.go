package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type ProjectFeatureType string

const (
	FeatureTypeAuth          ProjectFeatureType = "auth"
	FeatureTypeDatabase      ProjectFeatureType = "database"
	FeatureTypeStorage       ProjectFeatureType = "storage"
	FeatureTypeAnalytics     ProjectFeatureType = "analytics"
	FeatureTypeNotifications ProjectFeatureType = "notifications" // ✅ KAN CUSUB: Lagu daray
	FeatureTypeRealtime      ProjectFeatureType = "realtime"      // ✅ Lagu daray si uu ula mid noqdo SDK-ga
)

type ProjectFeature struct {
	ID        uuid.UUID          `gorm:"type:uuid;primaryKey" json:"id"`
	ProjectID string             `gorm:"type:uuid;not null;index;uniqueIndex:idx_project_type" json:"project_id"`
	Type      ProjectFeatureType `gorm:"type:varchar(50);not null;uniqueIndex:idx_project_type" json:"type"`
	Enabled   bool               `gorm:"default:false" json:"enabled"`
	Config    datatypes.JSON     `gorm:"type:jsonb;default:'{}'" json:"config"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}

// SeedDefaultFeatures: Kani wuxuu dhalinayaa dhammaan adeegyada mashruucu leeyahay marka la dhisayo
func SeedDefaultFeatures(projectID string) []ProjectFeature {
	featureTypes := []ProjectFeatureType{
		FeatureTypeAuth,
		FeatureTypeDatabase,
		FeatureTypeStorage,
		FeatureTypeAnalytics,
		FeatureTypeNotifications, // ✅ Hadda wuxuu ka mid yahay mishiinka bilowga ah
		FeatureTypeRealtime,
	}

	var features []ProjectFeature
	for _, t := range featureTypes {
		features = append(features, ProjectFeature{
			ID:        uuid.New(),
			ProjectID: projectID,
			Type:      t,
			Enabled:   false,                        // Developer-ka ha shito markuu rabo
			Config:    datatypes.JSON([]byte("{}")), // Bilowga waa empty {}
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	}
	return features
}
