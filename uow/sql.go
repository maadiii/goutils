package uow

import (
	"context"
	"database/sql"
	"fmt"
)

type uowSql[RepoFactory any] struct {
	db          *sql.DB
	constructor func(tx *sql.Tx) RepoFactory
}

func (u *uowSql[T]) Do(ctx context.Context, fn func(context.Context, T) error, opts ...*sql.TxOptions) error {
	var txOpts *sql.TxOptions
	if len(opts) > 0 {
		txOpts = opts[0]
	}

	tx, err := u.db.BeginTx(ctx, txOpts)
	if err != nil {
		return err
	}

	factory := u.constructor(tx)
	if err := fn(ctx, factory); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (u *uowSql[T]) Begin(ctx context.Context, opts ...*sql.TxOptions) (c context.Context, repoFactory T, err error) {
	var txOpts *sql.TxOptions
	if len(opts) > 0 {
		txOpts = opts[0]
	}

	tx, err := u.db.BeginTx(ctx, txOpts)
	if err != nil {
		return
	}

	factory := u.constructor(tx)
	ctx = context.WithValue(ctx, ctxTxKey{}, tx)

	return ctx, factory, nil
}

func (u *uowSql[T]) Commit(ctx context.Context) error {
	tx, ok := ctx.Value(ctxTxKey{}).(*sql.Tx)
	if !ok || tx == nil {
		return ErrTransactionNotFound
	}

	return tx.Commit()
}

func (u *uowSql[T]) Rollback(ctx context.Context) error {
	tx, ok := ctx.Value(ctxTxKey{}).(*sql.Tx)
	if !ok || tx == nil {
		return ErrTransactionNotFound
	}

	return tx.Rollback()
}

func (u *uowSql[T]) SavePoint(ctx context.Context, name string) error {
	tx, ok := ctx.Value(ctxTxKey{}).(*sql.Tx)
	if !ok || tx == nil {
		return ErrTransactionNotFound
	}

	_, err := tx.Exec(fmt.Sprintf("%s %s", savepoint, name))

	return err
}
