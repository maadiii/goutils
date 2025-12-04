package workerpool

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"
)

type Job[T any] interface {
	// Process executes the job's logic. It takes a context.Context as an argument,
	// which must be respected by checking ctx.Done(). Failing to do so may result
	// in the worker pool being blocked by the job or getting stuck in a deadlock.
	Process(ctx context.Context) (T, error)
	ID() int
	MaxRetries() int
	RetryDelay() time.Duration
}

type Result[T any] struct {
	JobID       int
	Value       T
	Error       error
	StartedAt   time.Time
	CompletedAt time.Time
	Duration    int64
}

func (r *Result[T]) process(ctx context.Context, job Job[T], timeout time.Duration) {
	now := time.Now()
	defer func() {
		if rec := recover(); rec != nil {
			r.Error = fmt.Errorf("%v\n%s", rec, debug.Stack())
			r.JobID = job.ID()
		}
	}()

	var (
		output T
		err    error
	)

LOOP:
	for attempt := 0; attempt <= job.MaxRetries(); attempt++ {
		select {
		case <-ctx.Done():
			err = ctx.Err()

			break LOOP
		default:
		}

		jobCtx, cancel := context.WithTimeout(ctx, timeout)
		output, err = job.Process(jobCtx)
		cancel()
		if err == nil {
			break
		}

		if attempt < job.MaxRetries() {
			select {
			case <-ctx.Done():
				err = ctx.Err()
			case <-time.After(job.RetryDelay() * (1 << attempt)):
			}
		}
	}

	r.JobID = job.ID()
	r.Value = output
	r.Error = err
	r.StartedAt = now
	r.CompletedAt = time.Now()
	r.Duration = r.CompletedAt.Sub(r.StartedAt).Milliseconds()
}
