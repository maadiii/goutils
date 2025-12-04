# Comprehensive Documentation: Generic Worker Pool

This package implements a robust, generic Worker Pool in Go, featuring built-in panic safety,
retry logic with exponential backoff, per-job timeouts, and controlled shutdown mechanisms
(Graceful and Forced).

The primary goal is to reliably execute asynchronous tasks while ensuring the stability and
controlled termination of the pool.

---

## 1. Architecture and Core Components

The Worker Pool utilizes buffered channels for efficient queuing of tasks and results, managed
by a configurable number of worker goroutines.

Key Components:

- **JobQueue**: A buffered channel (`chan Job[T]`) for storing pending tasks.
- **Worker Goroutines**: A configurable number of concurrent goroutines that consume jobs from the `JobQueue`.
- **ResultQueue**: A buffered channel (`chan *Result[T]`) for relaying the outcome (value or error) of processed jobs back to the user.
- **Context (`wp.ctx`)**: The central control mechanism for managing the entire pool's lifecycle, essential for forced shutdowns.

---

## 2. The Job Interface (`Job[T]`)

Any task submitted to the pool must implement the `Job[T]` interface, which requires specific methods for processing and configuration:

```go
type Job[T any] interface {
  // Process executes the job's logic. MUST respect the provided Context.
  Process(ctx context.Context) (T, error)
  ID() int
  MaxRetries() int
  RetryDelay() time.Duration
}
```

### Best Practices for Job Implementation

A. Context Respect (The Abort Mechanism) - MANDATORY

The `Process` method must check `ctx.Done()` to be interruptible. Failure to do so will result in:

- The job ignoring its own individual Timeout.
- The job preventing a Force Shutdown (`Close(true)`) from completing promptly.

Example from `timeoutJob`:

```go
func (j timeoutJob) Process(ctx context.Context) (int, error) {
  select {
  case <-time.After(50 * time.Millisecond):
    return 42, nil
  case <-ctx.Done():
    // Worker Pool (or Job Timeout) has been cancelled.
    return 0, ctx.Err()
  }
}
```

B. Idempotency (Retry Safety)

Since the pool supports automatic retries (`MaxRetries > 0`), the job's `Process` logic must be idempotent
or designed to handle multiple executions without adverse side effects.

Example advice: if connecting to a service, ensure the connection is reset or checked before re-use in subsequent attempts.
If writing to a database, use transactions or checks to prevent duplicate data insertion.

C. Resource Management and Atomic Operations (Cleanup/Rollback) - CRITICAL

In real-world applications (like database transactions or file processing), a job must ensure two things: guaranteed resource cleanup
and transactional atomicity upon cancellation.

1. Guaranteed Cleanup (Using `defer`)

Any resource acquired during `Process` must be released regardless of the job's outcome (success, error, or panic). This mechanism is often
referred to as Resource Cleanup.

Best Practice: Use `defer` immediately after resource acquisition (e.g., file handles, database connections, locks).

Example (Database Connection Cleanup):

```go
func (j DBJob) Process(ctx context.Context) (T, error) {
    conn := acquireDBConnection() // Resource Acquisition
    // IMPORTANT: The resource MUST be closed when Process exits.
    defer conn.Close()

    // ... Job Logic ...

    return result, nil
}
```

2. Transactional Rollback/Abort (Handling Cancellation)

If a job is performing multi-step, transactional work, and the context is cancelled (`ctx.Done()`) — typically due to a Force Shutdown
or Job Timeout — the job must undo the changes already made to maintain system integrity. This is the Abort Mechanism.

Best Practice (Database): If using a transaction, `defer Rollback()` and call `Commit()` only upon successful completion.

```go
func (j TransactionJob) Process(ctx context.Context) (T, error) {
    tx, err := db.BeginTx(ctx, nil)
    if err != nil { return nil, err }

    // Defer Rollback: If Commit() is not called, or if context is cancelled,
    // this will roll back the transaction when the function returns.
    defer tx.Rollback()

    // Perform multi-step work...

    select {
    case <-ctx.Done():
        // If cancelled, Rollback is automatically triggered by defer.
        return nil, ctx.Err()
    default:
    }

    // If all logic is successful, commit the transaction.
    if err := tx.Commit(); err != nil {
        return nil, err
    }
    return result, nil
}
```

Best Practice (File Processing): If writing a temporary file, delete the partial file if `ctx.Done()` is received.

```go
func (j FileJob) Process(ctx context.Context) (T, error) {
    tempFilePath := "/tmp/partial_file.dat"
    // ... open file ...

    select {
    case <-ctx.Done():
        // Abort: Delete the incomplete file before returning the context error.
        os.Remove(tempFilePath)
        return nil, ctx.Err()
    default:
    }

    // ... finish writing file ...
    return result, nil
}
```

---

## 3. Job Processing Mechanism (`Result[T].process`)

The `process` method handles execution, error management, and timing for every job.

A. Panic Safety and Cleanup Mechanism

The `defer`/`recover` block is critical for worker pool stability.

Mechanism: It catches any runtime panic originating from `job.Process()`.

Result: The panic is converted into a structured error, including the stack trace (`debug.Stack()`), and stored in `r.Error`.

Benefit: This prevents a single errant job from crashing the entire worker goroutine, ensuring the pool remains operational.

Reference Test Case (`panicJob`): The test confirms that submitting a `panicJob` results in a valid `Result` object containing a non-nil error,
and the worker goroutine remains alive to process subsequent jobs.

B. Timeout Management

For every attempt (including retries), the job is wrapped in a `context.WithTimeout`.

