package uow

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v3"
)

type pgxUser struct {
	ID   int
	Name string
}

type pgxRepoFactory struct {
	tx pgx.Tx
}

func (p pgxRepoFactory) CreateUser(u *pgxUser) error {
	_, err := p.tx.Exec(context.Background(), "INSERT INTO users (name) VALUES ($1)", u.Name)
	return err
}

func (p pgxRepoFactory) GetUser(id int) (*pgxUser, error) {
	row := p.tx.QueryRow(context.Background(), "SELECT id, name FROM users WHERE id = $1", id)
	var user pgxUser
	if err := row.Scan(&user.ID, &user.Name); err != nil {
		return nil, err
	}
	return &user, nil
}

func newPgxFactory(tx pgx.Tx) pgxRepoFactory {
	return pgxRepoFactory{tx: tx}
}

// ============================================================================
// pgxFactory Tests
// ============================================================================

func TestNewPgx_CreatesFactory(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock pool: %v", err)
	}
	defer mock.Close()

	factory := NewPgx(mock, newPgxFactory)

	if factory == nil {
		t.Fatalf("expected non-nil factory, got nil")
	}
}

func TestNewPgx_FactoryUoWReturnsUoW(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock pool: %v", err)
	}
	defer mock.Close()

	factory := NewPgx(mock, newPgxFactory)
	uow := factory.UoW()

	if uow == nil {
		t.Fatalf("expected non-nil UoW, got nil")
	}
}

func TestNewPgx_MultipleUoWCallsReturnDifferentInstances(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock pool: %v", err)
	}
	defer mock.Close()

	factory := NewPgx(mock, newPgxFactory)

	uow1 := factory.UoW()
	uow2 := factory.UoW()

	// They should be different instances (though they work with the same pool)
	if uow1 == uow2 {
		t.Fatalf("expected different UoW instances, got same instance")
	}
}

// ============================================================================
// uowPgx.Do() Tests
// ============================================================================

