# Worker Pool

Concurrent job processing with configurable worker pools for efficient parallel task execution.

## Overview

The workerpool package provides a robust implementation of the worker pool pattern for concurrent job processing in Go. It efficiently manages a pool of goroutines to process jobs in parallel.

## Installation

```bash
go get github.com/maadiii/goutils/workerpool
```

## Features

- ⚡ Concurrent job processing
- 🎯 Configurable worker count
- 📊 Job queue management
- ✅ Result collection
- ⏱️ Job timeout support
- 🛡️ Graceful shutdown
- 🔄 Generic type support

## Usage

### Basic Setup

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/maadiii/goutils/workerpool"
)

func main() {
    // Create worker pool
    pool := workerpool.New[string](workerpool.WorkerPoolConfig{
        NumWorkers:        5,              // 5 concurrent workers
        JobQueueSize:      100,            // Queue up to 100 jobs
        ResultQueueSize:   100,            // Buffer for 100 results
        JobProcessTimeout: 10 * time.Second, // 10s timeout per job
    })

    // Start the pool
    pool.Start()
    defer pool.Stop()

    // Submit jobs
    for i := 0; i < 10; i++ {
        jobID := i
        err := pool.Submit(workerpool.Job[string]{
            Execute: func(ctx context.Context) (string, error) {
                // Simulate work
                time.Sleep(100 * time.Millisecond)
                return fmt.Sprintf("Job %d completed", jobID), nil
            },
        })
        if err != nil {
            fmt.Println("Failed to submit job:", err)
        }
    }

    // Wait for all jobs to complete
    pool.Wait()

    // Collect results
    results := pool.Results()
    for _, result := range results {
        if result.Err != nil {
            fmt.Printf("Job failed: %v\n", result.Err)
        } else {
            fmt.Printf("Result: %v\n", result.Value)
        }
    }
}
```

### With Default Configuration

```go
// Use sensible defaults (NumCPU workers, 100 queue sizes, 5s timeout)
pool := workerpool.New[int](workerpool.DefaultWorkerPoolConfig())

pool.Start()
defer pool.Stop()

// Submit jobs
pool.Submit(workerpool.Job[int]{
    Execute: func(ctx context.Context) (int, error) {
        return 42, nil
    },
})
```

### Complete Example: Image Processing

```go
package main

import (
    "context"
    "fmt"
    "image"
    "time"

    "github.com/maadiii/goutils/workerpool"
)

type ImageResult struct {
    Filename string
    Width    int
    Height   int
}

func main() {
    // Create pool for image processing
    pool := workerpool.New[ImageResult](workerpool.WorkerPoolConfig{
        NumWorkers:        runtime.NumCPU(),
        JobQueueSize:      50,
        ResultQueueSize:   50,
        JobProcessTimeout: 30 * time.Second,
    })

    pool.Start()
    defer pool.Stop()

    // List of images to process
    images := []string{
        "photo1.jpg",
        "photo2.jpg",
        "photo3.jpg",
        // ... more images
    }

    // Submit processing jobs
    for _, filename := range images {
        imgFile := filename
        err := pool.Submit(workerpool.Job[ImageResult]{
            Execute: func(ctx context.Context) (ImageResult, error) {
                // Load image
                img, err := loadImage(imgFile)
                if err != nil {
                    return ImageResult{}, err
                }

                // Process image (resize, compress, etc.)
                processed := processImage(img)

                // Save processed image
                if err := saveImage(processed, "processed_"+imgFile); err != nil {
                    return ImageResult{}, err
                }

                bounds := processed.Bounds()
                return ImageResult{
                    Filename: imgFile,
                    Width:    bounds.Dx(),
                    Height:   bounds.Dy(),
                }, nil
            },
        })

        if err != nil {
            fmt.Printf("Failed to submit %s: %v\n", imgFile, err)
        }
    }

    // Wait for all processing to complete
    pool.Wait()

    // Collect results
    results := pool.Results()
    successCount := 0
    failCount := 0

    for _, result := range results {
        if result.Err != nil {
            fmt.Printf("Failed to process: %v\n", result.Err)
            failCount++
        } else {
            fmt.Printf("Processed %s: %dx%d\n",
                result.Value.Filename,
                result.Value.Width,
                result.Value.Height)
            successCount++
        }
    }

    fmt.Printf("\nProcessed %d images successfully, %d failed\n",
        successCount, failCount)
}
```

### Example: Data Processing Pipeline

```go
type DataRecord struct {
    ID   int
    Data string
}

