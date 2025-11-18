package example

import (
	"context"

	"gorm.io/gorm"
)

type Entity struct {
	*gorm.Model
}

type EntityRepo interface {
	Insert(ctx context.Context, entity *Entity) error
}

type EntityRepoImpl struct {
	db *gorm.DB
}

func NewEntityRepo(db *gorm.DB) *EntityRepoImpl {
	return &EntityRepoImpl{db}
}

func (e *EntityRepoImpl) Insert(ctx context.Context, entity *Entity) error {
	return e.db.Create(entity).Error
}
