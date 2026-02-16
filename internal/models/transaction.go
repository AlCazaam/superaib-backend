package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Transaction struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ProjectID   string    `gorm:"type:uuid;not null;index" json:"project_id"`
	DeveloperID string    `gorm:"type:uuid;not null" json:"developer_id"`
	PlanID      uuid.UUID `gorm:"type:uuid;not null" json:"plan_id"`

	Amount        float64 `gorm:"type:decimal(10,2)" json:"amount"`
	Currency      string  `gorm:"default:'USD'" json:"currency"`
	PhoneNumber   string  `json:"phone_number"`
	TransactionID string  `gorm:"unique" json:"transaction_id"`      // WaafiPay ID
	Status        string  `gorm:"default:'completed'" json:"status"` // completed, failed

	CreatedAt time.Time `json:"created_at"`
}

func (t *Transaction) BeforeCreate(tx *gorm.DB) (err error) {
	t.ID = uuid.New()
	return
}
