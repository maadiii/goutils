package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	psto "aidanwoods.dev/go-paseto"
)

// LocalTokenType distinguishes access vs refresh tokens.
type LocalTokenType string

const (
	AccessToken  LocalTokenType = "access"
	RefreshToken LocalTokenType = "refresh"
)

// LocalClaims captures the essential claims we expose to callers.
type LocalClaims struct {
	Subject      string
	Audience     string
	Issuer       string
	JTI          string
	ExpiresAt    time.Time
	IssuedAt     time.Time
	NotBefore    time.Time
	TokenType    LocalTokenType
	CustomClaims map[string]interface{}
}

// LocalPasetoConfig holds configuration for PASETO v4 local token generation.
type LocalPasetoConfig struct {
	Issuer     string
	AccessKey  []byte
	RefreshKey []byte
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

// LocalPaseto generates and validates PASETO v4 local tokens for access/refresh.
type LocalPaseto struct {
	issuer     string
	accessKey  psto.V4SymmetricKey
	refreshKey psto.V4SymmetricKey
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewLocalPaseto builds a LocalPaseto with distinct keys and TTLs for access and refresh tokens.
// Both keys must be 32 bytes for PASETO v4 local.
func NewLocalPaseto(cfg LocalPasetoConfig) (*LocalPaseto, error) {
	if len(cfg.AccessKey) != 32 || len(cfg.RefreshKey) != 32 {
		return nil, errors.New("paseto keys must be 32 bytes for v4 local")
	}

	ak, err := psto.V4SymmetricKeyFromBytes(cfg.AccessKey)
	if err != nil {
		return nil, err
	}
	rk, err := psto.V4SymmetricKeyFromBytes(cfg.RefreshKey)
	if err != nil {
		return nil, err
	}

	return &LocalPaseto{
		issuer:     cfg.Issuer,
		accessKey:  ak,
		refreshKey: rk,
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
	}, nil
}

// Tokens bundles generated access and refresh tokens.
type Tokens struct {
	Access  string
	Refresh string
}

// Generate issues a new pair of access and refresh tokens for the given subject and audience.
// customClaims is an optional map of additional claims to include in both tokens.
func (p *LocalPaseto) Generate(subject, audience string, customClaims map[string]interface{}) (Tokens, error) {
	access, err := p.makeToken(subject, audience, AccessToken, p.accessTTL, p.accessKey, customClaims)
	if err != nil {
		return Tokens{}, err
	}
	refresh, err := p.makeToken(subject, audience, RefreshToken, p.refreshTTL, p.refreshKey, customClaims)
	if err != nil {
		return Tokens{}, err
	}
	return Tokens{Access: access, Refresh: refresh}, nil
}

// ValidateAccess parses and validates an access token, returning its claims.
func (p *LocalPaseto) ValidateAccess(token string) (LocalClaims, error) {
	return p.parse(token, AccessToken, p.accessKey)
}

// ValidateRefresh parses and validates a refresh token, returning its claims.
func (p *LocalPaseto) ValidateRefresh(token string) (LocalClaims, error) {
	return p.parse(token, RefreshToken, p.refreshKey)
}

func (p *LocalPaseto) makeToken(subject, audience string, typ LocalTokenType, ttl time.Duration, key psto.V4SymmetricKey, customClaims map[string]interface{}) (string, error) {
	now := time.Now().UTC()
	exp := now.Add(ttl)

	tok := psto.NewToken()
	tok.SetSubject(subject)
	tok.SetAudience(audience)
	tok.SetIssuer(p.issuer)
	tok.SetIssuedAt(now)
	tok.SetNotBefore(now)
	tok.SetExpiration(exp)
	tok.SetJti(newJTI())
	_ = tok.Set("typ", string(typ))

	// Add custom claims
	for key, value := range customClaims {
		_ = tok.Set(key, value)
	}

	return tok.V4Encrypt(key, nil), nil
}

func (p *LocalPaseto) parse(token string, expected LocalTokenType, key psto.V4SymmetricKey) (LocalClaims, error) {
	parser := psto.NewParser()
	value, err := parser.ParseV4Local(key, token, nil)
	if err != nil {
		return LocalClaims{}, err
	}

	typ := ""
	_ = value.Get("typ", &typ)
	if typ != string(expected) {
		return LocalClaims{}, errors.New("paseto: unexpected token type")
	}

	exp, _ := value.GetExpiration()
	if time.Now().UTC().After(exp) {
		return LocalClaims{}, errors.New("paseto: token expired")
	}

	sub, _ := value.GetSubject()
	aud, _ := value.GetAudience()
	iss, _ := value.GetIssuer()
	jit, _ := value.GetJti()
	issuedAt, _ := value.GetIssuedAt()
	nbf, _ := value.GetNotBefore()

	// Extract custom claims
	allClaims := value.Claims()
	customClaims := make(map[string]interface{})
	reservedKeys := map[string]bool{
		"sub": true, "aud": true, "iss": true, "jti": true,
		"exp": true, "iat": true, "nbf": true, "typ": true,
	}

	for key, val := range allClaims {
		if !reservedKeys[key] {
			customClaims[key] = val
		}
	}

	return LocalClaims{
		Subject:      sub,
		Audience:     aud,
		Issuer:       iss,
		JTI:          jit,
		ExpiresAt:    exp,
		IssuedAt:     issuedAt,
		NotBefore:    nbf,
		TokenType:    expected,
		CustomClaims: customClaims,
	}, nil
}

func newJTI() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}
