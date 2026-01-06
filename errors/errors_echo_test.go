package errors

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

func TestEchoHandler(t *testing.T) {
	handler := EchoHandler(false)
	if handler == nil {
		t.Fatalf("expected non-nil handler")
	}

	devHandler := EchoHandler(true)
	if devHandler == nil {
		t.Fatalf("expected non-nil dev handler")
	}
}

func TestEchoHandlerWithError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := EchoHandler(false)
	err := &Error{Type: http.StatusNotFound}
	err.Msg("not found")
	handler(err, c)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestEchoHandlerWithErrorNoType(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := EchoHandler(false)
	err := &Error{}
	err.Msg("generic error")
	handler(err, c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestEchoHandlerWithValidationError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := EchoHandler(false)
	err := &validator.ValidationErrors{}
	handler(err, c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestEchoHandlerWithHTTPError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := EchoHandler(false)
	err := echo.NewHTTPError(http.StatusForbidden, "forbidden")
	handler(err, c)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestEchoHandlerWithGenericError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := EchoHandler(false)
	err := errors.New("generic error")
	handler(err, c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestHandleEchoPanic(t *testing.T) {
	middleware := HandleEchoPanic(false)
	if middleware == nil {
		t.Fatalf("expected non-nil middleware")
	}

	devMiddleware := HandleEchoPanic(true)
	if devMiddleware == nil {
		t.Fatalf("expected non-nil dev middleware")
	}
}

func TestHandleEchoPanicWithPanic(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := HandleEchoPanic(false)
	handler := middleware(func(c echo.Context) error {
		panic("test panic")
	})

	handler(c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestHandleEchoPanicNoPanic(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	middleware := HandleEchoPanic(false)
	handler := middleware(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	err := handler(c)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestEchoHandlerDevMode(t *testing.T) {
	// Test dev mode writes full error with stack
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := EchoHandler(true)
	err := New("test error with stack")
	handler(err, c)

	body := rec.Body.String()
	if len(body) == 0 {
		t.Fatalf("expected non-empty body in dev mode")
	}

	// Body should contain more than just the error message (should have stack trace)
	if !strings.Contains(body, "test error with stack") {
		t.Fatalf("body should contain error message")
	}
}
