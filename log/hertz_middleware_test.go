package log

import (
	"context"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/errors"
)

type mockLogger struct {
	infoCalled  bool
	errorCalled bool
	lastMsg     string
	lastFields  []any
}

func (m *mockLogger) Info(msg string, fields ...any) {
	m.infoCalled = true
	m.lastMsg = msg
	m.lastFields = fields
}

func (m *mockLogger) Error(msg string, fields ...any) {
	m.errorCalled = true
	m.lastMsg = msg
	m.lastFields = fields
}

func (m *mockLogger) Debug(msg string, fields ...any) {}

func (m *mockLogger) Warn(msg string, fields ...any) {}

func (m *mockLogger) Sync() error { return nil }

func TestHertzMiddleware_Success(t *testing.T) {
	logger := &mockLogger{}
	middleware := HertzMiddleware(logger)

	ctx := app.NewContext(0)
	ctx.Request.SetRequestURI("/api/test")
	ctx.Request.SetMethod("GET")
	ctx.Response.SetStatusCode(200)

	var nextCalled bool
	ctx.SetHandlers(app.HandlersChain{
		middleware,
		func(c context.Context, ctx *app.RequestContext) {
			nextCalled = true
		},
	})

	middleware(context.Background(), ctx)
	
	// Give goroutine time to execute
	time.Sleep(50 * time.Millisecond)

	if !nextCalled {
		t.Fatal("Next handler was not called")
	}

	if !logger.infoCalled {
		t.Fatal("Info logger should be called for 200 status")
	}

	if logger.errorCalled {
		t.Fatal("Error logger should not be called for 200 status")
	}
}

func TestHertzMiddleware_3xxRedirect(t *testing.T) {
	logger := &mockLogger{}
	middleware := HertzMiddleware(logger)

	ctx := app.NewContext(0)
	ctx.Request.SetRequestURI("/api/redirect")
	ctx.Request.SetMethod("POST")
	ctx.Response.SetStatusCode(302)

	var nextCalled bool
	ctx.SetHandlers(app.HandlersChain{
		middleware,
		func(c context.Context, ctx *app.RequestContext) {
			nextCalled = true
		},
	})

	middleware(context.Background(), ctx)
	
	time.Sleep(50 * time.Millisecond)

	if !nextCalled {
		t.Fatal("Next handler was not called")
	}

	if !logger.infoCalled {
		t.Fatal("Info logger should be called for 3xx status")
	}
}

func TestHertzMiddleware_4xxClientError(t *testing.T) {
	logger := &mockLogger{}
	middleware := HertzMiddleware(logger)

	ctx := app.NewContext(0)
	ctx.Request.SetRequestURI("/api/notfound")
	ctx.Request.SetMethod("GET")
	ctx.Response.SetStatusCode(404)
	
	// Add error to trigger error logging
	ctx.Errors = append(ctx.Errors, errors.NewPublic("not found error"))

	var nextCalled bool
	ctx.SetHandlers(app.HandlersChain{
		middleware,
		func(c context.Context, ctx *app.RequestContext) {
			nextCalled = true
		},
	})

	middleware(context.Background(), ctx)
	
	time.Sleep(50 * time.Millisecond)

	if !nextCalled {
		t.Fatal("Next handler was not called")
	}

	if !logger.errorCalled {
		t.Fatal("Error logger should be called for 4xx status with errors")
	}
}

func TestHertzMiddleware_4xxNoErrors(t *testing.T) {
	logger := &mockLogger{}
	middleware := HertzMiddleware(logger)

	ctx := app.NewContext(0)
	ctx.Request.SetRequestURI("/api/badrequest")
	ctx.Request.SetMethod("POST")
	ctx.Response.SetStatusCode(400)
	// No errors added

	var nextCalled bool
	ctx.SetHandlers(app.HandlersChain{
		middleware,
		func(c context.Context, ctx *app.RequestContext) {
			nextCalled = true
		},
	})

	middleware(context.Background(), ctx)
	
	time.Sleep(50 * time.Millisecond)

	if !nextCalled {
		t.Fatal("Next handler was not called")
	}

	if logger.infoCalled {
		t.Fatal("Info logger should not be called for 4xx status")
	}

	if logger.errorCalled {
		t.Fatal("Error logger should not be called for 4xx status without errors")
	}
}

