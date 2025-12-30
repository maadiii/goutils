# Errors

Enhanced error handling with HTTP status codes and framework-specific integrations for Echo and Hertz.

## Overview

The errors package provides structured error handling with support for error codes, metadata, stack traces, and seamless integration with popular Go web frameworks.

## Installation

```bash
go get github.com/maadiii/utils/errors
```

## Features

- 📋 HTTP status code mapping
- 🔍 Stack trace capture
- 🏷️ Error metadata support
- 🌐 Echo framework integration
- ⚡ Hertz framework integration
- 🔄 Error wrapping and chaining

## Usage

### Basic Error Creation

```go
package main

import (
    "fmt"
    "github.com/maadiii/utils/errors"
)

func main() {
    // Create a simple error
    err := errors.New("something went wrong")
    fmt.Println(err)
}
```

### Error with Type/Status Code

```go
// Create typed errors
err := errors.BadRequest().Msg("invalid request parameters")
err := errors.NotFound().Msg("user not found")
err := errors.Unauthorized().Msg("invalid credentials")
err := errors.InternalServerError().Msg("database connection failed")
```

### Error Wrapping

```go
// Wrap existing errors
dbErr := db.Query("SELECT * FROM users")
if dbErr != nil {
    err := errors.InternalServerError().Wrap(dbErr)
    return err
}
```

### Complete Example with HTTP Handler

```go
package main

import (
    "github.com/labstack/echo/v4"
    "github.com/maadiii/utils/errors"
)

type UserService struct {
    db Database
}

func (s *UserService) GetUser(userID string) (*User, error) {
    user, err := s.db.FindUser(userID)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, errors.NotFound().Msg("user %s not found", userID)
        }
        return nil, errors.InternalServerError().Wrap(err)
    }

    return user, nil
}

func (s *UserService) CreateUser(userData UserData) (*User, error) {
    if err := userData.Validate(); err != nil {
        return nil, errors.BadRequest().Wrap(err)
    }

    user, err := s.db.CreateUser(userData)
    if err != nil {
        return nil, errors.InternalServerError().Wrap(err)
    }

    return user, nil
}

func main() {
    e := echo.New()

    // Use custom error handler
    e.HTTPErrorHandler = errors.EchoErrorHandler

    service := &UserService{}

    e.GET("/users/:id", func(c echo.Context) error {
        user, err := service.GetUser(c.Param("id"))
        if err != nil {
            return err // Will be handled by EchoErrorHandler
        }
        return c.JSON(200, user)
    })

    e.Start(":8080")
}
```

## Echo Framework Integration

### Setup

```go
import (
    "github.com/labstack/echo/v4"
    "github.com/maadiii/utils/errors"
)

func main() {
    e := echo.New()

    // Set custom error handler
    e.HTTPErrorHandler = errors.EchoErrorHandler

    // Your routes
    e.Start(":8080")
}
```

### Handler Example

```go
func GetUser(c echo.Context) error {
    id := c.Param("id")

    user, err := userService.FindByID(id)
    if err != nil {
        // Return structured error
        return errors.NotFound().Msg("user with ID %s not found", id)
    }

    return c.JSON(http.StatusOK, user)
}
```

## Hertz Framework Integration

### Setup

```go
import (
    "github.com/cloudwego/hertz/pkg/app/server"
    "github.com/maadiii/utils/errors"
)

func main() {
    h := server.Default()

    // Add error handling middleware
    h.Use(errors.HertzErrorMiddleware())

    // Your routes
    h.Spin()
}
```

### Handler Example

```go
func GetUser(ctx context.Context, c *app.RequestContext) {
    id := c.Param("id")

    user, err := userService.FindByID(id)
    if err != nil {
        // Set error in context
        c.Error(errors.NotFound().Msg("user with ID %s not found", id))
        return
    }

    c.JSON(http.StatusOK, user)
}
```

## API Reference

### Error Types

The package provides common HTTP error types:

