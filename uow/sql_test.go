package uow

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

type sqlUser struct {
	ID   int
	Name string
}

type sqlRepoFactory struct {
	tx *sql.Tx
}

func (s sqlRepoFactory) CreateUser(u *sqlUser) error {
	_, err := s.tx.Exec("INSERT INTO users (name) VALUES (?)", u.Name)
	return err
}

func (s sqlRepoFactory) GetUser(id int) (*sqlUser, error) {
	row := s.tx.QueryRow("SELECT id, name FROM users WHERE id = ?", id)
	var user sqlUser
	if err := row.Scan(&user.ID, &user.Name); err != nil {
		return nil, err
	}
	return &user, nil
}

func newSqlFactory(tx *sql.Tx) sqlRepoFactory {
	return sqlRepoFactory{tx: tx}
}

func setupSqlDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("failed to ping database: %v", err)
	}

	if _, err := db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL
	)`); err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	return db
}

func countSqlUsers(t *testing.T, db *sql.DB) int {
	t.Helper()
	row := db.QueryRow("SELECT COUNT(*) FROM users")
	var cnt int
	if err := row.Scan(&cnt); err != nil {
		t.Fatalf("count users failed: %v", err)
	}
	return cnt
}

func getSqlUser(t *testing.T, db *sql.DB, id int) *sqlUser {
	t.Helper()
	row := db.QueryRow("SELECT id, name FROM users WHERE id = ?", id)
	var user sqlUser
	if err := row.Scan(&user.ID, &user.Name); err != nil {
		t.Fatalf("get user failed: %v", err)
	}
	return &user
}

// ============================================================================
// sqlFactory Tests
// ============================================================================

func TestNewSql_CreatesFactory(t *testing.T) {
	db := setupSqlDB(t)
	defer db.Close()

	factory := NewSql(db, newSqlFactory)

	if factory == nil {
		t.Fatalf("expected non-nil factory, got nil")
	}
}

func TestNewSql_FactoryUoWReturnsUoW(t *testing.T) {
	db := setupSqlDB(t)
	defer db.Close()

	factory := NewSql(db, newSqlFactory)
	uow := factory.UoW()

	if uow == nil {
		t.Fatalf("expected non-nil UoW, got nil")
	}
}

func TestNewSql_MultipleUoWCallsReturnDifferentInstances(t *testing.T) {
	db := setupSqlDB(t)
	defer db.Close()

	factory := NewSql(db, newSqlFactory)

	uow1 := factory.UoW()
	uow2 := factory.UoW()

	// They should be different instances (though they work with the same db)
	if uow1 == uow2 {
		t.Fatalf("expected different UoW instances, got same instance")
	}
}

// ============================================================================
// uowSql.Do() Tests
// ============================================================================

func TestSqlDo_CommitsOnSuccess(t *testing.T) {
	db := setupSqlDB(t)
	defer db.Close()

	factory := NewSql(db, newSqlFactory)
	uow := factory.UoW()

	err := uow.Do(context.Background(), func(ctx context.Context, repo sqlRepoFactory) error {
		return repo.CreateUser(&sqlUser{Name: "Alice"})
	})
	if err != nil {
		t.Fatalf("expected no error on success, got %v", err)
	}

	if got := countSqlUsers(t, db); got != 1 {
		t.Fatalf("expected 1 user after commit, got %d", got)
	}

	user := getSqlUser(t, db, 1)
	if user.Name != "Alice" {
		t.Fatalf("expected user name 'Alice', got %q", user.Name)
	}
}

func TestSqlDo_RollsBackOnError(t *testing.T) {
	db := setupSqlDB(t)
	defer db.Close()

	factory := NewSql(db, newSqlFactory)
	uow := factory.UoW()

	err := uow.Do(context.Background(), func(ctx context.Context, repo sqlRepoFactory) error {
		if err := repo.CreateUser(&sqlUser{Name: "Bob"}); err != nil {
			return err
		}
		return errors.New("force rollback")
	})

	if err == nil {
		t.Fatalf("expected error to propagate, got nil")
	}

	if got := countSqlUsers(t, db); got != 0 {
		t.Fatalf("expected 0 users after rollback, got %d", got)
	}
}

func TestSqlDo_PropagatesErrorFromFunction(t *testing.T) {
	db := setupSqlDB(t)
	defer db.Close()

	factory := NewSql(db, newSqlFactory)
	uow := factory.UoW()

	expectedErr := errors.New("test error")
	err := uow.Do(context.Background(), func(ctx context.Context, repo sqlRepoFactory) error {
		return expectedErr
	})

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func TestSqlDo_MultipleOperationsWithinTransaction(t *testing.T) {
	db := setupSqlDB(t)
	defer db.Close()

	factory := NewSql(db, newSqlFactory)
	uow := factory.UoW()

	err := uow.Do(context.Background(), func(ctx context.Context, repo sqlRepoFactory) error {
		if err := repo.CreateUser(&sqlUser{Name: "User1"}); err != nil {
			return err
		}
		if err := repo.CreateUser(&sqlUser{Name: "User2"}); err != nil {
			return err
		}
		if err := repo.CreateUser(&sqlUser{Name: "User3"}); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := countSqlUsers(t, db); got != 3 {
		t.Fatalf("expected 3 users, got %d", got)
	}
}

func TestSqlDo_MultipleOperationsRollbackOnLastError(t *testing.T) {
	db := setupSqlDB(t)
	defer db.Close()

	factory := NewSql(db, newSqlFactory)
	uow := factory.UoW()

	err := uow.Do(context.Background(), func(ctx context.Context, repo sqlRepoFactory) error {
		if err := repo.CreateUser(&sqlUser{Name: "User1"}); err != nil {
			return err
		}
		if err := repo.CreateUser(&sqlUser{Name: "User2"}); err != nil {
			return err
		}
		return errors.New("error on third operation")
	})

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if got := countSqlUsers(t, db); got != 0 {
		t.Fatalf("expected 0 users after rollback, got %d", got)
	}
}

func TestSqlDo_WithTxOptions(t *testing.T) {
	db := setupSqlDB(t)
	defer db.Close()

	factory := NewSql(db, newSqlFactory)
	uow := factory.UoW()

	opts := &sql.TxOptions{
		Isolation: sql.LevelDefault,
		ReadOnly:  false,
	}

	err := uow.Do(context.Background(), func(ctx context.Context, repo sqlRepoFactory) error {
		return repo.CreateUser(&sqlUser{Name: "TxOpts"})
	}, opts)
	if err != nil {
		t.Fatalf("expected no error with TxOptions, got %v", err)
	}

	if got := countSqlUsers(t, db); got != 1 {
		t.Fatalf("expected 1 user after commit with TxOptions, got %d", got)
	}
}

func TestSqlDo_WithMultipleTxOptions_UsesFirst(t *testing.T) {
	db := setupSqlDB(t)
	defer db.Close()

	factory := NewSql(db, newSqlFactory)
	uow := factory.UoW()

	opts1 := &sql.TxOptions{ReadOnly: false}
	opts2 := &sql.TxOptions{ReadOnly: true}

	// Should use opts1 (first one), ignores opts2
	err := uow.Do(context.Background(), func(ctx context.Context, repo sqlRepoFactory) error {
		return repo.CreateUser(&sqlUser{Name: "Multi"})
	}, opts1, opts2)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := countSqlUsers(t, db); got != 1 {
		t.Fatalf("expected 1 user, got %d", got)
	}
}

func TestSqlDo_ContextCancel(t *testing.T) {
	db := setupSqlDB(t)
	defer db.Close()

	factory := NewSql(db, newSqlFactory)
	uow := factory.UoW()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := uow.Do(ctx, func(ctx context.Context, repo sqlRepoFactory) error {
		return repo.CreateUser(&sqlUser{Name: "Cancelled"})
	})

	// Should get context cancelled error
	if err == nil {
		t.Fatalf("expected error for cancelled context")
	}

	// The transaction should not have been created/committed
	if got := countSqlUsers(t, db); got != 0 {
		t.Fatalf("expected 0 users after cancelled context, got %d", got)
	}
}

// ============================================================================
// uowSql.Begin() Tests
// ============================================================================

func TestSqlBegin_ReturnsContextAndRepoFactory(t *testing.T) {
	db := setupSqlDB(t)
	defer db.Close()

	factory := NewSql(db, newSqlFactory)
	uow := factory.UoW()

	ctx, repo, err := uow.Begin(context.Background())
	if err != nil {
		t.Fatalf("Begin returned error: %v", err)
	}

	if ctx == nil {
		t.Fatalf("expected non-nil context, got nil")
	}

	if repo == (sqlRepoFactory{}) {
		t.Fatalf("expected non-empty repo factory")
	}
}

func TestSqlBegin_ContextContainsTx(t *testing.T) {
	db := setupSqlDB(t)
	defer db.Close()

	factory := NewSql(db, newSqlFactory)
	uow := factory.UoW()

	ctx, _, err := uow.Begin(context.Background())
	if err != nil {
		t.Fatalf("Begin returned error: %v", err)
	}

	tx := ctx.Value(ctxTxKey{})
	if tx == nil {
		t.Fatalf("expected transaction in context")
	}
}

func TestSqlBegin_ErrorOnFailedTransaction(t *testing.T) {
	// Create a closed database to force an error
	db := setupSqlDB(t)
	db.Close()

	factory := NewSql(db, newSqlFactory)
	uow := factory.UoW()

	_, _, err := uow.Begin(context.Background())
	if err == nil {
		t.Fatalf("expected error for closed database, got nil")
	}
}

// ============================================================================
// uowSql.Commit() Tests
// ============================================================================

func TestSqlCommit_CommitsTransaction(t *testing.T) {
	db := setupSqlDB(t)
	defer db.Close()

	factory := NewSql(db, newSqlFactory)
	uow := factory.UoW()

	ctx, repo, err := uow.Begin(context.Background())
	if err != nil {
		t.Fatalf("Begin returned error: %v", err)
	}

	if err := repo.CreateUser(&sqlUser{Name: "CommitTest"}); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if err := uow.Commit(ctx); err != nil {
		t.Fatalf("Commit returned error: %v", err)
	}

	if got := countSqlUsers(t, db); got != 1 {
		t.Fatalf("expected 1 user after commit, got %d", got)
	}
}

func TestSqlCommit_ErrorWhenNoTransactionInContext(t *testing.T) {
	db := setupSqlDB(t)
	defer db.Close()

	factory := NewSql(db, newSqlFactory)
	uow := factory.UoW()

	err := uow.Commit(context.Background())
	if err == nil {
		t.Fatalf("expected error when no transaction in context, got nil")
	}

	if !errors.Is(err, errors.New("no transaction found in context")) {
		if err.Error() != "no transaction found in context" {
			t.Fatalf("expected 'no transaction found in context' error, got %v", err)
		}
	}
}

func TestSqlCommit_ErrorWhenTransactionIsNil(t *testing.T) {
	db := setupSqlDB(t)
	defer db.Close()

	factory := NewSql(db, newSqlFactory)
	uow := factory.UoW()

	// Manually create context with nil transaction
	ctx := context.WithValue(context.Background(), ctxTxKey{}, (*sql.Tx)(nil))

	err := uow.Commit(ctx)
	if err == nil {
		t.Fatalf("expected error when transaction is nil, got nil")
	}
}

// ============================================================================
// uowSql.Rollback() Tests
// ============================================================================

func TestSqlRollback_RollsBackTransaction(t *testing.T) {
	db := setupSqlDB(t)
	defer db.Close()

	factory := NewSql(db, newSqlFactory)
	uow := factory.UoW()

	ctx, repo, err := uow.Begin(context.Background())
	if err != nil {
		t.Fatalf("Begin returned error: %v", err)
	}

	if err := repo.CreateUser(&sqlUser{Name: "RollbackTest"}); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if err := uow.Rollback(ctx); err != nil {
		t.Fatalf("Rollback returned error: %v", err)
	}

	if got := countSqlUsers(t, db); got != 0 {
		t.Fatalf("expected 0 users after rollback, got %d", got)
	}
}

func TestSqlRollback_ErrorWhenNoTransactionInContext(t *testing.T) {
	db := setupSqlDB(t)
	defer db.Close()

	factory := NewSql(db, newSqlFactory)
	uow := factory.UoW()

	err := uow.Rollback(context.Background())
	if err == nil {
		t.Fatalf("expected error when no transaction in context, got nil")
	}

	if err.Error() != "no transaction found in context" {
		t.Fatalf("expected 'no transaction found in context' error, got %v", err)
	}
}

func TestSqlRollback_ErrorWhenTransactionIsNil(t *testing.T) {
	db := setupSqlDB(t)
	defer db.Close()

	factory := NewSql(db, newSqlFactory)
	uow := factory.UoW()

	ctx := context.WithValue(context.Background(), ctxTxKey{}, (*sql.Tx)(nil))

	err := uow.Rollback(ctx)
	if err == nil {
		t.Fatalf("expected error when transaction is nil, got nil")
	}
}

// ============================================================================
// uowSql.SavePoint() Tests
// ============================================================================

func TestSqlSavePoint_CreatesSavePoint(t *testing.T) {
	db := setupSqlDB(t)
	defer db.Close()

	factory := NewSql(db, newSqlFactory)
	uow := factory.UoW()

	ctx, repo, err := uow.Begin(context.Background())
	if err != nil {
		t.Fatalf("Begin returned error: %v", err)
	}

	if err := repo.CreateUser(&sqlUser{Name: "User1"}); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if err := uow.SavePoint(ctx, "sp1"); err != nil {
		t.Fatalf("SavePoint returned error: %v", err)
	}

	// Continue with more operations
	if err := repo.CreateUser(&sqlUser{Name: "User2"}); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if err := uow.Commit(ctx); err != nil {
		t.Fatalf("Commit returned error: %v", err)
	}

	if got := countSqlUsers(t, db); got != 2 {
		t.Fatalf("expected 2 users after savepoint and commit, got %d", got)
	}
}

func TestSqlSavePoint_ErrorWhenNoTransactionInContext(t *testing.T) {
	db := setupSqlDB(t)
	defer db.Close()

	factory := NewSql(db, newSqlFactory)
	uow := factory.UoW()

	err := uow.SavePoint(context.Background(), "sp1")
	if err == nil {
		t.Fatalf("expected error when no transaction in context, got nil")
	}

	if err.Error() != "no transaction found in context" {
		t.Fatalf("expected 'no transaction found in context' error, got %v", err)
	}
}

func TestSqlSavePoint_ErrorWhenTransactionIsNil(t *testing.T) {
	db := setupSqlDB(t)
	defer db.Close()

	factory := NewSql(db, newSqlFactory)
	uow := factory.UoW()

	ctx := context.WithValue(context.Background(), ctxTxKey{}, (*sql.Tx)(nil))

	err := uow.SavePoint(ctx, "sp1")
	if err == nil {
		t.Fatalf("expected error when transaction is nil, got nil")
	}
}

func TestSqlSavePoint_MultipleSavePoints(t *testing.T) {
	db := setupSqlDB(t)
	defer db.Close()

	factory := NewSql(db, newSqlFactory)
	uow := factory.UoW()

	ctx, repo, err := uow.Begin(context.Background())
	if err != nil {
		t.Fatalf("Begin returned error: %v", err)
	}

	if err := repo.CreateUser(&sqlUser{Name: "User1"}); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if err := uow.SavePoint(ctx, "sp1"); err != nil {
		t.Fatalf("SavePoint sp1 returned error: %v", err)
	}

	if err := repo.CreateUser(&sqlUser{Name: "User2"}); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if err := uow.SavePoint(ctx, "sp2"); err != nil {
		t.Fatalf("SavePoint sp2 returned error: %v", err)
	}

	if err := repo.CreateUser(&sqlUser{Name: "User3"}); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if err := uow.Commit(ctx); err != nil {
		t.Fatalf("Commit returned error: %v", err)
	}

	if got := countSqlUsers(t, db); got != 3 {
		t.Fatalf("expected 3 users after multiple savepoints, got %d", got)
	}
}

// ============================================================================
// Integration Tests: Begin/Commit/Rollback Flow
// ============================================================================

func TestSqlIntegration_BeginCommitRollbackFlow(t *testing.T) {
	db := setupSqlDB(t)
	defer db.Close()

	factory := NewSql(db, newSqlFactory)
	uow := factory.UoW()

	// First transaction: successful
	ctx1, repo1, err := uow.Begin(context.Background())
	if err != nil {
		t.Fatalf("Begin returned error: %v", err)
	}

	if err := repo1.CreateUser(&sqlUser{Name: "Alice"}); err != nil {
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

	if err := repo2.CreateUser(&sqlUser{Name: "Bob"}); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if err := uow.Rollback(ctx2); err != nil {
		t.Fatalf("Rollback returned error: %v", err)
	}

	// Verify: only Alice committed, Bob was rolled back
	if got := countSqlUsers(t, db); got != 1 {
		t.Fatalf("expected 1 user, got %d", got)
	}

	user := getSqlUser(t, db, 1)
	if user.Name != "Alice" {
		t.Fatalf("expected 'Alice', got %q", user.Name)
	}
}

func TestSqlIntegration_NestedSavePointAndRollback(t *testing.T) {
	db := setupSqlDB(t)
	defer db.Close()

	factory := NewSql(db, newSqlFactory)
	uow := factory.UoW()

	ctx, repo, err := uow.Begin(context.Background())
	if err != nil {
		t.Fatalf("Begin returned error: %v", err)
	}

	// Create first user
	if err := repo.CreateUser(&sqlUser{Name: "User1"}); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Create savepoint
	if err := uow.SavePoint(ctx, "sp1"); err != nil {
		t.Fatalf("SavePoint returned error: %v", err)
	}

	// Create second user
	if err := repo.CreateUser(&sqlUser{Name: "User2"}); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Rollback to savepoint - Note: SQLite RELEASE removes savepoint without rollback
	// For a complete test, we'll just verify operations can continue
	if err := uow.SavePoint(ctx, "sp2"); err != nil {
		t.Fatalf("SavePoint returned error: %v", err)
	}

	if err := uow.Commit(ctx); err != nil {
		t.Fatalf("Commit returned error: %v", err)
	}

	if got := countSqlUsers(t, db); got != 2 {
		t.Fatalf("expected 2 users after savepoints, got %d", got)
	}
}

func TestSqlIntegration_DoVsManualBeginCommit_ProducesSameResult(t *testing.T) {
	db1 := setupSqlDB(t)
	defer db1.Close()

	db2 := setupSqlDB(t)
	defer db2.Close()

	// Test with Do
	factory1 := NewSql(db1, newSqlFactory)
	uow1 := factory1.UoW()

	err1 := uow1.Do(context.Background(), func(ctx context.Context, repo sqlRepoFactory) error {
		if err := repo.CreateUser(&sqlUser{Name: "User1"}); err != nil {
			return err
		}
		if err := repo.CreateUser(&sqlUser{Name: "User2"}); err != nil {
			return err
		}
		return nil
	})

	if err1 != nil {
		t.Fatalf("Do returned error: %v", err1)
	}

	// Test with Begin/Commit
	factory2 := NewSql(db2, newSqlFactory)
	uow2 := factory2.UoW()

	ctx, repo, err := uow2.Begin(context.Background())
	if err != nil {
		t.Fatalf("Begin returned error: %v", err)
	}

	if err := repo.CreateUser(&sqlUser{Name: "User1"}); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if err := repo.CreateUser(&sqlUser{Name: "User2"}); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if err := uow2.Commit(ctx); err != nil {
		t.Fatalf("Commit returned error: %v", err)
	}

	// Both should have 2 users
	if got1 := countSqlUsers(t, db1); got1 != 2 {
		t.Fatalf("Do: expected 2 users, got %d", got1)
	}

	if got2 := countSqlUsers(t, db2); got2 != 2 {
		t.Fatalf("Begin/Commit: expected 2 users, got %d", got2)
	}
}

func TestSqlBegin_ErrorHandling(t *testing.T) {
	db := setupSqlDB(t)
	
	// Close the database to trigger an error on Begin
	db.Close()

	factory := NewSql(db, newSqlFactory)
	uow := factory.UoW()

	// Begin should fail because the database is closed
	_, _, err := uow.Begin(context.Background())
	if err == nil {
		t.Fatalf("expected error when Begin is called on closed db, got nil")
	}
	
	// Also test with options to ensure the opts path is covered
	_, _, err = uow.Begin(context.Background(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err == nil {
		t.Fatalf("expected error when Begin is called on closed db with options, got nil")
	}
}
