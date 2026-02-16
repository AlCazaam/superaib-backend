package repo

import (
	"context"
	"superaib/internal/models"

	"gorm.io/gorm"
)

type TransactionRepository interface {
	Create(ctx context.Context, tx *gorm.DB, transaction *models.Transaction) error
	GetByProject(ctx context.Context, projectID string) ([]models.Transaction, error)
}

type gormTransactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &gormTransactionRepository{db: db}
}

func (r *gormTransactionRepository) Create(ctx context.Context, tx *gorm.DB, t *models.Transaction) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.WithContext(ctx).Create(t).Error
}

func (r *gormTransactionRepository) GetByProject(ctx context.Context, projectID string) ([]models.Transaction, error) {
	var transactions []models.Transaction
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Find(&transactions).Error
	return transactions, err
}
