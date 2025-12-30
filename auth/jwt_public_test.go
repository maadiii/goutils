package auth

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

func TestNewJWTPublicRejectsMissingKeys(t *testing.T) {
	aPriv, aPub, err := GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("GenerateRSAKeyPair error: %v", err)
	}

	// Missing access keys
	if _, err := NewJWTPublic(JWTPublicConfig{
		Issuer:            "iss",
		RefreshPrivateKey: aPriv,
		RefreshPublicKey:  aPub,
		AccessTTL:         time.Minute,
		RefreshTTL:        time.Minute,
	}); err == nil {
		t.Fatalf("expected error for missing access keys")
	}

	// Missing refresh keys
	if _, err := NewJWTPublic(JWTPublicConfig{
		Issuer:           "iss",
		AccessPrivateKey: aPriv,
		AccessPublicKey:  aPub,
		AccessTTL:        time.Minute,
		RefreshTTL:       time.Minute,
	}); err == nil {
		t.Fatalf("expected error for missing refresh keys")
	}
}

func TestJWTPublicRejectsUnexpectedSigningMethod(t *testing.T) {
	// Build HMAC-signed token with JWTLocal
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
	ltoks, err := loc.Generate("sub", "aud", nil)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	// Setup JWTPublic to try validating HMAC token
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

	if _, err := pub.ValidateAccess(ltoks.Access); err == nil {
		t.Fatalf("expected ValidateAccess to fail due to unexpected signing method")
	}
}

func TestJWTPublicGenerateSignerErrors(t *testing.T) {
	aPriv, aPub, err := GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("GenerateRSAKeyPair error: %v", err)
	}
	rPriv, rPub, err := GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("GenerateRSAKeyPair error: %v", err)
	}

	j, err := NewJWTPublic(JWTPublicConfig{
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

	orig := jwtRS256Sign
	defer func() { jwtRS256Sign = orig }()

	// Fail access sign
	jwtRS256Sign = func(token *jwt.Token, key *rsa.PrivateKey) (string, error) {
		return "", errors.New("sign fail")
	}
	if _, err := j.Generate("sub", "aud", nil); err == nil {
		t.Fatalf("expected Generate to fail when access signing fails")
	}

	// Fail refresh sign only
	call := 0
	jwtRS256Sign = func(token *jwt.Token, key *rsa.PrivateKey) (string, error) {
		call++
		if call == 1 {
			return orig(token, key)
		}
		return "", errors.New("sign fail")
	}
	if _, err := j.Generate("sub", "aud", nil); err == nil {
		t.Fatalf("expected Generate to fail when refresh signing fails")
	}
}

func TestJWTPublicParsesTimesFromClaims(t *testing.T) {
	aPriv, aPub, err := GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("GenerateRSAKeyPair error: %v", err)
	}
	rPriv, rPub, err := GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("GenerateRSAKeyPair error: %v", err)
	}
	j, err := NewJWTPublic(JWTPublicConfig{
		Issuer:            "issuer",
		AccessPrivateKey:  aPriv,
		AccessPublicKey:   aPub,
		RefreshPrivateKey: rPriv,
		RefreshPublicKey:  rPub,
		AccessTTL:         time.Second,
		RefreshTTL:        2 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewJWTPublic error: %v", err)
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

func TestPublicNumericDateHelperCoversTypes(t *testing.T) {
	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"exp": jwt.NewNumericDate(now),
		"iat": json.Number("123"),
		"nbf": float64(456),
		"int64": int64(789),
		"int": 7,
		"bad": "noop",
	}
	if got := publicNumericDateFromClaims(claims, "exp"); got.IsZero() {
		t.Fatalf("expected exp from *NumericDate")
	}
	if got := publicNumericDateFromClaims(claims, "iat"); got.Unix() != 123 {
		t.Fatalf("expected iat=123, got %v", got.Unix())
	}
	if got := publicNumericDateFromClaims(claims, "nbf"); got.Unix() != 456 {
		t.Fatalf("expected nbf=456, got %v", got.Unix())
	}
	if got := publicNumericDateFromClaims(claims, "int64"); got.Unix() != 789 {
		t.Fatalf("expected int64=789, got %v", got.Unix())
	}
	if got := publicNumericDateFromClaims(claims, "int"); got.Unix() != 7 {
		t.Fatalf("expected int=7, got %v", got.Unix())
	}
	if got := publicNumericDateFromClaims(claims, "bad"); !got.IsZero() {
		t.Fatalf("expected zero time for unsupported type, got %v", got)
	}
	if got := publicNumericDateFromClaims(claims, "missing"); !got.IsZero() {
		t.Fatalf("expected zero time for missing key, got %v", got)
	}
}

func TestJWTPublicInvalidClaimsBranch(t *testing.T) {
	aPriv, aPub, err := GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("GenerateRSAKeyPair error: %v", err)
	}
	rPriv, rPub, err := GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("GenerateRSAKeyPair error: %v", err)
	}
	j, err := NewJWTPublic(JWTPublicConfig{Issuer: "iss", AccessPrivateKey: aPriv, AccessPublicKey: aPub, RefreshPrivateKey: rPriv, RefreshPublicKey: rPub, AccessTTL: time.Minute, RefreshTTL: time.Minute})
	if err != nil {
		t.Fatalf("NewJWTPublic error: %v", err)
	}
	toks, err := j.Generate("s", "a", nil)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	orig := jwtPublicParseWithClaims
	defer func() { jwtPublicParseWithClaims = orig }()
	jwtPublicParseWithClaims = func(tokenString string, claims jwt.Claims, keyFunc jwt.Keyfunc, options ...jwt.ParserOption) (*jwt.Token, error) {
		return &jwt.Token{Valid: false, Claims: jwt.MapClaims{}}, nil
	}
	if _, err := j.ValidateAccess(toks.Access); err == nil {
		t.Fatalf("expected error on invalid claims branch")
	}
}

