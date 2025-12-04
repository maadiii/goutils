package paseto

import (
	"testing"
	"time"
)

func TestClaimsSetters_ChainAndValues(t *testing.T) {
	now := time.Now()
	c := new(Claims).
		SetAudience("aud").
		SetExpiration(now).
		SetJti("jti").
		SetSubject("sub").
		SetPhoneNumber("+100200300").
		SetTotpSecret("totp123").
		SetStateToken("sttok").
		SetUserID("uid").
		SetRole("admin")

	if c.Audience != "aud" {
		t.Fatalf("Audience mismatch: got %q", c.Audience)
	}
	if !c.Expiration.Equal(now) {
		t.Fatalf("Expiration mismatch: got %v want %v", c.Expiration, now)
	}
	if c.Jti != "jti" || c.Subject != "sub" {
		t.Fatalf("Jti/Subject mismatch: %q %q", c.Jti, c.Subject)
	}
	if c.PhoneNumber != "+100200300" || c.TotpSecret != "totp123" || c.StateToken != "sttok" || c.UserID != "uid" {
		t.Fatalf("one of the string fields did not match expected values")
	}
	if c.Role != "admin" {
		t.Fatalf("Role mismatch: got %q", c.Role)
	}
}

func TestNew_PanicOnInvalidKey(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected New to panic on invalid key")
		}
	}()

	// very short key should cause V4SymmetricKeyFromBytes to return error -> New panics
	_ = New([]byte{1, 2, 3}, "issuer")
}

func TestGenerateAndValidate_SuccessAndRoleNotStored(t *testing.T) {
	// 32 bytes symmetric key for PASETO v4
	key := make([]byte, 32)
	for i := 0; i < 32; i++ {
		key[i] = byte(i + 1)
	}

	p := New(key, "test-issuer")

	exp := time.Now().Add(time.Hour)
	in := &Claims{
		Audience:    "aud",
		Expiration:  exp,
		Jti:         "jti-123",
		Subject:     "sub-1",
		PhoneNumber: "555-0100",
		TotpSecret:  "ts-xyz",
		StateToken:  "st-abc",
		UserID:      "user-1",
		Role:        "superuser",
	}

	tok, err := p.Generate(in)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if tok == "" {
		t.Fatalf("Generate returned empty token")
	}

	out, err := p.Validate(tok)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if out.Audience != in.Audience {
		t.Fatalf("Audience mismatch: got %q want %q", out.Audience, in.Audience)
	}
	if out.Expiration.Unix() != in.Expiration.Unix() {
		t.Fatalf("Expiration mismatch: got %v want %v", out.Expiration, in.Expiration)
	}
	if out.Jti != in.Jti || out.Subject != in.Subject {
		t.Fatalf("Jti/Subject mismatch: got %q %q want %q %q", out.Jti, out.Subject, in.Jti, in.Subject)
	}

	// fields stored via token.Set should roundtrip
	if out.PhoneNumber != in.PhoneNumber || out.TotpSecret != in.TotpSecret || out.StateToken != in.StateToken || out.UserID != in.UserID {
		t.Fatalf("one of the custom fields did not roundtrip as expected")
	}

	// Role is not serialized in Generate/Validate so it should be empty after Validate
	if out.Role != "" {
		t.Fatalf("expected Role to be empty after Validate, got %q", out.Role)
	}
}

func TestValidate_ReturnsErrorForInvalidToken(t *testing.T) {
	key := make([]byte, 32)
	p := New(key, "iss")

	if _, err := p.Validate("not-a-valid-token"); err == nil {
		t.Fatalf("expected error validating invalid token, got nil")
	}
}
