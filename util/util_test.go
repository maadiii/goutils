package util

import (
	"testing"
)

func TestGetPtrValue_WithPointer(t *testing.T) {
	val := 42
	result := GetPtrValue(&val)
	if result != 42 {
		t.Fatalf("expected 42, got %d", result)
	}
}

func TestGetPtrValue_WithNil(t *testing.T) {
	var p *int
	result := GetPtrValue(p)
	if result != 0 {
		t.Fatalf("expected 0 for nil pointer, got %d", result)
	}
}

func TestToPtr(t *testing.T) {
	val := 100
	ptr := ToPtr(val)

	if ptr == nil {
		t.Fatalf("expected non-nil pointer")
	}
	if *ptr != val {
		t.Fatalf("expected pointer to point to %d, got %d", val, *ptr)
	}
}

func TestToPtrOrNil_WithZero(t *testing.T) {
	var val int
	ptr := ToPtrOrNil(val)

	if ptr != nil {
		t.Fatalf("expected nil for zero value")
	}
}

func TestToPtrOrNil_WithNonZero(t *testing.T) {
	val := 50
	ptr := ToPtrOrNil(val)

	if ptr == nil {
		t.Fatalf("expected non-nil pointer")
	}
	if *ptr != val {
		t.Fatalf("expected pointer to point to %d, got %d", val, *ptr)
	}
}

func TestGetPtrValue_String(t *testing.T) {
	str := "hello"
	result := GetPtrValue(&str)
	if result != "hello" {
		t.Fatalf("expected 'hello', got %q", result)
	}
}

func TestToPtr_String(t *testing.T) {
	str := "world"
	ptr := ToPtr(str)

	if ptr == nil {
		t.Fatalf("expected non-nil pointer")
	}
	if *ptr != str {
		t.Fatalf("expected pointer to point to %q, got %q", str, *ptr)
	}
}

func TestToPtrOrNil_String_Empty(t *testing.T) {
	ptr := ToPtrOrNil("")

	if ptr != nil {
		t.Fatalf("expected nil for empty string")
	}
}

func TestToPtrOrNil_String_NonEmpty(t *testing.T) {
	str := "test"
	ptr := ToPtrOrNil(str)

	if ptr == nil {
		t.Fatalf("expected non-nil pointer")
	}
	if *ptr != str {
		t.Fatalf("expected pointer to point to %q, got %q", str, *ptr)
	}
}
