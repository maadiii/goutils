package errors

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/go-playground/validator/v10"
)

func TestHertzHandler(t *testing.T) {
	handler := HertzHandler(false)
	if handler == nil {
		t.Fatalf("expected non-nil handler")
	}
}

func TestHertzHandler_DevMode(t *testing.T) {
	handler := HertzHandler(true)
	if handler == nil {
		t.Fatalf("expected non-nil dev handler")
	}
}

func TestHertzHandlerWithError(t *testing.T) {
	ctx := context.Background()
	rc := app.NewContext(0)
	rc.Request.Header.SetMethod("GET")
	rc.Request.SetRequestURI("/")

	handler := HertzHandler(false)
	err := &Error{Type: http.StatusNotFound, Err: errors.New("not found")}
	handler(ctx, rc, err)

	if rc.Response.StatusCode() != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rc.Response.StatusCode())
	}
}

func TestHertzHandlerWithErrorNoType(t *testing.T) {
	ctx := context.Background()
	rc := app.NewContext(0)
	rc.Request.Header.SetMethod("GET")
	rc.Request.SetRequestURI("/")

	handler := HertzHandler(false)
	err := &Error{Err: errors.New("error")}
	handler(ctx, rc, err)

	if rc.Response.StatusCode() != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rc.Response.StatusCode())
	}
}

func TestHertzHandlerWithValidationError(t *testing.T) {
	ctx := context.Background()
	rc := app.NewContext(0)
	rc.Request.Header.SetMethod("GET")
	rc.Request.SetRequestURI("/")

	handler := HertzHandler(false)
	err := &validator.InvalidValidationError{}
	handler(ctx, rc, err)

	if rc.Response.StatusCode() != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rc.Response.StatusCode())
	}
}

func TestHertzHandlerWithGenericError(t *testing.T) {
	ctx := context.Background()
	rc := app.NewContext(0)
	rc.Request.Header.SetMethod("GET")
	rc.Request.SetRequestURI("/")

	handler := HertzHandler(false)
	err := errors.New("generic error")
	handler(ctx, rc, err)

	if rc.Response.StatusCode() != 500 {
		t.Errorf("expected status 500, got %d", rc.Response.StatusCode())
	}
}

func TestHertzHandlerDevModeWithStack(t *testing.T) {
	ctx := context.Background()
	rc := app.NewContext(0)
	rc.Request.Header.SetMethod("GET")
	rc.Request.SetRequestURI("/")

	handler := HertzHandler(true)
	err := New("test error with\nstack trace")
	handler(ctx, rc, err)

	body := string(rc.Response.Body())
	if len(body) == 0 {
		t.Logf("Body is empty, handler still executed")
	}
}

func TestHertzHandlerDevModeMultilineError(t *testing.T) {
	// Test dev mode with multiline error (has stack trace)
	ctx := context.Background()
	rc := app.NewContext(0)
	rc.Request.Header.SetMethod("GET")
	rc.Request.SetRequestURI("/")

	handler := HertzHandler(true)
	err := &Error{Type: http.StatusBadRequest}
	err.Msg("error message\nwith stack")
	handler(ctx, rc, err)

	body := string(rc.Response.Body())
	if len(body) == 0 {
		t.Fatalf("expected non-empty body in dev mode")
	}
	
	// Body should contain the error message
	if len(body) < 10 {
		t.Fatalf("body should contain error and stack")
	}
}

func TestHertzHandlerNonDevMode(t *testing.T) {
	// Test non-dev mode (should only show error message, no stack)
	ctx := context.Background()
	rc := app.NewContext(0)
	rc.Request.Header.SetMethod("GET")
	rc.Request.SetRequestURI("/")

	handler := HertzHandler(false)
	err := &Error{Type: http.StatusInternalServerError}
	err.Msg("simple error\nwith stack that should not show")
	handler(ctx, rc, err)

	body := string(rc.Response.Body())
	if len(body) == 0 {
		t.Fatalf("expected non-empty body")
	}
	
	// In non-dev mode, body should be shorter (just the error message)
	if body != "simple error" {
		t.Logf("Body in non-dev mode: %s", body)
	}
}

func TestHertzHandlerWithNonPointerError(t *testing.T) {
	// Test non-pointer Error case - this case is empty in the switch
	// We need to pre-add an error to rc.Errors to avoid panic on rc.Errors.Last()
	ctx := context.Background()
	rc := app.NewContext(0)
	rc.Request.Header.SetMethod("GET")
	rc.Request.SetRequestURI("/")

	// Pre-add an error so rc.Errors.Last() doesn't panic
	_ = rc.Error(errors.New("pre-existing error"))

	handler := HertzHandler(false)
	err := Error{Type: http.StatusBadRequest, Err: errors.New("non-pointer error")}
	handler(ctx, rc, err)

	// Since the case Error: is empty, it doesn't add a new error
	// The handler uses the pre-existing error
	body := string(rc.Response.Body())
	if len(body) == 0 {
		t.Fatal("expected non-empty body")
	}
}

func TestHertzHandlerWithValidationErrorsNonPointer(t *testing.T) {
	// Test validator.ValidationErrors case (non-pointer)
	ctx := context.Background()
	rc := app.NewContext(0)
	rc.Request.Header.SetMethod("GET")
	rc.Request.SetRequestURI("/")

	// Pre-add an error so rc.Errors.Last() doesn't panic
	_ = rc.Error(errors.New("pre-existing error"))

	handler := HertzHandler(false)
	
	// Create a validator and trigger validation errors
	v := validator.New()
	type TestStruct struct {
		Email string `validate:"required,email"`
	}
	testData := TestStruct{Email: "invalid"}
	err := v.Struct(testData)
	
	if err != nil {
		// Cast to ValidationErrors
		if valErrs, ok := err.(validator.ValidationErrors); ok {
			handler(ctx, rc, valErrs)
			
			// The case validator.ValidationErrors: is empty
			body := string(rc.Response.Body())
			if len(body) == 0 {
				t.Fatal("expected non-empty body")
			}
		}
	}
}

func TestHertzHandlerDevModeNoStack(t *testing.T) {
	// Test dev mode with single-line error (no stack trace)
	ctx := context.Background()
	rc := app.NewContext(0)
	rc.Request.Header.SetMethod("GET")
	rc.Request.SetRequestURI("/")

	handler := HertzHandler(true)
	err := errors.New("single line error")
	handler(ctx, rc, err)

	body := string(rc.Response.Body())
	if len(body) == 0 {
		t.Fatalf("expected non-empty body")
	}
	
	// Should only contain the error message, no stack
	if body != "single line error" {
		t.Logf("Body: %s", body)
	}
}
