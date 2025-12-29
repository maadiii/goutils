package auth

import (
	"fmt"
	"time"
)

// ExampleLocalPasetoWithCustomClaims demonstrates how to use custom claims with LocalPaseto.
func ExampleLocalPasetoWithCustomClaims() {
	// Generate keys (32 bytes for symmetric encryption)
	accessKey := make([]byte, 32)
	refreshKey := make([]byte, 32)
	// In production, load these from secure storage

	// Create LocalPaseto with configuration
	config := LocalPasetoConfig{
		Issuer:     "my-service",
		AccessKey:  accessKey,
		RefreshKey: refreshKey,
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 7 * 24 * time.Hour,
	}

	p, err := NewLocalPaseto(config)
	if err != nil {
		panic(err)
	}

	// Add custom claims for your application
	customClaims := map[string]interface{}{
		"user_id":     "user-12345",
		"role":        "admin",
		"org_id":      "org-67890",
		"permissions": []string{"read", "write", "delete"},
		"metadata": map[string]interface{}{
			"last_login": time.Now().Unix(),
			"ip_address": "192.168.1.100",
		},
	}

	// Generate tokens with custom claims
	tokens, err := p.Generate("user@example.com", "web-app", customClaims)
	if err != nil {
		panic(err)
	}

	fmt.Println("Access Token:", tokens.Access)
	fmt.Println("Refresh Token:", tokens.Refresh)

	// Validate and extract custom claims from access token
	claims, err := p.ValidateAccess(tokens.Access)
	if err != nil {
		panic(err)
	}

	// Access custom claims
	userID := claims.CustomClaims["user_id"].(string)
	role := claims.CustomClaims["role"].(string)
	orgID := claims.CustomClaims["org_id"].(string)

	fmt.Printf("User ID: %s, Role: %s, Org ID: %s\n", userID, role, orgID)

	// Handle permissions (slice will be []interface{} after JSON roundtrip)
	if perms, ok := claims.CustomClaims["permissions"].([]interface{}); ok {
		for _, perm := range perms {
			fmt.Printf("Permission: %s\n", perm.(string))
		}
	}
}

// ExamplePublicPasetoWithCustomClaims demonstrates how to use custom claims with PublicPaseto.
func ExamplePublicPasetoWithCustomClaims() {
	// Generate Ed25519 keys (64 bytes for private, 32 bytes for public)
	accessPrivateKey := make([]byte, 64)
	accessPublicKey := make([]byte, 32)
	refreshPrivateKey := make([]byte, 64)
	refreshPublicKey := make([]byte, 32)
	// In production, load these from secure storage

	// Create PublicPaseto with configuration
	config := PublicPasetoConfig{
		Issuer:            "my-service",
		AccessPrivateKey:  accessPrivateKey,
		AccessPublicKey:   accessPublicKey,
		RefreshPrivateKey: refreshPrivateKey,
		RefreshPublicKey:  refreshPublicKey,
		AccessTTL:         15 * time.Minute,
		RefreshTTL:        7 * 24 * time.Hour,
	}

	p, err := NewPublicPaseto(config)
	if err != nil {
		panic(err)
	}

	// Add custom claims
	customClaims := map[string]interface{}{
		"user_id":   "user-12345",
		"role":      "admin",
		"tenant_id": "tenant-99",
		"api_key":   "key-abc123",
	}

	// Generate tokens with custom claims
	tokens, err := p.Generate("user@example.com", "api-client", customClaims)
	if err != nil {
		panic(err)
	}

	fmt.Println("Access Token:", tokens.Access)
	fmt.Println("Refresh Token:", tokens.Refresh)

	// Validate and extract custom claims from access token
	claims, err := p.ValidateAccess(tokens.Access)
	if err != nil {
		panic(err)
	}

	// Access custom claims
	userID := claims.CustomClaims["user_id"].(string)
	role := claims.CustomClaims["role"].(string)
	tenantID := claims.CustomClaims["tenant_id"].(string)

	fmt.Printf("User ID: %s, Role: %s, Tenant ID: %s\n", userID, role, tenantID)
}
