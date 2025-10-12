package utils

import (
	"github.com/maadiii/utils/paseto"
	"github.com/maadiii/utils/totp"
)

type TOTP interface {
	Generate(opts totp.Opts) (secret string, code string, err error)
	Validate(code, secret string, opts totp.Opts) (err error)
}

type Paseto interface {
	Generate(claims *paseto.Claims) (string, error)
	Validate(token string) (claims *paseto.Claims, err error)
}