func ProcessRecords(records []DataRecord) {
    pool := workerpool.New[DataRecord](workerpool.WorkerPoolConfig{
        NumWorkers:      10,
        JobQueueSize:    1000,
        ResultQueueSize: 1000,
        JobProcessTimeout: 5 * time.Second,
    })

    pool.Start()
    defer pool.Stop()

    // Submit all records for processing
    for _, record := range records {
        rec := record
        pool.Submit(workerpool.Job[DataRecord]{
            Execute: func(ctx context.Context) (DataRecord, error) {
                // Transform data
                rec.Data = strings.ToUpper(rec.Data)

                // Validate
                if err := validate(rec); err != nil {
                    return DataRecord{}, err
                }

                // Save to database
                if err := saveToDatabase(rec); err != nil {
                    return DataRecord{}, err
                }

                return rec, nil
            },
        })
    }

    pool.Wait()

    // Check results
    results := pool.Results()
    for _, result := range results {
        if result.Err != nil {
            log.Printf("Failed to process record: %v\n", result.Err)
        }
    }
}
```

### Example: Web Scraping

```go
type ScrapedData struct {
    URL   string
    Title string
    Links []string
}

func ScrapeWebsites(urls []string) []ScrapedData {
    pool := workerpool.New[ScrapedData](workerpool.WorkerPoolConfig{
        NumWorkers:        20,
        JobQueueSize:      100,
        ResultQueueSize:   100,
        JobProcessTimeout: 30 * time.Second,
    })

    pool.Start()
    defer pool.Stop()

    for _, url := range urls {
        targetURL := url
        pool.Submit(workerpool.Job[ScrapedData]{
            Execute: func(ctx context.Context) (ScrapedData, error) {
                // Fetch page
                resp, err := http.Get(targetURL)
                if err != nil {
                    return ScrapedData{}, err
                }
                defer resp.Body.Close()

                // Parse HTML
                doc, err := goquery.NewDocumentFromReader(resp.Body)
                if err != nil {
                    return ScrapedData{}, err
                }

                // Extract data
                title := doc.Find("title").Text()
                var links []string
                doc.Find("a").Each(func(i int, s *goquery.Selection) {
                    href, exists := s.Attr("href")
                    if exists {
                        links = append(links, href)
                    }
                })

                return ScrapedData{
                    URL:   targetURL,
                    Title: title,
                    Links: links,
                }, nil
            },
        })
    }

    pool.Wait()

    // Collect successful results
    var scraped []ScrapedData
    for _, result := range pool.Results() {
        if result.Err == nil {
            scraped = append(scraped, result.Value)
        } else {
            log.Printf("Failed to scrape: %v\n", result.Err)
        }
    }

    return scraped
}
```

## API Reference

### Types

#### `WorkerPoolConfig`

Configuration for creating a worker pool.

```go
type WorkerPoolConfig struct {
    NumWorkers        int           // Number of worker goroutines
    JobQueueSize      int           // Size of job queue buffer
    ResultQueueSize   int           // Size of result queue buffer
    JobProcessTimeout time.Duration // Timeout for each job
}
```

#### `Job[T any]`

Represents a job to be executed.

```go
type Job[T any] struct {
    Execute func(context.Context) (T, error)
}
```

#### `Result[T any]`

Represents the result of a job execution.

```go
type Result[T any] struct {
    Value T
    Err   error
}
```

### Functions

#### `DefaultWorkerPoolConfig() WorkerPoolConfig`

Returns default configuration with sensible values.

```go
config := workerpool.DefaultWorkerPoolConfig()
// NumWorkers: runtime.NumCPU()
// JobQueueSize: 100
// ResultQueueSize: 100
// JobProcessTimeout: 5 seconds
```

#### `New[T any](config WorkerPoolConfig) *WorkerPool[T]`

Creates a new worker pool with the given configuration.

```go
pool := workerpool.New[string](config)
```

### Methods

#### `Start()`

Starts the worker pool and begins processing jobs.

```go
pool.Start()
```

#### `Submit(job Job[T]) error`

Submits a job to the worker pool. Returns `ErrClosedWorkerPool` if pool is closed.

```go
err := pool.Submit(workerpool.Job[int]{
    Execute: func(ctx context.Context) (int, error) {
        return 42, nil
    },
})
```

#### `Wait()`

Blocks until all submitted jobs are processed.

```go
pool.Wait()
```

#### `Stop()`

Stops the worker pool and waits for all jobs to complete.

```go
pool.Stop()
```

#### `Results() []Result[T]`

Returns all results from completed jobs.

```go
results := pool.Results()
for _, result := range results {
    if result.Err != nil {
        // Handle error
    } else {
        // Use result.Value
    }
}
```

## Configuration Guide

### Number of Workers

```go
// CPU-bound tasks: use CPU count
NumWorkers: runtime.NumCPU()

