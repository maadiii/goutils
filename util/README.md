# Util

Common utility functions for everyday Go programming tasks.

## Overview

The util package provides type-safe utility functions using Go generics for working with pointers, slices, and common operations.

## Installation

```bash
go get github.com/maadiii/goutils/util
```

## Features

- 🎯 Generic pointer utilities
- 🔄 Type-safe operations
- ✅ Null-safe value access
- 📦 Zero-allocation helpers

## Usage

### Pointer Utilities

#### Get Pointer Value Safely

```go
package main

import (
    "fmt"
    "github.com/maadiii/goutils/util"
)

func main() {
    // Get value from pointer, or zero value if nil
    var str *string
    value := util.GetPtrValue(str)
    fmt.Println(value) // Output: "" (empty string, not panic)

    // With actual value
    name := "John"
    namePtr := &name
    value = util.GetPtrValue(namePtr)
    fmt.Println(value) // Output: John
}
```

#### Create Pointer from Value

```go
import "github.com/maadiii/goutils/util"

// Create pointer to value
name := "Alice"
namePtr := util.ToPtr(name)
fmt.Println(*namePtr) // Output: Alice

// Works with any type
age := 25
agePtr := util.ToPtr(age)
fmt.Println(*agePtr) // Output: 25

// Useful for inline assignments
user := User{
    Name: util.ToPtr("Bob"),
    Age:  util.ToPtr(30),
}
```

#### Create Pointer or Nil for Zero Values

```go
import "github.com/maadiii/goutils/util"

// Returns pointer for non-zero values
name := "Charlie"
namePtr := util.ToPtrOrNil(name)
fmt.Println(*namePtr) // Output: Charlie

// Returns nil for zero values
emptyStr := ""
nilPtr := util.ToPtrOrNil(emptyStr)
fmt.Println(nilPtr == nil) // Output: true

// Works with numbers
age := 0
agePtr := util.ToPtrOrNil(age)
fmt.Println(agePtr == nil) // Output: true

nonZeroAge := 25
nonZeroPtr := util.ToPtrOrNil(nonZeroAge)
fmt.Println(*nonZeroPtr) // Output: 25
```

## Complete Examples

### Working with Optional Fields

```go
package main

import (
    "encoding/json"
    "fmt"
    "github.com/maadiii/goutils/util"
)

type User struct {
    ID       int     `json:"id"`
    Name     string  `json:"name"`
    Email    *string `json:"email,omitempty"`
    Age      *int    `json:"age,omitempty"`
    Bio      *string `json:"bio,omitempty"`
}

func CreateUser(id int, name string, email string, age int, bio string) User {
    return User{
        ID:    id,
        Name:  name,
        Email: util.ToPtrOrNil(email), // nil if empty
        Age:   util.ToPtrOrNil(age),   // nil if 0
        Bio:   util.ToPtrOrNil(bio),   // nil if empty
    }
}

func GetUserEmail(user User) string {
    // Safely get email or default value
    return util.GetPtrValue(user.Email)
}

func main() {
    // User with all fields
    user1 := CreateUser(1, "Alice", "alice@example.com", 25, "Developer")

    // User with only required fields
    user2 := CreateUser(2, "Bob", "", 0, "")

    // JSON output
    data1, _ := json.MarshalIndent(user1, "", "  ")
    fmt.Println(string(data1))
    // Output:
    // {
    //   "id": 1,
    //   "name": "Alice",
    //   "email": "alice@example.com",
    //   "age": 25,
    //   "bio": "Developer"
    // }

    data2, _ := json.MarshalIndent(user2, "", "  ")
    fmt.Println(string(data2))
    // Output:
    // {
    //   "id": 2,
    //   "name": "Bob"
    // }

    // Safe access
    fmt.Println(GetUserEmail(user1)) // alice@example.com
    fmt.Println(GetUserEmail(user2)) // (empty string)
}
```

### Database Query Results

