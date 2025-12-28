package validator

import (
	"github.com/go-playground/validator/v10"
	"github.com/maadiii/utils/errors"
)

type Validator struct {
	validator *validator.Validate
}

func NewValidator() *Validator {
	return &Validator{
		validator: validator.New(),
	}
}

func (v *Validator) Validate(i any) (err error) {
	if err = v.validator.Struct(i); err != nil {
		err = errors.BadRequest().Wrap(err)
	}

	return
}
