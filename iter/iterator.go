package iter

import (
	"iter"
	"sort"
)

// Pipe is a chainable wrapper around iter.Seq[T].
type Pipe[T any] struct {
	s iter.Seq[T]
}

// FromSlice constructs a Pipe from a slice.
func FromSlice[T any](in []T) Pipe[T] {
	return Pipe[T]{s: func(yield func(T) bool) {
		for _, v := range in {
			if !yield(v) {
				return
			}
		}
	}}
}

// Range returns numbers [start, end) in a sequence.
func Range(start, end int) Pipe[int] {
	return Pipe[int]{s: func(yield func(int) bool) {
		for i := start; i < end; i++ {
			if !yield(i) {
				return
			}
		}
	}}
}

// Seq exposes the underlying iter.Seq[T].
func (p Pipe[T]) Seq() iter.Seq[T] { return p.s }

// Map transforms values.
func (p Pipe[T]) Map(f func(T) T) Pipe[T] {
	return Pipe[T]{s: func(y func(T) bool) {
		p.s(func(v T) bool { return y(f(v)) })
	}}
}

// MapTo transforms values to a new type using a free function (since
// methods with type parameters may not be supported in your Go version).
func MapTo[T, U any](p Pipe[T], f func(T) U) Pipe[U] {
	return Pipe[U]{s: func(y func(U) bool) {
		p.s(func(v T) bool { return y(f(v)) })
	}}
}

// Filter keeps values that satisfy pred.
func (p Pipe[T]) Filter(pred func(T) bool) Pipe[T] {
	return Pipe[T]{s: func(y func(T) bool) {
		p.s(func(v T) bool {
			if pred(v) {
				return y(v)
			}
			return true
		})
	}}
}

// Take yields at most n values.
func (p Pipe[T]) Take(n int) Pipe[T] {
	return Pipe[T]{s: func(y func(T) bool) {
		if n <= 0 {
			return
		}
		count := 0
		p.s(func(v T) bool {
			if count >= n {
				return false
			}
			count++
			return y(v)
		})
	}}
}

// Drop skips the first n values.
func (p Pipe[T]) Drop(n int) Pipe[T] {
	return Pipe[T]{s: func(y func(T) bool) {
		count := 0
		p.s(func(v T) bool {
			if count < n {
				count++
				return true
			}
			return y(v)
		})
	}}
}

// ForEach consumes the sequence with side effects.
func (p Pipe[T]) ForEach(f func(T)) {
	p.s(func(v T) bool { f(v); return true })
}

// ToSlice collects all values into a slice.
func (p Pipe[T]) ToSlice() []T {
	out := []T{}
	p.s(func(v T) bool { out = append(out, v); return true })
	return out
}

// Reduce accumulates values using f starting with acc.
func (p Pipe[T]) Reduce(acc T, f func(T, T) T) T {
	p.s(func(v T) bool { acc = f(acc, v); return true })
	return acc
}

// ReduceTo reduces to a different accumulator type as a free function.
func ReduceTo[T, A any](p Pipe[T], acc A, f func(A, T) A) A {
	p.s(func(v T) bool { acc = f(acc, v); return true })
	return acc
}

// Concat yields p then other.
func (p Pipe[T]) Concat(other Pipe[T]) Pipe[T] {
	return Pipe[T]{s: func(y func(T) bool) {
		stopped := false
		p.s(func(v T) bool {
			if !y(v) {
				stopped = true
				return false
			}
			return true
		})
		if !stopped {
			other.s(y)
		}
	}}
}

// Some returns true if any value satisfies pred.
func (p Pipe[T]) Some(pred func(T) bool) bool {
	found := false
	p.s(func(v T) bool {
		if pred(v) {
			found = true
			return false
		}
		return true
	})
	return found
}

// Every returns true if all values satisfy pred (vacuously true for empty).
func (p Pipe[T]) Every(pred func(T) bool) bool {
	ok := true
	p.s(func(v T) bool {
		if !pred(v) {
			ok = false
			return false
		}
		return true
	})
	return ok
}

// Find returns the first value that satisfies pred and true; otherwise zero value and false.
func (p Pipe[T]) Find(pred func(T) bool) (T, bool) {
	var out T
	var ok bool
	p.s(func(v T) bool {
		if pred(v) {
			out = v
			ok = true
			return false
		}
		return true
	})
	return out, ok
}

// TakeWhile yields values while pred returns true.
func (p Pipe[T]) TakeWhile(pred func(T) bool) Pipe[T] {
	return Pipe[T]{s: func(y func(T) bool) {
		p.s(func(v T) bool {
			if !pred(v) {
				return false
			}
			return y(v)
		})
	}}
}

