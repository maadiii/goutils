package errors

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

func EchoHandler(devMode bool) func(err error, c echo.Context) {
	return func(err error, c echo.Context) {
		handleError(devMode, err, c)
	}
}

func HandleEchoPanic(devMode bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					err := fmt.Errorf("%v\n%s", r, string(debug.Stack()))
					handleError(devMode, err, c)
				}
			}()

			return next(c)
		}
	}
}

func handleError(devMode bool, err error, c echo.Context) {
	errhttp := &echo.HTTPError{
		Message: err.Error(),
	}

	switch e := err.(type) {
	case Error:
	case *Error:
		if e.Type != 0 {
			errhttp.Code = e.Type
		} else {
			errhttp.Code = http.StatusInternalServerError
		}

		if !devMode {
			errhttp.Message = http.StatusText(errhttp.Code)
		}
	case *validator.InvalidValidationError:
	case validator.ValidationErrors:
	case *validator.ValidationErrors:
		errhttp.Code = http.StatusBadRequest
		if !devMode {
			errhttp.Message = http.StatusText(http.StatusBadRequest)
		}
	case *echo.HTTPError:
		errhttp.Code = e.Code
		if !devMode {
			errhttp.Message = http.StatusText(e.Code)
		}
	default:
		errhttp.Code = http.StatusInternalServerError
		if !devMode {
			errhttp.Message = http.StatusText(http.StatusInternalServerError)
		}
	}

	errMsg := err.Error()
	lines := strings.Split(errMsg, "\n")
	errMsg = strings.ToLower(lines[0])
	stack := func() string {
		if len(lines) > 1 {
			return strings.Join(lines[1:], "\n")
		}

		return ""
	}()
	stackedErrMsg := errMsg + "\n" + stack

	if devMode {
		errhttp.Message = stackedErrMsg
	}

	c.Echo().Logger.Errorj(log.JSON{"code": errhttp.Code, "message": errMsg, "stack": stack})
	_ = c.String(errhttp.Code, errhttp.Message.(string))
}