func TestPgxDo_CommitsOnSuccess(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock pool: %v", err)
	}
	defer mock.Close()

	mock.ExpectBeginTx(pgx.TxOptions{AccessMode: pgx.ReadWrite})
	mock.ExpectExec("INSERT INTO users").
		WithArgs("Alice").
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectCommit()

	factory := NewPgx(mock, newPgxFactory)
	uow := factory.UoW()

	err = uow.Do(context.Background(), func(ctx context.Context, repo pgxRepoFactory) error {
		return repo.CreateUser(&pgxUser{Name: "Alice"})
	})
	if err != nil {
		t.Fatalf("expected no error on success, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPgxDo_RollsBackOnError(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock pool: %v", err)
	}
	defer mock.Close()

	mock.ExpectBeginTx(pgx.TxOptions{AccessMode: pgx.ReadWrite})
	mock.ExpectExec("INSERT INTO users").
		WithArgs("Bob").
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectRollback()

	factory := NewPgx(mock, newPgxFactory)
	uow := factory.UoW()

	err = uow.Do(context.Background(), func(ctx context.Context, repo pgxRepoFactory) error {
		if err := repo.CreateUser(&pgxUser{Name: "Bob"}); err != nil {
			return err
		}
		return errors.New("force rollback")
	})

	if err == nil {
		t.Fatalf("expected error to propagate, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPgxDo_PropagatesErrorFromFunction(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock pool: %v", err)
	}
	defer mock.Close()

	mock.ExpectBeginTx(pgx.TxOptions{AccessMode: pgx.ReadWrite})
	mock.ExpectRollback()

	factory := NewPgx(mock, newPgxFactory)
	uow := factory.UoW()

	expectedErr := errors.New("test error")
	err = uow.Do(context.Background(), func(ctx context.Context, repo pgxRepoFactory) error {
		return expectedErr
	})

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPgxDo_MultipleOperationsWithinTransaction(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock pool: %v", err)
	}
	defer mock.Close()

	mock.ExpectBeginTx(pgx.TxOptions{AccessMode: pgx.ReadWrite})
	mock.ExpectExec("INSERT INTO users").WithArgs("User1").WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectExec("INSERT INTO users").WithArgs("User2").WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectExec("INSERT INTO users").WithArgs("User3").WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectCommit()

	factory := NewPgx(mock, newPgxFactory)
	uow := factory.UoW()

	err = uow.Do(context.Background(), func(ctx context.Context, repo pgxRepoFactory) error {
		if err := repo.CreateUser(&pgxUser{Name: "User1"}); err != nil {
			return err
		}
		if err := repo.CreateUser(&pgxUser{Name: "User2"}); err != nil {
			return err
		}
		if err := repo.CreateUser(&pgxUser{Name: "User3"}); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPgxDo_MultipleOperationsRollbackOnLastError(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock pool: %v", err)
	}
	defer mock.Close()

	mock.ExpectBeginTx(pgx.TxOptions{AccessMode: pgx.ReadWrite})
	mock.ExpectExec("INSERT INTO users").WithArgs("User1").WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectExec("INSERT INTO users").WithArgs("User2").WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectRollback()

	factory := NewPgx(mock, newPgxFactory)
	uow := factory.UoW()

	err = uow.Do(context.Background(), func(ctx context.Context, repo pgxRepoFactory) error {
		if err := repo.CreateUser(&pgxUser{Name: "User1"}); err != nil {
			return err
		}
		if err := repo.CreateUser(&pgxUser{Name: "User2"}); err != nil {
			return err
		}
		return errors.New("error on third operation")
	})

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPgxDo_WithTxOptions(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock pool: %v", err)
	}
	defer mock.Close()

	mock.ExpectBeginTx(pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	})
	mock.ExpectExec("INSERT INTO users").WithArgs("TxOpts").WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectCommit()

	factory := NewPgx(mock, newPgxFactory)
	uow := factory.UoW()

	opts := &sql.TxOptions{
		Isolation: sql.LevelDefault,
		ReadOnly:  false,
	}

	err = uow.Do(context.Background(), func(ctx context.Context, repo pgxRepoFactory) error {
		return repo.CreateUser(&pgxUser{Name: "TxOpts"})
	}, opts)
	if err != nil {
		t.Fatalf("expected no error with TxOptions, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPgxDo_WithMultipleTxOptions_UsesFirst(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock pool: %v", err)
	}
	defer mock.Close()

	mock.ExpectBeginTx(pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	})
	mock.ExpectExec("INSERT INTO users").WithArgs("Multi").WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectCommit()

	factory := NewPgx(mock, newPgxFactory)
	uow := factory.UoW()

	opts1 := &sql.TxOptions{ReadOnly: false}
	opts2 := &sql.TxOptions{ReadOnly: true}

	// Should use opts1 (first one), ignores opts2
	err = uow.Do(context.Background(), func(ctx context.Context, repo pgxRepoFactory) error {
		return repo.CreateUser(&pgxUser{Name: "Multi"})
	}, opts1, opts2)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPgxDo_WithDifferentIsolationLevels(t *testing.T) {
	testCases := []struct {
		name        string
		isolation   sql.IsolationLevel
		expectedIso pgx.TxIsoLevel
	}{
		{"Default", sql.LevelDefault, pgx.ReadCommitted},
		{"ReadUncommitted", sql.LevelReadUncommitted, pgx.ReadCommitted},
		{"ReadCommitted", sql.LevelReadCommitted, pgx.ReadCommitted},
		{"RepeatableRead", sql.LevelRepeatableRead, pgx.RepeatableRead},
		{"Serializable", sql.LevelSerializable, pgx.Serializable},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			if err != nil {
				t.Fatalf("failed to create mock pool: %v", err)
			}
			defer mock.Close()

			mock.ExpectBeginTx(pgx.TxOptions{
				IsoLevel:   tc.expectedIso,
				AccessMode: pgx.ReadWrite,
			})
			mock.ExpectExec("INSERT INTO users").
				WithArgs(fmt.Sprintf("User%d", i)).
				WillReturnResult(pgxmock.NewResult("INSERT", 1))
			mock.ExpectCommit()

			factory := NewPgx(mock, newPgxFactory)
			uow := factory.UoW()

			opts := &sql.TxOptions{
				Isolation: tc.isolation,
				ReadOnly:  false,
			}

			err = uow.Do(context.Background(), func(ctx context.Context, repo pgxRepoFactory) error {
				return repo.CreateUser(&pgxUser{Name: fmt.Sprintf("User%d", i)})
			}, opts)
			if err != nil {
				t.Fatalf("expected no error with isolation %v, got %v", tc.isolation, err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestPgxDo_WithReadOnlyTransaction(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock pool: %v", err)
	}
	defer mock.Close()

	// Read operation should succeed
	mock.ExpectBeginTx(pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadOnly,
	})
	mock.ExpectQuery("SELECT id, name FROM users WHERE id").
		WithArgs(1).
		WillReturnRows(pgxmock.NewRows([]string{"id", "name"}).AddRow(1, "ExistingUser"))
	mock.ExpectCommit()

	factory := NewPgx(mock, newPgxFactory)
	uow := factory.UoW()

	opts := &sql.TxOptions{
		ReadOnly: true,
	}

	// Try to read in read-only transaction - should succeed
	err = uow.Do(context.Background(), func(ctx context.Context, repo pgxRepoFactory) error {
		user, err := repo.GetUser(1)
		if err != nil {
			return err
		}
		if user.Name != "ExistingUser" {
			return fmt.Errorf("expected 'ExistingUser', got %q", user.Name)
		}
		return nil
	}, opts)
	if err != nil {
		t.Fatalf("expected no error in read-only transaction, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}

	// Write operation should fail
	mock2, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock pool: %v", err)
	}
	defer mock2.Close()

	mock2.ExpectBeginTx(pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadOnly,
	})
	mock2.ExpectExec("INSERT INTO users").
		WithArgs("NewUser").
		WillReturnError(errors.New("cannot execute INSERT in a read-only transaction"))
	mock2.ExpectRollback()

	factory2 := NewPgx(mock2, newPgxFactory)
	uow2 := factory2.UoW()

	err = uow2.Do(context.Background(), func(ctx context.Context, repo pgxRepoFactory) error {
		return repo.CreateUser(&pgxUser{Name: "NewUser"})
	}, opts)

	if err == nil {
		t.Fatalf("expected error when writing in read-only transaction, got nil")
	}

	if err := mock2.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPgxDo_ContextCancel(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock pool: %v", err)
	}
	defer mock.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	mock.ExpectBeginTx(pgx.TxOptions{AccessMode: pgx.ReadWrite}).
		WillReturnError(context.Canceled)

	factory := NewPgx(mock, newPgxFactory)
	uow := factory.UoW()

	err = uow.Do(ctx, func(ctx context.Context, repo pgxRepoFactory) error {
		return repo.CreateUser(&pgxUser{Name: "Cancelled"})
	})

	// Should get context cancelled error
	if err == nil {
		t.Fatalf("expected error for cancelled context")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// ============================================================================
// uowPgx.Begin() Tests
// ============================================================================

func TestPgxBegin_ReturnsContextAndRepoFactory(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock pool: %v", err)
	}
	defer mock.Close()

	mock.ExpectBeginTx(pgx.TxOptions{AccessMode: pgx.ReadWrite})

	factory := NewPgx(mock, newPgxFactory)
	uow := factory.UoW()

	ctx, repo, err := uow.Begin(context.Background())
	if err != nil {
		t.Fatalf("Begin returned error: %v", err)
	}

	if ctx == nil {
		t.Fatalf("expected non-nil context, got nil")
	}

	if repo == (pgxRepoFactory{}) {
		t.Fatalf("expected non-empty repo factory")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPgxBegin_ContextContainsTx(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock pool: %v", err)
	}
	defer mock.Close()

	mock.ExpectBeginTx(pgx.TxOptions{AccessMode: pgx.ReadWrite})

	factory := NewPgx(mock, newPgxFactory)
	uow := factory.UoW()

	ctx, _, err := uow.Begin(context.Background())
	if err != nil {
		t.Fatalf("Begin returned error: %v", err)
	}

	tx := ctx.Value(ctxTxKey{})
	if tx == nil {
		t.Fatalf("expected transaction in context")
	}

	// Verify it's the correct type
	if _, ok := tx.(pgx.Tx); !ok {
		t.Fatalf("expected pgx.Tx in context, got %T", tx)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPgxBegin_WithTxOptions(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock pool: %v", err)
	}
	defer mock.Close()

	mock.ExpectBeginTx(pgx.TxOptions{
		IsoLevel:   pgx.Serializable,
		AccessMode: pgx.ReadWrite,
	})
	mock.ExpectExec("INSERT INTO users").WithArgs("TxOptsUser").WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectCommit()

	factory := NewPgx(mock, newPgxFactory)
	uow := factory.UoW()

	opts := &sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	}

	ctx, repo, err := uow.Begin(context.Background(), opts)
	if err != nil {
		t.Fatalf("Begin with TxOptions returned error: %v", err)
	}

	if err := repo.CreateUser(&pgxUser{Name: "TxOptsUser"}); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if err := uow.Commit(ctx); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// ============================================================================
// uowPgx.Commit() Tests
// ============================================================================

func TestPgxCommit_CommitsTransaction(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock pool: %v", err)
	}
	defer mock.Close()

	mock.ExpectBeginTx(pgx.TxOptions{AccessMode: pgx.ReadWrite})
	mock.ExpectExec("INSERT INTO users").WithArgs("CommitTest").WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectCommit()

	factory := NewPgx(mock, newPgxFactory)
	uow := factory.UoW()

	ctx, repo, err := uow.Begin(context.Background())
	if err != nil {
		t.Fatalf("Begin returned error: %v", err)
	}

	if err := repo.CreateUser(&pgxUser{Name: "CommitTest"}); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if err := uow.Commit(ctx); err != nil {
		t.Fatalf("Commit returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPgxCommit_ErrorWhenNoTransactionInContext(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock pool: %v", err)
	}
	defer mock.Close()

	factory := NewPgx(mock, newPgxFactory)
	uow := factory.UoW()

	err = uow.Commit(context.Background())
	if err == nil {
		t.Fatalf("expected error when no transaction in context, got nil")
	}

	if !errors.Is(err, ErrTransactionNotFound) {
		if err.Error() != "no transaction found in context" {
			t.Fatalf("expected 'no transaction found in context' error, got %v", err)
		}
	}
}

func TestPgxCommit_ErrorWhenTransactionIsNil(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock pool: %v", err)
	}
	defer mock.Close()

	factory := NewPgx(mock, newPgxFactory)
	uow := factory.UoW()

	// Manually create context with nil transaction
	ctx := context.WithValue(context.Background(), ctxTxKey{}, (pgx.Tx)(nil))

	err = uow.Commit(ctx)
	if err == nil {
		t.Fatalf("expected error when transaction is nil, got nil")
	}
}

// ============================================================================
// uowPgx.Rollback() Tests
// ============================================================================

func TestPgxRollback_RollsBackTransaction(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock pool: %v", err)
	}
	defer mock.Close()

	mock.ExpectBeginTx(pgx.TxOptions{AccessMode: pgx.ReadWrite})
	mock.ExpectExec("INSERT INTO users").WithArgs("RollbackTest").WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectRollback()

	factory := NewPgx(mock, newPgxFactory)
	uow := factory.UoW()

	ctx, repo, err := uow.Begin(context.Background())
	if err != nil {
		t.Fatalf("Begin returned error: %v", err)
	}

	if err := repo.CreateUser(&pgxUser{Name: "RollbackTest"}); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if err := uow.Rollback(ctx); err != nil {
		t.Fatalf("Rollback returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPgxRollback_ErrorWhenNoTransactionInContext(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock pool: %v", err)
	}
	defer mock.Close()

	factory := NewPgx(mock, newPgxFactory)
	uow := factory.UoW()

	err = uow.Rollback(context.Background())
	if err == nil {
		t.Fatalf("expected error when no transaction in context, got nil")
	}

	if err.Error() != "no transaction found in context" {
		t.Fatalf("expected 'no transaction found in context' error, got %v", err)
	}
}

func TestPgxRollback_ErrorWhenTransactionIsNil(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock pool: %v", err)
	}
	defer mock.Close()

	factory := NewPgx(mock, newPgxFactory)
	uow := factory.UoW()

	ctx := context.WithValue(context.Background(), ctxTxKey{}, (pgx.Tx)(nil))

	err = uow.Rollback(ctx)
	if err == nil {
		t.Fatalf("expected error when transaction is nil, got nil")
	}
}

// ============================================================================
// uowPgx.SavePoint() Tests
// ============================================================================

func TestPgxSavePoint_CreatesSavePoint(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock pool: %v", err)
	}
	defer mock.Close()

	mock.ExpectBeginTx(pgx.TxOptions{AccessMode: pgx.ReadWrite})
	mock.ExpectExec("INSERT INTO users").WithArgs("User1").WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectExec("SAVEPOINT sp1").WillReturnResult(pgxmock.NewResult("SAVEPOINT", 0))
	mock.ExpectExec("INSERT INTO users").WithArgs("User2").WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectCommit()

	factory := NewPgx(mock, newPgxFactory)
	uow := factory.UoW()

	ctx, repo, err := uow.Begin(context.Background())
	if err != nil {
		t.Fatalf("Begin returned error: %v", err)
	}

	if err := repo.CreateUser(&pgxUser{Name: "User1"}); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if err := uow.SavePoint(ctx, "sp1"); err != nil {
		t.Fatalf("SavePoint returned error: %v", err)
	}

	// Continue with more operations
	if err := repo.CreateUser(&pgxUser{Name: "User2"}); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if err := uow.Commit(ctx); err != nil {
		t.Fatalf("Commit returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPgxSavePoint_ErrorWhenNoTransactionInContext(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock pool: %v", err)
	}
	defer mock.Close()

	factory := NewPgx(mock, newPgxFactory)
	uow := factory.UoW()

	err = uow.SavePoint(context.Background(), "sp1")
	if err == nil {
		t.Fatalf("expected error when no transaction in context, got nil")
	}

	if err.Error() != "no transaction found in context" {
		t.Fatalf("expected 'no transaction found in context' error, got %v", err)
	}
}

