package utils

import (
	"context"
	"time"

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

type Cache interface {
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Get(ctx context.Context, key string) (value string, err error)
	Del(ctx context.Context, keys ...string) (err error)
}

type Random interface {
	String(length int) (string, error)
}

type Password interface {
	Generate(plain string) (hash string, err error)
	Compare(hash, plain string) bool
}
type Logger interface {
	Debug(ctx context.Context, msg string, fields ...any)
	Info(ctx context.Context, msg string, fields ...any)
	Warn(ctx context.Context, msg string, fields ...any)
	Error(ctx context.Context, msg string, fields ...any)
	Sync() error
}
