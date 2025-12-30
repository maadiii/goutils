# TOTP

Time-based One-Time Password (TOTP) implementation for two-factor authentication.

## Overview

The TOTP package provides a simple and secure way to implement time-based one-time passwords for two-factor authentication (2FA). It's compatible with authenticator apps like Google Authenticator, Authy, and Microsoft Authenticator.

## Installation

```bash
go get github.com/maadiii/utils/totp
```

## Features

- 🔐 RFC 6238 compliant TOTP implementation
- 📱 Compatible with all major authenticator apps
- ⏰ Customizable time periods and code lengths
- 🎯 Simple API for generation and validation
- ✅ QR code support via secret URL

## Usage

### Basic Setup

```go
package main

import (
    "fmt"
    "github.com/maadiii/utils/totp"
)

func main() {
    t := totp.New()

    // Generate TOTP secret and initial code
    secret, code, err := t.Generate(totp.Opts{
        Issuer:      "MyApp",
        AccountName: "user@example.com",
        Period:      30,  // 30 seconds
        Digits:      6,   // 6-digit codes
    })
    if err != nil {
        panic(err)
    }

    fmt.Println("Secret:", secret)
    fmt.Println("Initial code:", code)

    // User enters this code from their authenticator app
    userCode := "123456"

    // Validate the code
    err = t.Validate(userCode, secret, totp.Opts{
        Period: 30,
        Digits: 6,
    })

    if err != nil {
        fmt.Println("Invalid code")
    } else {
        fmt.Println("Valid code!")
    }
}
```

### Complete 2FA Implementation

```go
package main

import (
    "fmt"
    "github.com/maadiii/utils/totp"
)

type User struct {
    ID         string
    Email      string
    TOTPSecret string
    TOTPEnabled bool
}

// Enable 2FA for a user
func Enable2FA(user *User) (string, string, error) {
    t := totp.New()

    // Generate TOTP secret
    secret, initialCode, err := t.Generate(totp.Opts{
        Issuer:      "MyApp",
        AccountName: user.Email,
        Period:      30,
        Digits:      6,
    })
    if err != nil {
        return "", "", err
    }

    // Save secret to database (encrypted!)
    user.TOTPSecret = secret
    user.TOTPEnabled = true

    // Return QR code URL for user to scan
    qrURL := fmt.Sprintf("otpauth://totp/MyApp:%s?secret=%s&issuer=MyApp",
        user.Email, secret)

    return qrURL, initialCode, nil
}

// Verify 2FA code during login
func Verify2FA(user *User, code string) error {
    if !user.TOTPEnabled {
        return fmt.Errorf("2FA not enabled")
    }

    t := totp.New()

    err := t.Validate(code, user.TOTPSecret, totp.Opts{
        Period: 30,
        Digits: 6,
    })

    return err
}

func main() {
    user := &User{
        ID:    "123",
        Email: "user@example.com",
    }

    // Enable 2FA
    qrURL, initialCode, err := Enable2FA(user)
    if err != nil {
        panic(err)
    }

    fmt.Println("Scan this QR code:", qrURL)
    fmt.Println("Or enter this code:", initialCode)

    // Later, during login
    userEnteredCode := "123456" // From authenticator app

    if err := Verify2FA(user, userEnteredCode); err != nil {
        fmt.Println("2FA verification failed:", err)
    } else {
        fmt.Println("2FA verification successful!")
    }
}
```

### With HTTP Handlers

```go
import (
    "github.com/labstack/echo/v4"
    "github.com/maadiii/utils/totp"
)

type Enable2FAResponse struct {
    QRCodeURL string `json:"qr_code_url"`
    Secret    string `json:"secret"`
    Code      string `json:"initial_code"`
}

func HandleEnable2FA(c echo.Context) error {
    userID := c.Get("user_id").(string)
    user := getUserByID(userID)

    t := totp.New()

    secret, code, err := t.Generate(totp.Opts{
        Issuer:      "MyApp",
        AccountName: user.Email,
        Period:      30,
        Digits:      6,
    })
    if err != nil {
        return c.JSON(500, map[string]string{"error": "failed to generate TOTP"})
    }

    // Save secret (encrypted) to database
    saveUserTOTPSecret(userID, secret)

    qrURL := fmt.Sprintf("otpauth://totp/MyApp:%s?secret=%s&issuer=MyApp",
        user.Email, secret)

    return c.JSON(200, Enable2FAResponse{
        QRCodeURL: qrURL,
        Secret:    secret,
        Code:      code,
    })
}

type Verify2FARequest struct {
    Code string `json:"code" validate:"required,len=6"`
}

func HandleVerify2FA(c echo.Context) error {
    var req Verify2FARequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(400, map[string]string{"error": "invalid request"})
    }

    userID := c.Get("user_id").(string)
    secret := getUserTOTPSecret(userID)

    t := totp.New()

    err := t.Validate(req.Code, secret, totp.Opts{
        Period: 30,
        Digits: 6,
    })

    if err != nil {
        return c.JSON(401, map[string]string{"error": "invalid code"})
    }

    return c.JSON(200, map[string]string{"message": "verification successful"})
}
```

## API Reference

### Types

#### `Opts`

Configuration options for TOTP generation and validation.

```go
type Opts struct {
    Issuer      string // App or service name
    AccountName string // User identifier (email, username)
    Period      uint   // Time period in seconds (default: 30)
    Digits      int    // Number of digits in code (default: 6)
}
```

### Methods

#### `New() *totp`

Creates a new TOTP instance.

```go
t := totp.New()
```

#### `Generate(opts Opts) (secret string, code string, err error)`

Generates a new TOTP secret and initial code.

**Parameters:**

- `opts`: Configuration options

