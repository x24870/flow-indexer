package inscription

import (
	"context"
	"flow-indexer/internal/domain"
)

type Balance struct {
	domain.Base
	Account     string `gorm:"column:account;foreignKey;index;reference:Address"`
	Inscription string `gorm:"column:inscription;type:varchar(256);index"`
	Amount      uint64 `gorm:"column:amount;type:integer"`
}

type Repository interface {
	GetorCreateByInscriptionAndAddress(ctx context.Context, inscription, address string) (*Balance, error)
	Update(ctx context.Context, balance *Balance) error
}

func (Balance) TableName() string {
	return domain.FlowInscriptionPrefix + "balance"
}