- Duration: The timeout is controlled by the pool's configuration (`wp.timeout`).
- Mechanism: If `job.Process` exceeds this duration and respects the provided `jobCtx`, it will return `context.DeadlineExceeded`.

Reference Test Case (`timeoutJob`): Configuration sets `JobProcessTimeout = 10ms`. The `timeoutJob` attempts to sleep for 50ms. The test verifies that
the resulting error is indeed `context.DeadlineExceeded`.

C. Retry and Exponential Backoff

If `job.Process` returns an error, the retry logic engages if `attempt < MaxRetries()`.

Exponential Backoff Formula:

$$
\text{Delay} = \text{Job.RetryDelay} \times 2^{\text{Attempt}}
$$

This prevents overloading a potentially failing downstream service.

Context Abort During Delay: The `select` inside the retry block checks for `<-ctx.Done()`. If the main pool context is cancelled during the delay
(e.g., due to Force Shutdown), the retry is aborted immediately.

Reference Test Case (`retryJob`): The job is configured with `MaxRetries: 3`. The job intentionally fails on attempts 1 and 2 and succeeds on the 3rd attempt.
The test confirms that the final `Result.Value` is correct and `Result.Error` is `nil`.

---

## 4. Shutdown Management (Closing the Pool)

The `Close(force bool)` method determines the pool's termination behavior.

A. Graceful Shutdown (`Close(false)`)

This is the recommended default for ensuring data integrity.

Goal: Allow all jobs currently in the queue or being processed to complete normally.

Mechanism:

- Sets `wp.closed = true`.
- Closes `wp.jobQueue` (workers stop receiving new tasks).
- Calls `wp.wg.Wait()` to block until all worker goroutines finish their current tasks and exit.
- Closes `wp.resultQueue`.

Reference Test Case (`TestWorkerPool_GracefulShutdown`): The test submits 10 jobs, calls `Close(false)` in a separate goroutine, and then verifies that all 10 results
are successfully received before the result channel is closed.

B. Force Shutdown (`Close(true)`)

Used for rapid termination where completing pending jobs is less important than shutting down the application quickly.

Goal: Immediately stop workers and abort any running jobs that respect their context.

Mechanism (The Abort Mechanism):

- Sets `wp.closed = true`.
- Calls `wp.cancelFunc()`, which triggers `wp.ctx.Done()`.
- This context cancellation immediately terminates running jobs (due to Context Respect and the job's implementation of the Abort Mechanism) and causes workers waiting on channels (`<-wp.ctx.Done()`) to exit.
- Calls `wp.wg.Wait()` (this usually completes very quickly).
- Closes `wp.resultQueue`.

Reference Test Case (`TestWorkerPool_ForceShutdown_Safe`): The test submits long-running jobs (`blockingJobWithCtx`) and calls `Close(true)`. It verifies that the shutdown completes in a
fraction of the job's theoretical runtime, confirming the context cancellation worked.

C. Submission After Closure

Submitting a job after `Close()` has been called is prevented by a lock.

Mechanism: The `Submit` method first checks `wp.closed` under a read lock.

Reference Test Case (`TestWorkerPool_SubmitAfterClose`): The test confirms that attempting to submit a job after closing the pool returns `ErrClosedWorkerPool`.

---

## 5. Integrated Usage Example

This example demonstrates setting up and using the pool with a simple job type.

```go
package main

import (
    "context"
    "fmt"
    "time"
    "workerpool"
)

// SimpleTask implements workerpool.Job[string]
type SimpleTask struct { id int }

// Process respects the context and simulates work.
func (t SimpleTask) Process(ctx context.Context) (string, error) {
    select {
    case <-ctx.Done():
        // Mandatory context check for abort mechanism
        return "", ctx.Err()
    case <-time.After(10 * time.Millisecond):
        return fmt.Sprintf("Job %d completed successfully", t.id), nil
    }
}
func (t SimpleTask) ID() int                   { return t.id }
func (t SimpleTask) MaxRetries() int           { return 0 }
func (t SimpleTask) RetryDelay() time.Duration { return 0 }

func main() {
    ctx := context.Background()
    // Configure pool for 4 workers, 100ms max timeout per job
    cfg := workerpool.WorkerPoolConfig{
        NumWorkers:        4,
        JobQueueSize:      50,
        JobProcessTimeout: 100 * time.Millisecond,
    }

    // 1. Initialize Pool
    pool := workerpool.NewWorkerPool[string](ctx, cfg)

    // 2. Ensure Graceful Shutdown on exit
    defer pool.Close(false)

    // 3. Submit Jobs
    const numJobs = 10
    for i := 1; i <= numJobs; i++ {
        task := SimpleTask{id: i}
        if err := pool.Submit(task); err != nil {
            fmt.Printf("Warning: Failed to submit job %d: %v\n", i, err)
        }
    }

    // 4. Consume Results
    completedCount := 0
    for result := range pool.Results() {
        if result.Error != nil {
            fmt.Printf("❌ Job %d failed in %dms. Error: %v\n",
                result.JobID, result.Duration, result.Error)
        } else {
            fmt.Printf("✅ Job %d succeeded in %dms. Value: %s\n",
                result.JobID, result.Duration, result.Value)
        }
        completedCount++
    }

    fmt.Printf("\nTotal processed jobs: %d\n", completedCount)
}
```

---

## 6. Utility Functions

| Function | Description |
|---|---|
| `PendingJobs() int` | Returns the number of jobs currently waiting in the JobQueue. |
| `PendingResults() int` | Returns the number of processed results awaiting consumption in the ResultQueue. |
| `ErrClosedWorkerPool` | The error returned by `Submit()` when the pool has been closed. |

---

For implementation details, tests, and live examples, see the project files (`wp.go`, `job.go`, `wp_test.go`).
