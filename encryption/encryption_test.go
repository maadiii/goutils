package encryption

import (
	"errors"
	"testing"
)

// failingReader always returns an error
type failingReader struct{}

func (f *failingReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("random source failed")
}

func TestNewPassword(t *testing.T) {
	p := NewPassword()
	if p == nil {
		t.Fatalf("NewPassword returned nil")
	}
}

func TestPasswordGenerate(t *testing.T) {
	p := NewPassword()
	hash1, err := p.Generate("test_password")
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if hash1 == "" {
		t.Fatalf("expected non-empty hash")
	}
	if hash1 == "test_password" {
		t.Fatalf("hash should not equal password")
	}

	// Two hashes of same password should be different (bcrypt uses random salt)
	hash2, err := p.Generate("test_password")
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if hash1 == hash2 {
		t.Fatalf("bcrypt hashes should be different for same password")
	}
}

func TestPasswordCompare_Success(t *testing.T) {
	p := NewPassword()
	password := "secure_password_123"
	hash, err := p.Generate(password)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	match := p.Compare(hash, password)
	if !match {
		t.Fatalf("expected password to match hash")
	}
}

func TestPasswordCompare_Failed(t *testing.T) {
	p := NewPassword()
	hash, err := p.Generate("password1")
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	match := p.Compare(hash, "password2")
	if match {
		t.Fatalf("expected password mismatch")
	}
}

func TestPasswordCompare_InvalidHash(t *testing.T) {
	p := NewPassword()
	match := p.Compare("not_a_valid_bcrypt_hash", "password")
	if match {
		t.Fatalf("expected false for invalid hash")
	}
}

func TestNewRandom(t *testing.T) {
	r := NewRandom()
	if r == nil {
		t.Fatalf("NewRandom returned nil")
	}
}

func TestRandomString_Length(t *testing.T) {
	r := NewRandom()

	tests := []int{8, 16, 32, 64, 128}
	for _, length := range tests {
		s, err := r.String(length)
		if err != nil {
			t.Fatalf("String(%d) error: %v", length, err)
		}
		if len(s) != length {
			t.Fatalf("expected string of length %d, got %d", length, len(s))
		}
	}
}

func TestRandomString_Uniqueness(t *testing.T) {
	r := NewRandom()

	str1, err1 := r.String(32)
	if err1 != nil {
		t.Fatalf("String error: %v", err1)
	}

	str2, err2 := r.String(32)
	if err2 != nil {
		t.Fatalf("String error: %v", err2)
	}

	if str1 == str2 {
		t.Fatalf("random strings should be different")
	}
}

func TestRandomString_ValidCharacters(t *testing.T) {
	r := NewRandom()
	s, err := r.String(100)
	if err != nil {
		t.Fatalf("String error: %v", err)
	}

	// Check that string only contains valid characters from the charset
	// charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for _, ch := range s {
		isValid := (ch >= '0' && ch <= '9') ||
			(ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z')
		if !isValid {
			t.Fatalf("unexpected character in random string: %c", ch)
		}
	}
}

func TestPasswordGenerateEmpty(t *testing.T) {
	p := NewPassword()
	hash, err := p.Generate("")
	if err != nil {
		t.Fatalf("Generate with empty password returned error: %v", err)
	}
	if hash == "" {
		t.Fatalf("expected non-empty hash even for empty password")
	}
}

func TestPasswordCompareLongPassword(t *testing.T) {
	p := NewPassword()
	longPwd := string(make([]byte, 100))
	for i := range longPwd {
		longPwd = string(append([]byte(longPwd[:i]), 'a'))
	}
	hash, err := p.Generate(longPwd)
	if err != nil {
		t.Logf("Generate with long password returned error: %v", err)
		return
	}
	match := p.Compare(hash, longPwd)
	if !match {
		t.Fatalf("expected long password to match its hash")
	}
}

func TestRandomStringZeroLength(t *testing.T) {
	r := NewRandom()
	s, err := r.String(0)
	if err != nil {
		t.Fatalf("String(0) returned error: %v", err)
	}
	if len(s) != 0 {
		t.Fatalf("expected empty string for length 0")
	}
}

func TestRandomStringErrorHandling(t *testing.T) {
	// Create a random instance with a failing reader
	r := &random{reader: &failingReader{}}
	
	_, err := r.String(10)
	if err == nil {
		t.Fatalf("expected error from failing reader")
	}
	if err.Error() != "random source failed" {
		t.Fatalf("expected 'random source failed' error, got: %v", err)
	}
}
