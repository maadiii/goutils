package paseto

import (
	"time"

	psto "aidanwoods.dev/go-paseto"
)

func New(key []byte, issuer string) *paseto {
	symmKey, err := psto.V4SymmetricKeyFromBytes(key)
	if err != nil {
		panic(err)
	}

	return &paseto{symmKey, issuer}
}

type paseto struct {
	key    psto.V4SymmetricKey
	issuer string
}

func (p *paseto) Generate(claims *Claims) (string, error) {
	now := time.Now()

	token := psto.NewToken()
	token.SetAudience(claims.Audience)
	token.SetExpiration(claims.Expiration)
	token.SetIssuedAt(now)
	token.SetNotBefore(now)
	token.SetJti(claims.Jti)
	token.SetIssuer(p.issuer)
	token.SetSubject(claims.Subject)

	if len(claims.PhoneNumber) != 0 {
		_ = token.Set("phn", claims.PhoneNumber)
	}

	if len(claims.TotpSecret) != 0 {
		_ = token.Set("tpcrt", claims.TotpSecret)
	}

	if len(claims.StateToken) != 0 {
		_ = token.Set("stetkn", claims.StateToken)
	}

	if len(claims.UserID) != 0 {
		_ = token.Set("user_id", claims.UserID)
	}

	return token.V4Encrypt(p.key, nil), nil
}

func (p *paseto) Validate(token string) (claims *Claims, err error) {
	value, err := psto.NewParser().ParseV4Local(p.key, token, nil)
	if err != nil {
		return
	}

	claims = new(Claims)
	claims.Audience, _ = value.GetAudience()
	claims.Expiration, _ = value.GetExpiration()
	claims.Jti, _ = value.GetJti()
	claims.Subject, _ = value.GetSubject()

	_ = value.Get("phn", &claims.PhoneNumber)
	_ = value.Get("tpcrt", &claims.TotpSecret)
	_ = value.Get("stetkn", &claims.StateToken)
	_ = value.Get("user_id", &claims.UserID)

	return
}
