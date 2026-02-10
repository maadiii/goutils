package buffer

import (
	"sync"
)

type buffer[T any] struct {
	items     []T
	head, len int64
	mask      int64
}

type GrowingRing[T any] struct {
	content *buffer[T]
	mu      sync.Mutex
}

// NewGrowingRing creates a growing ring buffer with the given capacity (rounded up to nearest power of 2).
func NewGrowingRing[T any](capacity int64) *GrowingRing[T] {
	if capacity <= 0 {
		capacity = 1
	}
	// Round up to next power of 2
	size := int64(1)
	for size < capacity {
		size <<= 1
	}

	return &GrowingRing[T]{
		content: &buffer[T]{
			items: make([]T, size),
			head:  0,
			len:   0,
			mask:  size - 1,
		},
	}
}

// Push appends an item to the buffer, growing if necessary.
func (rb *GrowingRing[T]) Push(item T) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	c := rb.content

	// Grow if full
	if c.len == int64(len(c.items)) {
		newSize := int64(len(c.items)) << 1
		newItems := make([]T, newSize)
		// Copy existing items in order
		for i := int64(0); i < c.len; i++ {
			newItems[i] = c.items[(c.head+i)&c.mask]
		}
		c = &buffer[T]{
			items: newItems,
			head:  0,
			len:   c.len,
			mask:  newSize - 1,
		}
		rb.content = c
	}

	// Append at tail
	c.items[(c.head+c.len)&c.mask] = item
	c.len++
}

// Len returns the number of items in the buffer.
func (rb *GrowingRing[T]) Len() int64 {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	length := rb.content.len

	return length
}

// Pop removes and returns the first item, or false if empty.
func (rb *GrowingRing[T]) Pop() (T, bool) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	c := rb.content
	if c.len == 0 {
		var t T
		return t, false
	}
	item := c.items[c.head]
	var t T
	c.items[c.head] = t // clear for GC
	c.head = (c.head + 1) & c.mask
	c.len--

	return item, true
}

// PopN removes up to n items and returns them, or false if empty.
func (rb *GrowingRing[T]) PopN(n int64) ([]T, bool) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	c := rb.content
	if c.len == 0 {
		return nil, false
	}

	if n > c.len {
		n = c.len
	}

	items := make([]T, n)
	for i := int64(0); i < n; i++ {
		idx := (c.head + i) & c.mask
		items[i] = c.items[idx]
		var t T
		c.items[idx] = t // clear for GC
	}
	c.head = (c.head + n) & c.mask
	c.len -= n

	return items, true
}
