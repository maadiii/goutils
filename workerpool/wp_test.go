package workerpool_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	wp "workerpool"
)

// ---------------------- Simple Job ----------------------
type simpleJob struct {
	id int
}

func (j simpleJob) Process(ctx context.Context) (int, error) {
	return j.id, nil
}
func (j simpleJob) ID() int                   { return j.id }
func (j simpleJob) MaxRetries() int           { return 0 }
func (j simpleJob) RetryDelay() time.Duration { return 0 }

// ---------------------- Retry Job ----------------------
type retryJob struct {
	id       int
	failures int32
}

func (j *retryJob) Process(ctx context.Context) (int, error) {
	if atomic.AddInt32(&j.failures, 1) <= 2 {
		return 0, errors.New("temporary error")
	}
	return j.id, nil
}
func (j *retryJob) ID() int                   { return j.id }
func (j *retryJob) MaxRetries() int           { return 3 }
func (j *retryJob) RetryDelay() time.Duration { return 1 * time.Millisecond }

// ---------------------- Timeout Job ----------------------
type timeoutJob struct{}

func (j timeoutJob) Process(ctx context.Context) (int, error) {
	select {
	case <-time.After(50 * time.Millisecond):
		return 42, nil
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}
func (j timeoutJob) ID() int                   { return 1 }
func (j timeoutJob) MaxRetries() int           { return 0 }
func (j timeoutJob) RetryDelay() time.Duration { return 0 }

// ---------------------- Panic Job ----------------------
type panicJob struct {
	id int
}

func (j panicJob) Process(ctx context.Context) (int, error) {
	panic("job panicked")
}
func (j panicJob) ID() int                   { return j.id }
func (j panicJob) MaxRetries() int           { return 0 }
func (j panicJob) RetryDelay() time.Duration { return 0 }

// ---------------------- Blocking Job ----------------------
type blockingJob struct {
	id      int
	blocker chan struct{}
}

func (b blockingJob) Process(ctx context.Context) (int, error) {
	<-b.blocker
	return b.id, nil
}
func (blockingJob) ID() int                   { return -1 }
func (blockingJob) MaxRetries() int           { return 0 }
func (blockingJob) RetryDelay() time.Duration { return 0 }

// ---------------------- Tests ----------------------

func TestWorkerPool_Basic(t *testing.T) {
	ctx := context.Background()
	pool := wp.NewWorkerPool[int](ctx, wp.DefaultWorkerPoolConfig())
	defer pool.Close(false)

	job := simpleJob{id: 42}
	if err := pool.Submit(job); err != nil {
		t.Fatal(err)
	}

	result := <-pool.Results()
	if result.Value != 42 {
		t.Fatalf("expected 42, got %v", result.Value)
	}
	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}
}

func TestWorkerPool_Concurrency(t *testing.T) {
	ctx := context.Background()
	pool := wp.NewWorkerPool[int](ctx, wp.DefaultWorkerPoolConfig())
	defer pool.Close(false)

	const jobs = 50
	for i := range jobs {
		_ = pool.Submit(simpleJob{id: i})
	}

	results := make(map[int]bool)
	for range jobs {
		r := <-pool.Results()
		results[r.Value] = true
	}
	for i := range jobs {
		if !results[i] {
			t.Fatalf("missing result for job %d", i)
		}
	}
}

func TestWorkerPool_Retry(t *testing.T) {
	ctx := context.Background()
	pool := wp.NewWorkerPool[int](ctx, wp.DefaultWorkerPoolConfig())
	defer pool.Close(false)

	job := &retryJob{id: 7}
	_ = pool.Submit(job)

	result := <-pool.Results()
	if result.Value != 7 {
		t.Fatalf("expected 7, got %v", result.Value)
	}
	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}
}

func TestWorkerPool_Timeout(t *testing.T) {
	ctx := context.Background()
	cfg := wp.DefaultWorkerPoolConfig()
	cfg.JobProcessTimeout = 10 * time.Millisecond
	pool := wp.NewWorkerPool[int](ctx, cfg)
	defer pool.Close(false)

	_ = pool.Submit(timeoutJob{})

	result := <-pool.Results()
	if result.Error == nil || result.Error != context.DeadlineExceeded {
		t.Fatalf("expected DeadlineExceeded, got %v", result.Error)
	}
}

func TestWorkerPool_PanicSafety(t *testing.T) {
	ctx := context.Background()
	pool := wp.NewWorkerPool[int](ctx, wp.DefaultWorkerPoolConfig())
	defer pool.Close(false)

	_ = pool.Submit(panicJob{id: 99})

	result := <-pool.Results()
	if result.Error == nil {
		t.Fatalf("expected panic error, got nil")
	}
	if result.JobID != 99 {
		t.Fatalf("expected JobID 99, got %d", result.JobID)
	}
}

func TestWorkerPool_GracefulShutdown(t *testing.T) {
	ctx := context.Background()
	cfg := wp.DefaultWorkerPoolConfig()
	cfg.NumWorkers = 2
	pool := wp.NewWorkerPool[int](ctx, cfg)

	const jobs = 10
	for i := range jobs {
		_ = pool.Submit(simpleJob{id: i})
	}

	go pool.Close(false)

	count := 0
	for range pool.Results() {
		count++
	}
	if count != jobs {
		t.Fatalf("expected %d results, got %d", jobs, count)
	}
}

// blockingJobWithCtx respects ctx.Done()
type blockingJobWithCtx struct {
	id int
}

func (b *blockingJobWithCtx) Process(ctx context.Context) (int, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case <-time.After(5 * time.Second): // simulate work
	}
	return b.id, nil
}

func (b *blockingJobWithCtx) ID() int                   { return b.id }
func (b *blockingJobWithCtx) MaxRetries() int           { return 0 }
func (b *blockingJobWithCtx) RetryDelay() time.Duration { return 0 }

func TestWorkerPool_ForceShutdown_Safe(t *testing.T) {
	ctx := context.Background()
	cfg := wp.DefaultWorkerPoolConfig()
	cfg.NumWorkers = 2
	cfg.ResultQueueSize = 10
	pool := wp.NewWorkerPool[int](ctx, cfg)

	for i := range 5 {
		_ = pool.Submit(&blockingJobWithCtx{id: i})
	}

	done := make(chan struct{})
	go func() {
		pool.Close(true)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("force shutdown did not complete in time")
	}
}

func TestWorkerPool_PendingCounters(t *testing.T) {
	ctx := context.Background()
	cfg := wp.DefaultWorkerPoolConfig()
	cfg.NumWorkers = 1
	pool := wp.NewWorkerPool[int](ctx, cfg)
	defer pool.Close(false)

	blocker := make(chan struct{}, 5)
	const jobs = 5
	for i := range jobs {
		_ = pool.Submit(blockingJob{id: i, blocker: blocker})
	}

	pendingJobs := pool.PendingJobs()
	if pendingJobs != jobs {
		t.Fatalf("expected %d pending jobs, got %d", jobs, pendingJobs)
	}

	for range jobs {
		blocker <- struct{}{}
		<-pool.Results()
	}
}

func TestWorkerPool_SubmitAfterClose(t *testing.T) {
	ctx := context.Background()
	pool := wp.NewWorkerPool[int](ctx, wp.DefaultWorkerPoolConfig())
	pool.Close(false)

	err := pool.Submit(simpleJob{id: 1})
	if err != wp.ErrClosedWorkerPool {
		t.Fatalf("expected ErrClosedWorkerPool, got %v", err)
	}
}
