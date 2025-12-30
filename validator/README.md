# Validator

Request validation middleware for Echo framework using go-playground/validator.

## Overview

The validator package provides a simple wrapper around the popular go-playground/validator library with seamless integration into the Echo web framework.

## Installation

```bash
go get github.com/maadiii/utils/validator
```

## Features

- ✅ Struct validation
- 🌐 Echo framework integration
- 🎯 Built-in validation tags
- 🔄 Custom validation rules support
- 🚨 Automatic error handling

## Usage

### Basic Setup with Echo

```go
package main

import (
    "github.com/labstack/echo/v4"
    "github.com/maadiii/utils/validator"
)

func main() {
    e := echo.New()

    // Set custom validator
    e.Validator = validator.NewValidator()

    e.POST("/users", createUser)

    e.Start(":8080")
}
```

### Define Validation Rules

```go
type CreateUserRequest struct {
    Name     string `json:"name" validate:"required,min=3,max=50"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8,max=100"`
    Age      int    `json:"age" validate:"required,min=18,max=120"`
    Website  string `json:"website" validate:"omitempty,url"`
}

func createUser(c echo.Context) error {
    var req CreateUserRequest

    // Bind request body
    if err := c.Bind(&req); err != nil {
        return c.JSON(400, map[string]string{"error": "invalid request"})
    }

    // Validate - automatically uses the custom validator
    if err := c.Validate(&req); err != nil {
        return c.JSON(400, map[string]string{"error": err.Error()})
    }

    // Create user...
    return c.JSON(201, map[string]string{"message": "user created"})
}
```

### Complete Example

```go
package main

import (
    "net/http"

    "github.com/labstack/echo/v4"
    "github.com/maadiii/utils/validator"
)

type RegisterRequest struct {
    Username string `json:"username" validate:"required,alphanum,min=3,max=20"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8,containsany=!@#$%^&*"`
    Age      int    `json:"age" validate:"required,gte=18,lte=120"`
    Country  string `json:"country" validate:"required,iso3166_1_alpha2"`
    Website  string `json:"website" validate:"omitempty,url"`
    Terms    bool   `json:"terms" validate:"required,eq=true"`
}

type LoginRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required"`
}

type UpdateProfileRequest struct {
    Name    *string `json:"name" validate:"omitempty,min=2,max=50"`
    Bio     *string `json:"bio" validate:"omitempty,max=500"`
    Website *string `json:"website" validate:"omitempty,url"`
}

func main() {
    e := echo.New()

    // Set custom validator
    e.Validator = validator.NewValidator()

    // Routes
    e.POST("/register", handleRegister)
    e.POST("/login", handleLogin)
    e.PUT("/profile", handleUpdateProfile)

    e.Start(":8080")
}

func handleRegister(c echo.Context) error {
    var req RegisterRequest

    if err := c.Bind(&req); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
    }

    if err := c.Validate(&req); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    }

    // Register user logic...

    return c.JSON(http.StatusCreated, map[string]string{
        "message": "registration successful",
    })
}

func handleLogin(c echo.Context) error {
    var req LoginRequest

    if err := c.Bind(&req); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
    }

    if err := c.Validate(&req); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    }

    // Login logic...

    return c.JSON(http.StatusOK, map[string]string{
        "token": "jwt-token-here",
    })
}

