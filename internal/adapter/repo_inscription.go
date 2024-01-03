package adapter

import (
	"context"
	"flow-indexer/internal/domain/inscription"

	"gorm.io/gorm"
)

type inscriptionRepo struct {
	db *gorm.DB
}

func NewInscriptionRepo(db *gorm.DB) inscription.Repository {
	return &inscriptionRepo{db: db}
}

func (r *inscriptionRepo) GetByInscriptionAndAddress(ctx context.Context, insName, address string) (*inscription.Balance, error) {
	var balance inscription.Balance
	err := r.db.WithContext(ctx).Where("inscription = ? AND account = ?", insName, address).First(&balance).Error
	if err != nil {
		return nil, err
	}
	return &balance, nil
}
