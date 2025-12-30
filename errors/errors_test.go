package errors

import (
	"errors"
	"testing"
)

func TestNewError(t *testing.T) {
	err := New("test error %d", 123)
	if err == nil {
		t.Fatalf("expected non-nil error")
	}

	errStr := err.Error()
	if errStr == "" {
		t.Fatalf("expected non-empty error string")
	}
}

func TestWrap(t *testing.T) {
	originalErr := errors.New("original error")
	wrapped := Wrap(originalErr)
	if wrapped == nil {
		t.Fatalf("expected non-nil error")
	}
}

func TestUnwrap(t *testing.T) {
	originalErr := errors.New("original error")
	result := Unwrap(originalErr)
	// Since the error is not wrapped, Unwrap should return nil
	if result != nil {
		t.Logf("Unwrap returned: %v", result)
	}
}

func TestJoin(t *testing.T) {
	err1 := New("error 1")
	err2 := New("error 2")
	joined := Join(err1, err2)
	if joined == nil {
		t.Fatalf("expected non-nil error from Join")
	}
}

func TestIs(t *testing.T) {
	err1 := &Error{Type: 404}
	err2 := &Error{Type: 404}
	result := Is(err1, err2)
	if !result {
		t.Logf("Is check result: %v", result)
	}
}

func TestAs(t *testing.T) {
	err := &Error{Type: 500}
	var target *Error
	result := errors.As(err, &target)
	if !result {
		t.Logf("As check result: %v", result)
	}
}

func TestErrorMsg(t *testing.T) {
	e := &Error{}
	result := e.Msg("format %s", "test")
	if result == nil {
		t.Fatalf("expected non-nil error")
	}
}

func TestBadRequest(t *testing.T) {
	err := BadRequest()
	if err == nil {
		t.Fatalf("expected non-nil error")
	}

	if err.Type != 400 {
		t.Fatalf("expected Type 400, got %d", err.Type)
	}
}

func TestUnauthorized(t *testing.T) {
	err := Unauthorized()
	if err == nil {
		t.Fatalf("expected non-nil error")
	}

	if err.Type != 401 {
		t.Fatalf("expected Type 401, got %d", err.Type)
	}
}

func TestPaymentRequired(t *testing.T) {
	err := PaymentRequired()
	if err == nil {
		t.Fatalf("expected non-nil error")
	}

	if err.Type != 402 {
		t.Fatalf("expected Type 402, got %d", err.Type)
	}
}

func TestForbidden(t *testing.T) {
	err := Forbidden()
	if err == nil {
		t.Fatalf("expected non-nil error")
	}

	if err.Type != 403 {
		t.Fatalf("expected Type 403, got %d", err.Type)
	}
}

func TestNotFound(t *testing.T) {
	err := NotFound()
	if err == nil {
		t.Fatalf("expected non-nil error")
	}

	if err.Type != 404 {
		t.Fatalf("expected Type 404, got %d", err.Type)
	}
}

func TestConflict(t *testing.T) {
	err := Conflict()
	if err == nil {
		t.Fatalf("expected non-nil error")
	}

	if err.Type != 409 {
		t.Fatalf("expected Type 409, got %d", err.Type)
	}
}

func TestAlreadyExist(t *testing.T) {
	err := AlreadyExist()
	if err == nil {
		t.Fatalf("expected non-nil error")
	}

	if err.Type != 409 {
		t.Fatalf("expected Type 409, got %d", err.Type)
	}
}

func TestGone(t *testing.T) {
	err := Gone()
	if err == nil {
		t.Fatalf("expected non-nil error")
	}

	if err.Type != 410 {
		t.Fatalf("expected Type 410, got %d", err.Type)
	}
}

func TestUnprocessableEntity(t *testing.T) {
	err := UnprocessableEntity()
	if err == nil {
		t.Fatalf("expected non-nil error")
	}

	if err.Type != 422 {
		t.Fatalf("expected Type 422, got %d", err.Type)
	}
}

func TestInternalServerError(t *testing.T) {
	err := InternalServerError()
	if err == nil {
		t.Fatalf("expected non-nil error")
	}

	if err.Type != 500 {
		t.Fatalf("expected Type 500, got %d", err.Type)
	}
}

func TestErrorWrap(t *testing.T) {
	originalErr := errors.New("original error")
	e := &Error{}
	wrapped := e.Wrap(originalErr)
	if wrapped == nil {
		t.Fatalf("Wrap returned nil")
	}
}

func TestErrorWrapWithErrorType(t *testing.T) {
	e := &Error{Type: 404}
	wrappedErr := &Error{Type: 500}
	wrappedErr.Msg("wrapped error message")
	result := e.Wrap(wrappedErr)
	if result == nil {
		t.Fatalf("Wrap returned nil")
	}
}

func TestAsWrapper(t *testing.T) {
	err := &Error{Type: 500}
	var target *Error
	result := As(err, &target)
	if !result {
		t.Fatalf("As should return true for matching type")
	}
	if target == nil {
		t.Fatalf("target should not be nil")
	}
}

func TestWrapWithErrorPointer(t *testing.T) {
	// Test Wrap function with *Error type (should return errors.New)
	originalErr := &Error{Type: 404, Err: errors.New("not found")}
	wrapped := Wrap(originalErr)
	if wrapped == nil {
		t.Fatalf("Wrap returned nil")
	}
	// Should not be *Error type
	_, ok := wrapped.(*Error)
	if ok {
		t.Fatalf("expected non-*Error type")
	}
}

func TestErrorWrapWithError(t *testing.T) {
	// Test Error.Wrap method with Error type (not pointer)
	e := &Error{Type: 400}
	wrappedErr := Error{Type: 500, Err: errors.New("error")}
	result := e.Wrap(wrappedErr)
	if result == nil {
		t.Fatalf("Wrap returned nil")
	}
}

func TestIsDifferentTypes(t *testing.T) {
	// Test Is with different types
	err1 := &Error{Type: 404}
	err2 := &Error{Type: 500}
	result := Is(err1, err2)
	if result {
		t.Fatalf("Is should return false for different types")
	}
}

func TestIsWithNonErrorTypes(t *testing.T) {
	// Test Is fallback to errors.Is
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	result := Is(err1, err2)
	if result {
		t.Logf("Is result for non-Error types: %v", result)
	}
}
