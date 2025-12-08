package iter

import (
	"testing"
)

func BenchmarkRangeMapFilterToSlice(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Range(0, 10000).
			Map(func(n int) int { return n + 1 }).
			Filter(func(n int) bool { return n%3 != 0 }).
			ToSlice()
	}
}

func BenchmarkWindows(b *testing.B) {
	p := Range(0, 5000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Windows(p, 5).Drop(100).Take(100).ToSlice()
	}
}

func BenchmarkDropLast(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = DropLast(Range(0, 10000), 100).Take(100).ToSlice()
	}
}

func BenchmarkFlatMapTo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = FlatMapTo(Range(0, 5000), func(n int) Pipe[int] { return Range(n, n+2) }).Take(100).ToSlice()
	}
}
