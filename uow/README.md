# Unit of Work (UoW)

Database transaction patterns for GORM, PGX, and standard SQL with a clean, consistent interface.

## Overview

The Unit of Work pattern provides a clean abstraction for managing database transactions across multiple operations. This package implements the pattern for popular Go database libraries.

## Installation

```bash
go get github.com/maadiii/goutils/uow
```

## Features

- 🗄️ GORM support
- 🐘 PostgreSQL PGX support
- 💾 Standard database/sql support
- 🔄 Automatic transaction management
- 📦 Repository factory pattern
- ✅ Commit/Rollback handling
- 🎯 Context-aware operations
- 💾 Savepoint support

## Supported Databases

- **GORM**: Any GORM-supported database (PostgreSQL, MySQL, SQLite, SQL Server)
- **PGX**: PostgreSQL via pgx driver
- **SQL**: Any database/sql compatible driver

## Usage

### GORM

#### Basic Setup

```go
package main

import (
    "context"
    "github.com/maadiii/goutils/uow"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

type User struct {
    ID   uint
    Name string
}

type UserRepository struct {
    db *gorm.DB
}

func (r *UserRepository) Create(user *User) error {
    return r.db.Create(user).Error
}

func (r *UserRepository) FindByID(id uint) (*User, error) {
    var user User
    err := r.db.First(&user, id).Error
    return &user, err
}

func NewUserRepository(tx *gorm.DB) *UserRepository {
    return &UserRepository{db: tx}
}

func main() {
    // Connect to database
    db, err := gorm.Open(postgres.Open("dsn"), &gorm.Config{})
    if err != nil {
        panic(err)
    }

    // Create factory
    factory := uow.NewGorm(db, NewUserRepository)

    // Get Unit of Work
    uow := factory.UoW()

    // Use Do method (automatic commit/rollback)
    err = uow.Do(context.Background(), func(ctx context.Context, repo *UserRepository) error {
        user := &User{Name: "John Doe"}
        if err := repo.Create(user); err != nil {
            return err // Will rollback
        }

        // More operations...

        return nil // Will commit
    })

    if err != nil {
        // Transaction was rolled back
        panic(err)
    }
}
```

#### Manual Transaction Control

```go
// Begin transaction
ctx, repo, err := uow.Begin(context.Background())
if err != nil {
    panic(err)
}

// Perform operations
user := &User{Name: "Jane Doe"}
if err := repo.Create(user); err != nil {
    uow.Rollback(ctx)
    panic(err)
}

// Commit transaction
if err := uow.Commit(ctx); err != nil {
    panic(err)
}
```

#### With Savepoints

```go
err = uow.Do(context.Background(), func(ctx context.Context, repo *UserRepository) error {
    user1 := &User{Name: "User 1"}
    repo.Create(user1)

    // Create savepoint
    if err := uow.SavePoint(ctx, "sp1"); err != nil {
        return err
    }

    user2 := &User{Name: "User 2"}
    if err := repo.Create(user2); err != nil {
        // Rollback to savepoint
        return fmt.Errorf("ROLLBACK TO sp1: %w", err)
    }

    return nil
})
```

### PGX (PostgreSQL)

```go
package main

import (
    "context"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/maadiii/goutils/uow"
)

type AccountRepository struct {
    tx pgx.Tx
}

func (r *AccountRepository) Transfer(from, to int, amount decimal.Decimal) error {
    // Debit from account
    _, err := r.tx.Exec(context.Background(),
        "UPDATE accounts SET balance = balance - $1 WHERE id = $2",
        amount, from)
    if err != nil {
        return err
    }

    // Credit to account
    _, err = r.tx.Exec(context.Background(),
        "UPDATE accounts SET balance = balance + $1 WHERE id = $2",
        amount, to)
    return err
}

func NewAccountRepository(tx pgx.Tx) *AccountRepository {
    return &AccountRepository{tx: tx}
}

func main() {
    // Create connection pool
    pool, err := pgxpool.New(context.Background(), "postgres://...")
    if err != nil {
        panic(err)
    }
    defer pool.Close()

    // Create factory
    factory := uow.NewPgx(pool, NewAccountRepository)
    uow := factory.UoW()

    // Transfer money with transaction
    err = uow.Do(context.Background(), func(ctx context.Context, repo *AccountRepository) error {
        return repo.Transfer(1, 2, decimal.NewFromFloat(100.00))
    })

    if err != nil {
        log.Println("Transfer failed:", err)
    }
}
```

### Standard SQL

