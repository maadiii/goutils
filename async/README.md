# Async

Asynchronous task execution utilities for Go with Future-based patterns.

## Overview

The async package provides a simple and type-safe way to execute asynchronous operations in Go using the Future pattern. It handles panic recovery and context cancellation gracefully.

## Installation

```bash
go get github.com/maadiii/goutils/async
```

## Features

- 🔄 Future-based async execution
- 🛡️ Automatic panic recovery
- ⏱️ Context-aware cancellation
- 🎯 Type-safe generic implementation

## Usage

### Basic Async Operation

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/maadiii/goutils/async"
)

func main() {
    ctx := context.Background()

    // Spawn an async operation
    future := async.Spawn(ctx, func(ctx context.Context) (string, error) {
        // Simulate some work
        time.Sleep(100 * time.Millisecond)
        return "Hello from async!", nil
    })

    // Wait for the result
    result, err := future.Await()
    if err != nil {
        panic(err)
    }

    fmt.Println(result) // Output: Hello from async!
}
```

### Multiple Async Operations

```go
// Spawn multiple operations
future1 := async.Spawn(ctx, func(ctx context.Context) (int, error) {
    time.Sleep(100 * time.Millisecond)
    return 42, nil
})

future2 := async.Spawn(ctx, func(ctx context.Context) (int, error) {
    time.Sleep(50 * time.Millisecond)
    return 24, nil
})

// Await results
result1, err1 := future1.Await()
result2, err2 := future2.Await()

fmt.Printf("Results: %d, %d\n", result1, result2)
```

### With Context Cancellation

```go
ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
defer cancel()

future := async.Spawn(ctx, func(ctx context.Context) (string, error) {
    select {
    case <-time.After(200 * time.Millisecond):
        return "completed", nil
    case <-ctx.Done():
        return "", ctx.Err()
    }
})

result, err := future.Await()
if err != nil {
    fmt.Println("Operation cancelled:", err)
}
```

### Error Handling

```go
future := async.Spawn(ctx, func(ctx context.Context) (int, error) {
    return 0, errors.New("something went wrong")
})

result, err := future.Await()
if err != nil {
    fmt.Println("Error:", err)
    // Handle error
}
```

### Panic Recovery

The package automatically recovers from panics in async operations:

```go
future := async.Spawn(ctx, func(ctx context.Context) (string, error) {
    panic("unexpected error!")
    return "never reached", nil
})

result, err := future.Await()
// err will contain the panic information
```

## API Reference

### Types

#### `Future[E any]`

Interface representing an asynchronous operation that will produce a result of type E.

```go
type Future[E any] interface {
    Await() (E, error)
}
```

#### `FutureResult[E any]`

Holds the result value and error from an async operation.

```go
type FutureResult[E any] struct {
    Value E
    Err   error
}
```

### Functions

#### `Spawn[E any](ctx context.Context, f func(context.Context) (E, error)) Future[E]`

Spawns an asynchronous operation and returns a Future.

**Parameters:**

- `ctx`: Context for cancellation and timeout
- `f`: Function to execute asynchronously

**Returns:**

- `Future[E]`: A future that can be awaited for the result

## Best Practices

1. **Always handle errors**: Check the error returned from `Await()`
2. **Use contexts**: Pass appropriate contexts for cancellation and timeouts
3. **Don't block indefinitely**: Always use context timeouts for operations that might hang
4. **Type safety**: Leverage Go generics for type-safe async operations

## Thread Safety

The async package is thread-safe and handles concurrent access properly through channels.

## Performance Considerations

- Each `Spawn` call creates a new goroutine
- Futures use channels internally for result communication
- For CPU-bound work, consider using a worker pool instead

## License

MIT License - see [LICENSE](../LICENSE) for details
