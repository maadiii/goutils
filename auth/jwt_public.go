package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTPublicTokenType distinguishes access vs refresh tokens.
type JWTPublicTokenType string

const (
	JWTPublicAccessToken  JWTPublicTokenType = "access"
	JWTPublicRefreshToken JWTPublicTokenType = "refresh"
)

// JWTPublicClaims captures the essential claims we expose to callers.
type JWTPublicClaims struct {
	Subject      string
	Audience     string
	Issuer       string
	ExpiresAt    time.Time
	IssuedAt     time.Time
	NotBefore    time.Time
	TokenType    JWTPublicTokenType
	CustomClaims map[string]interface{}
}

// JWTPublicConfig holds configuration for JWT public (asymmetric) token generation using RSA.
type JWTPublicConfig struct {
	Issuer             string
	AccessPrivateKey   *rsa.PrivateKey
	AccessPublicKey    *rsa.PublicKey
	RefreshPrivateKey  *rsa.PrivateKey
	RefreshPublicKey   *rsa.PublicKey
	AccessTTL          time.Duration
	RefreshTTL         time.Duration
}

// JWTPublic generates and validates JWT tokens with RSA signatures for access/refresh.
type JWTPublic struct {
	issuer             string
	accessPrivateKey   *rsa.PrivateKey
	accessPublicKey    *rsa.PublicKey
	refreshPrivateKey  *rsa.PrivateKey
	refreshPublicKey   *rsa.PublicKey
	accessTTL          time.Duration
	refreshTTL         time.Duration
}

// NewJWTPublic builds a JWTPublic with distinct RSA key pairs for access and refresh tokens.
func NewJWTPublic(cfg JWTPublicConfig) (*JWTPublic, error) {
	if cfg.AccessPrivateKey == nil || cfg.AccessPublicKey == nil {
		return nil, errors.New("jwt: access keys required")
	}
	if cfg.RefreshPrivateKey == nil || cfg.RefreshPublicKey == nil {
		return nil, errors.New("jwt: refresh keys required")
	}

	return &JWTPublic{
		issuer:             cfg.Issuer,
		accessPrivateKey:   cfg.AccessPrivateKey,
		accessPublicKey:    cfg.AccessPublicKey,
		refreshPrivateKey:  cfg.RefreshPrivateKey,
		refreshPublicKey:   cfg.RefreshPublicKey,
		accessTTL:          cfg.AccessTTL,
		refreshTTL:         cfg.RefreshTTL,
	}, nil
}

// GenerateRSAKeyPair generates a new RSA 2048-bit key pair.
func GenerateRSAKeyPair() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	return privateKey, &privateKey.PublicKey, nil
}

// Generate issues a new pair of access and refresh JWT tokens for the given subject and audience.
// customClaims is an optional map of additional claims to include in both tokens.
func (j *JWTPublic) Generate(subject, audience string, customClaims map[string]interface{}) (Tokens, error) {
	access, err := j.makeToken(subject, audience, JWTPublicAccessToken, j.accessTTL, j.accessPrivateKey, customClaims)
	if err != nil {
		return Tokens{}, err
	}
	refresh, err := j.makeToken(subject, audience, JWTPublicRefreshToken, j.refreshTTL, j.refreshPrivateKey, customClaims)
	if err != nil {
		return Tokens{}, err
	}
	return Tokens{Access: access, Refresh: refresh}, nil
}

// ValidateAccess parses and validates an access token, returning its claims.
func (j *JWTPublic) ValidateAccess(token string) (JWTPublicClaims, error) {
	return j.parse(token, JWTPublicAccessToken, j.accessPublicKey)
}

// ValidateRefresh parses and validates a refresh token, returning its claims.
func (j *JWTPublic) ValidateRefresh(token string) (JWTPublicClaims, error) {
	return j.parse(token, JWTPublicRefreshToken, j.refreshPublicKey)
}

func (j *JWTPublic) makeToken(subject, audience string, typ JWTPublicTokenType, ttl time.Duration, privateKey *rsa.PrivateKey, customClaims map[string]interface{}) (string, error) {
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
		"jti": newJTI(),
	}

	// Add custom claims
	for k, v := range customClaims {
		claims[k] = v
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(privateKey)
	if err != nil {
		return "", err
	}

	return signed, nil
}

func (j *JWTPublic) parse(tokenString string, expected JWTPublicTokenType, publicKey *rsa.PublicKey) (JWTPublicClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("jwt: unexpected signing method")
		}
		return publicKey, nil
	})

	if err != nil {
		return JWTPublicClaims{}, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return JWTPublicClaims{}, errors.New("jwt: invalid claims")
	}

	// Extract standard claims
	sub, _ := claims["sub"].(string)
	aud, _ := claims["aud"].(string)
	iss, _ := claims["iss"].(string)
	typ, _ := claims["typ"].(string)

	var exp, iat, nbf time.Time
	if expNum, ok := claims["exp"].(*jwt.NumericDate); ok && expNum != nil {
		exp = expNum.Time
	}
	if iatNum, ok := claims["iat"].(*jwt.NumericDate); ok && iatNum != nil {
		iat = iatNum.Time
	}
	if nbfNum, ok := claims["nbf"].(*jwt.NumericDate); ok && nbfNum != nil {
		nbf = nbfNum.Time
	}

	if typ != string(expected) {
		return JWTPublicClaims{}, errors.New("jwt: unexpected token type")
	}

	// Extract custom claims (filter out standard claims)
	customClaims := make(map[string]interface{})
	reservedKeys := map[string]bool{
		"sub": true, "aud": true, "iss": true, "exp": true,
		"iat": true, "nbf": true, "typ": true, "jti": true,
	}

	for key, val := range claims {
		if !reservedKeys[key] {
			customClaims[key] = val
		}
	}

	return JWTPublicClaims{
		Subject:      sub,
		Audience:     aud,
		Issuer:       iss,
		ExpiresAt:    exp,
		IssuedAt:     iat,
		NotBefore:    nbf,
		TokenType:    JWTPublicTokenType(typ),
		CustomClaims: customClaims,
	}, nil
}
