package uow

import (
	"context"
	"database/sql"
)

type UoWFactory[RepoFactory any] interface {
	UoW() UoW[RepoFactory]
}

type UoW[RepoFactory any] interface {
	Do(ctx context.Context, fn func(ctx context.Context, repo RepoFactory) error, opts ...*sql.TxOptions) error
	Begin(ctx context.Context, opts ...*sql.TxOptions) (context.Context, RepoFactory, error)
	SavePoint(ctx context.Context, name string) error
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}