// I/O-bound tasks: use more workers
NumWorkers: runtime.NumCPU() * 2

// Mixed workload: experiment with values
NumWorkers: 10
```

### Queue Sizes

```go
// Small batches
JobQueueSize: 50
ResultQueueSize: 50

// Large batches
JobQueueSize: 1000
ResultQueueSize: 1000

// Memory-constrained
JobQueueSize: 10
ResultQueueSize: 10
```

### Timeouts

```go
// Quick operations
JobProcessTimeout: 1 * time.Second

// Standard operations
JobProcessTimeout: 5 * time.Second

// Long-running operations
JobProcessTimeout: 60 * time.Second

// No timeout (use with caution)
JobProcessTimeout: 0
```

## Best Practices

1. **Always defer Stop()**

   ```go
   pool.Start()
   defer pool.Stop() // Ensures cleanup
   ```

2. **Check Submit errors**

   ```go
   if err := pool.Submit(job); err != nil {
       log.Printf("Failed to submit: %v\n", err)
   }
   ```

3. **Handle job errors**

   ```go
   for _, result := range pool.Results() {
       if result.Err != nil {
           // Log or handle error
       }
   }
   ```

4. **Use appropriate worker count**

   - CPU-bound: `runtime.NumCPU()`
   - I/O-bound: `runtime.NumCPU() * 2` or more

5. **Set reasonable timeouts**

   - Prevents hanging jobs
   - Allows resource cleanup

6. **Monitor queue sizes**
   - Too small: jobs may be rejected
   - Too large: high memory usage

## Common Patterns

### Batch Processing

```go
func ProcessBatch(items []Item) {
    pool := workerpool.New[Result](config)
    pool.Start()
    defer pool.Stop()

    for _, item := range items {
        pool.Submit(createJob(item))
    }

    pool.Wait()
    handleResults(pool.Results())
}
```

### Retry Logic

```go
pool.Submit(workerpool.Job[Data]{
    Execute: func(ctx context.Context) (Data, error) {
        var data Data
        var err error

        for attempt := 0; attempt < 3; attempt++ {
            data, err = fetchData()
            if err == nil {
                break
            }
            time.Sleep(time.Second * time.Duration(attempt+1))
        }

        return data, err
    },
})
```

### Fan-Out Pattern

```go
// Submit many jobs
for _, task := range tasks {
    pool.Submit(createJob(task))
}

// Wait and collect (fan-in)
pool.Wait()
results := pool.Results()
```

## Performance Tips

- Use buffered channels (JobQueueSize, ResultQueueSize)
- Adjust worker count based on workload type
- Profile to find optimal configuration
- Consider memory usage with large result sets
- Use context for cancellation support

## Error Handling

### Pool Closed Error

```go
err := pool.Submit(job)
if err == workerpool.ErrClosedWorkerPool {
    log.Println("Pool is closed, cannot submit job")
}
```

### Job Timeout

```go
// Jobs exceeding JobProcessTimeout will return context.DeadlineExceeded
for _, result := range pool.Results() {
    if result.Err == context.DeadlineExceeded {
        log.Println("Job timed out")
    }
}
```

## Testing

```go
func TestWorkerPool(t *testing.T) {
    pool := workerpool.New[int](workerpool.DefaultWorkerPoolConfig())
    pool.Start()
    defer pool.Stop()

    // Submit test jobs
    for i := 0; i < 10; i++ {
        val := i
        pool.Submit(workerpool.Job[int]{
            Execute: func(ctx context.Context) (int, error) {
                return val * 2, nil
            },
        })
    }

    pool.Wait()

    results := pool.Results()
    if len(results) != 10 {
        t.Errorf("Expected 10 results, got %d", len(results))
    }
}
```

## License

MIT License - see [LICENSE](../LICENSE) for details
