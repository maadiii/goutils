package errors

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strings"
)

type Error struct {
	Type int
	Err  error
}

func (e Error) Error() string {
	return e.Err.Error()
}

func (e *Error) Msg(format string, a ...any) error {
	msg := fmt.Sprintf(format, a...)
	msg = fmt.Sprintf("%s\n%s", msg, stack())

	e.Err = errors.New(msg)

	return e
}

func (e *Error) Wrap(err error) error {
	_, ok := err.(Error)
	if !ok {
		msg := err.Error()
		msg = fmt.Sprintf("%s\n%s", msg, stack())

		e.Err = errors.New(msg)

		return e
	}

	e.Err = errors.New(err.Error())

	return e
}

func New(format string, a ...any) error {
	msg := fmt.Sprintf(format, a...)
	msg = fmt.Sprintf("%s\n%s", msg, stack())

	return &Error{Err: errors.New(msg)}
}

func Wrap(err error) error {
	_, ok := err.(*Error)
	if !ok {
		msg := err.Error()
		msg = fmt.Sprintf("%s\n%s", msg, stack())

		return &Error{Err: errors.New(msg)}
	}

	return errors.New(err.Error())
}

func Unwrap(err error) error {
	return errors.Unwrap(err)
}

func Join(errs ...error) error {
	return errors.Join(errs...)
}

func Is(err, target error) bool {
	e, ok := err.(*Error)
	if ok {
		t, ok := target.(*Error)
		if ok {
			return e.Type == t.Type
		}
	}

	return errors.Is(err, target)
}

func As(err error, target any) bool {
	return errors.As(err, target)
}

func BadRequest() *Error {
	return &Error{Type: http.StatusBadRequest}
}

func Unauthorized() *Error {
	return &Error{Type: http.StatusUnauthorized}
}

func PaymentRequired() *Error {
	return &Error{Type: http.StatusPaymentRequired}
}

func Forbidden() *Error {
	return &Error{Type: http.StatusForbidden}
}

func NotFound() *Error {
	return &Error{Type: http.StatusNotFound}
}

func Conflict() *Error {
	return &Error{Type: http.StatusConflict}
}

func AlreadyExist() *Error {
	return &Error{Type: http.StatusConflict}
}

func Gone() *Error {
	return &Error{Type: http.StatusGone}
}

func StatusUnprocessableEntity() *Error {
	return &Error{Type: http.StatusUnprocessableEntity}
}

func InternalServerError() *Error {
	return &Error{Type: http.StatusInternalServerError}
}

func stack() string {
	buf := make([]byte, 2048)
	runtime.Stack(buf, false)

	lines := strings.Split(string(buf), "\n")
	lines = lines[5 : len(lines)-1]

	return strings.Join(lines, "\n")
}
