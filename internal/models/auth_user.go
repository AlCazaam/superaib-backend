package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AuthUserStatus string

const (
	AuthUserActive  AuthUserStatus = "active"
	AuthUserBlocked AuthUserStatus = "blocked"
	AuthUserDeleted AuthUserStatus = "deleted"
	AuthUserInvited AuthUserStatus = "invited"
)

type AuthUser struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ProjectID string    `gorm:"type:uuid;index;not null" json:"project_id"`

	Email        string `gorm:"uniqueIndex:idx_email_project;not null" json:"email"`
	PasswordHash string `json:"-"`

	Roles    datatypes.JSON `gorm:"type:jsonb;default:'[]'" json:"roles"`
	Metadata datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"metadata"`

	// ✅ SAX: Waxaan ku qasabnay GORM inuu isticmaalo "o_auth_providers"
	OAuthProviders datatypes.JSON `gorm:"column:o_auth_providers;type:jsonb;default:'{}'" json:"o_auth_providers"`

	Status      AuthUserStatus `gorm:"type:varchar(20);default:'active'" json:"status"`
	LastLoginAt *time.Time     `json:"last_login_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

func (u *AuthUser) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return
}

func FromJSONBAuth(j datatypes.JSON) map[string]string {
	var m map[string]string
	if len(j) == 0 {
		return make(map[string]string)
	}
	if err := json.Unmarshal(j, &m); err != nil {
		return make(map[string]string)
	}
	return m
}

func ToJSONBAuth(m map[string]string) datatypes.JSON {
	if m == nil {
		return datatypes.JSON("{}")
	}
	b, _ := json.Marshal(m)
	return datatypes.JSON(b)
}
