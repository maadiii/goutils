package example

import (
	"gorm.io/gorm"
)

type RepoFactory interface {
	Entity() EntityRepo
}

type RepoFactoryImpl struct {
	entityRepo EntityRepo
}

func NewRepoFactory(db *gorm.DB) RepoFactory {
	return &RepoFactoryImpl{
		entityRepo: &EntityRepoImpl{db},
	}
}

func (r *RepoFactoryImpl) Entity() EntityRepo {
	return r.entityRepo
}
