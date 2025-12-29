## JWT Implementation Summary

This document summarizes the JWT (JSON Web Token) implementation added to the auth package, complementing the existing PASETO token system.

### Files Created

1. **jwt_local.go** - HMAC-based symmetric JWT tokens
2. **jwt_public.go** - RSA-based asymmetric JWT tokens
3. **jwt_local_test.go** - Comprehensive tests for JWT local
4. **jwt_public_test.go** - Comprehensive tests for JWT public
5. **example_jwt.go** - Usage examples and comparison with PASETO

### JWT Local (HMAC-SHA256)

**Use case**: Fast, symmetric token generation when both parties share a secret key.

```go
// Create JWT local with configuration
config := JWTLocalConfig{
    Issuer:     "my-service",
    AccessKey:  accessKey,      // 64+ bytes recommended
    RefreshKey: refreshKey,     // 64+ bytes recommended
    AccessTTL:  15 * time.Minute,
    RefreshTTL: 7 * 24 * time.Hour,
}

j, err := NewJWTLocal(config)

// Generate tokens with custom claims
customClaims := map[string]interface{}{
    "role":   "admin",
    "org_id": "org-123",
}

tokens, err := j.Generate("user@example.com", "web-app", customClaims)

// Validate and extract claims
claims, err := j.ValidateAccess(tokens.Access)
userRole := claims.CustomClaims["role"].(string)
```

**Features**:

- HMAC-SHA256 signing
- Custom claims support
- Access/Refresh token pair generation
- Type validation (access vs refresh)
- Expiration checking
- Issuer and Audience validation

### JWT Public (RSA-SHA256)

**Use case**: Asymmetric token signing when multiple services need to validate tokens without sharing the private key.

```go
// Generate RSA key pairs
accessPrivate, accessPublic, err := GenerateRSAKeyPair()   // 2048-bit
refreshPrivate, refreshPublic, err := GenerateRSAKeyPair()

// Create JWT public with configuration
config := JWTPublicConfig{
    Issuer:             "my-service",
    AccessPrivateKey:   accessPrivate,
    AccessPublicKey:    accessPublic,
    RefreshPrivateKey:  refreshPrivate,
    RefreshPublicKey:   refreshPublic,
    AccessTTL:          15 * time.Minute,
    RefreshTTL:         7 * 24 * time.Hour,
}

j, err := NewJWTPublic(config)

// Generate and validate same as JWTLocal
tokens, err := j.Generate("user@example.com", "api-service", customClaims)
claims, err := j.ValidateAccess(tokens.Access)
```

**Features**:

- RSA 2048-bit key pairs (RS256)
- Separate key pairs for access/refresh tokens
- Same custom claims and validation as JWTLocal
- Public key distribution for multi-service validation
- JTI (JWT ID) claim generation for tracking

### Standard Claims Structure

Both JWT implementations include these standard claims:

```go
type JWTLocalClaims struct {
    Subject      string                 // sub claim (user_id, email, etc)
    Audience     string                 // aud claim (web-app, api-service, etc)
    Issuer       string                 // iss claim (my-service)
    ExpiresAt    time.Time             // exp claim
    IssuedAt     time.Time             // iat claim
    NotBefore    time.Time             // nbf claim
    TokenType    JWTLocalTokenType     // typ claim (access or refresh)
    CustomClaims map[string]interface{} // Any additional claims
}
```

### Comparison: PASETO vs JWT

| Feature        | PASETO                            | JWT                          |
| -------------- | --------------------------------- | ---------------------------- |
| **Symmetric**  | LocalPaseto                       | JWTLocal                     |
| **Asymmetric** | PublicPaseto (Ed25519)            | JWTPublic (RSA)              |
| **Key Size**   | 32 bytes symmetric, 64/32 Ed25519 | 64+ bytes HMAC, 2048-bit RSA |
| **Standard**   | Modern, less widespread           | Industry standard            |
| **Complexity** | Simpler, explicit                 | More widely known            |
| **Crypto**     | ChaCha20-Poly1305 / EdDSA         | HMAC-SHA256 / RSA-SHA256     |
| **Token Size** | Generally smaller                 | Generally larger             |

### Test Coverage

All implementations have comprehensive test coverage:

**JWTLocal Tests (5 tests)**:

- ✅ Generate and validate success
- ✅ Rejects wrong key
- ✅ Rejects wrong token type
- ✅ Rejects expired tokens
- ✅ Custom claims roundtrip

**JWTPublic Tests (5 tests)**:

- ✅ Generate and validate success
- ✅ Rejects wrong public key
- ✅ Rejects wrong token type
- ✅ Rejects expired tokens
- ✅ Custom claims roundtrip

### Dependencies

- `github.com/golang-jwt/jwt/v5` - JWT library (v5.3.0)

### Usage Examples

See `example_jwt.go` for:

- Complete JWTLocal example with custom claims
- Complete JWTPublic example with RSA keys
- Comparison between PASETO and JWT approaches

### Subject and Audience Guidelines

- **Subject** (`sub`): Identifies who the token is about
  - User ID: `"user-12345"`
  - Email: `"user@example.com"`
  - Username: `"john.doe"`
- **Audience** (`aud`): Identifies who the token is for
  - Application: `"web-app"`, `"mobile-app"`
  - Service: `"api-service"`, `"payment-service"`
  - Endpoint: `"api.example.com"`

### Custom Claims Best Practices

Put user identifiers in **subject**, not custom claims:

```go
// ❌ Not recommended
customClaims := map[string]interface{}{
    "user_id": "user-123",  // Redundant
    "role": "admin",
}
j.Generate("user-123", "web-app", customClaims)

// ✅ Recommended
customClaims := map[string]interface{}{
    "role": "admin",
    "permissions": []string{"read", "write"},
}
j.Generate("user@example.com", "web-app", customClaims)
```

Then access via:

```go
claims, _ := j.ValidateAccess(token)
userID := claims.Subject        // Standard claim
role := claims.CustomClaims["role"].(string)  // Custom claim
```

### Complete Test Results

```
✅ TestJWTLocalGenerateAndValidateSuccess
✅ TestJWTLocalValidateRejectsWrongKey
✅ TestJWTLocalValidateRejectsWrongType
✅ TestJWTLocalValidateRejectsExpiredToken
✅ TestJWTLocalCustomClaims
✅ TestJWTPublicGenerateAndValidateSuccess
✅ TestJWTPublicValidateRejectsWrongKey
✅ TestJWTPublicValidateRejectsWrongType
✅ TestJWTPublicValidateRejectsExpiredToken
✅ TestJWTPublicCustomClaims

Total: 21 tests passing (including PASETO tests)
```

### Next Steps

1. Choose between PASETO (modern, simpler) or JWT (industry standard)
2. Generate keys appropriately (store securely)
3. Use subject for user identifiers
4. Add custom claims for application-specific data
5. Validate tokens on each request
6. Rotate keys periodically