```go
package main

import (
    "context"
    "database/sql"
    "github.com/maadiii/goutils/uow"
    _ "github.com/lib/pq"
)

type OrderRepository struct {
    tx *sql.Tx
}

func (r *OrderRepository) CreateOrder(order *Order) error {
    _, err := r.tx.Exec(
        "INSERT INTO orders (user_id, total) VALUES ($1, $2)",
        order.UserID, order.Total,
    )
    return err
}

func (r *OrderRepository) CreateOrderItems(orderID int, items []OrderItem) error {
    stmt, err := r.tx.Prepare("INSERT INTO order_items (order_id, product_id, quantity) VALUES ($1, $2, $3)")
    if err != nil {
        return err
    }
    defer stmt.Close()

    for _, item := range items {
        if _, err := stmt.Exec(orderID, item.ProductID, item.Quantity); err != nil {
            return err
        }
    }
    return nil
}

func NewOrderRepository(tx *sql.Tx) *OrderRepository {
    return &OrderRepository{tx: tx}
}

func main() {
    db, err := sql.Open("postgres", "postgres://...")
    if err != nil {
        panic(err)
    }
    defer db.Close()

    factory := uow.NewSql(db, NewOrderRepository)
    uow := factory.UoW()

    // Create order with items atomically
    err = uow.Do(context.Background(), func(ctx context.Context, repo *OrderRepository) error {
        order := &Order{UserID: 123, Total: 99.99}
        if err := repo.CreateOrder(order); err != nil {
            return err
        }

        items := []OrderItem{
            {ProductID: 1, Quantity: 2},
            {ProductID: 2, Quantity: 1},
        }

        return repo.CreateOrderItems(order.ID, items)
    })
}
```

## Complete Example: Multi-Repository Transaction

```go
type UserRepository struct {
    db *gorm.DB
}

func (r *UserRepository) Create(user *User) error {
    return r.db.Create(user).Error
}

type AccountRepository struct {
    db *gorm.DB
}

func (r *AccountRepository) Create(account *Account) error {
    return r.db.Create(account).Error
}

// Factory that creates multiple repositories
type RepositoryFactory struct {
    User    *UserRepository
    Account *AccountRepository
}

func NewRepositoryFactory(tx *gorm.DB) *RepositoryFactory {
    return &RepositoryFactory{
        User:    &UserRepository{db: tx},
        Account: &AccountRepository{db: tx},
    }
}

func main() {
    db, _ := gorm.Open(postgres.Open("dsn"), &gorm.Config{})

    factory := uow.NewGorm(db, NewRepositoryFactory)
    uow := factory.UoW()

    // Use multiple repositories in one transaction
    err := uow.Do(context.Background(), func(ctx context.Context, repos *RepositoryFactory) error {
        // Create user
        user := &User{Name: "John Doe", Email: "john@example.com"}
        if err := repos.User.Create(user); err != nil {
            return err
        }

        // Create account for user
        account := &Account{UserID: user.ID, Balance: 0}
        if err := repos.Account.Create(account); err != nil {
            return err // Both operations will rollback
        }

        return nil // Both operations will commit
    })
}
```

## API Reference

### Interfaces

#### `Factory[RepoFactory any]`

```go
type Factory[RepoFactory any] interface {
    UoW() UoW[RepoFactory]
}
```

#### `UoW[RepoFactory any]`

```go
type UoW[RepoFactory any] interface {
    Do(ctx context.Context, fn func(ctx context.Context, repo RepoFactory) error, opts ...*sql.TxOptions) error
    Begin(ctx context.Context, opts ...*sql.TxOptions) (context.Context, RepoFactory, error)
    SavePoint(ctx context.Context, name string) error
    Commit(ctx context.Context) error
    Rollback(ctx context.Context) error
}
```

### Functions

#### GORM

```go
func NewGorm[RepoFactory any](
    db *gorm.DB,
    factory func(*gorm.DB) RepoFactory,
) Factory[RepoFactory]
```

#### PGX

```go
func NewPgx[RepoFactory any](
    pool *pgxpool.Pool,
    factory func(pgx.Tx) RepoFactory,
) Factory[RepoFactory]
```

#### SQL

```go
func NewSql[RepoFactory any](
    db *sql.DB,
    factory func(*sql.Tx) RepoFactory,
) Factory[RepoFactory]
```

### Methods

#### `Do(ctx, fn, opts) error`

Executes a function within a transaction. Automatically commits on success or rolls back on error.

#### `Begin(ctx, opts) (context.Context, RepoFactory, error)`

Begins a new transaction and returns the context and repository factory.

#### `Commit(ctx) error`

Commits the transaction associated with the context.

#### `Rollback(ctx) error`

Rolls back the transaction associated with the context.

#### `SavePoint(ctx, name) error`

Creates a savepoint within the current transaction.

## Transaction Options

