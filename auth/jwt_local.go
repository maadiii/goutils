package auth

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// test hook for signing HS256 tokens
var jwtHS256Sign = func(token *jwt.Token, key []byte) (string, error) {
	return token.SignedString(key)
}

// test hook for parsing tokens
var jwtParseWithClaims = func(tokenString string, claims jwt.Claims, keyFunc jwt.Keyfunc, options ...jwt.ParserOption) (*jwt.Token, error) {
	return jwt.ParseWithClaims(tokenString, claims, keyFunc, options...)
}

// JWTLocalTokenType distinguishes access vs refresh tokens.
type JWTLocalTokenType string

const (
	JWTAccessToken  JWTLocalTokenType = "access"
	JWTRefreshToken JWTLocalTokenType = "refresh"
)

// JWTLocalClaims captures the essential claims we expose to callers.
type JWTLocalClaims struct {
	Subject      string
	Audience     string
	Issuer       string
	ExpiresAt    time.Time
	IssuedAt     time.Time
	NotBefore    time.Time
	TokenType    JWTLocalTokenType
	CustomClaims map[string]interface{}
}

// JWTLocalConfig holds configuration for JWT local token generation using HMAC.
type JWTLocalConfig struct {
	Issuer     string
	AccessKey  []byte
	RefreshKey []byte
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

// JWTLocal generates and validates JWT tokens with HMAC signatures for access/refresh.
type JWTLocal struct {
	issuer     string
	accessKey  []byte
	refreshKey []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewJWTLocal builds a JWTLocal with distinct keys and TTLs for access and refresh tokens.
// Keys can be any length but should be at least 32 bytes for HS256, 64 bytes for HS512.
func NewJWTLocal(cfg JWTLocalConfig) (*JWTLocal, error) {
	if len(cfg.AccessKey) == 0 || len(cfg.RefreshKey) == 0 {
		return nil, errors.New("jwt keys cannot be empty")
	}

	return &JWTLocal{
		issuer:     cfg.Issuer,
		accessKey:  cfg.AccessKey,
		refreshKey: cfg.RefreshKey,
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
	}, nil
}

// Generate issues a new pair of access and refresh JWT tokens for the given subject and audience.
// customClaims is an optional map of additional claims to include in both tokens.
func (j *JWTLocal) Generate(subject, audience string, customClaims map[string]any) (Tokens, error) {
	access, err := j.makeToken(subject, audience, JWTAccessToken, j.accessTTL, j.accessKey, customClaims)
	if err != nil {
		return Tokens{}, err
	}
	refresh, err := j.makeToken(subject, audience, JWTRefreshToken, j.refreshTTL, j.refreshKey, customClaims)
	if err != nil {
		return Tokens{}, err
	}
	return Tokens{Access: access, Refresh: refresh}, nil
}

// ValidateAccess parses and validates an access token, returning its claims.
func (j *JWTLocal) ValidateAccess(token string) (JWTLocalClaims, error) {
	return j.parse(token, JWTAccessToken, j.accessKey)
}

// ValidateRefresh parses and validates a refresh token, returning its claims.
func (j *JWTLocal) ValidateRefresh(token string) (JWTLocalClaims, error) {
	return j.parse(token, JWTRefreshToken, j.refreshKey)
}

func (j *JWTLocal) makeToken(subject, audience string, typ JWTLocalTokenType, ttl time.Duration, key []byte, customClaims map[string]interface{}) (string, error) {
	now := time.Now().UTC()
	exp := now.Add(ttl)

	// Build standard claims
	claims := jwt.MapClaims{
		"sub": subject,
		"aud": audience,
		"iss": j.issuer,
		"iat": now.Unix(),
		"exp": exp.Unix(),
		"nbf": now.Unix(),
		"typ": string(typ),
	}

	// Add custom claims
	for k, v := range customClaims {
		claims[k] = v
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := jwtHS256Sign(token, key)
	if err != nil {
		return "", err
	}

	return signed, nil
}

func (j *JWTLocal) parse(tokenString string, expected JWTLocalTokenType, key []byte) (JWTLocalClaims, error) {
	token, err := jwtParseWithClaims(tokenString, jwt.MapClaims{}, func(token *jwt.Token) (any, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("jwt: unexpected signing method")
		}
		return key, nil
	})
	if err != nil {
		return JWTLocalClaims{}, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return JWTLocalClaims{}, errors.New("jwt: invalid claims")
	}

	// Extract standard claims
	sub, _ := claims["sub"].(string)
	aud, _ := claims["aud"].(string)
	iss, _ := claims["iss"].(string)
	typ, _ := claims["typ"].(string)

	var exp, iat, nbf time.Time
	exp = numericDateFromClaims(claims, "exp")
	iat = numericDateFromClaims(claims, "iat")
	nbf = numericDateFromClaims(claims, "nbf")

	if typ != string(expected) {
		return JWTLocalClaims{}, errors.New("jwt: unexpected token type")
	}

	// Extract custom claims (filter out standard claims)
	customClaims := make(map[string]any)
	reservedKeys := map[string]bool{
		"sub": true, "aud": true, "iss": true, "exp": true,
		"iat": true, "nbf": true, "typ": true, "jti": true,
	}

	for key, val := range claims {
		if !reservedKeys[key] {
			customClaims[key] = val
		}
	}

	return JWTLocalClaims{
		Subject:      sub,
		Audience:     aud,
		Issuer:       iss,
		ExpiresAt:    exp,
		IssuedAt:     iat,
		NotBefore:    nbf,
		TokenType:    JWTLocalTokenType(typ),
		CustomClaims: customClaims,
	}, nil
}

// numericDateFromClaims extracts a time value from MapClaims supporting float64,
// json.Number, and *jwt.NumericDate values.
func numericDateFromClaims(claims jwt.MapClaims, key string) time.Time {
	v, ok := claims[key]
	if !ok || v == nil {
		return time.Time{}
	}
	switch tv := v.(type) {
	case *jwt.NumericDate:
		if tv != nil {
			return tv.Time.UTC()
		}
	case json.Number:
		if n, err := tv.Int64(); err == nil {
			return time.Unix(n, 0).UTC()
		}
	case float64:
		return time.Unix(int64(tv), 0).UTC()
	case int64:
		return time.Unix(tv, 0).UTC()
	case int:
		return time.Unix(int64(tv), 0).UTC()
	}
	return time.Time{}
}
