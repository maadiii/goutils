package auth

import (
	"testing"
	"time"

	psto "aidanwoods.dev/go-paseto"
)

func testPublicKeys() (aPvt, aPub, rPvt, rPub []byte) {
	// Generate proper Ed25519 keypairs for access and refresh
	accessKey := psto.NewV4AsymmetricSecretKey()
	refreshKey := psto.NewV4AsymmetricSecretKey()

	aPvt = accessKey.ExportBytes()
	aPub = accessKey.Public().ExportBytes()
	rPvt = refreshKey.ExportBytes()
	rPub = refreshKey.Public().ExportBytes()

	return
}

func TestPublicGenerateAndValidateSuccess(t *testing.T) {
	aPvt, aPub, rPvt, rPub := testPublicKeys()
	p, err := NewPublicPaseto(PublicPasetoConfig{
		Issuer:            "issuer",
		AccessPrivateKey:  aPvt,
		AccessPublicKey:   aPub,
		RefreshPrivateKey: rPvt,
		RefreshPublicKey:  rPub,
		AccessTTL:         time.Minute,
		RefreshTTL:        2 * time.Minute,
	})
	if err != nil {
		t.Fatalf("NewPublicPaseto error: %v", err)
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
	if acc.Subject != "subj" || acc.Audience != "aud" || acc.TokenType != PublicAccessToken {
		t.Fatalf("unexpected access claims: %+v", acc)
	}

	ref, err := p.ValidateRefresh(toks.Refresh)
	if err != nil {
		t.Fatalf("ValidateRefresh error: %v", err)
	}
	if ref.TokenType != PublicRefreshToken {
		t.Fatalf("unexpected refresh token type: %+v", ref)
	}
	if acc.JTI == "" || ref.JTI == "" || acc.JTI == ref.JTI {
		t.Fatalf("expected distinct non-empty JTIs: %q %q", acc.JTI, ref.JTI)
	}
}

func TestPublicValidateRejectsWrongKeyOrType(t *testing.T) {
	aPvt, aPub, rPvt, rPub := testPublicKeys()
	p, err := NewPublicPaseto(PublicPasetoConfig{
		Issuer:            "issuer",
		AccessPrivateKey:  aPvt,
		AccessPublicKey:   aPub,
		RefreshPrivateKey: rPvt,
		RefreshPublicKey:  rPub,
		AccessTTL:         time.Minute,
		RefreshTTL:        time.Minute,
	})
	if err != nil {
		t.Fatalf("NewPublicPaseto error: %v", err)
	}

	toks, err := p.Generate("subj", "aud", nil)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	// Try validating refresh token as access (wrong key)
	if _, err := p.ValidateAccess(toks.Refresh); err == nil {
		t.Fatalf("expected ValidateAccess to fail on refresh token")
	}
	// Try validating access token as refresh (wrong key)
	if _, err := p.ValidateRefresh(toks.Access); err == nil {
		t.Fatalf("expected ValidateRefresh to fail on access token")
	}
}

func TestPublicExpiryIsEnforced(t *testing.T) {
	aPvt, aPub, rPvt, rPub := testPublicKeys()
	p, err := NewPublicPaseto(PublicPasetoConfig{
		Issuer:            "issuer",
		AccessPrivateKey:  aPvt,
		AccessPublicKey:   aPub,
		RefreshPrivateKey: rPvt,
		RefreshPublicKey:  rPub,
		AccessTTL:         50 * time.Millisecond,
		RefreshTTL:        time.Second,
	})
	if err != nil {
		t.Fatalf("NewPublicPaseto error: %v", err)
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

func TestPublicConstructorRejectsBadKeyLength(t *testing.T) {
	_, aPub, _, rPub := testPublicKeys()

	// Bad private key length (access)
	if _, err := NewPublicPaseto(PublicPasetoConfig{
		Issuer:            "iss",
		AccessPrivateKey:  []byte{1, 2, 3},
		AccessPublicKey:   aPub,
		RefreshPrivateKey: make([]byte, 64),
		RefreshPublicKey:  rPub,
		AccessTTL:         time.Minute,
		RefreshTTL:        time.Minute,
	}); err == nil {
		t.Fatalf("expected error for short access private key")
	}

	// Bad public key length (access)
	if _, err := NewPublicPaseto(PublicPasetoConfig{
		Issuer:            "iss",
		AccessPrivateKey:  make([]byte, 64),
		AccessPublicKey:   []byte{1, 2, 3},
		RefreshPrivateKey: make([]byte, 64),
		RefreshPublicKey:  rPub,
		AccessTTL:         time.Minute,
		RefreshTTL:        time.Minute,
	}); err == nil {
		t.Fatalf("expected error for short access public key")
	}
}

func TestPublicTokensCannotBeValidatedWithWrongPublicKey(t *testing.T) {
	aPvt, aPub, rPvt, rPub := testPublicKeys()
	p, err := NewPublicPaseto(PublicPasetoConfig{
		Issuer:            "issuer",
		AccessPrivateKey:  aPvt,
		AccessPublicKey:   aPub,
		RefreshPrivateKey: rPvt,
		RefreshPublicKey:  rPub,
		AccessTTL:         time.Minute,
		RefreshTTL:        time.Minute,
	})
	if err != nil {
		t.Fatalf("NewPublicPaseto error: %v", err)
	}

	toks, err := p.Generate("subj", "aud", nil)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	// Create another instance with different keys
	aPvt2, aPub2, rPvt2, rPub2 := testPublicKeys()
	p2, err := NewPublicPaseto(PublicPasetoConfig{
		Issuer:            "issuer",
		AccessPrivateKey:  aPvt2,
		AccessPublicKey:   aPub2,
		RefreshPrivateKey: rPvt2,
		RefreshPublicKey:  rPub2,
		AccessTTL:         time.Minute,
		RefreshTTL:        time.Minute,
	})
	if err != nil {
		t.Fatalf("NewPublicPaseto error: %v", err)
	}

	// Should fail validation with wrong public key
	if _, err := p2.ValidateAccess(toks.Access); err == nil {
		t.Fatalf("expected ValidateAccess to fail with wrong public key")
	}
	if _, err := p2.ValidateRefresh(toks.Refresh); err == nil {
		t.Fatalf("expected ValidateRefresh to fail with wrong public key")
	}
}

func TestPublicCustomClaimsRoundtrip(t *testing.T) {
	aPvt, aPub, rPvt, rPub := testPublicKeys()
	p, err := NewPublicPaseto(PublicPasetoConfig{
		Issuer:            "issuer",
		AccessPrivateKey:  aPvt,
		AccessPublicKey:   aPub,
		RefreshPrivateKey: rPvt,
		RefreshPublicKey:  rPub,
		AccessTTL:         time.Minute,
		RefreshTTL:        time.Minute,
	})
	if err != nil {
		t.Fatalf("NewPublicPaseto error: %v", err)
	}

	customClaims := map[string]interface{}{
		"user_id":     "user-789",
		"role":        "viewer",
		"permissions": []string{"read", "write"},
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

	if acc.CustomClaims["user_id"] != "user-789" {
		t.Fatalf("expected user_id=user-789, got %v", acc.CustomClaims["user_id"])
	}
	if acc.CustomClaims["role"] != "viewer" {
		t.Fatalf("expected role=viewer, got %v", acc.CustomClaims["role"])
	}
}
