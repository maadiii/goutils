package uow

import (
	"context"
	"database/sql"
	"errors"

	"gorm.io/gorm"
)

type factory[RepoFactory any] struct {
	gormDb      *gorm.DB
	gormFactory func(*gorm.DB) RepoFactory
}

func NewGorm[RepoFactory any](db *gorm.DB, constructor func(tx *gorm.DB) RepoFactory) Factory[RepoFactory] {
	return &factory[RepoFactory]{
		gormDb:      db,
		gormFactory: constructor,
	}
}

func (u *factory[T]) UoW() UoW[T] {
	return &uow[T]{
		db:          u.gormDb,
		gormFactory: u.gormFactory,
	}
}

type uow[RepoFactory any] struct {
	db          *gorm.DB
	gormFactory func(*gorm.DB) RepoFactory
}

func (u *uow[T]) Do(ctx context.Context, fn func(context.Context, T) error, opts ...*sql.TxOptions) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		factory := u.gormFactory(tx)

		return fn(ctx, factory)
	}, opts...)
}

func (u *uow[T]) Begin(ctx context.Context, opts ...*sql.TxOptions) (c context.Context, repoFactory T, err error) {
	tx := u.db.Begin(opts...)
	if err = tx.Error; err != nil {
		return
	}

	factory := u.gormFactory(tx)
	ctx = context.WithValue(ctx, ctxTxKey{}, tx)

	return ctx, factory, nil
}

func (u *uow[T]) Commit(ctx context.Context) error {
	tx, ok := ctx.Value(ctxTxKey{}).(*gorm.DB)
	if !ok || tx == nil {
		return errors.New("no transaction found in context")
	}

	return tx.Commit().Error
}

func (u *uow[T]) Rollback(ctx context.Context) error {
	tx, ok := ctx.Value(ctxTxKey{}).(*gorm.DB)
	if !ok || tx == nil {
		return errors.New("no transaction found in context")
	}

	return tx.Rollback().Error
}

func (u *uow[T]) SavePoint(ctx context.Context, name string) error {
	tx, ok := ctx.Value(ctxTxKey{}).(*gorm.DB)
	if !ok || tx == nil {
		return errors.New("no transaction found in context")
	}

	return tx.SavePoint(name).Error
}

type ctxTxKey struct{}
