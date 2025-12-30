# Utils

[![Go Version](https://img.shields.io/github/go-mod/go-version/maadiii/utils)](https://golang.org/doc/devel/release.html)
[![Go Report Card](https://goreportcard.com/badge/github.com/maadiii/utils)](https://goreportcard.com/report/github.com/maadiii/utils)
[![GoDoc](https://godoc.org/github.com/maadiii/utils?status.svg)](https://godoc.org/github.com/maadiii/utils)
[![License](https://img.shields.io/github/license/maadiii/utils)](LICENSE)
A comprehensive collection of utility packages for Go applications, providing robust solutions for common development needs including authentication, caching, encryption, error handling, logging, and more.

## Features

- 🔐 **Authentication**: JWT and PASETO token support (local and public)
- 🔄 **Async Operations**: Asynchronous task execution utilities
- 💾 **Caching**: Redis-based caching with easy integration
- 🔒 **Encryption**: Password hashing and random string generation
- 🚨 **Error Handling**: Enhanced error handling for Echo and Hertz frameworks
- 📝 **Logging**: Zap logger integration with middleware support
- 🔑 **TOTP**: Time-based One-Time Password implementation
- 🗄️ **Unit of Work**: Database transaction patterns for GORM, PGX, and SQL
- 🛠️ **Utilities**: Common utility functions
- ✅ **Validation**: Request validation middleware for Echo
- ⚡ **Worker Pool**: Concurrent job processing with worker pools

## Installation

```bash
go get github.com/maadiii/utils
```

## Requirements

- Go 1.24.3 or higher

## Packages

### Authentication (`/auth`)

Comprehensive authentication package supporting both JWT and PASETO tokens with local (symmetric) and public (asymmetric) key operations.

#### JWT Local (Symmetric)

```go
import "github.com/maadiii/utils/auth"

// Create a JWT local authenticator
jwtLocal, err := auth.NewJWTLocal("your-secret-key")
if err != nil {
    log.Fatal(err)
}

// Generate token
token, err := jwtLocal.Generate(map[string]interface{}{
    "user_id": 123,
    "email": "user@example.com",
}, time.Hour*24)

// Verify token
claims, err := jwtLocal.Verify(token)
```

#### PASETO Local (Symmetric)

```go
// Create a PASETO local authenticator
pasetoLocal := auth.NewPasetoLocal("your-32-byte-secret-key-here")

// Generate token
token, err := pasetoLocal.Generate(map[string]interface{}{
    "user_id": 123,
}, time.Hour*24)

// Verify token
claims, err := pasetoLocal.Verify(token)
```

### Async (`/async`)

Execute asynchronous operations with error handling.

```go
import "github.com/maadiii/utils/async"

// Execute async operation
result, err := async.Do(context.Background(), func(ctx context.Context) (interface{}, error) {
    // Your async operation
    return fetchData(), nil
})
```

### Cache (`/cache`)

Redis-based caching solution with a clean interface.

```go
import "github.com/maadiii/utils/cache"

// Initialize cache
cache, err := cache.New(cache.Config{
    Addr: "localhost:6379",
    Password: "",
    DB: 0,
})

// Set value
err = cache.Set(ctx, "key", "value", time.Hour)

// Get value
value, err := cache.Get(ctx, "key")
```

### Encryption (`/encryption`)

Password hashing and random string generation utilities.

```go
import "github.com/maadiii/utils/encryption"

// Hash password
hashedPassword, err := encryption.HashPassword("my-password")

// Verify password
isValid := encryption.VerifyPassword("my-password", hashedPassword)

// Generate random string
randomStr, err := encryption.RandomString(32)
```

### Errors (`/errors`)

Enhanced error handling with framework-specific integrations.

```go
import "github.com/maadiii/utils/errors"

// Create custom error
err := errors.New("something went wrong").
    WithCode(errors.CodeNotFound).
    WithMetadata("key", "value")

// Use with Echo
app := echo.New()
app.HTTPErrorHandler = errors.EchoErrorHandler

// Use with Hertz
h := server.Default()
h.Use(errors.HertzErrorMiddleware())
```

### Log (`/log`)

Zap logger wrapper with middleware support for Echo and Hertz.

```go
import "github.com/maadiii/utils/log"

// Initialize logger
logger, err := log.NewZapLogger(log.Config{
    Level: "info",
    Encoding: "json",
    OutputPaths: []string{"stdout"},
})

// Use logger
logger.Info("application started", zap.String("version", "1.0.0"))

// Echo middleware
app.Use(log.ZapEchoMiddleware(logger))

// Hertz middleware
h.Use(log.HertzMiddleware(logger))
```

### TOTP (`/totp`)

Time-based One-Time Password implementation.

```go
import "github.com/maadiii/utils/totp"

// Generate TOTP key
key, err := totp.Generate(totp.GenerateOpts{
    Issuer: "MyApp",
    AccountName: "user@example.com",
})

// Generate OTP
code, err := totp.GenerateCode(key.Secret(), time.Now())

// Validate OTP
isValid := totp.Validate(code, key.Secret())
```

### Unit of Work (`/uow`)

Database transaction patterns for multiple ORMs and database drivers.

#### GORM

```go
import "github.com/maadiii/utils/uow"

uow := uow.NewGormUnitOfWork(db)
err := uow.Begin(ctx, func(ctx context.Context) error {
    // Your database operations
    return nil
})
```

#### PGX

```go
uow := uow.NewPgxUnitOfWork(pool)
err := uow.Begin(ctx, func(ctx context.Context) error {
    // Your database operations
    return nil
})
```

### Utilities (`/util`)

Common utility functions for everyday tasks.

```go
import "github.com/maadiii/utils/util"

// String pointer
strPtr := util.StringPtr("hello")

// Int pointer
intPtr := util.IntPtr(42)

// Contains check
exists := util.Contains([]string{"a", "b", "c"}, "b")
```

### Validator (`/validator`)

Request validation middleware for Echo framework.

```go
import "github.com/maadiii/utils/validator"

app := echo.New()
app.Validator = validator.NewValidator()

type RequestBody struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
}
```

### Worker Pool (`/workerpool`)

Concurrent job processing with configurable worker pools.

```go
import "github.com/maadiii/utils/workerpool"

// Create worker pool
pool := workerpool.New(workerpool.Config{
    Workers: 10,
    QueueSize: 100,
})

// Start pool
pool.Start()

// Submit job
pool.Submit(workerpool.Job{
    Execute: func(ctx context.Context) error {
        // Your job logic
        return nil
    },
})

// Shutdown
pool.Shutdown()
```

## Testing

Run all tests:

```bash
go test ./...
```

Run tests with coverage:

```bash
go test -race -coverprofile=coverage.txt -covermode=atomic ./...
```

Run tests for a specific package:

```bash
go test ./auth/...
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Hertz](https://github.com/cloudwego/hertz) - HTTP framework
- [Echo](https://github.com/labstack/echo) - HTTP framework
- [Zap](https://github.com/uber-go/zap) - Logging library
- [GORM](https://gorm.io/) - ORM library
- [PGX](https://github.com/jackc/pgx) - PostgreSQL driver
- [go-redis](https://github.com/redis/go-redis) - Redis client
- [PASETO](https://github.com/aidanwoods/go-paseto) - Token implementation

## Support

If you have any questions or need help, please open an issue on GitHub.

---

Made with ❤️ by [maadiii](https://github.com/maadiii)
