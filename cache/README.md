# Cache

Redis-based caching solution with a clean and simple interface.

## Overview

The cache package provides a straightforward wrapper around Redis for caching operations with context support and type-safe operations.

## Installation

```bash
go get github.com/maadiii/goutils/cache
```

## Dependencies

- Redis server (v6.0+)
- [go-redis/v9](https://github.com/redis/go-redis)

## Features

- 🚀 Simple Redis interface
- ⏰ TTL support
- 🎯 Context-aware operations
- 🔄 Connection pooling
- ✅ Automatic ping verification

## Usage

### Basic Setup

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/maadiii/goutils/cache"
    "github.com/redis/go-redis/v9"
)

func main() {
    // Initialize cache
    c := cache.NewRedis(&redis.Options{
        Addr:     "localhost:6379",
        Password: "", // no password set
        DB:       0,  // use default DB
    })

    ctx := context.Background()

    // Set a value with TTL
    err := c.Set(ctx, "user:123", "John Doe", 1*time.Hour)
    if err != nil {
        panic(err)
    }

    // Get a value
    value, err := c.Get(ctx, "user:123")
    if err != nil {
        panic(err)
    }

    fmt.Println("Value:", value)

    // Delete keys
    err = c.Del(ctx, "user:123")
    if err != nil {
        panic(err)
    }
}
```

### With TLS

```go
import "crypto/tls"

c := cache.NewRedis(&redis.Options{
    Addr:     "redis.example.com:6380",
    Password: "your-password",
    DB:       0,
    TLSConfig: &tls.Config{
        MinVersion: tls.VersionTLS12,
    },
})
```

### With Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

err := c.Set(ctx, "key", "value", time.Hour)
if err != nil {
    // Handle timeout or other errors
}
```

### Caching Patterns

#### Cache-Aside Pattern

```go
func GetUser(ctx context.Context, c *cache.cache, userID string) (*User, error) {
    // Try cache first
    cached, err := c.Get(ctx, "user:"+userID)
    if err == nil {
        var user User
        json.Unmarshal([]byte(cached), &user)
        return &user, nil
    }

    // Cache miss - fetch from database
    user, err := db.GetUser(userID)
    if err != nil {
        return nil, err
    }

    // Store in cache
    data, _ := json.Marshal(user)
    c.Set(ctx, "user:"+userID, string(data), 15*time.Minute)

    return user, nil
}
```

#### Write-Through Pattern

```go
func UpdateUser(ctx context.Context, c *cache.cache, user *User) error {
    // Update database
    if err := db.UpdateUser(user); err != nil {
        return err
    }

    // Update cache
    data, _ := json.Marshal(user)
    return c.Set(ctx, "user:"+user.ID, string(data), 15*time.Minute)
}
```

#### Cache Invalidation

```go
func DeleteUser(ctx context.Context, c *cache.cache, userID string) error {
    // Delete from database
    if err := db.DeleteUser(userID); err != nil {
        return err
    }

    // Invalidate cache
    return c.Del(ctx, "user:"+userID)
}
```

### Multiple Key Deletion

```go
// Delete multiple keys at once
err := c.Del(ctx, "user:1", "user:2", "user:3")
```

### Storing Complex Data

```go
import "encoding/json"

type Session struct {
    UserID    string
    CreatedAt time.Time
    Data      map[string]interface{}
}

// Store struct
session := Session{
    UserID:    "123",
    CreatedAt: time.Now(),
    Data:      map[string]interface{}{"role": "admin"},
}

data, _ := json.Marshal(session)
err := c.Set(ctx, "session:abc", string(data), 24*time.Hour)

// Retrieve struct
cached, err := c.Get(ctx, "session:abc")
if err == nil {
    var session Session
    json.Unmarshal([]byte(cached), &session)
}
```

## API Reference

### Functions

#### `NewRedis(opts *redis.Options) *cache`

Creates a new cache instance with Redis connection.

**Panics** if connection to Redis fails.

**Parameters:**

- `opts`: Redis connection options

**Returns:**

- `*cache`: Cache instance

### Methods

#### `Set(ctx context.Context, key string, value any, ttl time.Duration) error`

Sets a key-value pair in cache with TTL.

**Parameters:**

- `ctx`: Context for cancellation
- `key`: Cache key
- `value`: Value to store (will be serialized)
- `ttl`: Time-to-live duration

**Returns:**

- `error`: Error if operation fails

#### `Get(ctx context.Context, key string) (string, error)`

Retrieves a value from cache.

**Parameters:**

- `ctx`: Context for cancellation
- `key`: Cache key

**Returns:**

- `string`: Cached value
- `error`: Error if key doesn't exist or operation fails

#### `Del(ctx context.Context, keys ...string) error`

Deletes one or more keys from cache.

**Parameters:**

- `ctx`: Context for cancellation
- `keys`: Keys to delete

**Returns:**

- `error`: Error if operation fails

## Configuration

### Redis Options

```go
&redis.Options{
    // Network type: tcp or unix
    Network: "tcp",

    // Redis server address
    Addr: "localhost:6379",

    // Optional password
    Password: "",

    // Database to be selected
    DB: 0,

    // Maximum number of retries
    MaxRetries: 3,

    // Dial timeout
    DialTimeout: 5 * time.Second,

    // Read timeout
    ReadTimeout: 3 * time.Second,

    // Write timeout
    WriteTimeout: 3 * time.Second,

    // Maximum number of socket connections
    PoolSize: 10,

    // Minimum number of idle connections
    MinIdleConns: 5,

    // TLS configuration
    TLSConfig: &tls.Config{},
}
```

## Best Practices

1. **Use appropriate TTLs**

   - Short TTL for frequently changing data
   - Longer TTL for static data
   - No TTL for data that should persist

2. **Key naming conventions**

   - Use colon separators: `user:123`, `session:abc`
   - Include namespace: `app:user:123`
   - Be consistent

3. **Error handling**

   - Always check for cache miss errors
   - Have fallback to database
   - Don't let cache failures break your app

4. **Context usage**

   - Use context for timeouts
   - Propagate cancellation
   - Don't use `context.Background()` in HTTP handlers

5. **Serialization**
   - Use JSON for complex types
   - Consider msgpack for performance
   - Store simple strings when possible

## Performance Tips

- Use pipelining for bulk operations
- Set appropriate connection pool size
- Monitor Redis memory usage
- Use Redis Cluster for scalability
- Consider Redis Sentinel for high availability

## Common Patterns

### Session Storage

```go
func StoreSession(ctx context.Context, c *cache.cache, sessionID string, data SessionData) error {
    encoded, _ := json.Marshal(data)
    return c.Set(ctx, "session:"+sessionID, string(encoded), 30*time.Minute)
}
```

### Rate Limiting

```go
// Combine with Redis INCR for rate limiting
// (requires direct redis client access)
```

### Distributed Locks

```go
// Implement using Redis SETNX
// (requires direct redis client access)
```

## Testing

For testing, consider using [miniredis](https://github.com/alicebob/miniredis) for in-memory Redis simulation:

```go
import "github.com/alicebob/miniredis/v2"

func TestCache(t *testing.T) {
    mr, _ := miniredis.Run()
    defer mr.Close()

    c := cache.NewRedis(&redis.Options{
        Addr: mr.Addr(),
    })

    // Your tests here
}
```

## License

MIT License - see [LICENSE](../LICENSE) for details
