# Auth

Comprehensive authentication package supporting JWT and PASETO tokens with both symmetric (local) and asymmetric (public) key operations.

## Overview

This package provides robust token-based authentication solutions with support for:

- JWT (JSON Web Tokens) with HMAC and RSA/ECDSA signatures
- PASETO (Platform-Agnostic Security Tokens) v4
- Access and refresh token patterns
- Custom claims support

For detailed implementation information, see [JWT_IMPLEMENTATION.md](JWT_IMPLEMENTATION.md).

## Installation

```bash
go get github.com/maadiii/goutils/auth
```

## Features

- 🔐 JWT Local (HMAC-SHA256)
- 🔑 JWT Public (RSA/ECDSA)
- 🛡️ PASETO Local (symmetric)
- 📝 PASETO Public (asymmetric)
- ⏰ Access & Refresh token support
- 🎯 Custom claims
- ✅ Token validation

## Usage

### JWT Local (Symmetric - HMAC)

Best for single-server applications or when all services share the same secret.

```go
package main

import (
    "fmt"
    "time"

    "github.com/maadiii/goutils/auth"
)

func main() {
    // Configure JWT Local
    config := auth.JWTLocalConfig{
        Issuer:     "my-app",
        AccessKey:  []byte("your-secret-access-key-min-32-bytes"),
        RefreshKey: []byte("your-secret-refresh-key-min-32-bytes"),
        AccessTTL:  15 * time.Minute,
        RefreshTTL: 7 * 24 * time.Hour,
    }

    jwtLocal := auth.NewJWTLocal(config)

    // Generate access token
    accessToken, err := jwtLocal.GenerateAccessToken("user123", map[string]interface{}{
        "email": "user@example.com",
        "role":  "admin",
    })
    if err != nil {
        panic(err)
    }

    fmt.Println("Access Token:", accessToken)

    // Verify and extract claims
    claims, err := jwtLocal.VerifyAccessToken(accessToken)
    if err != nil {
        panic(err)
    }

    fmt.Printf("User ID: %s\n", claims.Subject)
    fmt.Printf("Email: %v\n", claims.CustomClaims["email"])
}
```

### JWT Public (Asymmetric - RSA/ECDSA)

Best for microservices where different services need to verify tokens without sharing secrets.

```go
// Generate RSA key pair
privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
if err != nil {
    panic(err)
}

config := auth.JWTPublicConfig{
    Issuer:         "my-app",
    AccessPrivate:  privateKey,
    RefreshPrivate: privateKey,
    AccessTTL:      15 * time.Minute,
    RefreshTTL:     7 * 24 * time.Hour,
}

jwtPublic := auth.NewJWTPublic(config)

// Generate token
token, err := jwtPublic.GenerateAccessToken("user123", map[string]interface{}{
    "permissions": []string{"read", "write"},
})

// Verify with public key
claims, err := jwtPublic.VerifyAccessToken(token)
```

### PASETO Local (Symmetric)

Modern alternative to JWT with better security defaults.

```go
// Create PASETO Local authenticator
// Key must be exactly 32 bytes
secretKey := []byte("your-32-byte-secret-key-here!")

pasetoLocal := auth.NewPasetoLocal(secretKey)

// Generate access token
accessToken, err := pasetoLocal.GenerateAccessToken(
    "user123",
    "my-app",
    "my-service",
    15*time.Minute,
    map[string]interface{}{
        "email": "user@example.com",
        "plan":  "premium",
    },
)

// Verify token
claims, err := pasetoLocal.VerifyAccessToken(accessToken)
if err != nil {
    panic(err)
}

fmt.Printf("Subject: %s\n", claims.Subject)
fmt.Printf("Email: %v\n", claims.CustomClaims["email"])
```

### PASETO Public (Asymmetric)

PASETO with public-key cryptography using Ed25519.

```go
import "crypto/ed25519"

// Generate Ed25519 key pair
publicKey, privateKey, err := ed25519.GenerateKey(nil)
if err != nil {
    panic(err)
}

pasetoPublic := auth.NewPasetoPublic(privateKey, publicKey)

// Generate token
token, err := pasetoPublic.GenerateAccessToken(
    "user123",
    "my-app",
    "my-service",
    15*time.Minute,
    map[string]interface{}{
        "department": "engineering",
    },
)

// Anyone with the public key can verify
claims, err := pasetoPublic.VerifyAccessToken(token)
```

### Refresh Tokens

All authenticators support refresh token patterns:

```go
// Generate refresh token (longer TTL)
refreshToken, err := jwtLocal.GenerateRefreshToken("user123", nil)

// Later, verify refresh token
claims, err := jwtLocal.VerifyRefreshToken(refreshToken)
if err != nil {
    // Refresh token expired or invalid
    // User needs to login again
}

// Generate new access token
newAccessToken, err := jwtLocal.GenerateAccessToken(
    claims.Subject,
    claims.CustomClaims,
)
```

## Token Types

### JWT Local

- **Algorithm**: HMAC-SHA256
- **Use Case**: Single server or shared secret environment
- **Key Type**: Symmetric (byte slice)
- **Performance**: Fast
- **Distribution**: Requires secure key distribution

### JWT Public

- **Algorithm**: RSA or ECDSA
- **Use Case**: Microservices, distributed systems
- **Key Type**: Asymmetric (private/public key pair)
- **Performance**: Slower than symmetric
- **Distribution**: Only public key needs distribution

### PASETO Local

- **Algorithm**: XChaCha20-Poly1305
- **Use Case**: Modern alternative to JWT
- **Key Type**: Symmetric (32 bytes)
- **Security**: Better defaults than JWT
- **Format**: Not compatible with JWT

### PASETO Public

- **Algorithm**: Ed25519
- **Use Case**: Modern distributed systems
- **Key Type**: Asymmetric (Ed25519 keys)
- **Performance**: Fast signature verification
- **Format**: Not compatible with JWT

## API Reference

### JWT Local

```go
type JWTLocal struct { }

func NewJWTLocal(config JWTLocalConfig) *JWTLocal
func (j *JWTLocal) GenerateAccessToken(subject string, claims map[string]interface{}) (string, error)
func (j *JWTLocal) GenerateRefreshToken(subject string, claims map[string]interface{}) (string, error)
func (j *JWTLocal) VerifyAccessToken(token string) (*JWTLocalClaims, error)
func (j *JWTLocal) VerifyRefreshToken(token string) (*JWTLocalClaims, error)
```

### JWT Public

```go
type JWTPublic struct { }

func NewJWTPublic(config JWTPublicConfig) *JWTPublic
func (j *JWTPublic) GenerateAccessToken(subject string, claims map[string]interface{}) (string, error)
func (j *JWTPublic) VerifyAccessToken(token string) (*JWTPublicClaims, error)
// ... refresh token methods
```

### PASETO Local

```go
type PasetoLocal struct { }

func NewPasetoLocal(secretKey []byte) *PasetoLocal
func (p *PasetoLocal) GenerateAccessToken(subject, issuer, audience string, ttl time.Duration, claims map[string]interface{}) (string, error)
func (p *PasetoLocal) VerifyAccessToken(token string) (*PasetoLocalClaims, error)
// ... refresh token methods
```

### PASETO Public

```go
type PasetoPublic struct { }

func NewPasetoPublic(privateKey ed25519.PrivateKey, publicKey ed25519.PublicKey) *PasetoPublic
func (p *PasetoPublic) GenerateAccessToken(subject, issuer, audience string, ttl time.Duration, claims map[string]interface{}) (string, error)
func (p *PasetoPublic) VerifyAccessToken(token string) (*PasetoPublicClaims, error)
```

## Security Best Practices

1. **Key Management**

   - Never commit secrets to version control
   - Use environment variables or secure vaults
   - Rotate keys periodically
   - Use minimum 32 bytes for symmetric keys

2. **Token Lifetime**

   - Keep access tokens short-lived (15-30 minutes)
   - Refresh tokens can be longer (days/weeks)
   - Implement token revocation for sensitive operations

3. **Claims**

   - Don't store sensitive data in tokens
   - Tokens are base64-encoded, not encrypted
   - Validate all claims on verification

4. **HTTPS Only**
   - Always transmit tokens over HTTPS
   - Use secure cookie flags in browsers

## Choosing the Right Authenticator

| Feature       | JWT Local | JWT Public | PASETO Local | PASETO Public |
| ------------- | --------- | ---------- | ------------ | ------------- |
| Speed         | ⚡⚡⚡    | ⚡⚡       | ⚡⚡⚡       | ⚡⚡⚡        |
| Security      | ✅        | ✅✅       | ✅✅✅       | ✅✅✅        |
| Microservices | ❌        | ✅         | ❌           | ✅            |
| Standard      | JWT       | JWT        | PASETO       | PASETO        |
| Compatibility | High      | High       | Low          | Low           |

## Testing

Run tests:

```bash
go test -v ./auth/...
```

## License

MIT License - see [LICENSE](../LICENSE) for details
