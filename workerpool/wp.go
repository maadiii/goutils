package workerpool

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

var ErrClosedWorkerPool = fmt.Errorf("worker pool is closed")

type WorkerPoolConfig struct {
	NumWorkers        int
	JobQueueSize      int
	ResultQueueSize   int
	JobProcessTimeout time.Duration
}

// DefaultWorkerPoolConfig provides default configuration for WorkerPool.
func DefaultWorkerPoolConfig() WorkerPoolConfig {
	cfg := WorkerPoolConfig{
		NumWorkers:        runtime.NumCPU(),
		JobQueueSize:      100,
		ResultQueueSize:   100,
		JobProcessTimeout: 5 * time.Second,
	}

	return cfg
}

// isZero checks if the WorkerPoolConfig is zero-valued.
func (cfg *WorkerPoolConfig) isZero() bool {
	return cfg.NumWorkers == 0 &&
		cfg.JobQueueSize == 0 &&
		cfg.ResultQueueSize == 0 &&
		cfg.JobProcessTimeout == 0
}

type WorkerPool[T any] struct {
	numWorkers  int
	jobQueue    chan Job[T]
	resultQueue chan *Result[T]
	timeout     time.Duration

	ctx        context.Context
	cancelFunc context.CancelFunc

	wg     sync.WaitGroup
	closed bool
	mu     sync.RWMutex
}

func NewWorkerPool[T any](ctx context.Context, cfg WorkerPoolConfig) *WorkerPool[T] {
	pool := new(WorkerPool[T]).construct(ctx, cfg)
	pool.start()

	return pool
}

func (wp *WorkerPool[T]) construct(ctx context.Context, cfg WorkerPoolConfig) *WorkerPool[T] {
	if cfg.isZero() {
		cfg = DefaultWorkerPoolConfig()
	}

	wp.ctx, wp.cancelFunc = context.WithCancel(ctx)
	wp.numWorkers = cfg.NumWorkers
	wp.jobQueue = make(chan Job[T], cfg.JobQueueSize)
	wp.resultQueue = make(chan *Result[T], cfg.ResultQueueSize)
	wp.timeout = cfg.JobProcessTimeout

	return wp
}

func (wp *WorkerPool[T]) start() {
	wp.wg.Add(wp.numWorkers)
	for i := 0; i < wp.numWorkers; i++ {
		go wp.worker()
	}
}

// Submit adds a job to the worker pool.
func (wp *WorkerPool[T]) Submit(job Job[T]) error {
	wp.mu.RLock()
	defer wp.mu.RUnlock()

	if wp.closed {
		return ErrClosedWorkerPool
	}

	select {
	case <-wp.ctx.Done():
		return wp.ctx.Err()
	case wp.jobQueue <- job:
		return nil
	}
}

// Results returns the result channel to receive processed job results.
func (wp *WorkerPool[T]) Results() <-chan *Result[T] {
	return wp.resultQueue
}

func (wp *WorkerPool[T]) Close(force bool) {
	wp.mu.Lock()
	if wp.closed {
		wp.mu.Unlock()
		return
	}
	wp.closed = true
	if force {
		wp.cancelFunc()
	}
	// close jobQueue while holding lock
	close(wp.jobQueue)
	wp.mu.Unlock()

	// wait for workers to finish processing
	wp.wg.Wait()

	// safe to close results after all workers exited
	close(wp.resultQueue)
}

func (wp *WorkerPool[T]) worker() {
	defer wp.wg.Done()

	for {
		select {
		case <-wp.ctx.Done():
			return
		case job, ok := <-wp.jobQueue:
			if !ok {
				return
			}
			result := new(Result[T])
			result.process(wp.ctx, job, wp.timeout)
			select {
			case <-wp.ctx.Done():
				return
			case wp.resultQueue <- result:
			}
		}
	}
}

func (wp *WorkerPool[T]) PendingJobs() int    { return len(wp.jobQueue) }
func (wp *WorkerPool[T]) PendingResults() int { return len(wp.resultQueue) }
