package iter

import (
	"reflect"
	"testing"
)

func TestDistinctAndDistinctBy(t *testing.T) {
	got := Distinct(FromSlice([]int{1, 2, 2, 3, 1})).ToSlice()
	want := []int{1, 2, 3}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("distinct got %v want %v", got, want)
	}

	type user struct {
		id   int
		name string
	}
	users := FromSlice([]user{{1, "a"}, {2, "b"}, {1, "x"}})
	u := DistinctBy(users, func(u user) int { return u.id }).ToSlice()
	if len(u) != 2 || u[0].id != 1 || u[1].id != 2 {
		t.Fatalf("distinctBy bad: %#v", u)
	}
}

func TestChunkAndWindows(t *testing.T) {
	ch := Chunk(Range(1, 9), 3).ToSlice() // [[1 2 3] [4 5 6] [7 8]]
	if len(ch) != 3 || len(ch[0]) != 3 || len(ch[2]) != 2 || ch[2][1] != 8 {
		t.Fatalf("chunk got %#v", ch)
	}
	wins := Windows(Range(1, 7), 3).ToSlice() // [1 2 3],[2 3 4],[3 4 5],[4 5 6]
	if len(wins) != 4 || wins[0][0] != 1 || wins[3][2] != 6 {
		t.Fatalf("windows got %#v", wins)
	}
}

func TestZipAndGroupBy(t *testing.T) {
	p := ZipTo(Range(1, 4), FromSlice([]string{"a", "b"})).ToSlice()
	if len(p) != 2 || p[1].First != 2 || p[1].Second != "b" {
		t.Fatalf("zip got %#v", p)
	}
	g := GroupBy(Range(1, 8), func(n int) int { return n % 2 }, func(n int) int { return n })
	if len(g[0]) != 3 || len(g[1]) != 4 { // evens: 2,4,6 | odds: 1,3,5,7
		t.Fatalf("groupBy got %#v", g)
	}
}

func TestIndexAndIncludesAndNth(t *testing.T) {
	p := Range(5, 10) // 5..9
	if !Includes(p, 7) {
		t.Fatalf("includes failed")
	}
	if idx := IndexOf(Range(5, 10), 9); idx != 4 {
		t.Fatalf("indexOf got %d", idx)
	}
	if v, ok := Nth(Range(5, 10), 2); !ok || v != 7 {
		t.Fatalf("nth got %v,%v", v, ok)
	}
}

func TestTapReverseTailHead(t *testing.T) {
	s := []int{}
	Tap(Range(1, 4), func(v int) { s = append(s, v) }).ToSlice()
	if !reflect.DeepEqual(s, []int{1, 2, 3}) {
		t.Fatalf("tap side effect bad: %v", s)
	}
	rev := Reverse(Range(1, 5)).Take(3).ToSlice()
	if !reflect.DeepEqual(rev, []int{4, 3, 2}) {
		t.Fatalf("reverse got %v", rev)
	}
	tail := TakeLast(Range(1, 7), 3).ToSlice()
	if !reflect.DeepEqual(tail, []int{4, 5, 6}) {
		t.Fatalf("takeLast got %v", tail)
	}
	front := DropLast(Range(1, 7), 3).ToSlice()
	if !reflect.DeepEqual(front, []int{1, 2, 3}) {
		t.Fatalf("dropLast got %v", front)
	}
}

func TestAggregatesAndSort(t *testing.T) {
	if c := Count(Range(10, 20)); c != 10 {
		t.Fatalf("count got %d", c)
	}
	if sum := SumInts(Range(1, 5)); sum != (1 + 2 + 3 + 4) {
		t.Fatalf("sum got %d", sum)
	}
	if mi, ok := MinInt(FromSlice([]int{5, -1, 10})); !ok || mi != -1 {
		t.Fatalf("min got %d %v", mi, ok)
	}
	if ma, ok := MaxInt(FromSlice([]int{5, -1, 10})); !ok || ma != 10 {
		t.Fatalf("max got %d %v", ma, ok)
	}
	out := SortBy(FromSlice([]int{3, 1, 2}), func(a, b int) bool { return a < b })
	if !reflect.DeepEqual(out, []int{1, 2, 3}) {
		t.Fatalf("sortBy got %v", out)
	}
}