// DropWhile skips values while pred returns true, then yields the rest.
func (p Pipe[T]) DropWhile(pred func(T) bool) Pipe[T] {
	return Pipe[T]{s: func(y func(T) bool) {
		dropping := true
		p.s(func(v T) bool {
			if dropping {
				if pred(v) {
					return true
				}
				dropping = false
			}
			return y(v)
		})
	}}
}

// Slice selects [start, end) by combining Drop and Take.
func (p Pipe[T]) Slice(start, end int) Pipe[T] {
	if start < 0 {
		start = 0
	}
	if end < start {
		end = start
	}
	return p.Drop(start).Take(end - start)
}

// Pair is returned by AsIndexedPairs.
type Pair[T any] struct {
	Index int
	Value T
}

// AsIndexedPairs attaches indices to values as free function.
func AsIndexedPairs[T any](p Pipe[T]) Pipe[Pair[T]] {
	return Pipe[Pair[T]]{s: func(y func(Pair[T]) bool) {
		i := 0
		p.s(func(v T) bool {
			ok := y(Pair[T]{Index: i, Value: v})
			i++
			return ok
		})
	}}
}

// FlatMapTo maps each value to a Pipe[U] and flattens.
func FlatMapTo[T, U any](p Pipe[T], f func(T) Pipe[U]) Pipe[U] {
	return Pipe[U]{s: func(y func(U) bool) {
		stopped := false
		p.s(func(v T) bool {
			inner := f(v)
			inner.s(func(u U) bool {
				if !y(u) {
					stopped = true
					return false
				}
				return true
			})
			return !stopped
		})
	}}
}

// FlatMapSliceTo maps to []U and flattens.
func FlatMapSliceTo[T, U any](p Pipe[T], f func(T) []U) Pipe[U] {
	return Pipe[U]{s: func(y func(U) bool) {
		p.s(func(v T) bool {
			for _, u := range f(v) {
				if !y(u) {
					return false
				}
			}
			return true
		})
	}}
}

// Flatten flattens one level from Pipe[Pipe[T]] to Pipe[T].
func Flatten[T any](pp Pipe[Pipe[T]]) Pipe[T] {
	return Pipe[T]{s: func(y func(T) bool) {
		stopped := false
		pp.s(func(inner Pipe[T]) bool {
			inner.s(func(v T) bool {
				if !y(v) {
					stopped = true
					return false
				}
				return true
			})
			return !stopped
		})
	}}
}

// FlattenSlice flattens Pipe[[]T] to Pipe[T].
func FlattenSlice[T any](p Pipe[[]T]) Pipe[T] {
	return Pipe[T]{s: func(y func(T) bool) {
		p.s(func(arr []T) bool {
			for _, v := range arr {
				if !y(v) {
					return false
				}
			}
			return true
		})
	}}
}

// Distinct keeps first occurrence by key.
func DistinctBy[T any, K comparable](p Pipe[T], key func(T) K) Pipe[T] {
	return Pipe[T]{s: func(y func(T) bool) {
		seen := make(map[K]struct{})
		p.s(func(v T) bool {
			k := key(v)
			if _, ok := seen[k]; ok {
				return true
			}
			seen[k] = struct{}{}
			return y(v)
		})
	}}
}

func Distinct[T comparable](p Pipe[T]) Pipe[T] {
	return DistinctBy(p, func(v T) T { return v })
}

// Chunk groups values into fixed-size slices; last chunk may be smaller.
func Chunk[T any](p Pipe[T], size int) Pipe[[]T] {
	if size <= 0 {
		size = 1
	}
	return Pipe[[]T]{s: func(y func([]T) bool) {
		buf := make([]T, 0, size)
		flush := func() bool {
			if len(buf) == 0 {
				return true
			}
			out := make([]T, len(buf))
			copy(out, buf)
			buf = buf[:0]
			return y(out)
		}
		p.s(func(v T) bool {
			buf = append(buf, v)
			if len(buf) == size {
				if !flush() {
					return false
				}
			}
			return true
		})
		flush()
	}}
}

// Windows emits sliding windows of given size (step=1). If size<=0, returns empty.
func Windows[T any](p Pipe[T], size int) Pipe[[]T] {
	if size <= 0 {
		return Pipe[[]T]{s: func(func([]T) bool) {}}
	}
	return Pipe[[]T]{s: func(y func([]T) bool) {
		buf := make([]T, 0, size)
		p.s(func(v T) bool {
			if len(buf) < size {
				buf = append(buf, v)
				if len(buf) == size {
					win := make([]T, size)
					copy(win, buf)
					if !y(win) {
						return false
					}
				}
				return true
			}
			// slide
			copy(buf, buf[1:])
			buf[len(buf)-1] = v
			win := make([]T, size)
			copy(win, buf)
			return y(win)
		})
	}}
}

// Pair2 holds two values for ZipTo.
type Pair2[A, B any] struct {
	First  A
	Second B
}

