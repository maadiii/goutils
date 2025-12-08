package iter

import (
	"reflect"
	"testing"
)

func TestConcat_EarlyStopDoesNotTouchSecond(t *testing.T) {
	secondTouched := false
	first := Range(1, 5)
	second := Tap(Range(100, 105), func(int) { secondTouched = true })
	out := first.Concat(second).Take(1).ToSlice()
	if !reflect.DeepEqual(out, []int{1}) {
		t.Fatalf("concat early stop out=%v", out)
	}
	if secondTouched {
		t.Fatalf("second sequence should not be touched on early stop")
	}
}

func TestFlatMapTo_EarlyStopStopsInnerAndOuter(t *testing.T) {
	innerYields := 0
	outerCalls := 0
	p := Range(1, 10)
	fm := FlatMapTo(p, func(n int) Pipe[int] {
		outerCalls++
		return Tap(FromSlice([]int{n, n + 1}), func(int) { innerYields++ })
	})
	out := fm.Take(1).ToSlice()
	if len(out) != 1 {
		t.Fatalf("flatMap early stop len=%d", len(out))
	}
	// With Take(1), inner will attempt a second yield which is rejected; Tap sees 2 events.
	if innerYields != 2 {
		t.Fatalf("expected inner to attempt two yields, got %d", innerYields)
	}
	if outerCalls != 1 {
		t.Fatalf("outer should be constructed once, got %d", outerCalls)
	}
}

func TestFlatten_EarlyStopStopsRemainingInners(t *testing.T) {
	firstCount := 0
	secondCount := 0
	pp := FromSlice([]Pipe[int]{
		Tap(FromSlice([]int{1, 2, 3}), func(int) { firstCount++ }),
		Tap(FromSlice([]int{4, 5, 6}), func(int) { secondCount++ }),
	})
	out := Flatten(pp).Take(1).ToSlice()
	if !reflect.DeepEqual(out, []int{1}) {
		t.Fatalf("flatten early stop out=%v", out)
	}
	// Similar to FlatMap, Take(1) rejects on second yield; Tap sees two attempts.
	if firstCount != 2 {
		t.Fatalf("first inner should attempt two yields, got %d", firstCount)
	}
	if secondCount != 0 {
		t.Fatalf("second inner should not run, got %d", secondCount)
	}
}