func TestPgxSavePoint_ErrorWhenTransactionIsNil(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock pool: %v", err)
	}
	defer mock.Close()

	factory := NewPgx(mock, newPgxFactory)
	uow := factory.UoW()

	ctx := context.WithValue(context.Background(), ctxTxKey{}, (pgx.Tx)(nil))

	err = uow.SavePoint(ctx, "sp1")
	if err == nil {
		t.Fatalf("expected error when transaction is nil, got nil")
	}
}

func TestPgxSavePoint_MultipleSavePoints(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock pool: %v", err)
	}
	defer mock.Close()

	mock.ExpectBeginTx(pgx.TxOptions{AccessMode: pgx.ReadWrite})
	mock.ExpectExec("INSERT INTO users").WithArgs("User1").WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectExec("SAVEPOINT sp1").WillReturnResult(pgxmock.NewResult("SAVEPOINT", 0))
	mock.ExpectExec("INSERT INTO users").WithArgs("User2").WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectExec("SAVEPOINT sp2").WillReturnResult(pgxmock.NewResult("SAVEPOINT", 0))
	mock.ExpectExec("INSERT INTO users").WithArgs("User3").WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectCommit()

	factory := NewPgx(mock, newPgxFactory)
	uow := factory.UoW()

	ctx, repo, err := uow.Begin(context.Background())
	if err != nil {
		t.Fatalf("Begin returned error: %v", err)
	}

	if err := repo.CreateUser(&pgxUser{Name: "User1"}); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if err := uow.SavePoint(ctx, "sp1"); err != nil {
		t.Fatalf("SavePoint sp1 returned error: %v", err)
	}

	if err := repo.CreateUser(&pgxUser{Name: "User2"}); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if err := uow.SavePoint(ctx, "sp2"); err != nil {
		t.Fatalf("SavePoint sp2 returned error: %v", err)
	}

	if err := repo.CreateUser(&pgxUser{Name: "User3"}); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if err := uow.Commit(ctx); err != nil {
		t.Fatalf("Commit returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// ============================================================================
// Integration Tests: Begin/Commit/Rollback Flow
// ============================================================================

func TestPgxIntegration_BeginCommitRollbackFlow(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock pool: %v", err)
	}
	defer mock.Close()

	// First transaction: successful
	mock.ExpectBeginTx(pgx.TxOptions{AccessMode: pgx.ReadWrite})
	mock.ExpectExec("INSERT INTO users").WithArgs("Alice").WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectCommit()

	// Second transaction: rollback
	mock.ExpectBeginTx(pgx.TxOptions{AccessMode: pgx.ReadWrite})
	mock.ExpectExec("INSERT INTO users").WithArgs("Bob").WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectRollback()

	factory := NewPgx(mock, newPgxFactory)
	uow := factory.UoW()

	// First transaction: successful
	ctx1, repo1, err := uow.Begin(context.Background())
	if err != nil {
		t.Fatalf("Begin returned error: %v", err)
	}

	if err := repo1.CreateUser(&pgxUser{Name: "Alice"}); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if err := uow.Commit(ctx1); err != nil {
		t.Fatalf("Commit returned error: %v", err)
	}

	// Second transaction: rollback
	ctx2, repo2, err := uow.Begin(context.Background())
	if err != nil {
		t.Fatalf("Begin returned error: %v", err)
	}

	if err := repo2.CreateUser(&pgxUser{Name: "Bob"}); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if err := uow.Rollback(ctx2); err != nil {
		t.Fatalf("Rollback returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPgxIntegration_NestedSavePointAndRollback(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock pool: %v", err)
	}
	defer mock.Close()

	mock.ExpectBeginTx(pgx.TxOptions{AccessMode: pgx.ReadWrite})
	mock.ExpectExec("INSERT INTO users").WithArgs("User1").WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectExec("SAVEPOINT sp1").WillReturnResult(pgxmock.NewResult("SAVEPOINT", 0))
	mock.ExpectExec("INSERT INTO users").WithArgs("User2").WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectExec("SAVEPOINT sp2").WillReturnResult(pgxmock.NewResult("SAVEPOINT", 0))
	mock.ExpectCommit()

	factory := NewPgx(mock, newPgxFactory)
	uow := factory.UoW()

	ctx, repo, err := uow.Begin(context.Background())
	if err != nil {
		t.Fatalf("Begin returned error: %v", err)
	}

	// Create first user
	if err := repo.CreateUser(&pgxUser{Name: "User1"}); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Create savepoint
	if err := uow.SavePoint(ctx, "sp1"); err != nil {
		t.Fatalf("SavePoint returned error: %v", err)
	}

	// Create second user
	if err := repo.CreateUser(&pgxUser{Name: "User2"}); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Create another savepoint
	if err := uow.SavePoint(ctx, "sp2"); err != nil {
		t.Fatalf("SavePoint returned error: %v", err)
	}

	if err := uow.Commit(ctx); err != nil {
		t.Fatalf("Commit returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestPgxIntegration_DoVsManualBeginCommit_ProducesSameResult(t *testing.T) {
	// Test with Do
	mock1, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock pool: %v", err)
	}
	defer mock1.Close()

	mock1.ExpectBeginTx(pgx.TxOptions{AccessMode: pgx.ReadWrite})
	mock1.ExpectExec("INSERT INTO users").WithArgs("User1").WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock1.ExpectExec("INSERT INTO users").WithArgs("User2").WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock1.ExpectCommit()

	factory1 := NewPgx(mock1, newPgxFactory)
	uow1 := factory1.UoW()

	err1 := uow1.Do(context.Background(), func(ctx context.Context, repo pgxRepoFactory) error {
		if err := repo.CreateUser(&pgxUser{Name: "User1"}); err != nil {
			return err
		}
		if err := repo.CreateUser(&pgxUser{Name: "User2"}); err != nil {
			return err
		}
		return nil
	})

	if err1 != nil {
		t.Fatalf("Do returned error: %v", err1)
	}

	if err := mock1.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}

	// Test with Begin/Commit
	mock2, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create mock pool: %v", err)
	}
	defer mock2.Close()

	mock2.ExpectBeginTx(pgx.TxOptions{AccessMode: pgx.ReadWrite})
	mock2.ExpectExec("INSERT INTO users").WithArgs("User1").WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock2.ExpectExec("INSERT INTO users").WithArgs("User2").WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock2.ExpectCommit()

	factory2 := NewPgx(mock2, newPgxFactory)
	uow2 := factory2.UoW()

	ctx, repo, err := uow2.Begin(context.Background())
	if err != nil {
		t.Fatalf("Begin returned error: %v", err)
	}

	if err := repo.CreateUser(&pgxUser{Name: "User1"}); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if err := repo.CreateUser(&pgxUser{Name: "User2"}); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if err := uow2.Commit(ctx); err != nil {
		t.Fatalf("Commit returned error: %v", err)
	}

	if err := mock2.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// ============================================================================
// convertPgxTxOptions Tests
// ============================================================================

func TestConvertPgxTxOptions_NilOptions(t *testing.T) {
	opts := convertPgxTxOptions(nil)

	if opts.IsoLevel != "" {
		t.Fatalf("expected empty IsoLevel for nil options, got %v", opts.IsoLevel)
	}

	if opts.AccessMode != pgx.ReadWrite {
		t.Fatalf("expected ReadWrite AccessMode for nil options, got %v", opts.AccessMode)
	}
}

func TestConvertPgxTxOptions_IsolationLevels(t *testing.T) {
	testCases := []struct {
		sqlLevel sql.IsolationLevel
		pgxLevel pgx.TxIsoLevel
	}{
		{sql.LevelDefault, pgx.ReadCommitted},
		{sql.LevelReadUncommitted, pgx.ReadCommitted}, // Postgres doesn't support dirty reads
		{sql.LevelReadCommitted, pgx.ReadCommitted},
		{sql.LevelRepeatableRead, pgx.RepeatableRead},
		{sql.LevelSerializable, pgx.Serializable},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("SQL_%v_to_PGX_%v", tc.sqlLevel, tc.pgxLevel), func(t *testing.T) {
			sqlOpts := &sql.TxOptions{Isolation: tc.sqlLevel}
			pgxOpts := convertPgxTxOptions(sqlOpts)

			if pgxOpts.IsoLevel != tc.pgxLevel {
				t.Fatalf("expected %v, got %v", tc.pgxLevel, pgxOpts.IsoLevel)
			}
		})
	}
}

