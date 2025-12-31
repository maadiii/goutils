package auth

import (
	"errors"
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
		RefreshTTL:        2 * time.Second,
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

func TestPublicConstructorRejectsBadRefreshKeyLength(t *testing.T) {
	aPvt, aPub, _, _ := testPublicKeys()
	// Bad refresh private key length
	if _, err := NewPublicPaseto(PublicPasetoConfig{
		Issuer:            "iss",
		AccessPrivateKey:  aPvt,
		AccessPublicKey:   aPub,
		RefreshPrivateKey: []byte{1, 2, 3},
		RefreshPublicKey:  make([]byte, 32),
		AccessTTL:         time.Minute,
		RefreshTTL:        time.Minute,
	}); err == nil {
		t.Fatalf("expected error for short refresh private key")
	}
	// Bad refresh public key length
	if _, err := NewPublicPaseto(PublicPasetoConfig{
		Issuer:            "iss",
		AccessPrivateKey:  aPvt,
		AccessPublicKey:   aPub,
		RefreshPrivateKey: make([]byte, 64),
		RefreshPublicKey:  []byte{1, 2, 3},
		AccessTTL:         time.Minute,
		RefreshTTL:        time.Minute,
	}); err == nil {
		t.Fatalf("expected error for short refresh public key")
	}
}

func TestNewPublicPasetoConstructorLibErrors(t *testing.T) {
	aPvt, aPub, rPvt, rPub := testPublicKeys()
	// Stub secret/public key constructors to fail
	origSec := newV4AsymmetricSecretKeyFromBytes
	origPub := newV4AsymmetricPublicKeyFromBytes
	defer func() {
		newV4AsymmetricSecretKeyFromBytes = origSec
		newV4AsymmetricPublicKeyFromBytes = origPub
	}()

	// Fail access private key
	newV4AsymmetricSecretKeyFromBytes = func(b []byte) (psto.V4AsymmetricSecretKey, error) {
		return psto.V4AsymmetricSecretKey{}, errors.New("fail")
	}
	if _, err := NewPublicPaseto(PublicPasetoConfig{
		Issuer:            "issuer",
		AccessPrivateKey:  aPvt,
		AccessPublicKey:   aPub,
		RefreshPrivateKey: rPvt,
		RefreshPublicKey:  rPub,
		AccessTTL:         time.Minute,
		RefreshTTL:        time.Minute,
	}); err == nil {
		t.Fatalf("expected error when access private key conversion fails")
	}

	// Succeed access, fail access public key
	newV4AsymmetricSecretKeyFromBytes = origSec
	newV4AsymmetricPublicKeyFromBytes = func(b []byte) (psto.V4AsymmetricPublicKey, error) {
		return psto.V4AsymmetricPublicKey{}, errors.New("fail")
	}
	if _, err := NewPublicPaseto(PublicPasetoConfig{
		Issuer:            "issuer",
		AccessPrivateKey:  aPvt,
		AccessPublicKey:   aPub,
		RefreshPrivateKey: rPvt,
		RefreshPublicKey:  rPub,
		AccessTTL:         time.Minute,
		RefreshTTL:        time.Minute,
	}); err == nil {
		t.Fatalf("expected error when access public key conversion fails")
	}

	// Succeed access pair, fail refresh private
	newV4AsymmetricPublicKeyFromBytes = origPub
	calls := 0
	newV4AsymmetricSecretKeyFromBytes = func(b []byte) (psto.V4AsymmetricSecretKey, error) {
		calls++
		if calls == 1 {
			return origSec(b)
		}
		return psto.V4AsymmetricSecretKey{}, errors.New("fail")
	}
	if _, err := NewPublicPaseto(PublicPasetoConfig{
		Issuer:            "issuer",
		AccessPrivateKey:  aPvt,
		AccessPublicKey:   aPub,
		RefreshPrivateKey: rPvt,
		RefreshPublicKey:  rPub,
		AccessTTL:         time.Minute,
		RefreshTTL:        time.Minute,
	}); err == nil {
		t.Fatalf("expected error when refresh private key conversion fails")
	}

	// Succeed refresh private, fail refresh public (need to succeed on first public key call for access, fail on second for refresh)
	newV4AsymmetricSecretKeyFromBytes = origSec
	pubCalls := 0
	newV4AsymmetricPublicKeyFromBytes = func(b []byte) (psto.V4AsymmetricPublicKey, error) {
		pubCalls++
		if pubCalls == 1 {
			// First call is access public key, succeed
			return origPub(b)
		}
		// Second call is refresh public key, fail
		return psto.V4AsymmetricPublicKey{}, errors.New("fail")
	}
	if _, err := NewPublicPaseto(PublicPasetoConfig{
		Issuer:            "issuer",
		AccessPrivateKey:  aPvt,
		AccessPublicKey:   aPub,
		RefreshPrivateKey: rPvt,
		RefreshPublicKey:  rPub,
		AccessTTL:         time.Minute,
		RefreshTTL:        time.Minute,
	}); err == nil {
		t.Fatalf("expected error when refresh public key conversion fails")
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

func TestValidatePublicTokenAllBranches(t *testing.T) {
	now := time.Now().UTC()
	// Success
	tok := psto.NewToken()
	tok.Set("typ", string(PublicAccessToken))
	tok.SetExpiration(now.Add(time.Minute))
	if err := validatePublicToken(&tok, PublicAccessToken); err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	// Wrong type
	tok2 := psto.NewToken()
	tok2.Set("typ", string(PublicRefreshToken))
	tok2.SetExpiration(now.Add(time.Minute))
	if err := validatePublicToken(&tok2, PublicAccessToken); err == nil {
		t.Fatalf("expected type error")
	}

	// Expired
	tok3 := psto.NewToken()
	tok3.Set("typ", string(PublicAccessToken))
	tok3.SetExpiration(now.Add(-time.Minute))
	if err := validatePublicToken(&tok3, PublicAccessToken); err == nil {
		t.Fatalf("expected expiry error")
	}
}

func TestPublicParse_AllPathsViaStub(t *testing.T) {
	aPvt, aPub, rPvt, rPub := testPublicKeys()
	p, err := NewPublicPaseto(PublicPasetoConfig{Issuer: "iss", AccessPrivateKey: aPvt, AccessPublicKey: aPub, RefreshPrivateKey: rPvt, RefreshPublicKey: rPub, AccessTTL: time.Minute, RefreshTTL: time.Minute})
	if err != nil {
		t.Fatalf("NewPublicPaseto error: %v", err)
	}

	// success via stub
	orig := parseV4PublicFunc
	defer func() { parseV4PublicFunc = orig }()
	goodTok := psto.NewToken()
	goodTok.Set("typ", string(PublicAccessToken))
	goodTok.SetExpiration(time.Now().UTC().Add(time.Minute))
	parseV4PublicFunc = func(key psto.V4AsymmetricPublicKey, token string) (*psto.Token, error) { return &goodTok, nil }
	if _, err := p.parse("dummy", PublicAccessToken, p.accessPublicKey); err != nil {
		t.Fatalf("expected success via stub: %v", err)
	}

	// parser error
	parseV4PublicFunc = func(key psto.V4AsymmetricPublicKey, token string) (*psto.Token, error) {
		return nil, errors.New("parse fail")
	}
	if _, err := p.parse("dummy", PublicAccessToken, p.accessPublicKey); err == nil {
		t.Fatalf("expected parse error via stub")
	}

	// success parse but validation error (wrong type)
	wrongTypeTok := psto.NewToken()
	wrongTypeTok.Set("typ", string(PublicRefreshToken)) // expect access, got refresh
	wrongTypeTok.SetExpiration(time.Now().UTC().Add(time.Minute))
	parseV4PublicFunc = func(key psto.V4AsymmetricPublicKey, token string) (*psto.Token, error) { return &wrongTypeTok, nil }
	if _, err := p.parse("dummy", PublicAccessToken, p.accessPublicKey); err == nil {
		t.Fatalf("expected validation error via stub (wrong type)")
	}

	// success parse but validation error (expired)
	expiredTok := psto.NewToken()
	expiredTok.Set("typ", string(PublicAccessToken))
	expiredTok.SetExpiration(time.Now().UTC().Add(-time.Minute))
	parseV4PublicFunc = func(key psto.V4AsymmetricPublicKey, token string) (*psto.Token, error) { return &expiredTok, nil }
	if _, err := p.parse("dummy", PublicAccessToken, p.accessPublicKey); err == nil {
		t.Fatalf("expected validation error via stub (expired)")
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

	customClaims := map[string]any{
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

func TestPublicPasetoGenerateSignErrors(t *testing.T) {
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

	orig := pasetoV4Sign
	defer func() { pasetoV4Sign = orig }()

	// Fail access sign
	pasetoV4Sign = func(tok psto.Token, key psto.V4AsymmetricSecretKey) (string, error) {
		return "", errors.New("sign fail")
	}
	if _, err := p.Generate("s", "a", nil); err == nil {
		t.Fatalf("expected Generate to fail when access sign fails")
	}

	// Fail refresh sign only
	call := 0
	pasetoV4Sign = func(tok psto.Token, key psto.V4AsymmetricSecretKey) (string, error) {
		call++
		if call == 1 {
			return orig(tok, key)
		}
		return "", errors.New("sign fail")
	}
	if _, err := p.Generate("s", "a", nil); err == nil {
		t.Fatalf("expected Generate to fail when refresh sign fails")
	}
}