func handleUpdateProfile(c echo.Context) error {
    var req UpdateProfileRequest

    if err := c.Bind(&req); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
    }

    if err := c.Validate(&req); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    }

    // Update profile logic...

    return c.JSON(http.StatusOK, map[string]string{
        "message": "profile updated",
    })
}
```

## Validation Tags

### Common Tags

#### Required

```go
Name string `validate:"required"`
```

#### String Length

```go
Username string `validate:"min=3,max=20"`
Password string `validate:"len=10"` // Exactly 10 characters
```

#### Numeric Ranges

```go
Age    int     `validate:"gte=18,lte=120"` // Greater than or equal, less than or equal
Price  float64 `validate:"gt=0,lt=1000"`   // Greater than, less than
Rating int     `validate:"min=1,max=5"`
```

#### Email

```go
Email string `validate:"email"`
```

#### URL

```go
Website string `validate:"url"`
Homepage string `validate:"http_url"` // Must include http:// or https://
```

#### Alphanumeric

```go
Username string `validate:"alphanum"`        // Only letters and numbers
Code     string `validate:"alpha"`           // Only letters
Number   string `validate:"numeric"`         // Only numbers
```

#### Contains

```go
Password string `validate:"containsany=!@#$%^&*"` // Must contain at least one special char
Text     string `validate:"contains=hello"`       // Must contain "hello"
```

#### Equality

```go
Terms    bool   `validate:"eq=true"`       // Must be true
Status   string `validate:"oneof=active inactive"` // Must be one of the values
```

#### Format Validation

```go
Phone   string `validate:"e164"`                    // E.164 phone format
UUID    string `validate:"uuid"`                    // UUID format
IPv4    string `validate:"ipv4"`                    // IPv4 address
MAC     string `validate:"mac"`                     // MAC address
Base64  string `validate:"base64"`                  // Base64 string
Country string `validate:"iso3166_1_alpha2"`        // Country code (US, GB, etc.)
```

#### Dates

```go
BirthDate string `validate:"datetime=2006-01-02"`
```

#### Optional Fields

```go
Bio     string `validate:"omitempty,max=500"` // Only validate if not empty
Website string `validate:"omitempty,url"`     // Optional but must be valid URL if provided
```

## Nested Struct Validation

```go
type Address struct {
    Street  string `validate:"required"`
    City    string `validate:"required"`
    Country string `validate:"required,iso3166_1_alpha2"`
    ZipCode string `validate:"required,numeric,len=5"`
}

type User struct {
    Name    string   `validate:"required"`
    Email   string   `validate:"required,email"`
    Address Address  `validate:"required,dive"` // Validate nested struct
}
```

## Slice Validation

```go
type CreateOrderRequest struct {
    Items []OrderItem `validate:"required,min=1,max=100,dive"`
}

