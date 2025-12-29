package auth

import (
	"testing"
	"time"
)

func testKeys() (ak, rk []byte) {
	ak = make([]byte, 32)
	rk = make([]byte, 32)
	for i := range 32 {
		ak[i] = byte(i + 1)
		rk[i] = byte(100 + i)
	}
	return
}

func TestGenerateAndValidateSuccess(t *testing.T) {
	ak, rk := testKeys()
	p, err := NewLocalPaseto(LocalPasetoConfig{
		Issuer:     "issuer",
		AccessKey:  ak,
		RefreshKey: rk,
		AccessTTL:  time.Minute,
		RefreshTTL: 2 * time.Minute,
	})
	if err != nil {
		t.Fatalf("NewLocalPaseto error: %v", err)
	}

	toks, err := p.Generate("subj", "aud", nil)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if toks.Access == "" || toks.Refresh == "" {
		t.Fatalf("expected non-empty tokens")
	}

	acc, err := p.ValidateAccess(toks.Access)
	if err != nil {
		t.Fatalf("ValidateAccess error: %v", err)
	}
	if acc.Subject != "subj" || acc.Audience != "aud" || acc.TokenType != AccessToken {
		t.Fatalf("unexpected access claims: %+v", acc)
	}

	ref, err := p.ValidateRefresh(toks.Refresh)
	if err != nil {
		t.Fatalf("ValidateRefresh error: %v", err)
	}
	if ref.TokenType != RefreshToken {
		t.Fatalf("unexpected refresh token type: %+v", ref)
	}
	if acc.JTI == "" || ref.JTI == "" || acc.JTI == ref.JTI {
		t.Fatalf("expected distinct non-empty JTIs: %q %q", acc.JTI, ref.JTI)
	}
}

func TestValidateRejectsWrongKeyOrType(t *testing.T) {
	ak, rk := testKeys()
	p, err := NewLocalPaseto(LocalPasetoConfig{
		Issuer:     "issuer",
		AccessKey:  ak,
		RefreshKey: rk,
		AccessTTL:  time.Minute,
		RefreshTTL: time.Minute,
	})
	if err != nil {
		t.Fatalf("NewLocalPaseto error: %v", err)
	}

	toks, err := p.Generate("subj", "aud", nil)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	if _, err := p.ValidateAccess(toks.Refresh); err == nil {
		t.Fatalf("expected ValidateAccess to fail on refresh token")
	}
	if _, err := p.ValidateRefresh(toks.Access); err == nil {
		t.Fatalf("expected ValidateRefresh to fail on access token")
	}
}

func TestExpiryIsEnforced(t *testing.T) {
	ak, rk := testKeys()
	p, err := NewLocalPaseto(LocalPasetoConfig{
		Issuer:     "issuer",
		AccessKey:  ak,
		RefreshKey: rk,
		AccessTTL:  50 * time.Millisecond,
		RefreshTTL: time.Second,
	})
	if err != nil {
		t.Fatalf("NewLocalPaseto error: %v", err)
	}

	toks, err := p.Generate("subj", "aud", nil)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	time.Sleep(80 * time.Millisecond)
	if _, err := p.ValidateAccess(toks.Access); err == nil {
		t.Fatalf("expected access token to expire")
	}
	// refresh still valid
	if _, err := p.ValidateRefresh(toks.Refresh); err != nil {
		t.Fatalf("expected refresh to remain valid: %v", err)
	}
}

func TestConstructorRejectsBadKeyLength(t *testing.T) {
	if _, err := NewLocalPaseto(LocalPasetoConfig{
		Issuer:     "iss",
		AccessKey:  []byte{1, 2, 3},
		RefreshKey: make([]byte, 32),
		AccessTTL:  time.Minute,
		RefreshTTL: time.Minute,
	}); err == nil {
		t.Fatalf("expected error for short access key")
	}
	if _, err := NewLocalPaseto(LocalPasetoConfig{
		Issuer:     "iss",
		AccessKey:  make([]byte, 32),
		RefreshKey: []byte{1, 2, 3},
		AccessTTL:  time.Minute,
		RefreshTTL: time.Minute,
	}); err == nil {
		t.Fatalf("expected error for short refresh key")
	}
}

func TestCustomClaimsRoundtrip(t *testing.T) {
	ak, rk := testKeys()
	p, err := NewLocalPaseto(LocalPasetoConfig{
		Issuer:     "issuer",
		AccessKey:  ak,
		RefreshKey: rk,
		AccessTTL:  time.Minute,
		RefreshTTL: time.Minute,
	})
	if err != nil {
		t.Fatalf("NewLocalPaseto error: %v", err)
	}

	customClaims := map[string]any{
		"user_id": "user-123",
		"role":    "admin",
		"org_id":  "org-456",
		"count":   42,
	}

	toks, err := p.Generate("subj", "aud", customClaims)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	acc, err := p.ValidateAccess(toks.Access)
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

	// Note: JSON numbers may come back as float64
	if count, ok := acc.CustomClaims["count"].(float64); !ok || count != 42 {
		t.Fatalf("expected count=42, got %v (%T)", acc.CustomClaims["count"], acc.CustomClaims["count"])
	}
}
