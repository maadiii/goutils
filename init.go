package utils

import (
	"github.com/maadiii/utils/totp"
)

type TOTP interface {
	Generate(opts totp.Opts) (secret string, code string, err error)
	Validate(code, secret string, opts totp.Opts) (err error)
}
