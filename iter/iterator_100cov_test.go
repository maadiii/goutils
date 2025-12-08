package iter

import (
	"reflect"
	"testing"
)

func TestFlatMapSliceTo_EarlyStop(t *testing.T) {
	// Inner loop should stop when downstream stops
	out := FlatMapSliceTo(Range(1, 10), func(n int) []int { return []int{n, n + 1, n + 2} }).
		Take(1).
		ToSlice()
	if !reflect.DeepEqual(out, []int{1}) {
		t.Fatalf("flatMapSlice early stop got %v", out)
	}
}

func TestFlattenSlice_EarlyStop(t *testing.T) {
	out := FlattenSlice(FromSlice([][]int{{1, 2}, {3, 4}})).
		Take(1).
		ToSlice()
	if !reflect.DeepEqual(out, []int{1}) {
		t.Fatalf("flattenSlice early stop got %v", out)
	}
}

func TestChunk_EarlyStopAndEmptyFlush(t *testing.T) {
	// Early stop after first chunk
	chunks := Chunk(Range(1, 10), 3).Take(1).ToSlice()
	if len(chunks) != 1 || !reflect.DeepEqual(chunks[0], []int{1, 2, 3}) {
		t.Fatalf("chunk early stop got %#v", chunks)
	}
	// Empty input triggers final flush with empty buffer path
	none := Chunk(FromSlice([]int{}), 3).ToSlice()
	if len(none) != 0 {
		t.Fatalf("chunk empty got %#v", none)
	}
}

func TestWindows_EarlyStopOnFirstWindow(t *testing.T) {
	wins := Windows(Range(1, 10), 3).Take(1).ToSlice()
	if len(wins) != 1 || !reflect.DeepEqual(wins[0], []int{1, 2, 3}) {
		t.Fatalf("windows early stop got %#v", wins)
	}
}

func TestWindows_EarlyStopOnSlide(t *testing.T) {
	// Skip first window, stop on first slide
	wins := Windows(Range(1, 10), 3).Drop(1).Take(1).ToSlice()
	if len(wins) != 1 || !reflect.DeepEqual(wins[0], []int{2, 3, 4}) {
		t.Fatalf("windows slide early stop got %#v", wins)
	}
}

func TestWindows_StopOnFirstYield(t *testing.T) {
	stopped := false
	Windows(Range(1, 10), 3).Seq()(func(win []int) bool {
		stopped = true
		return false
	})
	if !stopped {
		t.Fatalf("windows manual stop not triggered")
	}
}

func TestZipTo_VariousLengthsAndEarlyStop(t *testing.T) {
	// a shorter than b
	z1 := ZipTo(FromSlice([]int{1}), FromSlice([]string{"a", "b"})).ToSlice()
	if len(z1) != 1 || z1[0].First != 1 || z1[0].Second != "a" {
		t.Fatalf("zip a<b got %#v", z1)
	}
	// early stop
	z2 := ZipTo(Range(1, 10), FromSlice([]int{9, 8, 7})).Take(1).ToSlice()
	if len(z2) != 1 || z2[0].First != 1 || z2[0].Second != 9 {
		t.Fatalf("zip early stop got %#v", z2)
	}
}

func TestTakeLast_BiggerThanLen(t *testing.T) {
	out := TakeLast(Range(1, 3), 10).ToSlice()
	if !reflect.DeepEqual(out, []int{1, 2}) {
		t.Fatalf("takeLast n>len got %v", out)
	}
}

func TestDropLast_EarlyStopAndNoOutputWhenNTooBig(t *testing.T) {
	// Early stop after first emission
	out := DropLast(Range(1, 6), 2).Take(1).ToSlice()
	if !reflect.DeepEqual(out, []int{1}) {
		t.Fatalf("dropLast early stop got %v", out)
	}
	// n greater than length -> no output
	none := DropLast(Range(1, 2), 5).ToSlice()
	if len(none) != 0 {
		t.Fatalf("dropLast n>len should be empty, got %v", none)
	}
}

func TestChunk_SizeZeroActsAsOne(t *testing.T) {
	ch := Chunk(Range(1, 4), 0).ToSlice()
	if len(ch) != 3 || !reflect.DeepEqual(ch[0], []int{1}) || !reflect.DeepEqual(ch[2], []int{3}) {
		t.Fatalf("chunk size 0 got %#v", ch)
	}
}