```go
import (
    "database/sql"
    "github.com/maadiii/goutils/util"
)

type Product struct {
    ID          int
    Name        string
    Description *string
    Price       float64
    Discount    *float64
}

func ScanProduct(rows *sql.Rows) (*Product, error) {
    var p Product
    var description, discount sql.NullString

    err := rows.Scan(
        &p.ID,
        &p.Name,
        &description,
        &p.Price,
        &discount,
    )

    if err != nil {
        return nil, err
    }

    // Convert sql.NullString to *string
    if description.Valid {
        p.Description = util.ToPtr(description.String)
    }

    if discount.Valid {
        discountVal, _ := strconv.ParseFloat(discount.String, 64)
        p.Discount = util.ToPtr(discountVal)
    }

    return &p, nil
}

func GetProductDiscount(product *Product) float64 {
    // Get discount or default to 0
    return util.GetPtrValue(product.Discount)
}
```

### API Request/Response

```go
type UpdateUserRequest struct {
    Name  *string `json:"name,omitempty"`
    Email *string `json:"email,omitempty"`
    Bio   *string `json:"bio,omitempty"`
}

func UpdateUser(userID int, req UpdateUserRequest) error {
    updates := make(map[string]interface{})

    // Only update provided fields
    if req.Name != nil {
        updates["name"] = util.GetPtrValue(req.Name)
    }

    if req.Email != nil {
        updates["email"] = util.GetPtrValue(req.Email)
    }

    if req.Bio != nil {
        updates["bio"] = util.GetPtrValue(req.Bio)
    }

    return db.Model(&User{}).Where("id = ?", userID).Updates(updates).Error
}

// Client code
func main() {
    // Update only name
    req := UpdateUserRequest{
        Name: util.ToPtr("New Name"),
    }

    UpdateUser(123, req)
}
```

### Configuration with Defaults

```go
type Config struct {
    Host     string
    Port     *int
    Timeout  *time.Duration
    MaxConns *int
}

func LoadConfig() Config {
    cfg := Config{
        Host: getEnv("HOST", "localhost"),
    }

    // Optional configurations
    if port := getEnv("PORT", ""); port != "" {
        portInt, _ := strconv.Atoi(port)
        cfg.Port = util.ToPtr(portInt)
    }

    if timeout := getEnv("TIMEOUT", ""); timeout != "" {
        duration, _ := time.ParseDuration(timeout)
        cfg.Timeout = util.ToPtr(duration)
    }

    return cfg
}

func GetPort(cfg Config) int {
    // Get port or default to 8080
    port := util.GetPtrValue(cfg.Port)
    if port == 0 {
        return 8080
    }
    return port
}

func GetTimeout(cfg Config) time.Duration {
    // Get timeout or default to 30 seconds
    timeout := util.GetPtrValue(cfg.Timeout)
    if timeout == 0 {
        return 30 * time.Second
    }
    return timeout
}
```

## API Reference

### Types

#### `Pointer[T any]`

Interface constraint for pointer types.

```go
type Pointer[T any] interface {
    *T
}
```

### Functions

#### `GetPtrValue[T any, P Pointer[T]](p P) T`

Gets the value from a pointer, or returns zero value if pointer is nil.

**Parameters:**

- `p`: Pointer to dereference

**Returns:**

- `T`: Value pointed to, or zero value if nil

**Example:**

```go
var str *string
value := util.GetPtrValue(str) // Returns ""

name := "John"
value = util.GetPtrValue(&name) // Returns "John"
```

#### `ToPtr[T any](v T) *T`

Creates a pointer to the given value.

**Parameters:**

- `v`: Value to create pointer for

**Returns:**

- `*T`: Pointer to the value

**Example:**

```go
namePtr := util.ToPtr("Alice")
agePtr := util.ToPtr(25)
```

#### `ToPtrOrNil[T comparable](v T) *T`

Creates a pointer to the value, or returns nil if the value is zero.

**Parameters:**

- `v`: Value to create pointer for

**Returns:**

- `*T`: Pointer to the value, or nil if zero value

**Example:**

```go
ptr1 := util.ToPtrOrNil("") // Returns nil
ptr2 := util.ToPtrOrNil("hello") // Returns &"hello"
ptr3 := util.ToPtrOrNil(0) // Returns nil
ptr4 := util.ToPtrOrNil(42) // Returns &42
```