**Returns:**

- `secret`: Base32-encoded secret to store
- `code`: Initial TOTP code
- `err`: Error if generation fails

**Example:**

```go
secret, code, err := t.Generate(totp.Opts{
    Issuer:      "MyApp",
    AccountName: "user@example.com",
    Period:      30,
    Digits:      6,
})
```

#### `Validate(code string, secret string, opts Opts) error`

Validates a TOTP code against a secret.

**Parameters:**

- `code`: User-provided TOTP code
- `secret`: Stored secret (base32-encoded)
- `opts`: Same options used during generation

**Returns:**

- `error`: nil if valid, error if invalid

**Example:**

```go
err := t.Validate("123456", secret, totp.Opts{
    Period: 30,
    Digits: 6,
})
```

## Configuration Options

### Period

The time window for code validity (in seconds):

- **30 seconds** (recommended): Standard period, good balance
- **60 seconds**: Longer window, more user-friendly
- **15 seconds**: Shorter window, more secure

```go
Opts{
    Period: 30, // 30-second codes
}
```

### Digits

Number of digits in the generated code:

- **6 digits** (recommended): Standard, fits all authenticators
- **8 digits**: More secure, harder to guess

```go
Opts{
    Digits: 6, // 6-digit codes
}
```

## QR Code Generation

For a better user experience, generate QR codes that users can scan:

### Manual QR URL

```go
qrURL := fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&period=%d&digits=%d",
    issuer,
    accountName,
    secret,
    issuer,
    period,
    digits,
)
```

### Using QR Code Library

```go
import "github.com/skip2/go-qrcode"

func GenerateQRCode(secret, issuer, accountName string) ([]byte, error) {
    url := fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s",
        issuer, accountName, secret, issuer)

    // Generate QR code as PNG
    png, err := qrcode.Encode(url, qrcode.Medium, 256)
    return png, err
}
```

## Security Considerations

1. **Store Secrets Securely**

   ```go
   // Always encrypt TOTP secrets before storing
   encryptedSecret := encrypt(secret)
   db.Save(userID, encryptedSecret)
   ```

2. **Use HTTPS**

   - Always transmit secrets over HTTPS
   - Display QR codes only over secure connections

3. **Backup Codes**

   - Provide backup codes for account recovery
   - Store them hashed like passwords

4. **Rate Limiting**

   ```go
   // Limit verification attempts
   if attempts > 3 {
       return errors.New("too many attempts")
   }
   ```

5. **Time Synchronization**
   - Ensure server time is accurate (use NTP)
   - TOTP relies on synchronized time

## Common Patterns

### Enable 2FA Flow

1. User requests to enable 2FA
2. Generate secret and QR code
3. Display QR code to user
4. User scans with authenticator app
5. User enters verification code
6. Verify code before enabling
7. Save secret to database

### Login with 2FA Flow

1. User enters username/password
2. If credentials valid and 2FA enabled:
   - Prompt for TOTP code
   - Validate code
   - Grant access if valid
3. Implement rate limiting for attempts

### Backup Codes

```go
func GenerateBackupCodes() []string {
    codes := make([]string, 10)
    for i := range codes {
        code, _ := encryption.RandomString(8)
        codes[i] = code
    }
    return codes
}

func SaveBackupCodes(userID string, codes []string) {
    for _, code := range codes {
        hash, _ := encryption.HashPassword(code)
        db.SaveBackupCode(userID, hash)
    }
}
```

### Disable 2FA

```go
func Disable2FA(userID string, code string) error {
    secret := getUserTOTPSecret(userID)

    // Verify current code before disabling
    t := totp.New()
    if err := t.Validate(code, secret, totp.Opts{
        Period: 30,
        Digits: 6,
    }); err != nil {
        return errors.New("invalid code")
    }

    // Remove secret from database
    db.DeleteTOTPSecret(userID)
    db.UpdateUser(userID, map[string]bool{"totp_enabled": false})

    return nil
}
```

## Testing

### Time-Based Testing

```go
func TestTOTP(t *testing.T) {
    tp := totp.New()

    // Generate secret
    secret, code, err := tp.Generate(totp.Opts{
        Issuer:      "TestApp",
        AccountName: "test@example.com",
        Period:      30,
        Digits:      6,
    })

    if err != nil {
        t.Fatal(err)
    }

    // Validate immediately (should work)
    err = tp.Validate(code, secret, totp.Opts{
        Period: 30,
        Digits: 6,
    })

    if err != nil {
        t.Error("Valid code should pass")
    }

    // Invalid code should fail
    err = tp.Validate("000000", secret, totp.Opts{
        Period: 30,
        Digits: 6,
    })

    if err == nil {
        t.Error("Invalid code should fail")
    }
}
```

## Troubleshooting

### Code Always Invalid

- Check server time synchronization
- Verify Period and Digits match between generation and validation
- Ensure secret is stored correctly
- Check for time drift between server and authenticator

### Codes Expire Too Quickly

- Increase Period from 30 to 60 seconds
- Implement a time skew tolerance window

### User Lost Access

- Implement backup codes
- Provide account recovery mechanism
- Consider SMS fallback (though less secure)

## Compatible Authenticator Apps

- Google Authenticator
- Microsoft Authenticator
- Authy
- 1Password
- LastPass Authenticator
- FreeOTP
- Duo Mobile

## Best Practices

1. Always verify a code before enabling 2FA
2. Provide clear setup instructions with screenshots
3. Generate and display backup codes
4. Implement account recovery process
5. Allow users to disable 2FA (with verification)
6. Log 2FA events for security auditing
7. Send notification emails when 2FA is changed

## License

MIT License - see [LICENSE](../LICENSE) for details
