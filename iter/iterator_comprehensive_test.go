package iter

import (
	"fmt"
	"reflect"
	"testing"
)

func TestSeqAndFromSlice(t *testing.T) {
	p := FromSlice([]int{1, 2, 3})
	var out []int
	p.Seq()(func(v int) bool { out = append(out, v); return true })
	if !reflect.DeepEqual(out, []int{1, 2, 3}) {
		t.Fatalf("seq/fromSlice got %v", out)
	}

	var out2 []int
	Range(3, 6).Seq()(func(v int) bool { out2 = append(out2, v); return true })
	if !reflect.DeepEqual(out2, []int{3, 4, 5}) {
		t.Fatalf("seq/range got %v", out2)
	}
}

func TestReduceTo(t *testing.T) {
	// sum ints into float64 accumulator
	sum := ReduceTo(Range(1, 5), 0.0, func(a float64, v int) float64 { return a + float64(v) })
	if sum != 10.0 {
		t.Fatalf("reduceTo sum got %v", sum)
	}
}

func TestFlatMapTo(t *testing.T) {
	// For each n in 1..3, emit [n, n+1]
	fm := FlatMapTo(Range(1, 4), func(n int) Pipe[int] { return Range(n, n+2) }).ToSlice()
	want := []int{1, 2, 2, 3, 3, 4}
	if !reflect.DeepEqual(fm, want) {
		t.Fatalf("flatMapTo got %v want %v", fm, want)
	}
}

func TestFlattenSlice(t *testing.T) {
	p := FromSlice([][]int{{1, 2}, {3}, {}})
	flat := FlattenSlice(p).ToSlice()
	if !reflect.DeepEqual(flat, []int{1, 2, 3}) {
		t.Fatalf("flattenSlice got %v", flat)
	}
}

func TestFirstLast(t *testing.T) {
	if v, ok := First(Range(10, 15)); !ok || v != 10 {
		t.Fatalf("first got %v %v", v, ok)
	}
	if v, ok := Last(Range(10, 15)); !ok || v != 14 {
		t.Fatalf("last got %v %v", v, ok)
	}
	if _, ok := First(FromSlice([]int{})); ok {
		t.Fatalf("first on empty should be false")
	}
	if _, ok := Last(FromSlice([]int{})); ok {
		t.Fatalf("last on empty should be false")
	}
}

func TestIndexOfBy(t *testing.T) {
	idx := IndexOfBy(Range(1, 10), func(v int) bool { return v%4 == 0 }) // element 4 at index 3
	if idx != 3 {
		t.Fatalf("indexOfBy got %d", idx)
	}
	idx2 := IndexOfBy(Range(1, 5), func(v int) bool { return v > 10 })
	if idx2 != -1 {
		t.Fatalf("indexOfBy expected -1 got %d", idx2)
	}
}

func TestForEach(t *testing.T) {
	acc := 0
	Range(1, 5).ForEach(func(v int) { acc += v }) // 1+2+3+4
	if acc != 10 {
		t.Fatalf("forEach sum got %d", acc)
	}
}

func TestMapTo(t *testing.T) {
	ss := MapTo(Range(1, 4), func(n int) string { return fmt.Sprintf("n%02d", n) }).ToSlice()
	want := []string{"n01", "n02", "n03"}
	if !reflect.DeepEqual(ss, want) {
		t.Fatalf("mapTo got %v want %v", ss, want)
	}
}
