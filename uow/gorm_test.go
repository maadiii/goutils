package uow

import (
	"context"
	"errors"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	ID   uint
	Name string
}

type myFactory struct{ db *gorm.DB }

func (m myFactory) CreateUser(u *User) error {
	return m.db.Create(u).Error
}

func newMyFactory(tx *gorm.DB) myFactory { return myFactory{db: tx} }

func setupDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed open sqlite: %v", err)
	}

	if err := db.AutoMigrate(&User{}); err != nil {
		t.Fatalf("failed migrate: %v", err)
	}

	return db
}

func countUsers(t *testing.T, db *gorm.DB) int64 {
	t.Helper()
	var cnt int64
	if err := db.Model(&User{}).Count(&cnt).Error; err != nil {
		t.Fatalf("count users failed: %v", err)
	}
	return cnt
}

func TestDo_CommitsOnSuccessAndRollbacksOnError(t *testing.T) {
	db := setupDB(t)

	f := NewGorm(db, newMyFactory)
	u := f.UoW()

	// success path: should commit
	if err := u.Do(context.Background(), func(ctx context.Context, repo myFactory) error {
		return repo.CreateUser(&User{Name: "Alice"})
	}); err != nil {
		t.Fatalf("Do returned error on success path: %v", err)
	}

	if got := countUsers(t, db); got != 1 {
		t.Fatalf("expected 1 user after commit, got %d", got)
	}

	// failure path: should rollback
	if err := u.Do(context.Background(), func(ctx context.Context, repo myFactory) error {
		if err := repo.CreateUser(&User{Name: "Bob"}); err != nil {
			return err
		}
		return errors.New("force rollback")
	}); err == nil {
		t.Fatalf("expected error from Do to propagate, got nil")
	}

	if got := countUsers(t, db); got != 1 {
		t.Fatalf("expected still 1 user after rollback, got %d", got)
	}
}

func TestBegin_Commit_And_Begin_Rollback(t *testing.T) {
	db := setupDB(t)

	f := NewGorm(db, newMyFactory)
	u := f.UoW()

	// Begin + Commit
	ctx, repo, err := u.Begin(context.Background())
	if err != nil {
		t.Fatalf("Begin returned error: %v", err)
	}

	if err := repo.CreateUser(&User{Name: "Charlie"}); err != nil {
		t.Fatalf("create in tx failed: %v", err)
	}

	if err := u.Commit(ctx); err != nil {
		t.Fatalf("Commit returned error: %v", err)
	}

	if got := countUsers(t, db); got != 1 {
		t.Fatalf("expected 1 user after commit, got %d", got)
	}

	// Begin + Rollback
	ctx2, repo2, err := u.Begin(context.Background())
	if err != nil {
		t.Fatalf("Begin returned error: %v", err)
	}

	if err := repo2.CreateUser(&User{Name: "Dave"}); err != nil {
		t.Fatalf("create in tx failed: %v", err)
	}

	if err := u.Rollback(ctx2); err != nil {
		t.Fatalf("Rollback returned error: %v", err)
	}

	if got := countUsers(t, db); got != 1 {
		t.Fatalf("expected still 1 user after rollback, got %d", got)
	}
}

func TestSavePointAndErrorsWhenNoTx(t *testing.T) {
	db := setupDB(t)

	f := NewGorm(db, newMyFactory)
	u := f.UoW()

	// SavePoint within a transaction should not error
	ctx, repo, err := u.Begin(context.Background())
	if err != nil {
		t.Fatalf("Begin returned error: %v", err)
	}

	if err := repo.CreateUser(&User{Name: "Eve"}); err != nil {
		t.Fatalf("create in tx failed: %v", err)
	}

	if err := u.SavePoint(ctx, "sp1"); err != nil {
		t.Fatalf("SavePoint returned error: %v", err)
	}

	// Commit and check user exists
	if err := u.Commit(ctx); err != nil {
		t.Fatalf("Commit returned error: %v", err)
	}

	if got := countUsers(t, db); got != 1 {
		t.Fatalf("expected 1 user after commit, got %d", got)
	}

	// Calling Commit/Rollback/SavePoint without transaction should return specific error
	if err := u.Commit(context.Background()); err == nil {
		t.Fatalf("expected error when committing without tx, got nil")
	}

	if err := u.Rollback(context.Background()); err == nil {
		t.Fatalf("expected error when rolling back without tx, got nil")
	}

	if err := u.SavePoint(context.Background(), "nope"); err == nil {
		t.Fatalf("expected error when savepoint without tx, got nil")
	}
}

func TestBegin_ErrorHandling(t *testing.T) {
	// Test Begin with an already-closed database to trigger error
	db := setupDB(t)
	
	// Get the underlying sql.DB to close it
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get sql.DB: %v", err)
	}
	sqlDB.Close()

	f := NewGorm(db, newMyFactory)
	u := f.UoW()

	// Begin should fail because the database is closed
	_, _, err = u.Begin(context.Background())
	if err == nil {
		t.Fatalf("expected error when Begin is called on closed db, got nil")
	}
}