func TestConvertPgxTxOptions_AccessMode(t *testing.T) {
	testCases := []struct {
		name       string
		readOnly   bool
		accessMode pgx.TxAccessMode
	}{
		{"ReadWrite", false, pgx.ReadWrite},
		{"ReadOnly", true, pgx.ReadOnly},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sqlOpts := &sql.TxOptions{ReadOnly: tc.readOnly}
			pgxOpts := convertPgxTxOptions(sqlOpts)

			if pgxOpts.AccessMode != tc.accessMode {
				t.Fatalf("expected %v, got %v", tc.accessMode, pgxOpts.AccessMode)
			}
		})
	}
}

func TestChooseAccessMode_NilOptions(t *testing.T) {
	mode := chooseAccessMode(nil)

	if mode != pgx.ReadWrite {
		t.Fatalf("expected ReadWrite for nil options, got %v", mode)
	}
}

func TestChooseAccessMode_ReadOnly(t *testing.T) {
	opts := &sql.TxOptions{ReadOnly: true}
	mode := chooseAccessMode(opts)

	if mode != pgx.ReadOnly {
		t.Fatalf("expected ReadOnly, got %v", mode)
	}
}

func TestChooseAccessMode_ReadWrite(t *testing.T) {
	opts := &sql.TxOptions{ReadOnly: false}
	mode := chooseAccessMode(opts)

	if mode != pgx.ReadWrite {
		t.Fatalf("expected ReadWrite, got %v", mode)
	}
}
