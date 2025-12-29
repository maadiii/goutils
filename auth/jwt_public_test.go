package auth

import (
	"testing"
	"time"
)

func TestJWTPublicGenerateAndValidateSuccess(t *testing.T) {
	aPriv, aPub, err := GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("GenerateRSAKeyPair error: %v", err)
	}

	rPriv, rPub, err := GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("GenerateRSAKeyPair error: %v", err)
	}

	j, err := NewJWTPublic(JWTPublicConfig{
		Issuer:             "issuer",
		AccessPrivateKey:   aPriv,
		AccessPublicKey:    aPub,
		RefreshPrivateKey:  rPriv,
		RefreshPublicKey:   rPub,
		AccessTTL:          time.Minute,
		RefreshTTL:         2 * time.Minute,
	})
	if err != nil {
		t.Fatalf("NewJWTPublic error: %v", err)
	}

	toks, err := j.Generate("user@example.com", "api-service", nil)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if toks.Access == "" || toks.Refresh == "" {
		t.Fatalf("expected non-empty tokens")
	}

	acc, err := j.ValidateAccess(toks.Access)
	if err != nil {
		t.Fatalf("ValidateAccess error: %v", err)
	}
	if acc.Subject != "user@example.com" || acc.Audience != "api-service" || acc.TokenType != JWTPublicAccessToken {
		t.Fatalf("unexpected access claims: %+v", acc)
	}

	ref, err := j.ValidateRefresh(toks.Refresh)
	if err != nil {
		t.Fatalf("ValidateRefresh error: %v", err)
	}
	if ref.TokenType != JWTPublicRefreshToken {
		t.Fatalf("unexpected refresh token type: %+v", ref)
	}
}

func TestJWTPublicValidateRejectsWrongKey(t *testing.T) {
	aPriv, aPub, err := GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("GenerateRSAKeyPair error: %v", err)
	}

	rPriv, rPub, err := GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("GenerateRSAKeyPair error: %v", err)
	}

	j, err := NewJWTPublic(JWTPublicConfig{
		Issuer:             "issuer",
		AccessPrivateKey:   aPriv,
		AccessPublicKey:    aPub,
		RefreshPrivateKey:  rPriv,
		RefreshPublicKey:   rPub,
		AccessTTL:          time.Minute,
		RefreshTTL:         time.Minute,
	})
	if err != nil {
		t.Fatalf("NewJWTPublic error: %v", err)
	}

	toks, err := j.Generate("user", "aud", nil)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	// Try to validate with wrong public key
	wrongPriv, wrongPub, _ := GenerateRSAKeyPair()
	jWrong, _ := NewJWTPublic(JWTPublicConfig{
		Issuer:             "issuer",
		AccessPrivateKey:   wrongPriv,
		AccessPublicKey:    wrongPub,
		RefreshPrivateKey:  wrongPriv,
		RefreshPublicKey:   wrongPub,
		AccessTTL:          time.Minute,
		RefreshTTL:         time.Minute,
	})

	if _, err := jWrong.ValidateAccess(toks.Access); err == nil {
		t.Fatalf("expected ValidateAccess to fail with wrong key")
	}
}

func TestJWTPublicValidateRejectsWrongType(t *testing.T) {
	aPriv, aPub, err := GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("GenerateRSAKeyPair error: %v", err)
	}

	rPriv, rPub, err := GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("GenerateRSAKeyPair error: %v", err)
	}

	j, err := NewJWTPublic(JWTPublicConfig{
		Issuer:             "issuer",
		AccessPrivateKey:   aPriv,
		AccessPublicKey:    aPub,
		RefreshPrivateKey:  rPriv,
		RefreshPublicKey:   rPub,
		AccessTTL:          time.Minute,
		RefreshTTL:         time.Minute,
	})
	if err != nil {
		t.Fatalf("NewJWTPublic error: %v", err)
	}

	toks, err := j.Generate("user", "aud", nil)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	if _, err := j.ValidateAccess(toks.Refresh); err == nil {
		t.Fatalf("expected ValidateAccess to fail on refresh token")
	}
	if _, err := j.ValidateRefresh(toks.Access); err == nil {
		t.Fatalf("expected ValidateRefresh to fail on access token")
	}
}

func TestJWTPublicValidateRejectsExpiredToken(t *testing.T) {
	aPriv, aPub, err := GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("GenerateRSAKeyPair error: %v", err)
	}

	rPriv, rPub, err := GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("GenerateRSAKeyPair error: %v", err)
	}

	j, err := NewJWTPublic(JWTPublicConfig{
		Issuer:             "issuer",
		AccessPrivateKey:   aPriv,
		AccessPublicKey:    aPub,
		RefreshPrivateKey:  rPriv,
		RefreshPublicKey:   rPub,
		AccessTTL:          -time.Second, // Expired
		RefreshTTL:         time.Minute,
	})
	if err != nil {
		t.Fatalf("NewJWTPublic error: %v", err)
	}

	toks, err := j.Generate("user", "aud", nil)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	if _, err := j.ValidateAccess(toks.Access); err == nil {
		t.Fatalf("expected ValidateAccess to fail on expired token")
	}
}

func TestJWTPublicCustomClaims(t *testing.T) {
	aPriv, aPub, err := GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("GenerateRSAKeyPair error: %v", err)
	}

	rPriv, rPub, err := GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("GenerateRSAKeyPair error: %v", err)
	}

	j, err := NewJWTPublic(JWTPublicConfig{
		Issuer:             "issuer",
		AccessPrivateKey:   aPriv,
		AccessPublicKey:    aPub,
		RefreshPrivateKey:  rPriv,
		RefreshPublicKey:   rPub,
		AccessTTL:          time.Minute,
		RefreshTTL:         time.Minute,
	})
	if err != nil {
		t.Fatalf("NewJWTPublic error: %v", err)
	}

	customClaims := map[string]interface{}{
		"user_id":  "user-789",
		"role":     "user",
		"tenant":   "tenant-123",
		"perms":    42,
		"enabled":  true,
	}

	toks, err := j.Generate("john@example.com", "mobile-app", customClaims)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	acc, err := j.ValidateAccess(toks.Access)
	if err != nil {
		t.Fatalf("ValidateAccess error: %v", err)
	}

	if acc.CustomClaims == nil {
		t.Fatalf("expected custom claims, got nil")
	}

	if acc.CustomClaims["user_id"] != "user-789" {
		t.Fatalf("expected user_id=user-789, got %v", acc.CustomClaims["user_id"])
	}
	if acc.CustomClaims["role"] != "user" {
		t.Fatalf("expected role=user, got %v", acc.CustomClaims["role"])
	}
	if acc.CustomClaims["tenant"] != "tenant-123" {
		t.Fatalf("expected tenant=tenant-123, got %v", acc.CustomClaims["tenant"])
	}
	if perms, ok := acc.CustomClaims["perms"].(float64); !ok || perms != 42 {
		t.Fatalf("expected perms=42, got %v (%T)", acc.CustomClaims["perms"], acc.CustomClaims["perms"])
	}
	if enabled, ok := acc.CustomClaims["enabled"].(bool); !ok || !enabled {
		t.Fatalf("expected enabled=true, got %v (%T)", acc.CustomClaims["enabled"], acc.CustomClaims["enabled"])
	}
}
