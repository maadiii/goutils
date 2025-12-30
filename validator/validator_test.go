package validator

import (
	"testing"

	"github.com/labstack/echo/v4"
)

func TestNewValidator(t *testing.T) {
	validator := NewValidator()
	if validator == nil {
		t.Fatalf("NewValidator returned nil")
	}
}

func TestValidate_ValidEmail(t *testing.T) {
	validator := NewValidator()
	
	type Request struct {
		Email string `validate:"required,email"`
	}
	
	req := &Request{Email: "user@example.com"}
	err := validator.Validate(req)
	
	if err != nil {
		t.Fatalf("expected validation to pass for valid email, got error: %v", err)
	}
}

func TestValidate_InvalidEmail(t *testing.T) {
	validator := NewValidator()
	
	type Request struct {
		Email string `validate:"required,email"`
	}
	
	req := &Request{Email: "not-an-email"}
	err := validator.Validate(req)
	
	if err == nil {
		t.Fatalf("expected validation to fail for invalid email")
	}
}

func TestValidate_RequiredField(t *testing.T) {
	validator := NewValidator()
	
	type Request struct {
		Name string `validate:"required"`
	}
	
	req := &Request{Name: ""}
	err := validator.Validate(req)
	
	if err == nil {
		t.Fatalf("expected validation to fail for empty required field")
	}
}

func TestValidate_MinLength(t *testing.T) {
	validator := NewValidator()
	
	type Request struct {
		Password string `validate:"required,min=8"`
	}
	
	tests := []struct {
		password string
		valid    bool
	}{
		{"short", false},
		{"longpassword", true},
		{"12345678", true},
	}
	
	for _, test := range tests {
		req := &Request{Password: test.password}
		err := validator.Validate(req)
		
		if test.valid && err != nil {
			t.Fatalf("expected validation to pass for %q, got error: %v", test.password, err)
		}
		if !test.valid && err == nil {
			t.Fatalf("expected validation to fail for %q", test.password)
		}
	}
}

func TestValidate_Number(t *testing.T) {
	validator := NewValidator()
	
	type Request struct {
		Age int `validate:"min=0,max=150"`
	}

	tests := []struct {
		age   int
		valid bool
	}{
		{-1, false},
		{1, true},
	}

	for _, test := range tests {
		req := &Request{Age: test.age}
		err := validator.Validate(req)
		
		if test.valid && err != nil {
			t.Fatalf("expected validation to pass for age %d, got error: %v", test.age, err)
		}
		if !test.valid && err == nil {
			t.Fatalf("expected validation to fail for age %d", test.age)
		}
	}
}

func TestValidate_BindAndValidate(t *testing.T) {
	// Test that the validator can be used with echo context
	validator := NewValidator()
	
	type Request struct {
		Email string `json:"email" validate:"required,email"`
	}
	
	// Create a mock echo context would require more setup
	// For now, just test that the validator instance can be created
	// and used standalone
	req := &Request{Email: "test@example.com"}
	err := validator.Validate(req)
	
	if err != nil {
		t.Fatalf("validation failed: %v", err)
	}
}

// TestValidatorIsEchoValidator ensures the validator implements echo's Validator interface
func TestValidatorIsEchoValidator(t *testing.T) {
	validator := NewValidator()
	
	// Check if it implements the Validate method expected by echo
	var _ echo.Validator = validator
}
