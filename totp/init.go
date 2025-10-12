package totp

import (
	"context"
	"errors"
	"time"

	tp "github.com/pquerna/otp"
	ttp "github.com/pquerna/otp/totp"
)

type TOTP interface {
	Generate(ctx context.Context, opts Opts) (secret string, code string, err error)
	Validate(ctx context.Context, secret, code string) (err error)
}

func New() *totp {
	return &totp{}
}

type totp struct{}

func (t *totp) Generate(ctx context.Context, opts Opts) (secret string, code string, err error) {
	key, err := ttp.Generate(ttp.GenerateOpts{
		Issuer:      opts.Issuer,
		AccountName: opts.AccountName,
	})
	if err != nil {
		return
	}

	secret = key.Secret()
	code, err = ttp.GenerateCodeCustom(secret, time.Now(), ttp.ValidateOpts{
		Period: opts.Period,
		Digits: tp.Digits(opts.Digits),
	})

	return
}

func (t *totp) Validate(code string, secret string, opts Opts) (err error) {
	valid, err := ttp.ValidateCustom(code, secret, time.Now(), ttp.ValidateOpts{
		Period: opts.Period,
		Digits: tp.Digits(opts.Digits),
	})
	if err != nil {
		return
	}

	if !valid {
		return errors.New("invalid code")
	}

	return
}

type Opts struct {
	// Name of the issuing Organization/Company.
	Issuer string
	// Name of the User's Account (eg, email address)
	AccountName string
	// Periods before or after the current time to allow.  Value of 1 allows up to Period
	// of either side of the specified time.  Defaults to 0 allowed skews.  Values greater
	// than 1 are likely sketchy.
	Skew   uint
	Period uint
	Digits int
}
