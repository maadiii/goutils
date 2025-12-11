package uow

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
)

type uowGorm[RepoFactory any] struct {
	db          *gorm.DB
	gormFactory func(*gorm.DB) RepoFactory
}

func (u *uowGorm[T]) Do(ctx context.Context, fn func(context.Context, T) error, opts ...*sql.TxOptions) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		factory := u.gormFactory(tx)

		return fn(ctx, factory)
	}, opts...)
}

func (u *uowGorm[T]) Begin(ctx context.Context, opts ...*sql.TxOptions) (c context.Context, repoFactory T, err error) {
	tx := u.db.Begin(opts...)
	if err = tx.Error; err != nil {
		return
	}

	factory := u.gormFactory(tx)
	ctx = context.WithValue(ctx, ctxTxKey{}, tx)

	return ctx, factory, nil
}

func (u *uowGorm[T]) Commit(ctx context.Context) error {
	tx, ok := ctx.Value(ctxTxKey{}).(*gorm.DB)
	if !ok || tx == nil {
		return ErrTransactionNotFound
	}

	return tx.Commit().Error
}

func (u *uowGorm[T]) Rollback(ctx context.Context) error {
	tx, ok := ctx.Value(ctxTxKey{}).(*gorm.DB)
	if !ok || tx == nil {
		return ErrTransactionNotFound
	}

	return tx.Rollback().Error
}

func (u *uowGorm[T]) SavePoint(ctx context.Context, name string) error {
	tx, ok := ctx.Value(ctxTxKey{}).(*gorm.DB)
	if !ok || tx == nil {
		return ErrTransactionNotFound
	}

	return tx.SavePoint(name).Error
}

type ctxTxKey struct{}