func TestHertzMiddleware_5xxServerError(t *testing.T) {
	logger := &mockLogger{}
	middleware := HertzMiddleware(logger)

	ctx := app.NewContext(0)
	ctx.Request.SetRequestURI("/api/servererror")
	ctx.Request.SetMethod("PUT")
	ctx.Response.SetStatusCode(500)
	
	// Add error to trigger error logging
	ctx.Errors = append(ctx.Errors, errors.NewPublic("internal server error"))

	var nextCalled bool
	ctx.SetHandlers(app.HandlersChain{
		middleware,
		func(c context.Context, ctx *app.RequestContext) {
			nextCalled = true
		},
	})

	middleware(context.Background(), ctx)
	
	time.Sleep(50 * time.Millisecond)

	if !nextCalled {
		t.Fatal("Next handler was not called")
	}

	if !logger.errorCalled {
		t.Fatal("Error logger should be called for 5xx status")
	}
}

func TestHertzMiddleware_SwaggerPath(t *testing.T) {
	logger := &mockLogger{}
	middleware := HertzMiddleware(logger)

	ctx := app.NewContext(0)
	ctx.Request.SetRequestURI("/api/swagger/index.html")
	ctx.Request.SetMethod("GET")
	ctx.Response.SetStatusCode(200)

	var nextCalled bool
	ctx.SetHandlers(app.HandlersChain{
		middleware,
		func(c context.Context, ctx *app.RequestContext) {
			nextCalled = true
		},
	})

	middleware(context.Background(), ctx)
	
	time.Sleep(50 * time.Millisecond)

	if !nextCalled {
		t.Fatal("Next handler was not called")
	}

	// Swagger paths should not trigger logging
	if logger.infoCalled {
		t.Fatal("Info logger should not be called for swagger paths")
	}

	if logger.errorCalled {
		t.Fatal("Error logger should not be called for swagger paths")
	}
}

func TestHertzMiddleware_HealthPath(t *testing.T) {
	logger := &mockLogger{}
	middleware := HertzMiddleware(logger)

	ctx := app.NewContext(0)
	ctx.Request.SetRequestURI("/health")
	ctx.Request.SetMethod("GET")
	ctx.Response.SetStatusCode(200)

	var nextCalled bool
	ctx.SetHandlers(app.HandlersChain{
		middleware,
		func(c context.Context, ctx *app.RequestContext) {
			nextCalled = true
		},
	})

	middleware(context.Background(), ctx)
	
	time.Sleep(50 * time.Millisecond)

	if !nextCalled {
		t.Fatal("Next handler was not called")
	}

	// Health paths should not trigger logging
	if logger.infoCalled {
		t.Fatal("Info logger should not be called for health paths")
	}

	if logger.errorCalled {
		t.Fatal("Error logger should not be called for health paths")
	}
}

func TestHertzMiddleware_ErrorWithStack(t *testing.T) {
	logger := &mockLogger{}
	middleware := HertzMiddleware(logger)

	ctx := app.NewContext(0)
	ctx.Request.SetRequestURI("/api/error")
	ctx.Request.SetMethod("POST")
	ctx.Response.SetStatusCode(500)
	
	// Add error with stack trace (multiline error)
	ctx.Errors = append(ctx.Errors, errors.NewPublic("error message\nstack trace line 1\nstack trace line 2"))

	var nextCalled bool
	ctx.SetHandlers(app.HandlersChain{
		middleware,
		func(c context.Context, ctx *app.RequestContext) {
			nextCalled = true
		},
	})

	middleware(context.Background(), ctx)
	
	time.Sleep(50 * time.Millisecond)

	if !nextCalled {
		t.Fatal("Next handler was not called")
	}

	if !logger.errorCalled {
		t.Fatal("Error logger should be called for 500 status with errors")
	}

	if logger.lastMsg != "error message" {
		t.Fatalf("expected message 'error message', got %s", logger.lastMsg)
	}
}

func TestStack(t *testing.T) {
	ctx := app.NewContext(0)
	ctx.Errors = append(ctx.Errors, errors.NewPublic("simple error"))
	
	msg, stack := stack(ctx)
	if msg != "simple error" {
		t.Fatalf("expected message 'simple error', got %s", msg)
	}
	if stack != "" {
		t.Fatalf("expected empty stack, got %s", stack)
	}
}

func TestStackWithMultiline(t *testing.T) {
	ctx := app.NewContext(0)
	ctx.Errors = append(ctx.Errors, errors.NewPublic("error\nline1\nline2"))
	
	msg, stack := stack(ctx)
	if msg != "error" {
		t.Fatalf("expected message 'error', got %s", msg)
	}
	if stack != "line1\nline2" {
		t.Fatalf("expected stack 'line1\\nline2', got %s", stack)
	}
}
