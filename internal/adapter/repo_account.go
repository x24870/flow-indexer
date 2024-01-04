package adapter

import (
	"context"
	"flow-indexer/internal/domain/account"

	"gorm.io/gorm"
)

type accountRepo struct {
	db *gorm.DB
}

func NewAccountRepo(db *gorm.DB) account.Repository {
	return &accountRepo{db: db}
}

func (r *accountRepo) FirstOrCreate(ctx context.Context, address string) (*account.Account, error) {
	var account account.Account
	account.Address = address
	err := r.db.WithContext(ctx).Where("address = ?", address).FirstOrCreate(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *accountRepo) GetByAddress(ctx context.Context, address string) (*account.Account, error) {
	var account account.Account
	err := r.db.WithContext(ctx).Where("address = ?", address).First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}
