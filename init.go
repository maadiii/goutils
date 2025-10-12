package utils

import (
	"context"

	"github.com/maadiii/utils/totp"
)

type TOTP interface {
	Generate(ctx context.Context, opts totp.Opts) (secret string, code string, err error)
	Validate(ctx context.Context, secret, code string) (err error)
}
