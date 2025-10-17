package errors

import (
	"context"
	"net/http"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/go-playground/validator/v10"
)

func HertzHandler(devMode bool) func(c context.Context, rc *app.RequestContext, err error) {
	return func(c context.Context, rc *app.RequestContext, err error) {
		switch e := err.(type) {
		case Error:
		case *Error:
			if e.Type != 0 {
				_ = rc.Error(rc.AbortWithError(e.Type, e))
			} else {
				_ = rc.Error(rc.AbortWithError(http.StatusInternalServerError, e))
			}
		case validator.ValidationErrors:
		case *validator.InvalidValidationError:
			_ = rc.Error(rc.AbortWithError(http.StatusBadRequest, e))
		default:
			_ = rc.Error(rc.AbortWithError(500, err))
		}

		errMsg := rc.Errors.Last().Error()
		lines := strings.Split(errMsg, "\n")
		errMsg = strings.ToLower(lines[0])
		if devMode {
			stack := func() string {
				if len(lines) > 1 {
					return strings.Join(lines[1:], "\n")
				}

				return ""
			}()

			rc.SetBodyString(errMsg + "\n" + stack)
		} else {
			rc.SetBodyString(errMsg)
		}
	}
}
