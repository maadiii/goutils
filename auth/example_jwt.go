package auth

import (
	"fmt"
	"time"
)

// ExampleJWTLocalWithCustomClaims demonstrates how to use custom claims with JWTLocal (HMAC).
func ExampleJWTLocalWithCustomClaims() {
	// Generate HMAC keys (at least 32 bytes recommended)
	accessKey := make([]byte, 64)
	refreshKey := make([]byte, 64)
	// In production, load these from secure storage

	// Create JWTLocal with configuration
	config := JWTLocalConfig{
		Issuer:     "my-service",
		AccessKey:  accessKey,
		RefreshKey: refreshKey,
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 7 * 24 * time.Hour,
	}

	j, err := NewJWTLocal(config)
	if err != nil {
		panic(err)
	}

	// Add custom claims for your application
	customClaims := map[string]interface{}{
		"role":        "admin",
		"org_id":      "org-67890",
		"permissions": []string{"read", "write", "delete"},
	}

	// Generate tokens with custom claims
	// Note: Subject can be user_id, email, or username
	tokens, err := j.Generate("user@example.com", "web-app", customClaims)
	if err != nil {
		panic(err)
	}

	fmt.Println("JWT Access Token:", tokens.Access)
	fmt.Println("JWT Refresh Token:", tokens.Refresh)

	// Validate and extract custom claims from access token
	claims, err := j.ValidateAccess(tokens.Access)
	if err != nil {
		panic(err)
	}

	// Access standard claims
	fmt.Printf("Subject: %s, Audience: %s\n", claims.Subject, claims.Audience)

	// Access custom claims
	role := claims.CustomClaims["role"].(string)
	orgID := claims.CustomClaims["org_id"].(string)
	fmt.Printf("Role: %s, Org ID: %s\n", role, orgID)
}

// ExampleJWTPublicWithCustomClaims demonstrates how to use custom claims with JWTPublic (RSA).
func ExampleJWTPublicWithCustomClaims() {
	// Generate RSA key pairs for signing
	accessPrivate, accessPublic, err := GenerateRSAKeyPair()
	if err != nil {
		panic(err)
	}

	refreshPrivate, refreshPublic, err := GenerateRSAKeyPair()
	if err != nil {
		panic(err)
	}

	// Create JWTPublic with configuration
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
	if err != nil {
		panic(err)
	}

	// Add custom claims
	customClaims := map[string]interface{}{
		"role":      "user",
		"tenant_id": "tenant-99",
		"api_key":   "key-abc123",
	}

	// Generate tokens with custom claims
	// Note: Subject can be user_id, email, or username
	tokens, err := j.Generate("user@example.com", "api-client", customClaims)
	if err != nil {
		panic(err)
	}

	fmt.Println("JWT Access Token:", tokens.Access)
	fmt.Println("JWT Refresh Token:", tokens.Refresh)

	// Validate and extract custom claims from access token
	claims, err := j.ValidateAccess(tokens.Access)
	if err != nil {
		panic(err)
	}

	// Access standard claims
	fmt.Printf("Subject: %s, Audience: %s\n", claims.Subject, claims.Audience)

	// Access custom claims
	role := claims.CustomClaims["role"].(string)
	tenantID := claims.CustomClaims["tenant_id"].(string)
	fmt.Printf("Role: %s, Tenant ID: %s\n", role, tenantID)
}

// ExampleComparison shows the difference between PASETO and JWT approaches
func ExampleComparison() {
	// Both PASETO and JWT support the same pattern:
	// 1. Symmetric (local) vs Asymmetric (public) signing
	// 2. Custom claims support
	// 3. Access/Refresh token separation
	// 4. Expiration and validation

	// PASETO (Symmetric):
	// - LocalPaseto with 32-byte symmetric keys
	// - More limited standard claims
	// - Shorter tokens

	// JWT (Symmetric):
	// - JWTLocal with HMAC-SHA256 keys
	// - More widely supported
	// - Industry standard

	// PASETO (Asymmetric):
	// - PublicPaseto with Ed25519 keys (64-byte private, 32-byte public)
	// - Modern elliptic curve cryptography
	// - Better for large scale

	// JWT (Asymmetric):
	// - JWTPublic with RSA keys
	// - Most widely supported
	// - Easier key distribution

	fmt.Println("Choose PASETO for modern crypto, JWT for maximum compatibility")
}