All methods support SQL transaction options:

```go
import "database/sql"

opts := &sql.TxOptions{
    Isolation: sql.LevelSerializable,
    ReadOnly:  false,
}

err := uow.Do(ctx, func(ctx context.Context, repo *Repository) error {
    // Operations...
    return nil
}, opts)
```

### Isolation Levels

- `sql.LevelDefault`
- `sql.LevelReadUncommitted`
- `sql.LevelReadCommitted`
- `sql.LevelRepeatableRead`
- `sql.LevelSerializable`

## Common Patterns

### Nested Transactions with Savepoints

```go
err := uow.Do(ctx, func(ctx context.Context, repo *Repository) error {
    // First operation
    repo.CreateUser(user1)

    // Create savepoint before risky operation
    uow.SavePoint(ctx, "before_import")

    // Risky bulk operation
    if err := repo.BulkImport(data); err != nil {
        // Rollback to savepoint, keep user1
        return fmt.Errorf("ROLLBACK TO before_import: %w", err)
    }

    return nil
})
```

### Read-Only Transactions

```go
opts := &sql.TxOptions{ReadOnly: true}

err := uow.Do(ctx, func(ctx context.Context, repo *Repository) error {
    // Only read operations
    users, _ := repo.GetAllUsers()
    stats, _ := repo.GetStatistics()

    // Any write operation will fail
    return nil
}, opts)
```

### Distributed Transactions

```go
// Transaction across multiple database operations
err := uow.Do(ctx, func(ctx context.Context, repos *RepositoryFactory) error {
    // Update inventory
    if err := repos.Inventory.DecreaseStock(productID, quantity); err != nil {
        return err
    }

    // Create order
    if err := repos.Order.Create(order); err != nil {
        return err // Inventory change will rollback
    }

    // Send notification (non-transactional)
    go sendOrderConfirmation(order.ID)

    return nil
})
```

## Best Practices

1. **Use Do() for most cases**

   ```go
   // Preferred
   uow.Do(ctx, func(ctx context.Context, repo *Repository) error {
       // Operations
       return nil
   })
   ```

2. **Manual control only when needed**

   ```go
   // Only when you need fine-grained control
   ctx, repo, _ := uow.Begin(ctx)
   defer uow.Rollback(ctx) // Safety rollback
   // ... operations ...
   uow.Commit(ctx)
   ```

3. **Keep transactions short**

   - Don't perform I/O operations in transactions
   - Don't make HTTP calls
   - Don't send emails
   - Keep only database operations

4. **Error handling**

   ```go
   err := uow.Do(ctx, func(ctx context.Context, repo *Repository) error {
       if err := repo.Operation(); err != nil {
           return fmt.Errorf("operation failed: %w", err)
       }
       return nil
   })
   ```

5. **Context propagation**
   ```go
   // Pass context through
   err := uow.Do(ctx, func(ctx context.Context, repo *Repository) error {
       return repo.MethodWithContext(ctx, data)
   })
   ```

## Performance Considerations

- Transactions hold locks - keep them short
- Use appropriate isolation levels
- Consider read-only transactions for queries
- Use savepoints judiciously (they have overhead)
- Pool connections appropriately

## Testing

### Mock Repositories

```go
type MockUserRepository struct {
    CreateFunc func(*User) error
}

func (m *MockUserRepository) Create(user *User) error {
    return m.CreateFunc(user)
}

func TestService(t *testing.T) {
    // Use in-memory database or mocks
    db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

    factory := uow.NewGorm(db, NewUserRepository)
    service := NewService(factory)

    // Test service methods
}
```

## Error Handling

### Automatic Rollback

```go
err := uow.Do(ctx, func(ctx context.Context, repo *Repository) error {
    repo.Create(item1)

    if err := repo.Create(item2); err != nil {
        return err // Automatic rollback, item1 not saved
    }

    return nil // Automatic commit, both items saved
})
```

### Savepoint Rollback

```go
// Use special error format to rollback to savepoint
return fmt.Errorf("ROLLBACK TO %s: %w", savepointName, err)
```

## Migration from Direct DB Usage

### Before (Direct GORM)

```go
tx := db.Begin()
if err := tx.Create(&user).Error; err != nil {
    tx.Rollback()
    return err
}
if err := tx.Create(&account).Error; err != nil {
    tx.Rollback()
    return err
}
tx.Commit()
```

### After (UoW Pattern)

```go
uow.Do(ctx, func(ctx context.Context, repos *RepositoryFactory) error {
    if err := repos.User.Create(&user); err != nil {
        return err
    }
    return repos.Account.Create(&account)
})
```

## License

MIT License - see [LICENSE](../LICENSE) for details