## Common Use Cases

### JSON API Optional Fields

```go
// Request with optional fields
type CreatePostRequest struct {
    Title       string  `json:"title" validate:"required"`
    Content     string  `json:"content" validate:"required"`
    PublishedAt *string `json:"published_at,omitempty"`
    Tags        *[]string `json:"tags,omitempty"`
}

// Only include non-nil fields in JSON
post := CreatePostRequest{
    Title:   "My Post",
    Content: "Content here",
    Tags:    util.ToPtrOrNil([]string{}), // nil for empty slice
}
```

### Database Nullable Columns

```go
type Article struct {
    ID          int
    Title       string
    PublishedAt *time.Time
    UpdatedAt   *time.Time
}

// Check if article is published
func IsPublished(article Article) bool {
    publishedAt := util.GetPtrValue(article.PublishedAt)
    return !publishedAt.IsZero()
}
```

### Form Data Processing

```go
type ProfileUpdateForm struct {
    Name     *string
    Bio      *string
    Website  *string
    Location *string
}

func ProcessForm(form ProfileUpdateForm) map[string]string {
    updates := make(map[string]string)

    if form.Name != nil {
        updates["name"] = *form.Name
    }
    if form.Bio != nil {
        updates["bio"] = *form.Bio
    }
    // Only update provided fields

    return updates
}
```

### Configuration Merging

```go
type Settings struct {
    Theme      *string
    Language   *string
    PageSize   *int
    AutoSave   *bool
}

func MergeSettings(base, override Settings) Settings {
    merged := base

    // Override only non-nil fields
    if override.Theme != nil {
        merged.Theme = override.Theme
    }
    if override.Language != nil {
        merged.Language = override.Language
    }
    if override.PageSize != nil {
        merged.PageSize = override.PageSize
    }
    if override.AutoSave != nil {
        merged.AutoSave = override.AutoSave
    }

    return merged
}
```

## Best Practices

1. **Use `ToPtrOrNil` for optional fields**

   ```go
   // Good - omits zero values from JSON
   user := User{
       Name:  "Alice",
       Email: util.ToPtrOrNil(emailInput),
   }
   ```

2. **Use `GetPtrValue` for safe access**

   ```go
   // Good - no panic if nil
   email := util.GetPtrValue(user.Email)

   // Avoid - can panic
   email := *user.Email
   ```

3. **Inline pointer creation**

   ```go
   // Good - concise
   user := User{
       Name: util.ToPtr("Bob"),
       Age:  util.ToPtr(30),
   }

   // Avoid - verbose
   name := "Bob"
   age := 30
   user := User{
       Name: &name,
       Age:  &age,
   }
   ```

4. **Check for meaningful zero values**

   ```go
   // Be careful with ToPtrOrNil for numbers
   // 0 might be a valid value!
   quantity := 0
   ptr := util.ToPtrOrNil(quantity) // nil - might not be desired

   // Better: use ToPtr if 0 is valid
   ptr := util.ToPtr(quantity) // &0
   ```

## Type Safety

All functions use Go generics for type safety:

```go
// Compile-time type checking
namePtr := util.ToPtr("Alice")  // *string
agePtr := util.ToPtr(25)        // *int

// No casting needed
name := util.GetPtrValue(namePtr) // string
age := util.GetPtrValue(agePtr)   // int
```

## Testing

```go
func TestGetPtrValue(t *testing.T) {
    // Test with nil
    var nilPtr *string
    if util.GetPtrValue(nilPtr) != "" {
        t.Error("Expected empty string for nil pointer")
    }

    // Test with value
    value := "test"
    if util.GetPtrValue(&value) != "test" {
        t.Error("Expected 'test'")
    }
}

func TestToPtrOrNil(t *testing.T) {
    // Test zero value
    if util.ToPtrOrNil("") != nil {
        t.Error("Expected nil for empty string")
    }

    // Test non-zero value
    ptr := util.ToPtrOrNil("test")
    if ptr == nil || *ptr != "test" {
        t.Error("Expected pointer to 'test'")
    }
}
```

## License

MIT License - see [LICENSE](../LICENSE) for details
