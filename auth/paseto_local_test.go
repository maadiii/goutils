package auth

import (
	"errors"
	"testing"
	"time"

	psto "aidanwoods.dev/go-paseto"
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
		AccessTTL:  100 * time.Millisecond,
		RefreshTTL: 2 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewLocalPaseto error: %v", err)
	}

	toks, err := p.Generate("subj", "aud", nil)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	time.Sleep(150 * time.Millisecond)
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

func TestNewLocalPasetoConstructorLibErrors(t *testing.T) {
	ak, rk := testKeys()
	// Ensure lengths are correct so we hit the FromBytes paths
	orig := v4SymmetricKeyFromBytes
	defer func() { v4SymmetricKeyFromBytes = orig }()
	// Make access key conversion fail
	v4SymmetricKeyFromBytes = func(b []byte) (psto.V4SymmetricKey, error) {
		return psto.V4SymmetricKey{}, errors.New("fail")
	}
	if _, err := NewLocalPaseto(LocalPasetoConfig{
		Issuer:     "issuer",
		AccessKey:  ak,
		RefreshKey: rk,
		AccessTTL:  time.Minute,
		RefreshTTL: time.Minute,
	}); err == nil {
		t.Fatalf("expected error when access key conversion fails")
	}

	// Succeed first, then fail for refresh key
	call := 0
	v4SymmetricKeyFromBytes = func(b []byte) (psto.V4SymmetricKey, error) {
		call++
		if call == 1 {
			return orig(b)
		}
		return psto.V4SymmetricKey{}, errors.New("fail")
	}
	if _, err := NewLocalPaseto(LocalPasetoConfig{
		Issuer:     "issuer",
		AccessKey:  ak,
		RefreshKey: rk,
		AccessTTL:  time.Minute,
		RefreshTTL: time.Minute,
	}); err == nil {
		t.Fatalf("expected error when refresh key conversion fails")
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

func TestLocalPasetoValidateWithWrongKeyFailsParse(t *testing.T) {
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

	// Different keys to force parse error (decryption fails)
	ak2 := make([]byte, 32)
	rk2 := make([]byte, 32)
	for i := 0; i < 32; i++ {
		ak2[i] = byte(200 + i)
		rk2[i] = byte(50 + i)
	}
	p2, err := NewLocalPaseto(LocalPasetoConfig{
		Issuer:     "issuer",
		AccessKey:  ak2,
		RefreshKey: rk2,
		AccessTTL:  time.Minute,
		RefreshTTL: time.Minute,
	})
	if err != nil {
		t.Fatalf("NewLocalPaseto error: %v", err)
	}
	if _, err := p2.ValidateAccess(toks.Access); err == nil {
		t.Fatalf("expected ValidateAccess to fail with wrong key")
	}
}

func TestValidateLocalTokenAllBranches(t *testing.T) {
	now := time.Now().UTC()
	// Success path
	tok := psto.NewToken()
	tok.Set("typ", string(AccessToken))
	tok.SetExpiration(now.Add(time.Minute))
	if err := validateLocalToken(&tok, AccessToken); err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	// Wrong type
	tok2 := psto.NewToken()
	tok2.Set("typ", string(RefreshToken))
	tok2.SetExpiration(now.Add(time.Minute))
	if err := validateLocalToken(&tok2, AccessToken); err == nil {
		t.Fatalf("expected type error")
	}

	// Expired
	tok3 := psto.NewToken()
	tok3.Set("typ", string(AccessToken))
	tok3.SetExpiration(now.Add(-time.Minute))
	if err := validateLocalToken(&tok3, AccessToken); err == nil {
		t.Fatalf("expected expiry error")
	}
}

func TestLocalParse_AllPathsViaStub(t *testing.T) {
	ak, rk := testKeys()
	p, err := NewLocalPaseto(LocalPasetoConfig{Issuer: "iss", AccessKey: ak, RefreshKey: rk, AccessTTL: time.Minute, RefreshTTL: time.Minute})
	if err != nil { t.Fatalf("NewLocalPaseto error: %v", err) }

	// success case stub
	orig := parseV4LocalFunc
	defer func() { parseV4LocalFunc = orig }()
	goodTok := psto.NewToken()
	goodTok.Set("typ", string(AccessToken))
	goodTok.SetExpiration(time.Now().UTC().Add(time.Minute))
	parseV4LocalFunc = func(key psto.V4SymmetricKey, token string) (*psto.Token, error) { return &goodTok, nil }
	if _, err := p.parse("dummy", AccessToken, p.accessKey); err != nil {
		t.Fatalf("expected success via stub: %v", err)
	}

	// error from parser
	parseV4LocalFunc = func(key psto.V4SymmetricKey, token string) (*psto.Token, error) { return nil, errors.New("parse fail") }
	if _, err := p.parse("dummy", AccessToken, p.accessKey); err == nil {
		t.Fatalf("expected parse error via stub")
	}

	// success parse but validation error (e.g. wrong type)
	wrongTypeTok := psto.NewToken()
	wrongTypeTok.Set("typ", string(RefreshToken))  // expect access, got refresh
	wrongTypeTok.SetExpiration(time.Now().UTC().Add(time.Minute))
	parseV4LocalFunc = func(key psto.V4SymmetricKey, token string) (*psto.Token, error) { return &wrongTypeTok, nil }
	if _, err := p.parse("dummy", AccessToken, p.accessKey); err == nil {
		t.Fatalf("expected validation error via stub (wrong type)")
	}

	// success parse but validation error (expired)
	expiredTok := psto.NewToken()
	expiredTok.Set("typ", string(AccessToken))
	expiredTok.SetExpiration(time.Now().UTC().Add(-time.Minute))
	parseV4LocalFunc = func(key psto.V4SymmetricKey, token string) (*psto.Token, error) { return &expiredTok, nil }
	if _, err := p.parse("dummy", AccessToken, p.accessKey); err == nil {
		t.Fatalf("expected validation error via stub (expired)")
	}
}

func TestLocalPasetoGenerateEncryptErrors(t *testing.T) {
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

	// Stub encrypt to fail for access
	orig := pasetoV4Encrypt
	defer func() { pasetoV4Encrypt = orig }()
	pasetoV4Encrypt = func(tok psto.Token, key psto.V4SymmetricKey) (string, error) {
		return "", errors.New("encrypt fail")
	}
	if _, err := p.Generate("s", "a", nil); err == nil {
		t.Fatalf("expected Generate to fail when access encryption fails")
	}

	// Succeed once, then fail for refresh
	call := 0
	pasetoV4Encrypt = func(tok psto.Token, key psto.V4SymmetricKey) (string, error) {
		call++
		if call == 1 {
			return orig(tok, key)
		}
		return "", errors.New("encrypt fail")
	}
	if _, err := p.Generate("s", "a", nil); err == nil {
		t.Fatalf("expected Generate to fail when refresh encryption fails")
	}
}
