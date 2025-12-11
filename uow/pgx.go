package uow

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// PgxPool interface abstracts pgxpool.Pool for testing with mocks
type PgxPool interface {
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

type uowPgx[RepoFactory any] struct {
	pool        PgxPool
	constructor func(tx pgx.Tx) RepoFactory
}

func (u *uowPgx[T]) Do(ctx context.Context, fn func(context.Context, T) error, opts ...*sql.TxOptions) error {
	var txOpts *sql.TxOptions
	if len(opts) > 0 {
		txOpts = opts[0]
	}

	pgxTxOpts := convertPgxTxOptions(txOpts)

	tx, err := u.pool.BeginTx(ctx, pgxTxOpts)
	if err != nil {
		return err
	}

	factory := u.constructor(tx)
	if err := fn(ctx, factory); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}

func (u *uowPgx[T]) Begin(ctx context.Context, opts ...*sql.TxOptions) (c context.Context, repoFactory T, err error) {
	var txOpts *sql.TxOptions
	if len(opts) > 0 {
		txOpts = opts[0]
	}

	pgxTxOpts := convertPgxTxOptions(txOpts)

	tx, err := u.pool.BeginTx(ctx, pgxTxOpts)
	if err != nil {
		return
	}

	ctx = context.WithValue(ctx, ctxTxKey{}, tx)

	return ctx, u.constructor(tx), nil
}

func (u *uowPgx[RepoFactory]) Commit(ctx context.Context) error {
	tx, ok := ctx.Value(ctxTxKey{}).(pgx.Tx)
	if !ok || tx == nil {
		return ErrTransactionNotFound
	}

	return tx.Commit(ctx)
}

func (u *uowPgx[RepoFactory]) Rollback(ctx context.Context) error {
	tx, ok := ctx.Value(ctxTxKey{}).(pgx.Tx)
	if !ok || tx == nil {
		return ErrTransactionNotFound
	}

	return tx.Rollback(ctx)
}

func (u *uowPgx[T]) SavePoint(ctx context.Context, name string) error {
	tx, ok := ctx.Value(ctxTxKey{}).(pgx.Tx)
	if !ok || tx == nil {
		return ErrTransactionNotFound
	}

	_, err := tx.Exec(ctx, fmt.Sprintf("%s %s", savepoint, name))
	return err
}

func convertPgxTxOptions(opts *sql.TxOptions) pgx.TxOptions {
	var pgxOpts pgx.TxOptions

	// Set default access mode for nil options
	if opts == nil {
		pgxOpts.AccessMode = pgx.ReadWrite
		return pgxOpts
	}

	var iso pgx.TxIsoLevel

	switch opts.Isolation {
	case sql.LevelDefault:
		iso = pgx.ReadCommitted
	case sql.LevelReadUncommitted:
		// Postgres doesn't actually support dirty reads
		iso = pgx.ReadCommitted
	case sql.LevelReadCommitted:
		iso = pgx.ReadCommitted
	case sql.LevelRepeatableRead:
		iso = pgx.RepeatableRead
	case sql.LevelSerializable:
		iso = pgx.Serializable
	default:
		iso = pgx.ReadCommitted
	}

	pgxOpts.IsoLevel = iso
	pgxOpts.AccessMode = chooseAccessMode(opts)

	return pgxOpts
}

func chooseAccessMode(opts *sql.TxOptions) pgx.TxAccessMode {
	if opts == nil {
		return pgx.ReadWrite
	}

	if opts.ReadOnly {
		return pgx.ReadOnly
	}

	return pgx.ReadWrite
}