```go
// 4xx Client Errors
errors.BadRequest()          // 400
errors.Unauthorized()        // 401
errors.PaymentRequired()     // 402
errors.Forbidden()           // 403
errors.NotFound()            // 404
errors.MethodNotAllowed()    // 405
errors.Conflict()            // 409
errors.Gone()                // 410
errors.UnprocessableEntity() // 422
errors.TooManyRequests()     // 429

// 5xx Server Errors
errors.InternalServerError() // 500
errors.NotImplemented()      // 501
errors.BadGateway()          // 502
errors.ServiceUnavailable()  // 503
errors.GatewayTimeout()      // 504
```

### Error Methods

#### `New(format string, a ...any) error`

Creates a new error with formatted message and stack trace.

```go
err := errors.New("user %s not found", userID)
```

#### `Msg(format string, a ...any) error`

Sets the error message with formatting.

```go
err := errors.NotFound().Msg("resource %s not found", resourceID)
```

#### `Wrap(err error) error`

Wraps an existing error while preserving the type.

```go
err := errors.InternalServerError().Wrap(dbError)
```

### Error Structure

```go
type Error struct {
    Type int   // HTTP status code
    Err  error // Underlying error
}
```

## Common Patterns

### Validation Errors

```go
func ValidateUser(user *User) error {
    if user.Email == "" {
        return errors.BadRequest().Msg("email is required")
    }

    if !isValidEmail(user.Email) {
        return errors.BadRequest().Msg("invalid email format")
    }

    if len(user.Password) < 8 {
        return errors.BadRequest().Msg("password must be at least 8 characters")
    }

    return nil
}
```

### Database Errors

```go
func GetUserByID(id string) (*User, error) {
    user, err := db.QueryUser(id)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, errors.NotFound().Msg("user not found")
        }
        return nil, errors.InternalServerError().Wrap(err)
    }
    return user, nil
}
```

### Authentication Errors

```go
func AuthenticateUser(username, password string) (*User, error) {
    user, err := db.FindUserByUsername(username)
    if err != nil {
        return nil, errors.Unauthorized().Msg("invalid credentials")
    }

    if !verifyPassword(password, user.PasswordHash) {
        return nil, errors.Unauthorized().Msg("invalid credentials")
    }

    return user, nil
}
```

### Authorization Errors

```go
func RequireAdmin(userID string) error {
    user, err := GetUser(userID)
    if err != nil {
        return err
    }

    if !user.IsAdmin {
        return errors.Forbidden().Msg("admin access required")
    }

    return nil
}
```

### Rate Limiting

```go
func CheckRateLimit(userID string) error {
    allowed, err := rateLimiter.Allow(userID)
    if err != nil {
        return errors.InternalServerError().Wrap(err)
    }

    if !allowed {
        return errors.TooManyRequests().Msg("rate limit exceeded")
    }

    return nil
}
```

## Stack Traces

Errors automatically capture stack traces for debugging:

```go
err := errors.New("something went wrong")
// err.Error() includes stack trace information
```

The stack trace helps identify exactly where the error originated.

## Best Practices

1. **Use Specific Error Types**

   ```go
   // Good
   return errors.NotFound().Msg("user not found")

   // Avoid
   return errors.InternalServerError().Msg("user not found")
   ```

2. **Don't Expose Internal Errors**

   ```go
   // Good
   return errors.InternalServerError().Msg("failed to process request")

   // Avoid - exposes internal details
   return errors.InternalServerError().Msg("database connection failed: %v", dbErr)
   ```

3. **Wrap Lower-Level Errors**

   ```go
   if err := db.Save(user); err != nil {
       return errors.InternalServerError().Wrap(err)
   }
   ```

4. **Use Consistent Messages**

   ```go
   // Define common messages
   const (
       ErrUserNotFound = "user not found"
       ErrInvalidInput = "invalid input"
   )

   return errors.NotFound().Msg(ErrUserNotFound)
   ```

5. **Log Errors Appropriately**
   ```go
   if err != nil {
       logger.Error("failed to get user", zap.Error(err))
       return errors.InternalServerError().Msg("failed to retrieve user")
   }
   ```

## Testing

```go
func TestErrorTypes(t *testing.T) {
    err := errors.NotFound().Msg("test error")

    if e, ok := err.(*errors.Error); ok {
        if e.Type != http.StatusNotFound {
            t.Errorf("expected 404, got %d", e.Type)
        }
    }
}
```

## License

MIT License - see [LICENSE](../LICENSE) for details
