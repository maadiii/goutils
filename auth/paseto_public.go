package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	psto "aidanwoods.dev/go-paseto"
)

// test hook for PASETO v4 public signing
var pasetoV4Sign = func(tok psto.Token, key psto.V4AsymmetricSecretKey) (string, error) {
	return tok.V4Sign(key, nil), nil
}

// PublicTokenType distinguishes access vs refresh tokens for public PASETO.
type PublicTokenType string

const (
	PublicAccessToken  PublicTokenType = "access"
	PublicRefreshToken PublicTokenType = "refresh"
)

// PublicClaims captures the essential claims for public PASETO tokens.
type PublicClaims struct {
	Subject      string
	Audience     string
	Issuer       string
	JTI          string
	ExpiresAt    time.Time
	IssuedAt     time.Time
	NotBefore    time.Time
	TokenType    PublicTokenType
	CustomClaims map[string]any
}

// PublicPasetoConfig holds configuration for PASETO v4 public token generation.
type PublicPasetoConfig struct {
	Issuer            string
	AccessPrivateKey  []byte
	AccessPublicKey   []byte
	RefreshPrivateKey []byte
	RefreshPublicKey  []byte
	AccessTTL         time.Duration
	RefreshTTL        time.Duration
}

// PublicPaseto generates and validates PASETO v4 public tokens for access/refresh.
type PublicPaseto struct {
	issuer            string
	accessPrivateKey  psto.V4AsymmetricSecretKey
	accessPublicKey   psto.V4AsymmetricPublicKey
	refreshPrivateKey psto.V4AsymmetricSecretKey
	refreshPublicKey  psto.V4AsymmetricPublicKey
	accessTTL         time.Duration
	refreshTTL        time.Duration
}

// NewPublicPaseto builds a PublicPaseto with distinct key pairs and TTLs for access and refresh tokens.
// Private keys must be 64 bytes and public keys must be 32 bytes for PASETO v4 public.
var (
	newV4AsymmetricSecretKeyFromBytes = psto.NewV4AsymmetricSecretKeyFromBytes
	newV4AsymmetricPublicKeyFromBytes = psto.NewV4AsymmetricPublicKeyFromBytes
)

func NewPublicPaseto(cfg PublicPasetoConfig) (*PublicPaseto, error) {
	if len(cfg.AccessPrivateKey) != 64 || len(cfg.RefreshPrivateKey) != 64 {
		return nil, errors.New("paseto private keys must be 64 bytes for v4 public")
	}
	if len(cfg.AccessPublicKey) != 32 || len(cfg.RefreshPublicKey) != 32 {
		return nil, errors.New("paseto public keys must be 32 bytes for v4 public")
	}

	aPvt, err := newV4AsymmetricSecretKeyFromBytes(cfg.AccessPrivateKey)
	if err != nil {
		return nil, err
	}
	aPub, err := newV4AsymmetricPublicKeyFromBytes(cfg.AccessPublicKey)
	if err != nil {
		return nil, err
	}

	rPvt, err := newV4AsymmetricSecretKeyFromBytes(cfg.RefreshPrivateKey)
	if err != nil {
		return nil, err
	}
	rPub, err := newV4AsymmetricPublicKeyFromBytes(cfg.RefreshPublicKey)
	if err != nil {
		return nil, err
	}

	return &PublicPaseto{
		issuer:            cfg.Issuer,
		accessPrivateKey:  aPvt,
		accessPublicKey:   aPub,
		refreshPrivateKey: rPvt,
		refreshPublicKey:  rPub,
		accessTTL:         cfg.AccessTTL,
		refreshTTL:        cfg.RefreshTTL,
	}, nil
}

// PublicTokens bundles generated access and refresh public tokens.
type PublicTokens struct {
	Access  string
	Refresh string
}

