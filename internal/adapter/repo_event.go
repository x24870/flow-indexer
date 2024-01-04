package adapter

import (
	"context"
	"flow-indexer/internal/domain/event"

	"gorm.io/gorm"
)

type eventRepo struct {
	db *gorm.DB
}

func NewEventRepo(db *gorm.DB) event.Repository {
	return &eventRepo{db: db}
}

func (r *eventRepo) Create(ctx context.Context, event *event.FlowEvent) error {
	return r.db.WithContext(ctx).Create(&event).Error
}
