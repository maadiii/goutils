package buffer_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/maadiii/goutils/buffer"
)

type Item struct {
	i int
}

func TestPushPop(t *testing.T) {
	rb := buffer.NewGrowingRing[Item](1024)
	for i := range 5000 {
		rb.Push(Item{i})
		item, ok := rb.Pop()
		if ok {
			if item.i != i {
				t.Fatal("invalid item popped")
			}
		}
	}
}

func TestPushPopN(t *testing.T) {
	rb := buffer.NewGrowingRing[Item](1024)
	n := 5000
	for i := range n {
		rb.Push(Item{i})
	}
	items, ok := rb.PopN(int64(n))
	if !ok {
		t.Fatal("expected to pop many items")
	}
	for i := range n {
		if items[i].i != i {
			t.Fatal("invalid item popped")
		}
	}
}

func TestPopThreadSafety(t *testing.T) {
	t.Run("Pop should be thread-safe", func(t *testing.T) {
		testCase := func() {
			rb := buffer.NewGrowingRing[int](4)
			rb.Push(1)
			wg := sync.WaitGroup{}
			for range 2 {
				wg.Add(1)
				go func() {
					defer wg.Done()
					rb.Pop()
				}()
			}
			wg.Wait()
			if rb.Len() == -1 {
				t.Fatal("item popped twice")
			}
		}

		// Increase the number of iterations to raise the likelihood of reproducing the race condition
		for range 100_000 {
			testCase()
		}
	})

	t.Run("PopN should be thread-safe", func(t *testing.T) {
		testCase := func() {
			rb := buffer.NewGrowingRing[int](4)
			rb.Push(1)
			counter := atomic.Int32{}
			wg := sync.WaitGroup{}
			for range 2 {
				wg.Add(1)
				go func() {
					defer wg.Done()
					_, ok := rb.PopN(1)
					if ok {
						counter.Add(1)
					}
				}()
			}
			wg.Wait()
			if counter.Load() > 1 {
				t.Fatal("false positive item removal")
			}
		}

		// Increase the number of iterations to raise the likelihood of reproducing the race condition
		for range 100_000 {
			testCase()
		}
	})
}

func TestConcurrentStress(t *testing.T) {
	t.Run("Concurrent push/pop under high load", func(t *testing.T) {
		rb := buffer.NewGrowingRing[int](16)
		const numGoroutines = 10
		const opsPerGoroutine = 1000

		wg := sync.WaitGroup{}

		// Producers: push items
		for g := range numGoroutines {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for i := range opsPerGoroutine {
					rb.Push(id*opsPerGoroutine + i)
				}
			}(g)
		}

		// Consumers: pop items
		for g := range numGoroutines {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				popCount := 0
				for popCount < opsPerGoroutine {
					if _, ok := rb.Pop(); ok {
						popCount++
					}
				}
			}(g)
		}

		wg.Wait()

		// Final state: all items should be consumed
		if rb.Len() != 0 {
			t.Fatalf("expected empty buffer, got len=%d", rb.Len())
		}
	})
}

func TestConcurrentGrowth(t *testing.T) {
	t.Run("Buffer grows safely under concurrent push", func(t *testing.T) {
		rb := buffer.NewGrowingRing[int](2) // Start small to force growth
		const numGoroutines = 8
		const itemsPerGoroutine = 500

		wg := sync.WaitGroup{}
		for g := range numGoroutines {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for i := range itemsPerGoroutine {
					defer func() {
						if r := recover(); r != nil {
							t.Errorf("panic during growth in goroutine %d: %v", id, r)
						}
					}()
					rb.Push(id*itemsPerGoroutine + i)
				}
			}(g)
		}

		wg.Wait()

		// Verify total count
		expectedLen := int64(numGoroutines * itemsPerGoroutine)
		if rb.Len() != expectedLen {
			t.Fatalf("expected len=%d, got len=%d after growth", expectedLen, rb.Len())
		}

		// Drain and verify no corruption
		count := int64(0)
		for {
			_, ok := rb.Pop()
			if !ok {
				break
			}
			count++
		}
		if count != expectedLen {
			t.Fatalf("drained %d items, expected %d", count, expectedLen)
		}
	})
}

func TestConcurrentPopN(t *testing.T) {
	t.Run("PopN is safe with concurrent pushes", func(t *testing.T) {
		rb := buffer.NewGrowingRing[int](32)
		const totalItems = 5000

		// Push all items first
		for i := range totalItems {
			rb.Push(i)
		}

		wg := sync.WaitGroup{}
		popCount := atomic.Int64{}

		// Multiple consumers using PopN
		for g := range 4 {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for {
					items, ok := rb.PopN(100)
					if !ok {
						break
					}
					popCount.Add(int64(len(items)))
					if len(items) == 0 {
						break
					}
				}
			}(g)
		}

		wg.Wait()

		if popCount.Load() != totalItems {
			t.Fatalf("expected to pop %d items, got %d", totalItems, popCount.Load())
		}

		if rb.Len() != 0 {
			t.Fatalf("buffer should be empty, len=%d", rb.Len())
		}
	})
}

func TestConcurrentLenSafety(t *testing.T) {
	t.Run("Len() is safe and consistent", func(t *testing.T) {
		rb := buffer.NewGrowingRing[int](64)
		const iterations = 10000

		wg := sync.WaitGroup{}

		// Pusher
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range iterations {
				rb.Push(i)
			}
		}()

		// Popper
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range iterations {
				rb.Pop()
			}
		}()

		// Observer: Len() should never be negative or exceed pushed count
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range iterations * 2 {
				len := rb.Len()
				if len < 0 {
					t.Errorf("negative length observed: %d", len)
				}
				if len > iterations {
					t.Errorf("length exceeded total items: %d > %d", len, iterations)
				}
			}
		}()

		wg.Wait()
	})
}

func BenchmarkPush(b *testing.B) {
	rb := buffer.NewGrowingRing[int](1024)

	for i := 0; b.Loop(); i++ {
		rb.Push(i)
	}
}

func BenchmarkPop(b *testing.B) {
	rb := buffer.NewGrowingRing[int](1024)
	for i := 0; i < b.N; i++ {
		rb.Push(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.Pop()
	}
}

func BenchmarkPopN(b *testing.B) {
	rb := buffer.NewGrowingRing[int](1024)
	items := 100
	for i := 0; i < b.N*items; i++ {
		rb.Push(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.PopN(int64(items))
	}
}

func BenchmarkGrowth(b *testing.B) {
	rb := buffer.NewGrowingRing[int](16)

	for i := 0; b.Loop(); i++ {
		rb.Push(i)
	}
}
