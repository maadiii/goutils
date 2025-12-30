package auth

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

func TestNewJWTLocalRejectsEmptyKeys(t *testing.T) {
	// Empty access key
	if _, err := NewJWTLocal(JWTLocalConfig{
		Issuer:     "iss",
		AccessKey:  nil,
		RefreshKey: []byte{1, 2, 3},
		AccessTTL:  time.Minute,
		RefreshTTL: time.Minute,
	}); err == nil {
		t.Fatalf("expected error for empty access key")
	}

	// Empty refresh key
	if _, err := NewJWTLocal(JWTLocalConfig{
		Issuer:     "iss",
		AccessKey:  []byte{1, 2, 3},
		RefreshKey: nil,
		AccessTTL:  time.Minute,
		RefreshTTL: time.Minute,
	}); err == nil {
		t.Fatalf("expected error for empty refresh key")
	}
}

func TestJWTLocalRejectsUnexpectedSigningMethod(t *testing.T) {
	// Create a JWTPublic instance to generate an RSA-signed token
	aPriv, aPub, err := GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("GenerateRSAKeyPair error: %v", err)
	}
	rPriv, rPub, err := GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("GenerateRSAKeyPair error: %v", err)
	}
	pub, err := NewJWTPublic(JWTPublicConfig{
		Issuer:            "iss",
		AccessPrivateKey:  aPriv,
		AccessPublicKey:   aPub,
		RefreshPrivateKey: rPriv,
		RefreshPublicKey:  rPub,
		AccessTTL:         time.Minute,
		RefreshTTL:        time.Minute,
	})
	if err != nil {
		t.Fatalf("NewJWTPublic error: %v", err)
	}

	toks, err := pub.Generate("sub", "aud", nil)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	// Now validate the RSA token with HMAC validator -> should fail signing method check
	ak, rk := testHMACKeys()
	loc, err := NewJWTLocal(JWTLocalConfig{
		Issuer:     "iss",
		AccessKey:  ak,
		RefreshKey: rk,
		AccessTTL:  time.Minute,
		RefreshTTL: time.Minute,
	})
	if err != nil {
		t.Fatalf("NewJWTLocal error: %v", err)
	}

	if _, err := loc.ValidateAccess(toks.Access); err == nil {
		t.Fatalf("expected ValidateAccess to fail due to unexpected signing method")
	}
}

func TestJWTLocalGenerateSignerErrors(t *testing.T) {
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

	// Stub signer to fail for access creation
	orig := jwtHS256Sign
	defer func() { jwtHS256Sign = orig }()
	jwtHS256Sign = func(token *jwt.Token, key []byte) (string, error) {
		return "", errors.New("signing failed")
	}
	if _, err := j.Generate("sub", "aud", nil); err == nil {
		t.Fatalf("expected Generate to fail when access token signing fails")
	}

	// Make signer fail only on refresh by restoring then stubbing during second call
	call := 0
	jwtHS256Sign = func(token *jwt.Token, key []byte) (string, error) {
		call++
		if call == 1 {
			// succeed for access
			return orig(token, key)
		}
		return "", errors.New("refresh signing failed")
	}
	if _, err := j.Generate("sub", "aud", nil); err == nil {
		t.Fatalf("expected Generate to fail when refresh token signing fails")
	}
}