func TestJWTPublicParseWrongTypeDirect(t *testing.T) {
	aPriv, aPub, err := GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("GenerateRSAKeyPair error: %v", err)
	}
	rPriv, rPub, err := GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("GenerateRSAKeyPair error: %v", err)
	}
	j, err := NewJWTPublic(JWTPublicConfig{Issuer: "iss", AccessPrivateKey: aPriv, AccessPublicKey: aPub, RefreshPrivateKey: rPriv, RefreshPublicKey: rPub, AccessTTL: time.Minute, RefreshTTL: time.Minute})
	if err != nil {
		t.Fatalf("NewJWTPublic error: %v", err)
	}
	toks, err := j.Generate("s", "a", nil)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	// Use parser stub to return valid token but with wrong typ claim
	orig := jwtPublicParseWithClaims
	defer func() { jwtPublicParseWithClaims = orig }()
	jwtPublicParseWithClaims = func(tokenString string, claims jwt.Claims, keyFunc jwt.Keyfunc, options ...jwt.ParserOption) (*jwt.Token, error) {
		// Return valid token with wrong typ
		wrongClaims := jwt.MapClaims{
			"sub": "s",
			"aud": "a",
			"iss": "iss",
			"iat": float64(time.Now().Unix()),
			"exp": float64(time.Now().Add(time.Minute).Unix()),
			"nbf": float64(time.Now().Unix()),
			"typ": string(JWTPublicRefreshToken), // wrong type for access validation
		}
		return &jwt.Token{Valid: true, Claims: wrongClaims, Method: jwt.SigningMethodRS256}, nil
	}
	if _, err := j.ValidateAccess(toks.Access); err == nil {
		t.Fatalf("expected error on wrong type in parse")
	}
}

func TestGenerateRSAKeyPairErrorPath(t *testing.T) {
	orig := rsaGenerateKey
	defer func() { rsaGenerateKey = orig }()
	rsaGenerateKey = func(reader io.Reader, bits int) (*rsa.PrivateKey, error) {
		return nil, errors.New("gen fail")
	}
	if _, _, err := GenerateRSAKeyPair(); err == nil {
		t.Fatalf("expected error from GenerateRSAKeyPair when rsaGenerateKey fails")
	}
}
