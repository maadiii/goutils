package iter

import "testing"

func TestPipeline_SomeEveryFind(t *testing.T) {
	p := Range(1, 10)
	if !p.Some(func(v int) bool { return v == 7 }) {
		t.Fatalf("some should find 7")
	}
	if p.Every(func(v int) bool { return v < 9 }) {
		t.Fatalf("every should be false since 9 is not < 9")
	}
	if v, ok := p.Find(func(v int) bool { return v%4 == 0 }); !ok || v != 4 {
		t.Fatalf("find expected 4, got %v %v", v, ok)
	}
}

func TestPipeline_TakeDropWhileSliceConcat(t *testing.T) {
	p := Range(1, 30)
	got := p.DropWhile(func(v int) bool { return v < 10 }).
		TakeWhile(func(v int) bool { return v < 15 }).
		Concat(FromSlice([]int{99, 100})).
		ToSlice()
	want := []int{10, 11, 12, 13, 14, 99, 100}
	if len(got) != len(want) {
		t.Fatalf("len %d want %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("%d: got %d want %d", i, got[i], want[i])
		}
	}

	// Slice equals Drop+Take
	got2 := Range(1, 100).Slice(5, 9).ToSlice() // [start,end) => 6..9
	want2 := []int{6, 7, 8, 9}
	if len(got2) != len(want2) {
		t.Fatalf("len %d want %d", len(got2), len(want2))
	}
	for i := range got2 {
		if got2[i] != want2[i] {
			t.Fatalf("%d: got %d want %d", i, got2[i], want2[i])
		}
	}
}

func TestPipeline_AsIndexedPairs_FlatMap(t *testing.T) {
	pairs := AsIndexedPairs(Range(5, 8)).ToSlice()
	if len(pairs) != 3 || pairs[0].Index != 0 || pairs[0].Value != 5 || pairs[2].Index != 2 || pairs[2].Value != 7 {
		t.Fatalf("pairs mismatch: %#v", pairs)
	}

	// FlatMapSliceTo: for each n, emit [n, n*10]
	fm := FlatMapSliceTo(Range(1, 4), func(n int) []int { return []int{n, n * 10} }).ToSlice()
	want := []int{1, 10, 2, 20, 3, 30}
	if len(fm) != len(want) {
		t.Fatalf("len %d want %d", len(fm), len(want))
	}
	for i := range fm {
		if fm[i] != want[i] {
			t.Fatalf("%d: got %d want %d", i, fm[i], want[i])
		}
	}

	// Flatten over Pipe[Pipe[int]]
	pp := MapTo(Range(1, 3), func(n int) Pipe[int] { return FromSlice([]int{n, -n}) })
	flat := Flatten(pp).ToSlice()
	wantFlat := []int{1, -1, 2, -2}
	if len(flat) != len(wantFlat) {
		t.Fatalf("len %d want %d", len(flat), len(wantFlat))
	}
	for i := range flat {
		if flat[i] != wantFlat[i] {
			t.Fatalf("%d: got %d want %d", i, flat[i], wantFlat[i])
		}
	}
}
