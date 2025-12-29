package errors

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
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
	switch e := err.(type) {
	case Error:
	case *Error:
		if e.Type != 0 {
			_ = c.NoContent(e.Type)
		} else {
			_ = c.NoContent(http.StatusInternalServerError)
		}
	case validator.ValidationErrors:
	case *validator.InvalidValidationError:
		_ = c.NoContent(http.StatusBadRequest)
	case *echo.HTTPError:
		_ = c.NoContent(e.Code)
	default:
		_ = c.NoContent(http.StatusInternalServerError)
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
	final := errMsg + "\n" + stack
	c.Echo().Logger.Error(final)

	if devMode {
		_, _ = c.Response().Write([]byte(final))
	} else {
		_, _ = c.Response().Write([]byte(errMsg))
	}
}
