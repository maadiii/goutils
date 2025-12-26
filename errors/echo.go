package errors

import (
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

func EchoHandler(devMode bool) func(err error, c echo.Context) {
	return func(err error, c echo.Context) {
		switch e := err.(type) {
		case Error:
		case *Error:
			if e.Type != 0 {
				c.Error(c.NoContent(e.Type))
			} else {
				c.Error(c.NoContent(http.StatusInternalServerError))
			}
		case validator.ValidationErrors:
		case *validator.InvalidValidationError:
			c.Error(c.NoContent(http.StatusBadRequest))
		default:
			c.Error(c.NoContent(http.StatusInternalServerError))
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
			c.Error(err)
		} else {
			_, err = c.Response().Write([]byte(errMsg))
			c.Error(err)
		}
	}
}
