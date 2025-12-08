package iter

import "testing"

func TestPipeline_Basics(t *testing.T) {
	got := Range(1, 11).
		Filter(func(n int) bool { return n%3 != 0 }).
		Map(func(n int) int { return n * 2 }).
		Take(4).
		ToSlice()
	want := []int{2, 4, 8, 10}
	if len(got) != len(want) {
		t.Fatalf("len(got)=%d len(want)=%d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("idx %d: got %d want %d", i, got[i], want[i])
		}
	}
}

func TestPipeline_DropAndReduce(t *testing.T) {
	sum := Range(1, 100).
		Drop(10). // drop 1..10
		Take(5).  // 11..15
		Reduce(0, func(a, v int) int { return a + v })
	if sum != (11 + 12 + 13 + 14 + 15) {
		t.Fatalf("got %d, want %d", sum, 65)
	}
}