// ZipTo zips two sequences by materializing them and pairing up to min length.
func ZipTo[A, B any](a Pipe[A], b Pipe[B]) Pipe[Pair2[A, B]] {
	as := a.ToSlice()
	bs := b.ToSlice()
	n := len(as)
	if len(bs) < n {
		n = len(bs)
	}
	return Pipe[Pair2[A, B]]{s: func(y func(Pair2[A, B]) bool) {
		for i := 0; i < n; i++ {
			if !y(Pair2[A, B]{First: as[i], Second: bs[i]}) {
				return
			}
		}
	}}
}

// Count counts values.
func Count[T any](p Pipe[T]) int {
	c := 0
	p.s(func(T) bool { c++; return true })
	return c
}

// First returns first element if any.
func First[T any](p Pipe[T]) (T, bool) {
	var out T
	ok := false
	p.s(func(v T) bool { out = v; ok = true; return false })
	return out, ok
}

// Last returns last element if any.
func Last[T any](p Pipe[T]) (T, bool) {
	var out T
	ok := false
	p.s(func(v T) bool { out = v; ok = true; return true })
	return out, ok
}

// Nth returns element at zero-based index n.
func Nth[T any](p Pipe[T], n int) (T, bool) {
	var out T
	if n < 0 {
		return out, false
	}
	idx := 0
	ok := false
	p.s(func(v T) bool {
		if idx == n {
			out = v
			ok = true
			return false
		}
		idx++
		return true
	})
	return out, ok
}

// Includes reports whether target exists (requires comparable).
func Includes[T comparable](p Pipe[T], target T) bool {
	return p.Some(func(v T) bool { return v == target })
}

// IndexOfBy returns index of first value satisfying pred, or -1.
func IndexOfBy[T any](p Pipe[T], pred func(T) bool) int {
	idx := -1
	i := 0
	p.s(func(v T) bool {
		if pred(v) {
			idx = i
			return false
		}
		i++
		return true
	})
	return idx
}

// IndexOf returns index of target or -1.
func IndexOf[T comparable](p Pipe[T], target T) int {
	return IndexOfBy(p, func(v T) bool { return v == target })
}

// GroupBy groups into map[K][]V with key/value selectors.
func GroupBy[T any, K comparable, V any](p Pipe[T], key func(T) K, val func(T) V) map[K][]V {
	m := make(map[K][]V)
	p.s(func(v T) bool {
		k := key(v)
		m[k] = append(m[k], val(v))
		return true
	})
	return m
}

// Tap applies a side effect and passes values through.
func Tap[T any](p Pipe[T], f func(T)) Pipe[T] {
	return Pipe[T]{s: func(y func(T) bool) {
		p.s(func(v T) bool { f(v); return y(v) })
	}}
}

// Reverse materializes and yields reversed.
func Reverse[T any](p Pipe[T]) Pipe[T] {
	s := p.ToSlice()
	return Pipe[T]{s: func(y func(T) bool) {
		for i := len(s) - 1; i >= 0; i-- {
			if !y(s[i]) {
				return
			}
		}
	}}
}

// TakeLast materializes and yields last n elements.
func TakeLast[T any](p Pipe[T], n int) Pipe[T] {
	if n <= 0 {
		return Pipe[T]{s: func(func(T) bool) {}}
	}
	s := p.ToSlice()
	if n > len(s) {
		n = len(s)
	}
	tail := make([]T, n)
	copy(tail, s[len(s)-n:])
	return FromSlice(tail)
}

// DropLast yields all but the last n values using a ring buffer.
func DropLast[T any](p Pipe[T], n int) Pipe[T] {
	if n <= 0 {
		return p
	}
	return Pipe[T]{s: func(y func(T) bool) {
		buf := make([]T, 0, n)
		p.s(func(v T) bool {
			if len(buf) < n {
				buf = append(buf, v)
				return true
			}
			// emit oldest, then push v
			if !y(buf[0]) {
				return false
			}
			copy(buf, buf[1:])
			buf[len(buf)-1] = v
			return true
		})
	}}
}

// SumInts sums ints.
func SumInts(p Pipe[int]) int { return p.Reduce(0, func(a, v int) int { return a + v }) }

// MinInt returns min and ok.
func MinInt(p Pipe[int]) (int, bool) {
	var min int
	ok := false
	p.s(func(v int) bool {
		if !ok || v < min {
			min = v
			ok = true
		}
		return true
	})
	return min, ok
}

// MaxInt returns max and ok.
func MaxInt(p Pipe[int]) (int, bool) {
	var max int
	ok := false
	p.s(func(v int) bool {
		if !ok || v > max {
			max = v
			ok = true
		}
		return true
	})
	return max, ok
}

// SortBy materializes and sorts with less comparator, returning a slice.
func SortBy[T any](p Pipe[T], less func(a, b T) bool) []T {
	out := p.ToSlice()
	sort.Slice(out, func(i, j int) bool { return less(out[i], out[j]) })
	return out
}