type OrderItem struct {
    ProductID string `validate:"required,uuid"`
    Quantity  int    `validate:"required,min=1,max=999"`
}
```

## API Reference

### Types

#### `Validator`

```go
type Validator struct {
    validator *validator.Validate
}
```

### Functions

#### `NewValidator() *Validator`

Creates a new validator instance.

```go
v := validator.NewValidator()
```

### Methods

#### `Validate(i any) error`

Validates a struct and returns an error if validation fails.

**Parameters:**

- `i`: Struct to validate

**Returns:**

- `error`: Validation error wrapped as BadRequest, or nil if valid

**Example:**

```go
err := v.Validate(&req)
if err != nil {
    // Handle validation error
}
```

## Error Handling

The validator automatically wraps validation errors with `errors.BadRequest()`:

```go
if err := c.Validate(&req); err != nil {
    // err is already a BadRequest error
    return err // Returns 400 status code
}
```

### Custom Error Handling

```go
if err := c.Validate(&req); err != nil {
    // Parse validation errors
    if ve, ok := err.(validator.ValidationErrors); ok {
        for _, fe := range ve {
            fmt.Printf("Field: %s, Tag: %s\n", fe.Field(), fe.Tag())
        }
    }

    return c.JSON(400, map[string]string{
        "error": "validation failed",
        "details": err.Error(),
    })
}
```

## Common Validation Patterns

### User Registration

```go
type RegisterRequest struct {
    Email           string `validate:"required,email"`
    Password        string `validate:"required,min=8,max=100"`
    PasswordConfirm string `validate:"required,eqfield=Password"`
    Username        string `validate:"required,alphanum,min=3,max=20"`
    Terms           bool   `validate:"required,eq=true"`
}
```

### Credit Card (for display only - never process real CC data without PCI compliance)

```go
type PaymentRequest struct {
    CardNumber string `validate:"required,creditcard"`
    CVV        string `validate:"required,numeric,len=3"`
    ExpiryDate string `validate:"required,datetime=01/06"`
}
```

### Contact Form

```go
type ContactRequest struct {
    Name    string `validate:"required,min=2,max=100"`
    Email   string `validate:"required,email"`
    Subject string `validate:"required,min=5,max=200"`
    Message string `validate:"required,min=10,max=2000"`
}
```

### Search Query

```go
type SearchRequest struct {
    Query  string `validate:"required,min=2,max=100"`
    Page   int    `validate:"required,min=1,max=1000"`
    Limit  int    `validate:"required,oneof=10 25 50 100"`
    SortBy string `validate:"omitempty,oneof=date relevance title"`
}
```

### File Upload

```go
type UploadRequest struct {
    Filename    string `validate:"required,max=255"`
    ContentType string `validate:"required,oneof=image/jpeg image/png image/gif"`
    Size        int64  `validate:"required,max=5242880"` // 5MB
}
```

## Advanced Usage

### Cross-Field Validation

```go
type ChangePasswordRequest struct {
    CurrentPassword string `validate:"required"`
    NewPassword     string `validate:"required,min=8,nefield=CurrentPassword"`
    ConfirmPassword string `validate:"required,eqfield=NewPassword"`
}
```

### Conditional Validation

```go
type ShippingRequest struct {
    ShippingMethod string  `validate:"required,oneof=pickup delivery"`
    Address        *string `validate:"required_if=ShippingMethod delivery"`
}
```

### Multiple Validations

```go
type Product struct {
    SKU   string  `validate:"required,alphanum,len=8"`
    Name  string  `validate:"required,min=3,max=100"`
    Price float64 `validate:"required,gt=0,lt=10000"`
    Tags  []string `validate:"omitempty,min=1,max=10,dive,min=2,max=20"`
}
```

## Testing

```go
func TestValidation(t *testing.T) {
    v := validator.NewValidator()

    // Valid request
    validReq := CreateUserRequest{
        Name:     "John Doe",
        Email:    "john@example.com",
        Password: "SecurePass123!",
        Age:      25,
    }

    if err := v.Validate(&validReq); err != nil {
        t.Error("Valid request should pass validation")
    }

    // Invalid request
    invalidReq := CreateUserRequest{
        Name:     "Jo", // Too short
        Email:    "invalid-email",
        Password: "123", // Too short
        Age:      15, // Under 18
    }

    if err := v.Validate(&invalidReq); err == nil {
        t.Error("Invalid request should fail validation")
    }
}
```

## Best Practices

1. **Use descriptive field names**

   ```go
   // Good
   Email string `validate:"required,email"`

   // Avoid
   E string `validate:"required,email"`
   ```

2. **Combine related validations**

   ```go
   Password string `validate:"required,min=8,max=100,containsany=!@#$%^&*"`
   ```

3. **Use omitempty for optional fields**

   ```go
   Website string `validate:"omitempty,url"` // Optional but must be valid if provided
   ```

4. **Validate early**

   ```go
   // Validate immediately after binding
   if err := c.Bind(&req); err != nil {
       return err
   }
   if err := c.Validate(&req); err != nil {
       return err
   }
   ```

5. **Document validation rules**
   ```go
   // CreateUserRequest represents user registration data
   type CreateUserRequest struct {
       // Username must be 3-20 alphanumeric characters
       Username string `json:"username" validate:"required,alphanum,min=3,max=20"`
       // Email must be valid email format
       Email string `json:"email" validate:"required,email"`
   }
   ```

## Available Validation Tags

For a complete list of validation tags, see the [go-playground/validator documentation](https://github.com/go-playground/validator).

## Performance

- Validation is fast (microseconds for simple structs)
- Validation rules are compiled once
- Suitable for high-throughput APIs

## License

MIT License - see [LICENSE](../LICENSE) for details
