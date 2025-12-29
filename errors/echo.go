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

func HandleEchoPanic(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("%v\n%s", r, string(debug.Stack()))
				handleError(false, err, c)
			}
		}()

		return next(c)
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
	if devMode {
		stack := func() string {
			if len(lines) > 1 {
				return strings.Join(lines[1:], "\n")
			}

			return ""
		}()

		_, err = c.Response().Write([]byte(errMsg + "\n" + stack))
		if err != nil {
			panic(err)
		}
	}
}
