package totp

import (
	"testing"
)

func TestGenerateAndValidate_Success(t *testing.T) {
	svc := New()
	opts := Opts{
		Issuer:      "TestIssuer",
		AccountName: "user@example.com",
		Period:      30,
		Digits:      6,
	}

	secret, code, err := svc.Generate(opts)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if secret == "" {
		t.Fatalf("Generate returned empty secret")
	}
	if len(code) != opts.Digits {
		t.Fatalf("expected code length %d, got %d (code: %s)", opts.Digits, len(code), code)
	}

	if err := svc.Validate(code, secret, opts); err != nil {
		t.Fatalf("Validate returned error for valid code: %v", err)
	}
}

func TestValidate_FailsOnWrongCode(t *testing.T) {
	svc := New()
	opts := Opts{
		Issuer:      "TestIssuer",
		AccountName: "user2@example.com",
		Period:      30,
		Digits:      6,
	}

	secret, code, err := svc.Generate(opts)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	// tamper one digit to produce wrong code
	wrong := code
	if len(code) > 0 {
		if code[0] == '0' {
			wrong = "1" + code[1:]
		} else {
			wrong = "0" + code[1:]
		}
	}

	if err := svc.Validate(wrong, secret, opts); err == nil {
		t.Fatalf("Validate did not return error for wrong code; wrong=%s secret=%s", wrong, secret)
	}
}

func TestGenerateWithDifferentDigits(t *testing.T) {
	svc := New()
	opts := Opts{
		Issuer:      "TestIssuer",
		AccountName: "user3@example.com",
		Period:      30,
		Digits:      8,
	}

	secret, code, err := svc.Generate(opts)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if secret == "" {
		t.Fatalf("Generate returned empty secret")
	}
	if len(code) != opts.Digits {
		t.Fatalf("expected code length %d, got %d (code: %s)", opts.Digits, len(code), code)
	}

	if err := svc.Validate(code, secret, opts); err != nil {
		t.Fatalf("Validate returned error for valid 8-digit code: %v", err)
	}
}