// Generate issues a new pair of access and refresh public tokens for the given subject and audience.
// customClaims is an optional map of additional claims to include in both tokens.
func (p *PublicPaseto) Generate(subject, audience string, customClaims map[string]any) (PublicTokens, error) {
	access, err := p.makeToken(subject, audience, PublicAccessToken, p.accessTTL, p.accessPrivateKey, customClaims)
	if err != nil {
		return PublicTokens{}, err
	}
	refresh, err := p.makeToken(subject, audience, PublicRefreshToken, p.refreshTTL, p.refreshPrivateKey, customClaims)
	if err != nil {
		return PublicTokens{}, err
	}
	return PublicTokens{Access: access, Refresh: refresh}, nil
}

// ValidateAccess parses and validates an access public token, returning its claims.
func (p *PublicPaseto) ValidateAccess(token string) (PublicClaims, error) {
	return p.parse(token, PublicAccessToken, p.accessPublicKey)
}

// ValidateRefresh parses and validates a refresh public token, returning its claims.
func (p *PublicPaseto) ValidateRefresh(token string) (PublicClaims, error) {
	return p.parse(token, PublicRefreshToken, p.refreshPublicKey)
}

func (p *PublicPaseto) makeToken(
	subject, audience string, typ PublicTokenType, ttl time.Duration, key psto.V4AsymmetricSecretKey, customClaims map[string]any,
) (string, error) {
	now := time.Now().UTC()
	exp := now.Add(ttl)

	tok := psto.NewToken()
	tok.SetSubject(subject)
	tok.SetAudience(audience)
	tok.SetIssuer(p.issuer)
	tok.SetIssuedAt(now)
	tok.SetNotBefore(now)
	tok.SetExpiration(exp)
	tok.SetJti(newPublicJTI())
	_ = tok.Set("typ", string(typ))

	// Add custom claims
	for key, value := range customClaims {
		_ = tok.Set(key, value)
	}

	return pasetoV4Sign(tok, key)
}

func (p *PublicPaseto) parse(token string, expected PublicTokenType, key psto.V4AsymmetricPublicKey) (PublicClaims, error) {
	value, err := parseV4PublicFunc(key, token)
	if err != nil {
		return PublicClaims{}, err
	}
	if err := validatePublicToken(value, expected); err != nil {
		return PublicClaims{}, err
	}
	return buildPublicClaims(value, expected), nil
}

// test hook for parsing v4 public PASETO tokens
var parseV4PublicFunc = func(key psto.V4AsymmetricPublicKey, token string) (*psto.Token, error) {
	parser := psto.NewParser()
	return parser.ParseV4Public(key, token, nil)
}

func buildPublicClaims(value *psto.Token, expected PublicTokenType) PublicClaims {
	sub, _ := value.GetSubject()
	aud, _ := value.GetAudience()
	iss, _ := value.GetIssuer()
	jit, _ := value.GetJti()
	issuedAt, _ := value.GetIssuedAt()
	nbf, _ := value.GetNotBefore()
	exp, _ := value.GetExpiration()

	allClaims := value.Claims()
	customClaims := make(map[string]any)
	reservedKeys := map[string]bool{
		"sub": true, "aud": true, "iss": true, "jti": true,
		"exp": true, "iat": true, "nbf": true, "typ": true,
	}
	for key, val := range allClaims {
		if !reservedKeys[key] {
			customClaims[key] = val
		}
	}

	return PublicClaims{
		Subject:      sub,
		Audience:     aud,
		Issuer:       iss,
		JTI:          jit,
		ExpiresAt:    exp,
		IssuedAt:     issuedAt,
		NotBefore:    nbf,
		TokenType:    expected,
		CustomClaims: customClaims,
	}
}

func validatePublicToken(value *psto.Token, expected PublicTokenType) error {
	typ := ""
	_ = value.Get("typ", &typ)
	if typ != string(expected) {
		return errors.New("paseto: unexpected token type")
	}
	exp, _ := value.GetExpiration()
	if time.Now().UTC().After(exp) {
		return errors.New("paseto: token expired")
	}
	return nil
}

func newPublicJTI() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}
