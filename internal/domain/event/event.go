package event

import (
	"context"
	"flow-indexer/internal/domain"

	uuid "github.com/satori/go.uuid"
)

type FlowEvent struct {
	domain.Base
	ID      uuid.UUID `gorm:"column:id;primaryKey;type:uuid;default:uuid_generate_v4()"`
	Account string    `gorm:"column:account;foreignKey;index;reference:Address"`
	Event   string    `gorm:"column:event;type:varchar(256);index"`
	Block   uint64    `gorm:"column:block;type:integer;default:0"`
}

type Repository interface {
	Create(ctx context.Context, event *FlowEvent) error
}

func (FlowEvent) TableName() string {
	return domain.FlowInscriptionPrefix + "event"
}
