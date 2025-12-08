package iter

import (
	"reflect"
	"testing"
)

func TestEdgeBranches_Coverage(t *testing.T) {
	// Take with n<=0
	if got := Range(1, 5).Take(0).ToSlice(); len(got) != 0 {
		t.Fatalf("take(0) should be empty: %v", got)
	}
	// Drop with n<=0 (no-op)
	if got := Range(1, 4).Drop(0).ToSlice(); !reflect.DeepEqual(got, []int{1, 2, 3}) {
		t.Fatalf("drop(0) no-op: %v", got)
	}
	// Slice with negative start and end<start
	if got := Range(1, 5).Slice(-3, -1).ToSlice(); len(got) != 0 {
		t.Fatalf("slice negative should be empty: %v", got)
	}
	// Windows with size<=0
	if w := Windows(Range(1, 5), 0).ToSlice(); len(w) != 0 {
		t.Fatalf("windows(0) should be empty: %v", w)
	}
	// Zip with empty second
	z := ZipTo(Range(1, 4), FromSlice([]int{})).ToSlice()
	if len(z) != 0 {
		t.Fatalf("zip with empty second should be empty: %v", z)
	}
	// Nth with negative index
	if _, ok := Nth(Range(1, 5), -1); ok {
		t.Fatalf("nth(-1) should be false")
	}
	// TakeLast with n<=0 -> empty
	if got := TakeLast(Range(1, 5), 0).ToSlice(); len(got) != 0 {
		t.Fatalf("takeLast(0) should be empty: %v", got)
	}
	// DropLast with n<=0 -> identity
	if got := DropLast(Range(1, 5), 0).ToSlice(); !reflect.DeepEqual(got, []int{1, 2, 3, 4}) {
		t.Fatalf("dropLast(0) identity, got %v", got)
	}
}
