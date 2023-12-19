package domain

import (
	"flow-indexer/internal/domain"

	uuid "github.com/satori/go.uuid"
)

type Inscription struct {
	domain.Base
	ID           uuid.UUID `gorm:"column:id;type:uuid;primary_key"`
	Creator      string    `gorm:"column:creator;foreignKey;index;reference:Address"`
	Tick         string    `gorm:"column:tick;type:varchar(64);primary_key"`
	TotalSupply  uint64    `gorm:"column:total_supply;type:integer"`
	LimitPerMint uint64    `gorm:"column:limit_per_mint;type:integer"`
}

func (Inscription) TableName() string {
	return domain.FlowInscriptionPrefix + "inscription"
}

type Balance struct {
	domain.Base
	Account       string    `gorm:"column:account;foreignKey;index;reference:Address"`
	InscriptionID uuid.UUID `gorm:"column:inscription_id;type:uuid"`
	Amount        uint64    `gorm:"column:amount;type:integer"`
}

func (Balance) TableName() string {
	return domain.FlowInscriptionPrefix + "balance"
}

type OperationLog struct {
	domain.Base
	ID            uuid.UUID `gorm:"column:id;type:uuid;primary_key"`
	InscriptionID uuid.UUID `gorm:"column:inscription_id;type:uuid;"`
	Operation     string    `gorm:"column:operation;type:varchar(64);"`
	From          string    `gorm:"column:from;foreignKey;index;reference:Address"`
	To            string    `gorm:"column:to;foreignKey;index;reference:Address"`
	Amount        uint64    `gorm:"column:amount;type:integer"`
	Valid         bool      `gorm:"column:valid;type:boolean"`
}

func (OperationLog) TableName() string {
	return domain.FlowInscriptionPrefix + "operation_log"
}
