package example

import (
	"context"

	"github.com/maadiii/utils/uow"
)

type service struct {
	entities   EntityRepo
	uowFactory uow.UoWFactory[RepoFactory]
}

func (s *service) Create(ctx context.Context) error {
	entity := &Entity{}

	if err := s.entities.Insert(ctx, entity); err != nil {
		return err
	}

	uow := s.uowFactory.Tx()
	err := uow.Do(ctx, func(ctx context.Context, repo RepoFactory) error {
		return repo.Entity().Insert(ctx, entity)
	})
	if err != nil {
		return err
	}

	ctx, repo, err := uow.Begin(ctx)
	if err != nil {
		return err
	}

	if err := repo.Entity().Insert(ctx, entity); err != nil {
		uow.Rollback(ctx)

		return err
	}

	return uow.Commit(ctx)
}
