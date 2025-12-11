package uow

import (
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5"
	"gorm.io/gorm"
)

var (
	ErrTransactionNotFound = errors.New("no transaction found in context")
	savepoint              = "SAVEPOINT"
)

type gormFactory[RepoFactory any] struct {
	gormDb      *gorm.DB
	gormFactory func(*gorm.DB) RepoFactory
}

func NewGorm[RepoFactory any](db *gorm.DB, constructor func(tx *gorm.DB) RepoFactory) Factory[RepoFactory] {
	return &gormFactory[RepoFactory]{
		gormDb:      db,
		gormFactory: constructor,
	}
}

func (u *gormFactory[T]) UoW() UoW[T] {
	return &uowGorm[T]{
		db:          u.gormDb,
		gormFactory: u.gormFactory,
	}
}

type pgxFactory[RepoFactory any] struct {
	pool        PgxPool
	constructor func(tx pgx.Tx) RepoFactory
}

func NewPgx[RepoFactory any](pool PgxPool, constructor func(tx pgx.Tx) RepoFactory) Factory[RepoFactory] {
	return &pgxFactory[RepoFactory]{
		pool:        pool,
		constructor: constructor,
	}
}

func (f *pgxFactory[RepoFactory]) UoW() UoW[RepoFactory] {
	return &uowPgx[RepoFactory]{
		pool:        f.pool,
		constructor: f.constructor,
	}
}

type sqlFactory[RepoFactory any] struct {
	db          *sql.DB
	constructor func(tx *sql.Tx) RepoFactory
}

func NewSql[RepoFactory any](db *sql.DB, constructor func(tx *sql.Tx) RepoFactory) Factory[RepoFactory] {
	return &sqlFactory[RepoFactory]{
		db:          db,
		constructor: constructor,
	}
}

func (f *sqlFactory[RepoFactory]) UoW() UoW[RepoFactory] {
	return &uowSql[RepoFactory]{
		db:          f.db,
		constructor: f.constructor,
	}
}
