package log

import (
	"context"
	"strings"
	"sync"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/maadiii/goutils"
)

var mu sync.Mutex

func HertzMiddleware(logger goutils.Logger) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		c = WithContext(c, logger)

		ctx.Next(c)

		ctxCopy := ctx.Copy()
		ctxCopy.Errors = ctx.Errors

		go logRequest(c, ctxCopy, logger)
	}
}

func logRequest(c context.Context, ctx *app.RequestContext, logger goutils.Logger) { //nolint
	mu.Lock()
	defer mu.Unlock()

	if strings.Contains(string(ctx.URI().Path()), "swagger") ||
		strings.Contains(string(ctx.URI().Path()), "health") {
		return
	}

	statusCode := ctx.Response.StatusCode()
	statusClass := func() string {
		switch {
		case statusCode < 300:
			return "2xx"
		case statusCode < 400:
			return "3xx"
		case statusCode < 500:
			return "4xx"
		}

		return "5xx"
	}()

	if ctx.Response.StatusCode() < 400 {
		logger.Info(
			"success api call",
			"route", string(ctx.URI().Path()),
			"method", string(ctx.Method()),
			"status", statusCode,
			"status_class", statusClass,
		)

		return
	}

	if ctx.Response.StatusCode() < 500 {
		if len(ctx.Errors) == 0 {
			return
		}

		msg, stack := stack(ctx)
		logger.Error(
			msg,
			"route", string(ctx.URI().Path()),
			"method", string(ctx.Method()),
			"status", statusCode,
			"status_class", statusClass,
			"stack", stack,
		)

		return
	}

	if ctx.Response.StatusCode() >= 500 {
		msg, stack := stack(ctx)
		logger.Error(
			msg,
			"route", string(ctx.URI().Path()),
			"method", string(ctx.Method()),
			"status", statusCode,
			"status_class", statusClass,
			"stack", stack,
		)

		return
	}
}

func stack(rc *app.RequestContext) (msg, stack string) {
	errMsg := rc.Errors.Last().Error()
	lines := strings.Split(errMsg, "\n")
	msg = lines[0]
	if len(lines) > 1 {
		stack = strings.Join(lines[1:], "\n")
	}

	return
}
