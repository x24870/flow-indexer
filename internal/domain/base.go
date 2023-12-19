package domain

import (
	"time"
)

const (
	FlowInscriptionPrefix = "flow_inscription_"
)

// ForeignKeyConstraint defines the required arguments to the AddForeignKey call.
type ForeignKeyConstraint struct {
	Field    string
	Dest     string
	OnDelete string
	OnUpdate string
}

// ForeignKeyConstrainer defines a interface for models that support creating foreign key constraints.
type ForeignKeyConstrainer interface {
	ForeignKeyConstraints() []ForeignKeyConstraint
}

// CustomIndex defines index information
type CustomIndex struct {
	Name      string
	Unique    bool
	Fields    []string
	Type      string
	Condition string
}

// CustomIndexer definces a interface for models that decouples creating index
// from Gorm tag functionality
type CustomIndexer interface {
	Indexes() []CustomIndex
}

// Base is the base model for all data model.
type Base struct {
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamp with time zone" json:"-"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamp with time zone" json:"-"`
}
