package auth

import (
	"testing"
	"time"
)

func testHMACKeys() (ak, rk []byte) {
	ak = make([]byte, 64)
	rk = make([]byte, 64)
	for i := range 64 {
		ak[i] = byte(i + 1)
		rk[i] = byte(100 + i)
	}
	return
}

func TestJWTLocalGenerateAndValidateSuccess(t *testing.T) {
	ak, rk := testHMACKeys()
	j, err := NewJWTLocal(JWTLocalConfig{
		Issuer:     "issuer",
		AccessKey:  ak,
		RefreshKey: rk,
		AccessTTL:  time.Minute,
		RefreshTTL: 2 * time.Minute,
	})
	if err != nil {
		t.Fatalf("NewJWTLocal error: %v", err)
	}

	toks, err := j.Generate("user@example.com", "web-app", nil)
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
	if acc.Subject != "user@example.com" || acc.Audience != "web-app" || acc.TokenType != JWTAccessToken {
		t.Fatalf("unexpected access claims: %+v", acc)
	}

	ref, err := j.ValidateRefresh(toks.Refresh)
	if err != nil {
		t.Fatalf("ValidateRefresh error: %v", err)
	}
	if ref.TokenType != JWTRefreshToken {
		t.Fatalf("unexpected refresh token type: %+v", ref)
	}
}

func TestJWTLocalValidateRejectsWrongKey(t *testing.T) {
	ak, rk := testHMACKeys()
	j, err := NewJWTLocal(JWTLocalConfig{
		Issuer:     "issuer",
		AccessKey:  ak,
		RefreshKey: rk,
		AccessTTL:  time.Minute,
		RefreshTTL: time.Minute,
	})
	if err != nil {
		t.Fatalf("NewJWTLocal error: %v", err)
	}

	toks, err := j.Generate("user", "aud", nil)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	// Try to validate with wrong key
	wrongKey := make([]byte, 64)
	for i := range 64 {
		wrongKey[i] = byte(i + 200)
	}
	jWrong, _ := NewJWTLocal(JWTLocalConfig{
		Issuer:     "issuer",
		AccessKey:  wrongKey,
		RefreshKey: wrongKey,
		AccessTTL:  time.Minute,
		RefreshTTL: time.Minute,
	})

	if _, err := jWrong.ValidateAccess(toks.Access); err == nil {
		t.Fatalf("expected ValidateAccess to fail with wrong key")
	}
}

func TestJWTLocalValidateRejectsWrongType(t *testing.T) {
	ak, rk := testHMACKeys()
	j, err := NewJWTLocal(JWTLocalConfig{
		Issuer:     "issuer",
		AccessKey:  ak,
		RefreshKey: rk,
		AccessTTL:  time.Minute,
		RefreshTTL: time.Minute,
	})
	if err != nil {
		t.Fatalf("NewJWTLocal error: %v", err)
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

func TestJWTLocalValidateRejectsExpiredToken(t *testing.T) {
	ak, rk := testHMACKeys()
	j, err := NewJWTLocal(JWTLocalConfig{
		Issuer:     "issuer",
		AccessKey:  ak,
		RefreshKey: rk,
		AccessTTL:  -time.Second, // Expired
		RefreshTTL: time.Minute,
	})
	if err != nil {
		t.Fatalf("NewJWTLocal error: %v", err)
	}

	toks, err := j.Generate("user", "aud", nil)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	if _, err := j.ValidateAccess(toks.Access); err == nil {
		t.Fatalf("expected ValidateAccess to fail on expired token")
	}
}

func TestJWTLocalCustomClaims(t *testing.T) {
	ak, rk := testHMACKeys()
	j, err := NewJWTLocal(JWTLocalConfig{
		Issuer:     "issuer",
		AccessKey:  ak,
		RefreshKey: rk,
		AccessTTL:  time.Minute,
		RefreshTTL: time.Minute,
	})
	if err != nil {
		t.Fatalf("NewJWTLocal error: %v", err)
	}

	customClaims := map[string]interface{}{
		"user_id":  "user-123",
		"role":     "admin",
		"org_id":   "org-456",
		"count":    42,
		"verified": true,
	}

	toks, err := j.Generate("user@example.com", "web-app", customClaims)
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

	if acc.CustomClaims["user_id"] != "user-123" {
		t.Fatalf("expected user_id=user-123, got %v", acc.CustomClaims["user_id"])
	}
	if acc.CustomClaims["role"] != "admin" {
		t.Fatalf("expected role=admin, got %v", acc.CustomClaims["role"])
	}
	if acc.CustomClaims["org_id"] != "org-456" {
		t.Fatalf("expected org_id=org-456, got %v", acc.CustomClaims["org_id"])
	}
	if count, ok := acc.CustomClaims["count"].(float64); !ok || count != 42 {
		t.Fatalf("expected count=42, got %v (%T)", acc.CustomClaims["count"], acc.CustomClaims["count"])
	}
	if verified, ok := acc.CustomClaims["verified"].(bool); !ok || !verified {
		t.Fatalf("expected verified=true, got %v (%T)", acc.CustomClaims["verified"], acc.CustomClaims["verified"])
	}
}
