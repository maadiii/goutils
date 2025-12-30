# Encryption

Password hashing and random string generation utilities using industry-standard cryptographic functions.

## Overview

The encryption package provides secure password hashing with bcrypt and cryptographically secure random string generation.

## Installation

```bash
go get github.com/maadiii/utils/encryption
```

## Features

- 🔒 Bcrypt password hashing
- ✅ Password verification
- 🎲 Cryptographically secure random strings
- 🛡️ Industry-standard security

## Usage

### Password Hashing

```go
package main

import (
    "fmt"

    "github.com/maadiii/utils/encryption"
)

func main() {
    p := encryption.NewPassword()

    // Hash a password
    hashedPassword, err := p.Generate("my-secure-password")
    if err != nil {
        panic(err)
    }

    fmt.Println("Hashed:", hashedPassword)

    // Verify password
    isValid := p.Compare(hashedPassword, "my-secure-password")
    fmt.Println("Password valid:", isValid) // true

    // Wrong password
    isValid = p.Compare(hashedPassword, "wrong-password")
    fmt.Println("Wrong password:", isValid) // false
}
```

### Random String Generation

```go
import "github.com/maadiii/utils/encryption"

// Generate a random string (for tokens, IDs, etc.)
randomStr, err := encryption.RandomString(32)
if err != nil {
    panic(err)
}

fmt.Println("Random string:", randomStr)
// Output: Random string: a8f7d9e2b4c6... (64 hex characters)
```

## Complete Example: User Registration

```go
package main

import (
    "fmt"
    "github.com/maadiii/utils/encryption"
)

type User struct {
    ID           string
    Email        string
    PasswordHash string
}

func RegisterUser(email, password string) (*User, error) {
    p := encryption.NewPassword()

    // Generate unique user ID
    userID, err := encryption.RandomString(16)
    if err != nil {
        return nil, err
    }

    // Hash password
    hashedPassword, err := p.Generate(password)
    if err != nil {
        return nil, err
    }

    user := &User{
        ID:           userID,
        Email:        email,
        PasswordHash: hashedPassword,
    }

    return user, nil
}

func AuthenticateUser(user *User, password string) bool {
    p := encryption.NewPassword()
    return p.Compare(user.PasswordHash, password)
}

func main() {
    // Register
    user, err := RegisterUser("user@example.com", "SecurePass123!")
    if err != nil {
        panic(err)
    }

    fmt.Printf("Registered user: %s\n", user.ID)

    // Login
    if AuthenticateUser(user, "SecurePass123!") {
        fmt.Println("Login successful!")
    } else {
        fmt.Println("Login failed!")
    }
}
```

## API Reference

### Password

#### `NewPassword() *password`

Creates a new password hasher instance.

```go
p := encryption.NewPassword()
```

#### `Generate(plain string) (hash string, err error)`

Generates a bcrypt hash from a plain text password.

**Parameters:**

- `plain`: Plain text password

**Returns:**

- `hash`: Bcrypt hash string
- `err`: Error if hashing fails

**Example:**

```go
hash, err := p.Generate("mypassword")
```

#### `Compare(hash string, plain string) bool`

Compares a bcrypt hash with a plain text password.

**Parameters:**

- `hash`: Bcrypt hash to compare against
- `plain`: Plain text password to verify

**Returns:**

- `bool`: true if password matches, false otherwise

**Example:**

```go
isValid := p.Compare(hash, "mypassword")
```

### Random String

#### `RandomString(length int) (string, error)`

Generates a cryptographically secure random hex string.

**Parameters:**

- `length`: Number of random bytes (output will be 2x length in hex)

**Returns:**

- `string`: Hex-encoded random string (length \* 2 characters)
- `error`: Error if random generation fails

**Example:**

```go
// Generates 64 character hex string (32 bytes)
token, err := encryption.RandomString(32)
```

## Security Considerations

### Password Hashing

1. **Bcrypt Cost**: The package uses `bcrypt.DefaultCost` (10)

   - Provides good balance between security and performance
   - Each increment doubles the time to hash

2. **Salt**: Bcrypt automatically handles salt generation

   - Each password gets a unique salt
   - Salt is embedded in the hash output

3. **Hash Format**: Bcrypt produces hashes like:
   ```
   $2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy
   ```
   - `$2a$`: Bcrypt identifier
   - `10`: Cost factor
   - Next 22 chars: Salt
   - Remaining: Actual hash

### Random Strings

1. **Cryptographic Security**: Uses `crypto/rand`, not `math/rand`
2. **Suitable For**:

   - Session tokens
   - API keys
   - Password reset tokens
   - CSRF tokens
   - Unique identifiers

3. **Output Length**: Returns hex-encoded string (2x input length)
   ```go
   RandomString(16) // Returns 32 character string
   RandomString(32) // Returns 64 character string
   ```

## Common Use Cases

### API Key Generation

```go
apiKey, err := encryption.RandomString(32)
if err != nil {
    return err
}
// Store apiKey in database
```

### Password Reset Token

```go
resetToken, err := encryption.RandomString(32)
if err != nil {
    return err
}

// Store token with expiration
db.StoreResetToken(userID, resetToken, time.Now().Add(1*time.Hour))
```

### Session ID

```go
sessionID, err := encryption.RandomString(24)
if err != nil {
    return err
}

// Use as session identifier
session := NewSession(sessionID, userData)
```

### CSRF Token

```go
csrfToken, err := encryption.RandomString(32)
if err != nil {
    return err
}

// Store in session or cookie
http.SetCookie(w, &http.Cookie{
    Name:     "csrf_token",
    Value:    csrfToken,
    HttpOnly: true,
    Secure:   true,
})
```

## Best Practices

1. **Password Policies**

   - Enforce minimum length (8+ characters)
   - Require mix of character types
   - Check against common password lists
   - Implement rate limiting on login attempts

2. **Password Storage**

   - Never log passwords
   - Never store plain text passwords
   - Only store bcrypt hashes
   - Use prepared statements to prevent SQL injection

3. **Token Generation**

   - Use sufficient length (32+ bytes for tokens)
   - Set appropriate expiration times
   - Implement one-time use for sensitive operations
   - Store securely and transmit over HTTPS only

4. **Error Handling**
   - Don't reveal whether username or password was wrong
   - Use constant-time comparison for tokens
   - Implement account lockout after failed attempts

## Performance

### Bcrypt

- **Speed**: Intentionally slow (~200-300ms per hash)
- **Purpose**: Makes brute force attacks impractical
- **Scalability**: Consider async hashing for high-traffic apps

### Random Generation

- **Speed**: Very fast (~microseconds)
- **No caching needed**: Each call generates new random data

## Testing

```go
func TestPasswordHashing(t *testing.T) {
    p := encryption.NewPassword()

    password := "test-password"
    hash, err := p.Generate(password)
    if err != nil {
        t.Fatal(err)
    }

    // Should match
    if !p.Compare(hash, password) {
        t.Error("Password should match")
    }

    // Should not match
    if p.Compare(hash, "wrong-password") {
        t.Error("Wrong password should not match")
    }
}

func TestRandomString(t *testing.T) {
    s1, err := encryption.RandomString(32)
    if err != nil {
        t.Fatal(err)
    }

    s2, err := encryption.RandomString(32)
    if err != nil {
        t.Fatal(err)
    }

    // Should be unique
    if s1 == s2 {
        t.Error("Random strings should be unique")
    }

    // Should be correct length (64 hex chars from 32 bytes)
    if len(s1) != 64 {
        t.Errorf("Expected length 64, got %d", len(s1))
    }
}
```

## License

MIT License - see [LICENSE](../LICENSE) for details