func TestJWTLocalParsesTimesFromClaims(t *testing.T) {
	ak, rk := testHMACKeys()
	j, err := NewJWTLocal(JWTLocalConfig{
		Issuer:     "issuer",
		AccessKey:  ak,
		RefreshKey: rk,
		AccessTTL:  time.Second,
		RefreshTTL: 2 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewJWTLocal error: %v", err)
	}
	toks, err := j.Generate("s", "a", nil)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	c, err := j.ValidateAccess(toks.Access)
	if err != nil {
		t.Fatalf("ValidateAccess error: %v", err)
	}
	if c.ExpiresAt.IsZero() || c.IssuedAt.IsZero() || c.NotBefore.IsZero() {
		t.Fatalf("expected non-zero time claims: %+v", c)
	}
}

func TestNumericDateHelperCoversTypes(t *testing.T) {
	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"exp": jwt.NewNumericDate(now),
		"iat": json.Number("123"),
		"nbf": float64(456),
		"int64": int64(789),
		"int": 7,
		"bad": "noop",
	}
	if got := numericDateFromClaims(claims, "exp"); got.IsZero() {
		t.Fatalf("expected exp from *NumericDate")
	}
	if got := numericDateFromClaims(claims, "iat"); got.Unix() != 123 {
		t.Fatalf("expected iat=123, got %v", got.Unix())
	}
	if got := numericDateFromClaims(claims, "nbf"); got.Unix() != 456 {
		t.Fatalf("expected nbf=456, got %v", got.Unix())
	}
	if got := numericDateFromClaims(claims, "int64"); got.Unix() != 789 {
		t.Fatalf("expected int64=789, got %v", got.Unix())
	}
	if got := numericDateFromClaims(claims, "int"); got.Unix() != 7 {
		t.Fatalf("expected int=7, got %v", got.Unix())
	}
	if got := numericDateFromClaims(claims, "bad"); !got.IsZero() {
		t.Fatalf("expected zero time for unsupported type, got %v", got)
	}
	if got := numericDateFromClaims(claims, "missing"); !got.IsZero() {
		t.Fatalf("expected zero time for missing key, got %v", got)
	}
}

func TestJWTLocalInvalidClaimsBranch(t *testing.T) {
	ak, rk := testHMACKeys()
	j, err := NewJWTLocal(JWTLocalConfig{Issuer: "iss", AccessKey: ak, RefreshKey: rk, AccessTTL: time.Minute, RefreshTTL: time.Minute})
	if err != nil {
		t.Fatalf("NewJWTLocal error: %v", err)
	}
	toks, err := j.Generate("s", "a", nil)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	orig := jwtParseWithClaims
	defer func() { jwtParseWithClaims = orig }()
	jwtParseWithClaims = func(tokenString string, claims jwt.Claims, keyFunc jwt.Keyfunc, options ...jwt.ParserOption) (*jwt.Token, error) {
		return &jwt.Token{Valid: false, Claims: jwt.MapClaims{}}, nil
	}
	if _, err := j.ValidateAccess(toks.Access); err == nil {
		t.Fatalf("expected error on invalid claims branch")
	}
}

func TestJWTLocalParseWrongTypeDirect(t *testing.T) {
	ak, rk := testHMACKeys()
	j, err := NewJWTLocal(JWTLocalConfig{Issuer: "iss", AccessKey: ak, RefreshKey: rk, AccessTTL: time.Minute, RefreshTTL: time.Minute})
	if err != nil {
		t.Fatalf("NewJWTLocal error: %v", err)
	}
	toks, err := j.Generate("s", "a", nil)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	// Use parser stub to return valid token but with wrong typ claim
	orig := jwtParseWithClaims
	defer func() { jwtParseWithClaims = orig }()
	jwtParseWithClaims = func(tokenString string, claims jwt.Claims, keyFunc jwt.Keyfunc, options ...jwt.ParserOption) (*jwt.Token, error) {
		// Return valid token with wrong typ
		wrongClaims := jwt.MapClaims{
			"sub": "s",
			"aud": "a",
			"iss": "iss",
			"iat": float64(time.Now().Unix()),
			"exp": float64(time.Now().Add(time.Minute).Unix()),
			"nbf": float64(time.Now().Unix()),
			"typ": string(JWTRefreshToken), // wrong type for access validation
		}
		return &jwt.Token{Valid: true, Claims: wrongClaims, Method: jwt.SigningMethodHS256}, nil
	}
	if _, err := j.ValidateAccess(toks.Access); err == nil {
		t.Fatalf("expected error on wrong type in parse")
	}
}
